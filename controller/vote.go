package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/service"
	"github.com/namelyzz/sayit/utils/api"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// PostVoteController 帖子投票接口
// 路由: POST /api/v1/vote (需要JWT认证)
// 请求体: JSON { "post_id": "123456", "direction": 1 }
// 流程: 参数绑定 -> 获取当前用户ID -> 调用 service.VoteForPost -> 返回结果
//
// 投票方向 (direction):
//   - 1: 赞成票（upvote）
//   - 0: 取消投票（unvote）
//   - -1: 反对票（downvote）
//
// 业务错误:
//   - 投票时间已过（帖子发布超过7天）
//   - 重复投票（新旧投票方向相同）
func PostVoteController(c *gin.Context) {
	// 1. 绑定并校验 JSON 请求参数
	// post_id 为必填，direction 必须是 1、0、-1 之一
	p := new(models.ParamVote)
	if err := c.ShouldBindJSON(p); err != nil {
		handleBindError(c, err)
		return
	}

	// 2. 从 JWT 上下文获取当前登录用户 ID
	userID, err := api.GetCurrentUserID(c)
	if err != nil {
		api.ResponseError(c, api.CodeNeedLogin)
		return
	}

	// 3. 调用 service 层执行投票逻辑
	if err := service.VoteForPost(c.Request.Context(), userID, p); err != nil {
		// 区分业务错误和系统错误
		// 业务错误: 投票时间过期、重复投票 → 返回具体错误信息
		// 系统错误: Redis 异常等 → 返回服务器繁忙
		if errors.Is(err, api.ErrorVoteTimeExpire) || errors.Is(err, api.ErrorVoteRepeated) {
			api.ResponseErrorWithMsg(c, api.CodeInvalidParam, err.Error())
		} else {
			zap.L().Error("service.VoteForPost failed",
				zap.Int64("userID", userID),
				zap.String("postID", p.PostID),
				zap.Error(err))
			api.ResponseError(c, api.CodeServerBusy)
		}
		return
	}

	// 4. 投票成功
	api.ResponseSuccess(c, nil)
}
