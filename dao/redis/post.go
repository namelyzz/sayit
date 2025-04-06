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
func GetPostIDsInOrder(ctx context.Context, p *models.ParamPostList, offset, limit int) (res []string, err error) {
	if limit <= 0 {
		return []string{}, nil
	}

	targetKey, cleanup, err := getPostQueryKey(ctx, p.CommunityID, p.SortBy)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	if p.SortBy == models.SortFieldCreateTime && (p.StartTime != nil || p.EndTime != nil) {
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
			Offset: int64(offset),
			Count:  int64(limit),
		}

		if p.Order == models.SortDirectionDesc {
			return client.ZRevRangeByScore(ctx, targetKey, opt).Result()
		}
		return client.ZRangeByScore(ctx, targetKey, opt).Result()
	}

	start := int64(offset)
	stop := int64(offset + limit - 1)
	if p.Order == models.SortDirectionDesc {
		return client.ZRevRange(ctx, targetKey, start, stop).Result()
	}
	return client.ZRange(ctx, targetKey, start, stop).Result()
}

func getPostQueryKey(ctx context.Context, commID int64, sortBy models.SortField) (targetKey string, cleanup func(), err error) {
	baseKey := getRedisKey(KeyPostScoreZset)
	if sortBy == models.SortFieldCreateTime {
		baseKey = getRedisKey(KeyPostTimeZset)
	}

	cleanup = func() {}
	targetKey = baseKey
	if commID <= 0 {
		return targetKey, cleanup, nil
	}

	communityKey := getRedisKey(KeyCommunitySetPF + strconv.FormatInt(commID, 10))
	tempKey := getRedisKey(fmt.Sprintf("temp:post:%d:%d:%s", commID, time.Now().UnixNano(), sortBy))
	cleanup = func() {
		_ = client.Del(context.Background(), tempKey).Err()
	}

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
		return "", cleanup, err
	}

	return tempKey, cleanup, nil
}
