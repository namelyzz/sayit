package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/service"
	"github.com/namelyzz/sayit/utils/api"
	"github.com/pkg/errors"
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

// GetCommunitiesHandler 获取社区列表（支持搜索和分页）
func GetCommunitiesHandler(c *gin.Context) {
	keyword := c.Query("keyword")
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size <= 0 || size > 50 {
		size = 20
	}

	data, err := service.GetCommunityListWithSearch(keyword, page, size)
	if err != nil {
		zap.L().Error("service.GetCommunityListWithSearch() failed", zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	api.ResponseSuccess(c, data)
}

// CreateCommunityHandler 创建社区
func CreateCommunityHandler(c *gin.Context) {
	// 检查用户是否登录
	_, err := api.GetCurrentUserID(c)
	if err != nil {
		api.ResponseError(c, api.CodeNeedLogin)
		return
	}

	var req models.ParamCreateCommunity
	if err := c.ShouldBindJSON(&req); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	community, err := service.CreateCommunity(req.Name, req.Introduction)
	if err != nil {
		if errors.Is(err, api.ErrorCommunityExist) {
			api.ResponseError(c, api.CodeCommunityExist)
			return
		}
		zap.L().Error("service.CreateCommunity() failed", zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	api.ResponseSuccess(c, community)
}
