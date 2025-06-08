package service

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/namelyzz/sayit/dao/mysql"
	"github.com/namelyzz/sayit/dao/redis"
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/utils/snowflake"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	notificationWorkerBatchSize = 50
	notificationWorkerBlock     = 5 * time.Second
	notificationWorkerIdleSleep = time.Second
)

var (
	ensureNotificationConsumerGroupFunc = redis.EnsureNotificationConsumerGroup
	readNotificationEventsFunc          = redis.ReadNotificationEvents
	ackNotificationEventFunc            = redis.AckNotificationEvent
	incrNotificationUnreadFunc          = redis.IncrNotificationUnread
	getNotificationUnreadFunc           = redis.GetNotificationUnread
	setNotificationUnreadFunc           = redis.SetNotificationUnread
	decrNotificationUnreadFunc          = redis.DecrNotificationUnread
	publishNotificationEventFunc        = redis.PublishNotificationEvent

	createNotificationFunc       = mysql.CreateNotification
	listNotificationsFunc        = mysql.ListNotifications
	countUnreadNotificationsFunc = mysql.CountUnreadNotifications
	markNotificationReadFunc     = mysql.MarkNotificationRead
	markAllNotificationsReadFunc = mysql.MarkAllNotificationsRead
	genNotificationIDFunc        = snowflake.GenID
)

// PublishCommentLikedNotification 发布评论被点赞通知。通知失败只记录日志，不影响主业务。
func PublishCommentLikedNotification(ctx context.Context, actorID int64, comment *models.Comment) {
	if comment == nil || int64(comment.AuthorID) == 0 || actorID == int64(comment.AuthorID) {
		return
	}
	event := newNotificationEvent(
		models.NotificationTypeCommentLiked,
		int64(comment.AuthorID),
		actorID,
		int64(comment.PostID),
		int64(comment.CommentID),
		0,
		0,
		fmt.Sprintf("comment_liked:%d:%d", comment.CommentID, actorID),
	)
	publishNotification(ctx, event)
}

// PublishCommentRepliedNotification 发布评论被直接回复通知。
func PublishCommentRepliedNotification(ctx context.Context, actorID int64, comment *models.Comment, parent *models.Comment) {
	if comment == nil || parent == nil || int64(parent.AuthorID) == 0 || actorID == int64(parent.AuthorID) {
		return
	}
	event := newNotificationEvent(
		models.NotificationTypeCommentReplied,
		int64(parent.AuthorID),
		actorID,
		int64(comment.PostID),
		int64(comment.CommentID),
		int64(parent.CommentID),
		0,
		fmt.Sprintf("comment_replied:%d", comment.CommentID),
	)
	publishNotification(ctx, event)
}

// PublishPostVotedNotification 发布帖子被投票通知。
func PublishPostVotedNotification(ctx context.Context, actorID int64, post *models.Post, direction int8) {
	if post == nil || int64(post.AuthorID) == 0 || direction == 0 || actorID == int64(post.AuthorID) {
		return
	}
	event := newNotificationEvent(
		models.NotificationTypePostVoted,
		int64(post.AuthorID),
		actorID,
		int64(post.PostID),
		0,
		0,
		direction,
		fmt.Sprintf("post_voted:%d:%d", post.PostID, actorID),
	)
	publishNotification(ctx, event)
}

// PublishUserFollowedNotification 发布用户被关注通知。
func PublishUserFollowedNotification(ctx context.Context, actorID, recipientID int64) {
	if actorID == 0 || recipientID == 0 || actorID == recipientID {
		return
	}
	event := newNotificationEvent(
		models.NotificationTypeUserFollowed,
		recipientID,
		actorID,
		0,
		0,
		0,
		0,
		fmt.Sprintf("user_followed:%d:%d", recipientID, actorID),
	)
	publishNotification(ctx, event)
}

func newNotificationEvent(eventType string, recipientID, actorID, postID, commentID, parentID int64, direction int8, dedupeKey string) *models.NotificationEvent {
	return &models.NotificationEvent{
		EventID:     genNotificationIDFunc(),
		Type:        eventType,
		RecipientID: recipientID,
		ActorID:     actorID,
		PostID:      postID,
		CommentID:   commentID,
		ParentID:    parentID,
		Direction:   direction,
		CreatedAt:   nowFunc().Unix(),
		DedupeKey:   dedupeKey,
	}
}

