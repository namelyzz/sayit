package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/namelyzz/sayit/models"
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
	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}

	data, err := service.GetHotCommunityList(limit)
	if err != nil {
		zap.L().Error("service.GetHotCommunityList() failed", zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}
	api.ResponseSuccess(c, data)
}

func RandomCommunityHandler(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}

	data, err := service.GetRandomCommunityList(limit)
	if err != nil {
		zap.L().Error("service.GetRandomCommunityList() failed", zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}
	api.ResponseSuccess(c, data)
}

// FollowCommunityHandler 关注社区
func FollowCommunityHandler(c *gin.Context) {
	userID, err := api.GetCurrentUserID(c)
	if err != nil {
		api.ResponseError(c, api.CodeInvalidToken)
		return
	}

	var req struct {
		CommunityID string `json:"community_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	communityID, err := strconv.ParseInt(req.CommunityID, 10, 64)
	if err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	if err := service.FollowCommunity(userID, communityID); err != nil {
		zap.L().Error("service.FollowCommunity() failed", zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	api.ResponseSuccess(c, nil)
}

// UnfollowCommunityHandler 取消关注社区
func UnfollowCommunityHandler(c *gin.Context) {
	userID, err := api.GetCurrentUserID(c)
	if err != nil {
		api.ResponseError(c, api.CodeInvalidToken)
		return
	}

	var req struct {
		CommunityID string `json:"community_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	communityID, err := strconv.ParseInt(req.CommunityID, 10, 64)
	if err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	if err := service.UnfollowCommunity(userID, communityID); err != nil {
		zap.L().Error("service.UnfollowCommunity() failed", zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	api.ResponseSuccess(c, nil)
}

// IsFollowedCommunityHandler 检查是否已关注社区
func IsFollowedCommunityHandler(c *gin.Context) {
	userID, err := api.GetCurrentUserID(c)
	if err != nil {
		api.ResponseError(c, api.CodeInvalidToken)
		return
	}

	communityIDStr := c.Query("community_id")
	if communityIDStr == "" {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	communityID, err := strconv.ParseInt(communityIDStr, 10, 64)
	if err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	isFollowed, err := service.IsFollowingCommunity(userID, communityID)
	if err != nil {
		zap.L().Error("service.IsFollowingCommunity() failed", zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	api.ResponseSuccess(c, &models.FollowStatus{IsFollowed: isFollowed})
}

// GetFollowedCommunityListHandler 获取已关注社区列表
func GetFollowedCommunityListHandler(c *gin.Context) {
	userID, err := api.GetCurrentUserID(c)
	if err != nil {
		api.ResponseError(c, api.CodeInvalidToken)
		return
	}

	communities, err := service.GetFollowedCommunityList(userID)
	if err != nil {
		zap.L().Error("service.GetFollowedCommunityList() failed", zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	api.ResponseSuccess(c, communities)
}
