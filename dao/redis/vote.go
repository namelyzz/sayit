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

// UpdatePostVote 更新帖子分数与用户投票记录
// 该函数在一个 Redis Pipeline 中原子性地完成以下操作：
// 1. ZIncrBy: 根据 operate 和 diff 更新帖子的全局分数 (KeyPostScoreZset)
// 2. ZAdd/ZRem: 更新或移除用户在帖子下的投票状态 (KeyPostVotedZsetPF)
//
// 参数说明:
//   - ctx:      上下文
//   - userID:   用户 ID
//   - postID:   帖子 ID
//   - voteVal:  用户当前的最终投票状态。
//     取值: 1(赞成), -1(反对), 0(取消投票/无记录)。
//     如果为 0，会从 Redis 记录中移除该用户；否则更新为对应值。
//   - operate:  分数变化的方向系数。
//     取值: 1 (表示分数增加), -1 (表示分数减少)。
//     推导逻辑: if newVote > curVote then 1 else -1。
//   - diff:     分数变化的幅度系数（绝对值）。
//     取值: 1 (普通投票/取消), 2 (从赞成改反对，或反之)。
//     推导逻辑: |newVote - curVote|。
//
// 示例场景:
//   - 没投过 -> 投赞成: voteVal=1, operate=1, diff=1  (总分 +432)
//   - 投赞成 -> 投反对: voteVal=-1, operate=-1, diff=2 (总分 -864)
//   - 投反对 -> 取消:   voteVal=0, operate=1, diff=1  (总分 +432，即把之前扣的补回来)
func UpdatePostVote(ctx context.Context, userID, postID string, voteVal, operate, diff float64) error {
	// 开启 Redis 事务管道，确保后续操作要么全成功，要么全失败
	pipe := client.TxPipeline()

	// 更新帖子分数
	// 计算公式：操作方向 * 差值 * 每票的分值
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