func publishNotification(ctx context.Context, event *models.NotificationEvent) {
	if _, err := publishNotificationEventFunc(ctx, event); err != nil {
		zap.L().Warn("publish notification event failed",
			zap.String("type", event.Type),
			zap.Int64("recipientID", event.RecipientID),
			zap.Int64("actorID", event.ActorID),
			zap.String("dedupeKey", event.DedupeKey),
			zap.Error(err))
	}
}

// StartNotificationWorker 启动通知消费 worker。
func StartNotificationWorker(ctx context.Context) {
	if err := ensureNotificationConsumerGroupFunc(ctx); err != nil {
		zap.L().Error("ensure notification consumer group failed", zap.Error(err))
		return
	}

	consumer := notificationConsumerName()
	for {
		if err := consumeNotificationBatch(ctx, consumer); err != nil {
			zap.L().Warn("consume notification batch failed", zap.Error(err))
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(notificationWorkerIdleSleep):
		}
	}
}

func consumeNotificationBatch(ctx context.Context, consumer string) error {
	messages, err := readNotificationEventsFunc(ctx, consumer, notificationWorkerBatchSize, notificationWorkerBlock)
	if err != nil {
		return err
	}
	for _, msg := range messages {
		if err := processNotificationMessage(ctx, msg); err != nil {
			zap.L().Warn("process notification message failed",
				zap.String("messageID", msg.ID),
				zap.Error(err))
		}
	}
	return nil
}

func processNotificationMessage(ctx context.Context, msg goredis.XMessage) error {
	event, err := notificationEventFromStreamValues(msg.Values)
	if err != nil {
		_ = ackNotificationEventFunc(ctx, msg.ID)
		return err
	}

	n, err := buildNotificationFromEvent(event)
	if err != nil {
		_ = ackNotificationEventFunc(ctx, msg.ID)
		return err
	}

	created, err := createNotificationFunc(ctx, n)
	if err != nil {
		return err
	}
	if created {
		if err := incrNotificationUnreadFunc(ctx, event.RecipientID); err != nil {
			return err
		}
	}
	return ackNotificationEventFunc(ctx, msg.ID)
}

// GetNotificationUnreadCount 获取用户未读通知数，Redis 未命中时从 MySQL 回填。
func GetNotificationUnreadCount(ctx context.Context, userID int64) (int64, error) {
	count, err := getNotificationUnreadFunc(ctx, userID)
	if err == nil {
		return count, nil
	}
	if err != goredis.Nil {
		zap.L().Warn("get notification unread cache failed", zap.Int64("userID", userID), zap.Error(err))
	}

	count, err = countUnreadNotificationsFunc(ctx, userID)
	if err != nil {
		return 0, err
	}
	if err := setNotificationUnreadFunc(ctx, userID, count); err != nil {
		zap.L().Warn("set notification unread cache failed", zap.Int64("userID", userID), zap.Int64("count", count), zap.Error(err))
	}
	return count, nil
}

// GetNotificationList 获取用户通知列表。
func GetNotificationList(ctx context.Context, userID int64, p *models.ParamNotificationList) (*models.NotificationListResponse, error) {
	if p == nil {
		p = &models.ParamNotificationList{}
	}
	status := mysql.NormalizeNotificationStatus(p.Status)
	list, total, err := listNotificationsFunc(ctx, userID, p.Page, p.Size, status)
	if err != nil {
		return nil, err
	}
	unreadCount, err := GetNotificationUnreadCount(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &models.NotificationListResponse{List: list, Total: total, UnreadCount: unreadCount}, nil
}

// MarkNotificationRead 标记单条通知已读。
func MarkNotificationRead(ctx context.Context, userID, notificationID int64) error {
	affected, err := markNotificationReadFunc(ctx, userID, notificationID)
	if err != nil {
		return err
	}
	return decrNotificationUnreadFunc(ctx, userID, affected)
}

// MarkAllNotificationsRead 标记所有通知已读。
func MarkAllNotificationsRead(ctx context.Context, userID int64) error {
	affected, err := markAllNotificationsReadFunc(ctx, userID)
	if err != nil {
		return err
	}
	return decrNotificationUnreadFunc(ctx, userID, affected)
}

func notificationConsumerName() string {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "unknown"
	}
	return fmt.Sprintf("notification-worker-%s-%d", hostname, os.Getpid())
}

