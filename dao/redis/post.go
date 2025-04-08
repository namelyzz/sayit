package redis

import (
	"context"
	"fmt"
	"github.com/namelyzz/sayit/models"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

// CreatePost 将帖子写入 Redis 排行榜（事务性管道操作）
// 一次事务中写入三个数据结构:
//   1. sayit:post:time (ZSet) - 时间排行榜，score 为创建时间戳，用于"最新发布"排序
//   2. sayit:post:score (ZSet) - 热度排行榜，score 初始为时间戳，后续叠加投票分数
//   3. sayit:community:<id> (Set) - 社区帖子集合，用于按社区筛选
//
// 为什么热度排行榜的初始分数是时间戳？
//   - 新帖子刚发布时没有投票，如果分数为 0 会沉底
//   - 用时间戳作为初始分，保证新帖子暂时排在前面
//   - 随着时间推移，旧帖子如果投票分很高，总分仍可超过新帖子
func CreatePost(ctx context.Context, postID, communityID int64, score float64) error {
	// 使用事务管道，保证三个操作的原子性
	pipe := client.TxPipeline()

	// 1. 按时间入榜：将帖子加入"最新发布"排行榜
	pipe.ZAdd(ctx, getRedisKey(KeyPostTimeZset), redis.Z{
		Score:  score,
		Member: postID,
	})

	// 2. 按分数入榜：将帖子加入"综合热度"排行榜
	pipe.ZAdd(ctx, getRedisKey(KeyPostScoreZset), redis.Z{
		Score:  score,
		Member: postID,
	})

	// 3. 社区关联：将帖子 ID 记录到对应社区的集合中
	cKey := getRedisKey(KeyCommunitySetPF + strconv.Itoa(int(communityID)))
	pipe.SAdd(ctx, cKey, postID)

	// 执行事务管道（原子性提交）
	_, err := pipe.Exec(ctx)
	return err
}

// GetPostIDsInOrder 从 Redis 排行榜中获取排序后的帖子 ID 列表
// 支持两种排序: create_time（时间榜）和 score（热度榜）
// 支持按社区筛选: 通过 ZInterStore 将排行榜与社区集合求交集
// 支持时间范围筛选: 通过 ZRangeByScore 的 Min/Max 参数
//
// 返回值: 帖子 ID 的字符串列表（Redis ZSet 的 member 类型为 string）
func GetPostIDsInOrder(ctx context.Context, p *models.ParamPostList, offset, limit int) (res []string, err error) {
	if limit <= 0 {
		return []string{}, nil
	}

	// 1. 获取查询目标 Key（可能需要 ZInterStore 生成临时 Key）
	targetKey, cleanup, err := getPostQueryKey(ctx, p.CommunityID, p.SortBy)
	if err != nil {
		return nil, err
	}
	defer cleanup() // 确保临时 Key 被清理

	// 2. 按创建时间排序 + 有时间范围筛选: 使用 ZRangeByScore
	if p.SortBy == models.SortFieldCreateTime && (p.StartTime != nil || p.EndTime != nil) {
		minTime, maxTime := "-inf", "+inf"
		if p.StartTime != nil {
			minTime = strconv.FormatInt(*p.StartTime, 10)
		}
		if p.EndTime != nil {
			maxTime = strconv.FormatInt(*p.EndTime, 10)
		}

		opt := &redis.ZRangeBy{
			Min:    minTime,      // 时间范围下界
			Max:    maxTime,      // 时间范围上界
			Offset: int64(offset), // 跳过前面的记录（分页）
			Count:  int64(limit),  // 返回的记录数
		}

		// 根据排序方向选择 ZRevRangeByScore（倒序）或 ZRangeByScore（正序）
		if p.Order == models.SortDirectionDesc {
			return client.ZRevRangeByScore(ctx, targetKey, opt).Result()
		}
		return client.ZRangeByScore(ctx, targetKey, opt).Result()
	}

	// 3. 无时间范围筛选: 使用 ZRange/ZRevRange（效率更高）
	start := int64(offset)
	stop := int64(offset + limit - 1)
	if p.Order == models.SortDirectionDesc {
		return client.ZRevRange(ctx, targetKey, start, stop).Result()
	}
	return client.ZRange(ctx, targetKey, start, stop).Result()
}

// getPostQueryKey 获取帖子查询的 Redis Key
// 如果指定了社区 ID，需要通过 ZInterStore 将排行榜与社区集合求交集
//
// 返回值:
//   - targetKey: 实际查询的 Key（原始排行榜 Key 或临时交集 Key）
//   - cleanup: 清理临时 Key 的函数（defer 调用）
//   - err: 错误信息
//
// ZInterStore 原理:
//   - 将排行榜 ZSet 和社区 Set 求交集
//   - 权重: ZSet=1（保留原分数）, Set=0（忽略 Set 的分数）
//   - 结果存入临时 Key，查询完毕后清理
func getPostQueryKey(ctx context.Context, commID int64, sortBy models.SortField) (targetKey string, cleanup func(), err error) {
	// 1. 根据排序方式选择基础排行榜 Key
	baseKey := getRedisKey(KeyPostScoreZset)
	if sortBy == models.SortFieldCreateTime {
		baseKey = getRedisKey(KeyPostTimeZset)
	}

	// 2. 如果没有指定社区 ID，直接返回基础排行榜 Key
	cleanup = func() {}
	targetKey = baseKey
	if commID <= 0 {
		return targetKey, cleanup, nil
	}

	// 3. 指定了社区 ID，需要求交集
	// 社区帖子集合 Key
	communityKey := getRedisKey(KeyCommunitySetPF + strconv.FormatInt(commID, 10))
	// 临时 Key（使用时间戳保证唯一性）
	tempKey := getRedisKey(fmt.Sprintf("temp:post:%d:%d:%s", commID, time.Now().UnixNano(), sortBy))
	// 清理函数: 查询完毕后删除临时 Key
	cleanup = func() {
		_ = client.Del(context.Background(), tempKey).Err()
	}

	// 4. 执行 ZInterStore: 排行榜 ∩ 社区集合 → 临时 Key
	// Weights: [1, 0] 表示保留排行榜的分数，忽略社区集合的分数
	// Aggregate MAX: 取最大值（由于 Set 分数为 0，实际等于保留 ZSet 分数）
	err = client.ZInterStore(ctx, tempKey, &redis.ZStore{
		Keys:      []string{baseKey, communityKey},
		Weights:   []float64{1, 0},
		Aggregate: "MAX",
	}).Err()
	if err != nil {
		return "", cleanup, err
	}

	return tempKey, cleanup, nil
}
