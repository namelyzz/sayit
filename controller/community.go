package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/namelyzz/sayit/service"
	"github.com/namelyzz/sayit/utils/api"
	"go.uber.org/zap"
	"strconv"
)

func CommunityHandler(c *gin.Context) {
	data, err := service.GetCommunityList()
	if err != nil {
		zap.L().Error("service.GetCommunityList() failed", zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}
	api.ResponseSuccess(c, data)
}

func CommunityDetailHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	data, err := service.GetCommunityDetailByID(id)
	if err != nil {
		zap.L().Error("service.GetCommunityDetailByID() failed", zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}
	api.ResponseSuccess(c, data)
}

func HotCommunityHandler(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		zap.L().Warn("invalid limit parameter, use default",
			zap.String("limit", limitStr), zap.Error(err))
		limit = 10
	} else if limit <= 0 {
		zap.L().Warn("limit must be positive, use default",
			zap.String("limit", limitStr), zap.Int("parsed", limit))
		limit = 10
	}

	data, err := service.GetHotCommunityList(limit)
	if err != nil {
		zap.L().Error("service.GetHotCommunityList() failed", zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}
	api.ResponseSuccess(c, data)
}
