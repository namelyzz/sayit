package models

import "time"

const (
	NotificationTypeCommentLiked   = "comment_liked"
	NotificationTypePostCommented  = "post_commented"
	NotificationTypePostVoted      = "post_voted"
	NotificationTypeCommentReplied = "comment_replied"
	NotificationTypeUserFollowed   = "user_followed"
)

// Notification 站内通知模型，对应 notifications 表。
type Notification struct {
	ID             int64        `json:"-" gorm:"column:id;primaryKey"`
	NotificationID SnowflakeID  `json:"notification_id" gorm:"column:notification_id"`
	RecipientID    SnowflakeID  `json:"recipient_id" gorm:"column:recipient_id"`
	ActorID        SnowflakeID  `json:"actor_id" gorm:"column:actor_id"`
	Type           string       `json:"type" gorm:"column:type"`
	PostID         *SnowflakeID `json:"post_id,omitempty" gorm:"column:post_id"`
	CommentID      *SnowflakeID `json:"comment_id,omitempty" gorm:"column:comment_id"`
	ParentID       *SnowflakeID `json:"parent_id,omitempty" gorm:"column:parent_id"`
	Direction      *int8        `json:"direction,omitempty" gorm:"column:direction"`
	Title          string       `json:"title" gorm:"column:title"`
	Content        string       `json:"content" gorm:"column:content"`
	Link           string       `json:"link" gorm:"column:link"`
	IsRead         bool         `json:"is_read" gorm:"column:is_read"`
	DedupeKey      string       `json:"-" gorm:"column:dedupe_key"`
	CreateTime     time.Time    `json:"create_time" gorm:"column:create_time;autoCreateTime"`
	ReadTime       *time.Time   `json:"read_time,omitempty" gorm:"column:read_time"`
}

func (Notification) TableName() string {
	return "notifications"
}

// NotificationEvent 是 Redis Stream 中传递的通知事件。
type NotificationEvent struct {
	EventID     int64  `json:"event_id"`
	Type        string `json:"type"`
	RecipientID int64  `json:"recipient_id"`
	ActorID     int64  `json:"actor_id"`
	PostID      int64  `json:"post_id,omitempty"`
	CommentID   int64  `json:"comment_id,omitempty"`
	ParentID    int64  `json:"parent_id,omitempty"`
	Direction   int8   `json:"direction,omitempty"`
	CreatedAt   int64  `json:"created_at"`
	DedupeKey   string `json:"dedupe_key"`
}

// NotificationItem 是通知列表接口返回项，额外包含动作发起人的用户名。
type NotificationItem struct {
	Notification
	ActorName string `json:"actor_name" gorm:"column:actor_name"`
}

type NotificationListResponse struct {
	List        []*NotificationItem `json:"list"`
	Total       int64               `json:"total"`
	UnreadCount int64               `json:"unread_count"`
}

type NotificationUnreadCountResponse struct {
	Count int64 `json:"count"`
}

type ParamNotificationList struct {
	Page   int    `json:"page" form:"page"`
	Size   int    `json:"size" form:"size"`
	Status string `json:"status" form:"status"`
}