func notificationEventFromStreamValues(values map[string]any) (*models.NotificationEvent, error) {
	event := &models.NotificationEvent{}
	var err error
	if event.EventID, err = streamInt64(values, "event_id", true); err != nil {
		return nil, err
	}
	if event.Type, err = streamString(values, "type", true); err != nil {
		return nil, err
	}
	if event.RecipientID, err = streamInt64(values, "recipient_id", true); err != nil {
		return nil, err
	}
	if event.ActorID, err = streamInt64(values, "actor_id", true); err != nil {
		return nil, err
	}
	if event.PostID, err = streamInt64(values, "post_id", false); err != nil {
		return nil, err
	}
	if event.CommentID, err = streamInt64(values, "comment_id", false); err != nil {
		return nil, err
	}
	if event.ParentID, err = streamInt64(values, "parent_id", false); err != nil {
		return nil, err
	}
	direction, err := streamInt64(values, "direction", false)
	if err != nil {
		return nil, err
	}
	event.Direction = int8(direction)
	if event.CreatedAt, err = streamInt64(values, "created_at", true); err != nil {
		return nil, err
	}
	if event.DedupeKey, err = streamString(values, "dedupe_key", true); err != nil {
		return nil, err
	}
	return event, validateNotificationEvent(event)
}

func validateNotificationEvent(event *models.NotificationEvent) error {
	if event.Type == "" || event.RecipientID == 0 || event.ActorID == 0 || event.DedupeKey == "" {
		return fmt.Errorf("invalid notification event required fields")
	}
	if event.Type == models.NotificationTypePostVoted && event.Direction == 0 {
		return fmt.Errorf("post_voted notification requires direction")
	}
	return nil
}

func buildNotificationFromEvent(event *models.NotificationEvent) (*models.Notification, error) {
	title, content, link, err := notificationPresentation(event)
	if err != nil {
		return nil, err
	}
	return &models.Notification{
		NotificationID: models.SnowflakeID(genNotificationIDFunc()),
		RecipientID:    models.SnowflakeID(event.RecipientID),
		ActorID:        models.SnowflakeID(event.ActorID),
		Type:           event.Type,
		PostID:         snowflakeIDPtr(event.PostID),
		CommentID:      snowflakeIDPtr(event.CommentID),
		ParentID:       snowflakeIDPtr(event.ParentID),
		Direction:      directionPtr(event.Direction),
		Title:          title,
		Content:        content,
		Link:           link,
		DedupeKey:      event.DedupeKey,
	}, nil
}

func notificationPresentation(event *models.NotificationEvent) (string, string, string, error) {
	switch event.Type {
	case models.NotificationTypeCommentLiked:
		return "有人点赞了你的评论", "点击查看评论", fmt.Sprintf("/post/%d?comment=%d", event.PostID, event.CommentID), nil
	case models.NotificationTypePostVoted:
		if event.Direction > 0 {
			return "有人赞成了你的帖子", "点击查看帖子", fmt.Sprintf("/post/%d", event.PostID), nil
		}
		return "有人反对了你的帖子", "点击查看帖子", fmt.Sprintf("/post/%d", event.PostID), nil
	case models.NotificationTypeCommentReplied:
		return "有人回复了你的评论", "点击查看回复", fmt.Sprintf("/post/%d?comment=%d", event.PostID, event.CommentID), nil
	case models.NotificationTypeUserFollowed:
		return "有人关注了你", "点击查看主页", fmt.Sprintf("/user/%d", event.ActorID), nil
	default:
		return "", "", "", fmt.Errorf("unsupported notification type: %s", event.Type)
	}
}

func streamString(values map[string]any, key string, required bool) (string, error) {
	val, ok := values[key]
	if !ok {
		if required {
			return "", fmt.Errorf("missing stream field: %s", key)
		}
		return "", nil
	}
	if s, ok := val.(string); ok {
		return s, nil
	}
	return fmt.Sprint(val), nil
}

func streamInt64(values map[string]any, key string, required bool) (int64, error) {
	s, err := streamString(values, key, required)
	if err != nil || s == "" {
		return 0, err
	}
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid stream field %s: %w", key, err)
	}
	return val, nil
}

func snowflakeIDPtr(id int64) *models.SnowflakeID {
	if id == 0 {
		return nil
	}
	v := models.SnowflakeID(id)
	return &v
}

func directionPtr(direction int8) *int8 {
	if direction == 0 {
		return nil
	}
	return &direction
}
