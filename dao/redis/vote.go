package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

// 投票系统常量
const (
	oneWeekInSeconds = 7 * 24 * 3600 // 7天的秒数，投票时间窗口
	ScorePerVote     = 432           // 每票的基础分值
)

// GetPostCreateTime 从 Redis 时间排行榜获取帖子的创建时间戳
// 原理: 从 sayit:post:time ZSet 中查询帖子 ID 对应的 score（即创建时间戳）
// 返回值: 创建时间的 Unix 时间戳（秒级），如果帖子不存在返回 0
func GetPostCreateTime(ctx context.Context, postID string) float64 {
	return client.ZScore(ctx, getRedisKey(KeyPostTimeZset), postID).Val()
}

// IsPostCreatedWithinOneWeek 检查帖子是否在发布 7 天内
// 用于判断帖子是否还允许投票
// 原理: 当前时间 - 帖子创建时间 < 7天(604800秒)
// 返回值: true=可以投票, false=已过期或帖子不存在
func IsPostCreatedWithinOneWeek(ctx context.Context, postID string) bool {
	createTime := GetPostCreateTime(ctx, postID)
	if createTime == 0 {
		return false // 帖子不在 Redis 时间排行榜中（可能未创建或数据丢失）
	}
	return time.Now().Unix()-int64(createTime) < oneWeekInSeconds
}

// GetPostVoteScore 获取用户对某帖子的当前投票状态
// 原理: 从 sayit:post:voted:<postID> ZSet 中查询用户 ID 对应的 score
// 返回值: 1(赞成), -1(反对), 0(未投过票或记录不存在)
func GetPostVoteScore(ctx context.Context, postID, userID string) float64 {
	return client.ZScore(ctx, getRedisKey(KeyPostVotedZsetPF+postID), userID).Val()
}

// GetPostVoteCount 获取帖子的净投票数（赞成票数 - 反对票数）。
func GetPostVoteCount(ctx context.Context, postID string) int64 {
	key := getRedisKey(KeyPostVotedZsetPF + postID)
	upCount := client.ZCount(ctx, key, "1", "1").Val()
	downCount := client.ZCount(ctx, key, "-1", "-1").Val()
	return upCount - downCount
}

// GetPostVoteValue 获取帖子的投票分值。
func GetPostVoteValue(ctx context.Context, postID string) int64 {
	return GetPostVoteCount(ctx, postID) * ScorePerVote
}

// UpdatePostVote 更新帖子分数与用户投票记录
// ═══════════════════════════════════════════════════════════════════════════════
//
// 该函数在一个 Redis 事务管道（TxPipeline）中原子性地完成以下操作:
//
//	操作1: ZIncrBy - 更新帖子的全局热度分数 (sayit:post:score)
//	操作2: ZAdd/ZRem - 更新或移除用户的投票记录 (sayit:post:voted:<postID>)
//
// ═══════════════════════════════════════════════════════════════════════════════
//
// 参数说明:
//
//	ctx      - 上下文，支持超时和取消
//	userID   - 用户 ID（字符串形式）
//	postID   - 帖子 ID（字符串形式）
//	voteVal  - 用户最终的投票状态
//	           1: 赞成 → ZAdd 记录
//	           -1: 反对 → ZAdd 记录
//	           0: 取消投票 → ZRem 删除记录
//	operate  - 分数变化方向系数
//	           1: 分数增加（投赞成、取消反对）
//	           -1: 分数减少（投反对、取消赞成）
//	diff     - 分数变化幅度系数
//	           1: 普通投票/取消（如 0→1, 1→0, 0→-1, -1→0）
//	           2: 反向改票（如 1→-1, -1→1）
//
// ═══════════════════════════════════════════════════════════════════════════════
//
// 分数计算公式:
//
//	实际分数变化 = operate × diff × 432
//
// 示例:
//
//	没投过 → 赞成:   voteVal=1,  operate=+1, diff=1 → +432
//	赞成 → 取消:     voteVal=0,  operate=-1, diff=1 → -432
//	赞成 → 反对:     voteVal=-1, operate=-1, diff=2 → -864
//	反对 → 赞成:     voteVal=1,  operate=+1, diff=2 → +864
//	反对 → 取消:     voteVal=0,  operate=+1, diff=1 → +432
//
// ═══════════════════════════════════════════════════════════════════════════════
func UpdatePostVote(ctx context.Context, userID, postID string, voteVal, operate, diff float64) error {
	// 开启 Redis 事务管道，确保后续操作要么全成功，要么全失败
	pipe := client.TxPipeline()

	// ── 操作1: 更新帖子热度分数 ──
	// Key: sayit:post:score (ZSet)
	// 计算: 当前分数 + (operate × diff × 432)
	// 例如: 赞成票 → 分数 +432，反对改赞成 → 分数 +864
	pipe.ZIncrBy(ctx, getRedisKey(KeyPostScoreZset), operate*diff*ScorePerVote, postID)

	// ── 操作2: 更新用户投票记录 ──
	// Key: sayit:post:voted:<postID> (ZSet)
	if voteVal == 0 {
		// 取消投票: 从投票记录中移除该用户
		pipe.ZRem(ctx, getRedisKey(KeyPostVotedZsetPF+postID), userID)
	} else {
		// 投票/改票: 记录用户的投票方向（1 或 -1）
		pipe.ZAdd(ctx, getRedisKey(KeyPostVotedZsetPF+postID), redis.Z{
			Score:  voteVal, // 投票方向: 1 或 -1
			Member: userID,  // 用户ID
		})
	}

	// 执行事务管道（原子性提交）
	_, err := pipe.Exec(ctx)
	return err
}
