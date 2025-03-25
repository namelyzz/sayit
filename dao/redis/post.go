package redis

import (
	"context"
	"fmt"
	"github.com/namelyzz/sayit/models"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

func CreatePost(ctx context.Context, postID, communityID int64, score float64) error {
	pipe := client.TxPipeline()

	// 按时间入榜：将帖子加入“最新发布”排行榜
	pipe.ZAdd(ctx, getRedisKey(KeyPostTimeZset), redis.Z{
		Score:  score,
		Member: postID,
	})

	// 按分数入榜：将帖子加入“综合热度”排行榜（初始分数设为当前时间戳）。
	// 为什么初始分数是时间戳？ 最终分数 = 基础时间分 + 投票加权分
	// 1.为了保证新帖子有曝光机会：
	// 刚发布的帖子没有投票，如果分数为 0，它会沉底。将分数初始化为当前时间戳，能保证新发布的帖子暂时排在前面（比旧帖子分数高）。
	// 2.随着时间推移，旧帖子的“时间分”虽然小，但如果它的“投票分”很高，总分就会超过这个新帖子。
	pipe.ZAdd(ctx, getRedisKey(KeyPostScoreZset), redis.Z{
		Score:  score,
		Member: postID,
	})

	// 社区关联：将帖子 ID 记录到对应社区的集合中。
	// 用于快速查找某个社区内的帖子列表
	cKey := getRedisKey(KeyCommunitySetPF + strconv.Itoa(int(communityID)))
	pipe.SAdd(ctx, cKey, postID)
	_, err := pipe.Exec(ctx)
	return err
}

// GetPostIDsInOrder 从Redis中获取排序后的帖子ID列表
func GetPostIDsInOrder(ctx context.Context, p *models.ParamPostList) (res []string, err error) {
	targetKey, err := genPostKey(ctx, p.CommunityID, p.SortBy)
	if err != nil {
		return nil, err
	}

	if p.SortBy == models.SortFieldCreateTime && (p.StartTime != nil || p.EndTime != nil) {
		// 场景1：按时间排序 且 有时间范围限制 -> 使用 ZRangeByScore
		minTime, maxTime := "-inf", "+inf"
		if p.StartTime != nil {
			minTime = strconv.FormatInt(*p.StartTime, 10)
		}
		if p.EndTime != nil {
			maxTime = strconv.FormatInt(*p.EndTime, 10)
		}

		opt := &redis.ZRangeBy{
			Min:    minTime,
			Max:    maxTime,
			Offset: int64((p.Page - 1) * p.Size),
			Count:  int64(p.Size),
		}

		if p.Order == models.SortDirectionDesc {
			res, err = client.ZRevRangeByScore(ctx, targetKey, opt).Result()
		} else {
			res, err = client.ZRangeByScore(ctx, targetKey, opt).Result()
		}
	} else {
		// 场景 2: 普通翻页 (无时间范围，纯按排名) -> 使用 ZRange
		start := int64((p.Page - 1) * p.Size)
		stop := int64(p.Page*p.Size - 1)

		if p.Order == models.SortDirectionDesc {
			res, err = client.ZRevRange(ctx, targetKey, start, stop).Result()
		} else {
			res, err = client.ZRange(ctx, targetKey, start, stop).Result()
		}
	}

	return res, err
}

// genPostKey 确定基础 key 以及是否需要聚合计算
func genPostKey(ctx context.Context, commID int64, sortBy models.SortField) (targetKey string, err error) {
	baseKey := getRedisKey(KeyPostScoreZset)
	if sortBy == models.SortFieldCreateTime {
		baseKey = getRedisKey(KeyPostTimeZset)
	}

	targetKey = baseKey
	if commID > 0 {
		communityKey := getRedisKey(KeyCommunitySetPF + strconv.Itoa(int(commID)))

		tempKey := fmt.Sprintf("temp:post:%d:%d", commID, time.Now().UnixNano())
		defer client.Del(ctx, tempKey)

		// AGGREGATE MAX:
		// 社区 Set 里的分数通常是 0 或无关紧要。
		// ZSet 里的分数是时间戳或热度。
		// 取 MAX 或 SUM 都能保留原 ZSet 的分数特性（前提是 Set 里分数不干扰）。
		err = client.ZInterStore(ctx, tempKey, &redis.ZStore{
			Keys:      []string{baseKey, communityKey},
			Weights:   []float64{1, 0}, // 权重: ZSet=1, CommunitySet=0 (忽略社区Set原本的分数)
			Aggregate: "MAX",
		}).Err()
		if err != nil {
			return "", err
		}
		targetKey = tempKey
	}

	return targetKey, nil
}
