package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

const (
	oneWeekInSeconds = 7 * 24 * 3600
	scorePerVore     = 432
)

func GetPostCreateTime(ctx context.Context, postID string) float64 {
	return client.ZScore(ctx, getRedisKey(KeyPostTimeZset), postID).Val()
}

func IsPostCreatedWithinOneWeek(ctx context.Context, postID string) bool {
	createTime := GetPostCreateTime(ctx, postID)
	if createTime == 0 {
		return false // 帖子不存在
	}
	return time.Now().Unix()-int64(createTime) < oneWeekInSeconds
}

func GetPostVoteScore(ctx context.Context, postID, userID string) float64 {
	return client.ZScore(ctx, getRedisKey(KeyPostVotedZsetPF+postID), userID).Val()
}

func UpdatePostVote(ctx context.Context, userID, postID string, voteVal, operate, diff float64) error {
	pipe := client.TxPipeline()

	// 更新帖子分数
	pipe.ZIncrBy(ctx, getRedisKey(KeyPostScoreZset), operate*diff*scorePerVore, postID)

	// 记录用户投票数据
	if voteVal == 0 {
		pipe.ZRem(ctx, getRedisKey(KeyPostVotedZsetPF+postID), userID)
	} else {
		pipe.ZAdd(ctx, getRedisKey(KeyPostVotedZsetPF+postID), redis.Z{
			Score:  voteVal,
			Member: userID,
		})
	}

	_, err := pipe.Exec(ctx)
	return err
}
