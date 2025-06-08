package redis

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/namelyzz/sayit/models"
	goredis "github.com/redis/go-redis/v9"
)

const NotificationConsumerGroup = "sayit:notification:group"

const notificationStreamMaxLen int64 = 100000

func notificationStreamKey() string {
	return getRedisKey(KeyNotificationStream)
}

func notificationUnreadKey(userID int64) string {
	return getRedisKey(KeyNotificationUnreadPF + strconv.FormatInt(userID, 10))
}

// PublishNotificationEvent 将通知事件写入 Redis Stream。
func PublishNotificationEvent(ctx context.Context, event *models.NotificationEvent) (string, error) {
	return client.XAdd(ctx, &goredis.XAddArgs{
		Stream: notificationStreamKey(),
		MaxLen: notificationStreamMaxLen,
		Approx: true,
		Values: notificationEventToStreamValues(event),
	}).Result()
}

// EnsureNotificationConsumerGroup 确保通知 Consumer Group 存在。
func EnsureNotificationConsumerGroup(ctx context.Context) error {
	err := client.XGroupCreateMkStream(ctx, notificationStreamKey(), NotificationConsumerGroup, "0").Err()
	if err != nil && !isBusyGroupError(err) {
		return err
	}
	return nil
}

// ReadNotificationEvents 读取新的通知事件。
func ReadNotificationEvents(ctx context.Context, consumer string, count int64, block time.Duration) ([]goredis.XMessage, error) {
	streams, err := client.XReadGroup(ctx, &goredis.XReadGroupArgs{
		Group:    NotificationConsumerGroup,
		Consumer: consumer,
		Streams:  []string{notificationStreamKey(), ">"},
		Count:    count,
		Block:    block,
	}).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	if len(streams) == 0 {
		return nil, nil
	}
	return streams[0].Messages, nil
}

// AckNotificationEvent 确认通知事件已处理。
func AckNotificationEvent(ctx context.Context, messageID string) error {
	return client.XAck(ctx, notificationStreamKey(), NotificationConsumerGroup, messageID).Err()
}

// IncrNotificationUnread 增加用户未读通知数。
func IncrNotificationUnread(ctx context.Context, userID int64) error {
	return client.Incr(ctx, notificationUnreadKey(userID)).Err()
}

// GetNotificationUnread 获取用户未读通知数。未命中时返回 redis.Nil。
func GetNotificationUnread(ctx context.Context, userID int64) (int64, error) {
	return client.Get(ctx, notificationUnreadKey(userID)).Int64()
}

// SetNotificationUnread 设置用户未读通知数。
func SetNotificationUnread(ctx context.Context, userID, count int64) error {
	if count < 0 {
		count = 0
	}
	return client.Set(ctx, notificationUnreadKey(userID), count, 0).Err()
}

// DecrNotificationUnread 安全减少用户未读通知数，不会减成负数。
func DecrNotificationUnread(ctx context.Context, userID, count int64) error {
	if count <= 0 {
		return nil
	}
	_, err := decrNotificationUnreadScript.Run(ctx, client, []string{notificationUnreadKey(userID)}, count).Result()
	if errors.Is(err, goredis.Nil) {
		return nil
	}
	return err
}

var decrNotificationUnreadScript = goredis.NewScript(`
local val = tonumber(redis.call('GET', KEYS[1]))
local dec = tonumber(ARGV[1])
if val == nil then return nil end
if val <= dec then
  redis.call('SET', KEYS[1], 0)
  return 0
end
return redis.call('DECRBY', KEYS[1], dec)
`)

func notificationEventToStreamValues(event *models.NotificationEvent) map[string]any {
	values := map[string]any{
		"event_id":     strconv.FormatInt(event.EventID, 10),
		"type":         event.Type,
		"recipient_id": strconv.FormatInt(event.RecipientID, 10),
		"actor_id":     strconv.FormatInt(event.ActorID, 10),
		"created_at":   strconv.FormatInt(event.CreatedAt, 10),
		"dedupe_key":   event.DedupeKey,
	}
	if event.PostID != 0 {
		values["post_id"] = strconv.FormatInt(event.PostID, 10)
	}
	if event.CommentID != 0 {
		values["comment_id"] = strconv.FormatInt(event.CommentID, 10)
	}
	if event.ParentID != 0 {
		values["parent_id"] = strconv.FormatInt(event.ParentID, 10)
	}
	if event.Direction != 0 {
		values["direction"] = strconv.FormatInt(int64(event.Direction), 10)
	}
	return values
}

func isBusyGroupError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "BUSYGROUP")
}
