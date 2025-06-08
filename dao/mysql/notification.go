package mysql

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/namelyzz/sayit/models"
	"gorm.io/gorm"
)

const (
	notificationStatusAll    = "all"
	notificationStatusUnread = "unread"
	maxNotificationPageSize  = 50
)

// CreateNotification 写入通知记录。唯一键冲突表示重复事件，返回 created=false。
func CreateNotification(ctx context.Context, n *models.Notification) (created bool, err error) {
	err = db.WithContext(ctx).Omit("ID", "ReadTime").Create(n).Error
	if err != nil {
		if IsDuplicateNotificationError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ListNotifications 分页查询用户通知列表。
func ListNotifications(ctx context.Context, userID int64, page, size int, status string) ([]*models.NotificationItem, int64, error) {
	page, size = normalizeNotificationPagination(page, size)
	query := buildNotificationListQuery(ctx, userID, status)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*models.NotificationItem
	err := query.
		Order("n.create_time DESC").
		Offset((page - 1) * size).
		Limit(size).
		Scan(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// CountUnreadNotifications 统计用户未读通知数。
func CountUnreadNotifications(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("recipient_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

// MarkNotificationRead 标记单条通知已读，仅允许用户操作自己的通知。
func MarkNotificationRead(ctx context.Context, userID, notificationID int64) (int64, error) {
	now := time.Now()
	res := db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("recipient_id = ? AND notification_id = ? AND is_read = ?", userID, notificationID, false).
		Updates(map[string]any{
			"is_read":   true,
			"read_time": now,
		})
	return res.RowsAffected, res.Error
}

// MarkAllNotificationsRead 标记用户所有未读通知为已读。
func MarkAllNotificationsRead(ctx context.Context, userID int64) (int64, error) {
	now := time.Now()
	res := db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("recipient_id = ? AND is_read = ?", userID, false).
		Updates(map[string]any{
			"is_read":   true,
			"read_time": now,
		})
	return res.RowsAffected, res.Error
}

func buildNotificationListQuery(ctx context.Context, userID int64, status string) *gorm.DB {
	query := db.WithContext(ctx).
		Table("notifications n").
		Select(`n.id, n.notification_id, n.recipient_id, n.actor_id, u.username AS actor_name,
			n.type, n.post_id, n.comment_id, n.parent_id, n.direction, n.title, n.content,
			n.link, n.is_read, n.dedupe_key, n.create_time, n.read_time`).
		Joins("LEFT JOIN users u ON n.actor_id = u.user_id").
		Where("n.recipient_id = ?", userID)

	if status == notificationStatusUnread {
		query = query.Where("n.is_read = ?", false)
	}
	return query
}

func normalizeNotificationPagination(page, size int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > maxNotificationPageSize {
		size = maxNotificationPageSize
	}
	return page, size
}

// NormalizeNotificationStatus 归一化通知列表状态参数。
func NormalizeNotificationStatus(status string) string {
	if status == notificationStatusUnread {
		return notificationStatusUnread
	}
	return notificationStatusAll
}

// IsDuplicateNotificationError 判断是否为通知唯一键冲突错误。
func IsDuplicateNotificationError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "duplicated")
}
