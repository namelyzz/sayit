package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/service"
	"github.com/namelyzz/sayit/utils/api"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func PostVoteController(c *gin.Context) {
	p := new(models.ParamVote)
	if err := c.ShouldBindJSON(p); err != nil {
		handleBindError(c, err)
		return
	}

	userID, err := api.GetCurrentUserID(c)
	if err != nil {
		api.ResponseError(c, api.CodeNeedLogin)
		return
	}

	if err := service.VoteForPost(c.Request.Context(), userID, p); err != nil {
		// 区分业务错误和系统错误
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

	api.ResponseSuccess(c, nil)
}
