package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/service"
	"github.com/namelyzz/sayit/utils/api"
	"go.uber.org/zap"
)

// GetNotificationUnreadCountHandler 获取当前用户未读通知数。
// 路由: GET /api/v1/notifications/unread_count (需要JWT认证)
func GetNotificationUnreadCountHandler(c *gin.Context) {
	userID, err := api.GetCurrentUserID(c)
	if err != nil {
		api.ResponseError(c, api.CodeNeedLogin)
		return
	}

	count, err := service.GetNotificationUnreadCount(c.Request.Context(), userID)
	if err != nil {
		zap.L().Error("get notification unread count failed", zap.Int64("userID", userID), zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	api.ResponseSuccess(c, &models.NotificationUnreadCountResponse{Count: count})
}

// GetNotificationListHandler 获取当前用户通知列表。
// 路由: GET /api/v1/notifications?page=1&size=20&status=all (需要JWT认证)
func GetNotificationListHandler(c *gin.Context) {
	userID, err := api.GetCurrentUserID(c)
	if err != nil {
		api.ResponseError(c, api.CodeNeedLogin)
		return
	}

	p := new(models.ParamNotificationList)
	if err := c.ShouldBindQuery(p); err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	data, err := service.GetNotificationList(c.Request.Context(), userID, p)
	if err != nil {
		zap.L().Error("get notification list failed", zap.Int64("userID", userID), zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	api.ResponseSuccess(c, data)
}

// MarkNotificationReadHandler 标记单条通知已读。
// 路由: POST /api/v1/notifications/:id/read (需要JWT认证)
func MarkNotificationReadHandler(c *gin.Context) {
	userID, err := api.GetCurrentUserID(c)
	if err != nil {
		api.ResponseError(c, api.CodeNeedLogin)
		return
	}

	notificationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	if err := service.MarkNotificationRead(c.Request.Context(), userID, notificationID); err != nil {
		zap.L().Error("mark notification read failed",
			zap.Int64("userID", userID),
			zap.Int64("notificationID", notificationID),
			zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	api.ResponseSuccess(c, nil)
}

// MarkAllNotificationsReadHandler 标记当前用户所有通知已读。
// 路由: POST /api/v1/notifications/read_all (需要JWT认证)
func MarkAllNotificationsReadHandler(c *gin.Context) {
	userID, err := api.GetCurrentUserID(c)
	if err != nil {
		api.ResponseError(c, api.CodeNeedLogin)
		return
	}

	if err := service.MarkAllNotificationsRead(c.Request.Context(), userID); err != nil {
		zap.L().Error("mark all notifications read failed", zap.Int64("userID", userID), zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	api.ResponseSuccess(c, nil)
}
