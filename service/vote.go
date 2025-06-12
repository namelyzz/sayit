package service

import (
	"context"
	"github.com/namelyzz/sayit/dao/mysql"
	"github.com/namelyzz/sayit/dao/redis"
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/utils/api"
	"go.uber.org/zap"
	"math"
	"strconv"
)

// 投票系统的分数常量
const scorePerVote = 432 // 每票的基础分值（已在 dao/redis/vote.go 中定义，此处保留说明）

var (
	isPostCreatedWithinOneWeekFunc = redis.IsPostCreatedWithinOneWeek
	getPostVoteScoreFunc           = redis.GetPostVoteScore
	updatePostVoteFunc             = redis.UpdatePostVote
	getPostForVoteFunc             = mysql.GetPostByID
)

// VoteForPost 为帖子投票的核心业务逻辑
//
// ═══════════════════════════════════════════════════════════════════════════════
// 投票数轴模型:
//
//	-1 (反对) ◄──────── 0 (无票) ────────► 1 (赞成)
//
// ═══════════════════════════════════════════════════════════════════════════════
//
// 投票规则与分数变化（每票 432 分）:
//
//	┌─────────────────────────────────────────────────────────────────────────┐
//	│ 当前状态     │ 新操作        │ 分数变化  │ 说明                        │
//	├─────────────────────────────────────────────────────────────────────────┤
//	│ 无票 (0)     │ 赞成 (+1)     │ +432     │ 新投赞成票                  │
//	│ 无票 (0)     │ 反对 (-1)     │ -432     │ 新投反对票                  │
//	│ 赞成 (+1)    │ 取消 (0)      │ -432     │ 取消赞成票                  │
//	│ 赞成 (+1)    │ 反对 (-1)     │ -864     │ 从赞成改为反对（差值为2）   │
//	│ 反对 (-1)    │ 取消 (0)      │ +432     │ 取消反对票（分数恢复）      │
//	│ 反对 (-1)    │ 赞成 (+1)     │ +864     │ 从反对改为赞成（差值为2）   │
//	└─────────────────────────────────────────────────────────────────────────┘
//
// 计算公式:
//
//	diff = |newVote - curVote|        (差值绝对值: 1 或 2)
//	operate = sign(newVote - curVote)  (方向: +1 加分, -1 减分)
//	分数变化 = operate × diff × 432
//
// ═══════════════════════════════════════════════════════════════════════════════
//
// 投票限制:
//
//	帖子发布超过 7 天后不允许投票
//	（通过检查 Redis 时间排行榜中的帖子创建时间判断）
//
// ═══════════════════════════════════════════════════════════════════════════════
//
// 数据存储:
//
//	投票记录: Redis ZSet, Key=sayit:post:voted:<postID>, Score=投票方向, Member=用户ID
//	帖子分数: Redis ZSet, Key=sayit:post:score, Score=热度分数, Member=帖子ID
//
// ═══════════════════════════════════════════════════════════════════════════════
func VoteForPost(ctx context.Context, userID int64, p *models.ParamVote) (err error) {
	postID := p.PostID

	// ─── Step 1: 投票时间校验 ───
	// 检查帖子是否在发布 7 天内，超过 7 天不允许投票
	// 原理: 从 Redis 时间排行榜获取帖子创建时间，与当前时间比较
	if !isPostCreatedWithinOneWeekFunc(ctx, postID) {
		return api.ErrorVoteTimeExpire
	}

	// ─── Step 2: 获取当前投票状态 ───
	userIDStr := strconv.FormatInt(userID, 10)

	// 用户想要投的新票值: 1(赞成), 0(取消), -1(反对)
	newVote := float64(p.Direction)

	// 查询用户之前对该帖子的投票记录
	// 如果没投过票，返回 0（Redis ZScore 对不存在的 key 返回 0）
	curVote := getPostVoteScoreFunc(ctx, postID, userIDStr)

	// ─── Step 3: 重复投票校验 ───
	// 如果新票值与旧票值相同，说明是重复操作，拒绝处理
	if newVote == curVote {
		return api.ErrorVoteRepeated
	}

	// ─── Step 4: 计算分数变化参数 ───
	// operate: 分数变化方向，+1 表示加分，-1 表示减分
	operate := 1
	if newVote < curVote {
		operate = -1 // 新值 < 旧值，说明在减少权重（向左移动），即减分
	}

	// diff: 票值的绝对差值，决定了分数变化的倍数
	// 差值为 1: 普通投票/取消 (如 0→1, 1→0, 0→-1, -1→0) → 变动 432 分
	// 差值为 2: 反向改票 (如 -1→1, 1→-1) → 变动 432×2 = 864 分
	diff := math.Abs(newVote - curVote)

	// ─── Step 5: 更新 Redis ───
	// 在同一个 Redis 事务管道中同时更新帖子分数和投票记录
	if err := updatePostVoteFunc(ctx, userIDStr, postID, newVote, float64(operate), diff); err != nil {
		return err
	}
	if newVote != 0 {
		postIDInt, parseErr := strconv.ParseInt(postID, 10, 64)
		if parseErr != nil {
			return api.ErrorInvalidID
		}
		post, err := getPostForVoteFunc(postIDInt)
		if err != nil {
			zap.L().Warn("get post for vote notification failed", zap.String("postID", postID), zap.Error(err))
			return nil
		}
		PublishPostVotedNotification(ctx, userID, post, p.Direction)
	}
	return nil
}
