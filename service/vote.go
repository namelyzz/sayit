package service

import (
	"context"
	"github.com/namelyzz/sayit/dao/redis"
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/utils/api"
	"math"
	"strconv"
)

/*
VoteForPost 为帖子投票，每一票的分值是 432分

投票规则：数轴：-1 (反对) <---> 0 (无) <---> 1 (赞成)

	   direction=1时，有两种情况：
	   	1. 之前没有投过票，现在投赞成票    --> 更新分数和投票记录  差值的绝对值：1  +432
	   	2. 之前投反对票，现在改投赞成票    --> 更新分数和投票记录  差值的绝对值：2  +432*2
	   direction=0时，有两种情况：
	   	1. 之前投过反对票，现在要取消投票  --> 更新分数和投票记录  差值的绝对值：1  +432
		2. 之前投过赞成票，现在要取消投票  --> 更新分数和投票记录  差值的绝对值：1  -432
	   direction=-1时，有两种情况：
	   	1. 之前没有投过票，现在投反对票    --> 更新分数和投票记录  差值的绝对值：1  -432
	   	2. 之前投赞成票，现在改投反对票    --> 更新分数和投票记录  差值的绝对值：2  -432*2

投票的限制：

	每个贴子自发表之日起一个星期之内允许用户投票，超过一个星期就不允许再投票了。
		1. 到期之后将redis中保存的赞成票数及反对票数存储到mysql表中
		2. 到期之后删除那个 KeyPostVotedZSetPF
*/
func VoteForPost(ctx context.Context, userID int64, p *models.ParamVote) (err error) {
	postID := p.PostID

	// 判断当前帖子是否可以投票，超过时间则不能再投票了
	if !redis.IsPostCreatedWithinOneWeek(ctx, postID) {
		return api.ErrorVoteTimeExpire
	}

	// 获取当前投票值, 验证投票有效性
	userIDStr := strconv.FormatInt(userID, 10)

	// 用户当前想要投的票：1(赞), 0(取消), -1(踩)
	newVote := float64(p.Direction)

	// 查询当前用户之前对该帖子的投票记录：1, 0, 或 -1
	// 如果没投过票，GetPostVoteScore 应返回 0
	curVote := redis.GetPostVoteScore(ctx, postID, userIDStr)

	// 如果用户的新票值和旧票值一致，说明是重复操作，直接返回
	if newVote == curVote {
		return api.ErrorVoteRepeated
	}

	// 计算分数更新的方向 (operate) 和权重 (diff)
	operate := 1
	// 如果新值 < 旧值 (例如 1变0, 0变-1, 1变-1)，说明是在减少权重（向左移动），即减分
	if newVote < curVote {
		operate = -1
	}

	// 计算票值的绝对差值，决定了基础分值的倍数
	// 差值为 1：表示 新投/取消 (如 0->1, 1->0) -> 变动 432分
	// 差值为 2：表示 反向改票 (如 -1->1, 1->-1) -> 变动 432*2 分
	diff := math.Abs(newVote - curVote)

	return redis.UpdatePostVote(ctx, userIDStr, postID, newVote, float64(operate), diff)
}
