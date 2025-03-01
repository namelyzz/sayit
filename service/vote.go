package service

import (
	"context"
	"github.com/namelyzz/sayit/dao/redis"
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/utils/api"
	"math"
	"strconv"
)

func VoteForPost(ctx context.Context, userID int64, p *models.ParamVote) (err error) {
	postID := p.PostID

	// 判断当前帖子是否可以投票
	if !redis.IsPostCreatedWithinOneWeek(ctx, postID) {
		return api.ErrorVoteTimeExpire
	}

	// 获取当前投票值, 验证投票有效性
	userIDStr := strconv.FormatInt(userID, 10)
	newVote := float64(p.Direction)
	curVote := redis.GetPostVoteScore(ctx, postID, userIDStr)
	if newVote == curVote {
		return api.ErrorVoteRepeated
	}

	// 计算投票变化
	operate := 1
	if newVote < curVote {
		operate = -1
	}
	diff := math.Abs(newVote - curVote)

	return redis.UpdatePostVote(ctx, userIDStr, postID, newVote, float64(operate), diff)
}
