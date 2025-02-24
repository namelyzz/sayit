package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/service"
	"github.com/namelyzz/sayit/utils/api"
	"go.uber.org/zap"
	"strconv"
)

func CreatePostHandler(c *gin.Context) {
	p := new(models.Post)
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Error("create post with invalid param")
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	userID, err := api.GetCurrentUserID(c)
	if err != nil {
		api.ResponseError(c, api.CodeNeedLogin)
		return
	}

	p.AuthorID = userID
	if err = service.CreatePost(p); err != nil {
		zap.L().Error("service.CreatePost() failed",
			zap.Error(err),
			zap.Int64("userID", userID),
		)
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	api.ResponseSuccess(c, nil)
}

func GetPostDetailHandler(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		zap.L().Error(
			"get post detail with invalid param",
			zap.String("invalid post id", postIDStr),
			zap.Error(err),
		)
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	data, err := service.GetPostDetailByID(postID)
	if err != nil {
		zap.L().Error("service.GetPostDetailByID failed", zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	api.ResponseSuccess(c, data)
}
