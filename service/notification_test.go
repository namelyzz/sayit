package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/namelyzz/sayit/models"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationEventFromStreamValues(t *testing.T) {
	values := map[string]any{
		"event_id":     "1001",
		"type":         models.NotificationTypePostVoted,
		"recipient_id": "2001",
		"actor_id":     "3001",
		"post_id":      "4001",
		"direction":    "-1",
		"created_at":   "1700000000",
		"dedupe_key":   "post_voted:4001:3001",
	}

	event, err := notificationEventFromStreamValues(values)
	require.NoError(t, err)
	assert.Equal(t, int64(1001), event.EventID)
	assert.Equal(t, models.NotificationTypePostVoted, event.Type)
	assert.Equal(t, int64(2001), event.RecipientID)
	assert.Equal(t, int64(3001), event.ActorID)
	assert.Equal(t, int64(4001), event.PostID)
	assert.Equal(t, int8(-1), event.Direction)
	assert.Equal(t, int64(1700000000), event.CreatedAt)
	assert.Equal(t, "post_voted:4001:3001", event.DedupeKey)
}

func TestNotificationEventFromStreamValues_InvalidMissingRequired(t *testing.T) {
	_, err := notificationEventFromStreamValues(map[string]any{
		"event_id":     "1001",
		"type":         models.NotificationTypeUserFollowed,
		"recipient_id": "2001",
		"created_at":   "1700000000",
		"dedupe_key":   "user_followed:2001:3001",
	})
	assert.Error(t, err)
}

func TestNotificationPresentation(t *testing.T) {
	tests := []struct {
		name        string
		event       *models.NotificationEvent
		wantTitle   string
		wantContent string
		wantLink    string
	}{
		{
			name: "comment liked",
			event: &models.NotificationEvent{
				Type:      models.NotificationTypeCommentLiked,
				PostID:    10,
				CommentID: 20,
			},
			wantTitle:   "有人点赞了你的评论",
			wantContent: "点击查看评论",
			wantLink:    "/post/10?comment=20",
		},
		{
			name: "post upvoted",
			event: &models.NotificationEvent{
				Type:      models.NotificationTypePostVoted,
				PostID:    10,
				Direction: 1,
			},
			wantTitle:   "有人赞成了你的帖子",
			wantContent: "点击查看帖子",
			wantLink:    "/post/10",
		},
		{
			name: "post downvoted",
			event: &models.NotificationEvent{
				Type:      models.NotificationTypePostVoted,
				PostID:    10,
				Direction: -1,
			},
			wantTitle:   "有人反对了你的帖子",
			wantContent: "点击查看帖子",
			wantLink:    "/post/10",
		},
		{
			name: "comment replied",
			event: &models.NotificationEvent{
				Type:      models.NotificationTypeCommentReplied,
				PostID:    10,
				CommentID: 20,
			},
			wantTitle:   "有人回复了你的评论",
			wantContent: "点击查看回复",
			wantLink:    "/post/10?comment=20",
		},
		{
			name: "user followed",
			event: &models.NotificationEvent{
				Type:    models.NotificationTypeUserFollowed,
				ActorID: 30,
			},
			wantTitle:   "有人关注了你",
			wantContent: "点击查看主页",
			wantLink:    "/user/30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, content, link, err := notificationPresentation(tt.event)
			require.NoError(t, err)
			assert.Equal(t, tt.wantTitle, title)
			assert.Equal(t, tt.wantContent, content)
			assert.Equal(t, tt.wantLink, link)
		})
	}
}

func TestPublishUserFollowedNotification(t *testing.T) {
	restore := mockNotificationDeps()
	defer restore()

	var published *models.NotificationEvent
	publishNotificationEventFunc = func(ctx context.Context, event *models.NotificationEvent) (string, error) {
		published = event
		return "1-0", nil
	}

	PublishUserFollowedNotification(context.Background(), 42, 77)

	require.NotNil(t, published)
	assert.Equal(t, models.NotificationTypeUserFollowed, published.Type)
	assert.Equal(t, int64(77), published.RecipientID)
	assert.Equal(t, int64(42), published.ActorID)
	assert.Equal(t, "user_followed:77:42", published.DedupeKey)
}

func TestPublishUserFollowedNotification_SelfSkipped(t *testing.T) {
	restore := mockNotificationDeps()
	defer restore()

	publishNotificationEventFunc = func(ctx context.Context, event *models.NotificationEvent) (string, error) {
		t.Fatal("self follow should not publish notification")
		return "", nil
	}

	PublishUserFollowedNotification(context.Background(), 42, 42)
}

func TestPublishPostVotedNotification(t *testing.T) {
	restore := mockNotificationDeps()
	defer restore()

	var published *models.NotificationEvent
	publishNotificationEventFunc = func(ctx context.Context, event *models.NotificationEvent) (string, error) {
		published = event
		return "1-0", nil
	}

	PublishPostVotedNotification(context.Background(), 42, &models.Post{PostID: 100, AuthorID: 77}, -1)

	require.NotNil(t, published)
	assert.Equal(t, models.NotificationTypePostVoted, published.Type)
	assert.Equal(t, int64(77), published.RecipientID)
	assert.Equal(t, int64(42), published.ActorID)
	assert.Equal(t, int64(100), published.PostID)
	assert.Equal(t, int8(-1), published.Direction)
	assert.Equal(t, "post_voted:100:42", published.DedupeKey)
}

func TestPublishPostVotedNotification_SelfSkipped(t *testing.T) {
	restore := mockNotificationDeps()
	defer restore()

	publishNotificationEventFunc = func(ctx context.Context, event *models.NotificationEvent) (string, error) {
		t.Fatal("self vote should not publish notification")
		return "", nil
	}

	PublishPostVotedNotification(context.Background(), 42, &models.Post{PostID: 100, AuthorID: 42}, 1)
}

func TestNotificationCooldownScope(t *testing.T) {
	tests := []struct {
		name  string
		event *models.NotificationEvent
		want  string
	}{
		{
			name:  "follow",
			event: &models.NotificationEvent{Type: models.NotificationTypeUserFollowed, ActorID: 42, RecipientID: 77},
			want:  "user_followed:42:77:user:77",
		},
		{
			name:  "comment liked",
			event: &models.NotificationEvent{Type: models.NotificationTypeCommentLiked, ActorID: 42, RecipientID: 77, CommentID: 1001},
			want:  "comment_liked:42:77:comment:1001",
		},
		{
			name:  "comment replied",
			event: &models.NotificationEvent{Type: models.NotificationTypeCommentReplied, ActorID: 42, RecipientID: 77, CommentID: 1002},
			want:  "comment_replied:42:77:comment:1002",
		},
		{
			name:  "post voted",
			event: &models.NotificationEvent{Type: models.NotificationTypePostVoted, ActorID: 42, RecipientID: 77, PostID: 2001},
			want:  "post_voted:42:77:post:2001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, notificationCooldownScope(tt.event))
		})
	}
}

func TestPublishNotification_CooldownHitSkipsPublish(t *testing.T) {
	restore := mockNotificationDeps()
	defer restore()

	published := false
	acquireNotificationCooldownFunc = func(ctx context.Context, scope string, ttl time.Duration) (bool, error) {
		assert.Equal(t, notificationCooldownTTL, ttl)
		assert.Equal(t, "user_followed:42:77:user:77", scope)
		return false, nil
	}
	publishNotificationEventFunc = func(ctx context.Context, event *models.NotificationEvent) (string, error) {
		published = true
		return "1-0", nil
	}

	PublishUserFollowedNotification(context.Background(), 42, 77)

	assert.False(t, published)
}

func TestPublishNotification_CooldownErrorFallsBackToPublish(t *testing.T) {
	restore := mockNotificationDeps()
	defer restore()

	published := false
	acquireNotificationCooldownFunc = func(ctx context.Context, scope string, ttl time.Duration) (bool, error) {
		return false, errors.New("redis down")
	}
	publishNotificationEventFunc = func(ctx context.Context, event *models.NotificationEvent) (string, error) {
		published = true
		return "1-0", nil
	}

	PublishUserFollowedNotification(context.Background(), 42, 77)

	assert.True(t, published)
}

func TestProcessNotificationMessage_CreatedIncrementsUnreadAndAck(t *testing.T) {
	restore := mockNotificationDeps()
	defer restore()

	var incrUserID int64
	var ackID string
	createNotificationFunc = func(ctx context.Context, n *models.Notification) (bool, error) {
		assert.Equal(t, models.NotificationTypeUserFollowed, n.Type)
		assert.Equal(t, int64(2001), int64(n.RecipientID))
		return true, nil
	}
	incrNotificationUnreadFunc = func(ctx context.Context, userID int64) error {
		incrUserID = userID
		return nil
	}
	ackNotificationEventFunc = func(ctx context.Context, messageID string) error {
		ackID = messageID
		return nil
	}

	err := processNotificationMessage(context.Background(), goredis.XMessage{
		ID: "1-0",
		Values: map[string]any{
			"event_id":     "1001",
			"type":         models.NotificationTypeUserFollowed,
			"recipient_id": "2001",
			"actor_id":     "3001",
			"created_at":   "1700000000",
			"dedupe_key":   "user_followed:2001:3001",
		},
	})

	require.NoError(t, err)
	assert.Equal(t, int64(2001), incrUserID)
	assert.Equal(t, "1-0", ackID)
}

func TestProcessNotificationMessage_DuplicateDoesNotIncrementUnread(t *testing.T) {
	restore := mockNotificationDeps()
	defer restore()

	incrCalled := false
	ackCalled := false
	createNotificationFunc = func(ctx context.Context, n *models.Notification) (bool, error) {
		return false, nil
	}
	incrNotificationUnreadFunc = func(ctx context.Context, userID int64) error {
		incrCalled = true
		return nil
	}
	ackNotificationEventFunc = func(ctx context.Context, messageID string) error {
		ackCalled = true
		return nil
	}

	err := processNotificationMessage(context.Background(), goredis.XMessage{
		ID: "1-0",
		Values: map[string]any{
			"event_id":     "1001",
			"type":         models.NotificationTypeUserFollowed,
			"recipient_id": "2001",
			"actor_id":     "3001",
			"created_at":   "1700000000",
			"dedupe_key":   "user_followed:2001:3001",
		},
	})

	require.NoError(t, err)
	assert.False(t, incrCalled)
	assert.True(t, ackCalled)
}

func TestProcessNotificationMessage_InvalidMessageAcked(t *testing.T) {
	restore := mockNotificationDeps()
	defer restore()

	ackCalled := false
	createCalled := false
	createNotificationFunc = func(ctx context.Context, n *models.Notification) (bool, error) {
		createCalled = true
		return true, nil
	}
	ackNotificationEventFunc = func(ctx context.Context, messageID string) error {
		ackCalled = true
		return nil
	}

	err := processNotificationMessage(context.Background(), goredis.XMessage{
		ID: "1-0",
		Values: map[string]any{
			"event_id": "1001",
		},
	})

	assert.Error(t, err)
	assert.True(t, ackCalled)
	assert.False(t, createCalled)
}

func TestGetNotificationUnreadCount_CacheHit(t *testing.T) {
	restore := mockNotificationDeps()
	defer restore()

	countDBCalled := false
	getNotificationUnreadFunc = func(ctx context.Context, userID int64) (int64, error) {
		return 7, nil
	}
	countUnreadNotificationsFunc = func(ctx context.Context, userID int64) (int64, error) {
		countDBCalled = true
		return 0, nil
	}

	count, err := GetNotificationUnreadCount(context.Background(), 1001)
	require.NoError(t, err)
	assert.Equal(t, int64(7), count)
	assert.False(t, countDBCalled)
}

func TestGetNotificationUnreadCount_CacheMissBackfills(t *testing.T) {
	restore := mockNotificationDeps()
	defer restore()

	var backfillCount int64
	getNotificationUnreadFunc = func(ctx context.Context, userID int64) (int64, error) {
		return 0, goredis.Nil
	}
	countUnreadNotificationsFunc = func(ctx context.Context, userID int64) (int64, error) {
		return 9, nil
	}
	setNotificationUnreadFunc = func(ctx context.Context, userID, count int64) error {
		backfillCount = count
		return nil
	}

	count, err := GetNotificationUnreadCount(context.Background(), 1001)
	require.NoError(t, err)
	assert.Equal(t, int64(9), count)
	assert.Equal(t, int64(9), backfillCount)
}

func TestMarkNotificationRead_DecrementsAffectedUnread(t *testing.T) {
	restore := mockNotificationDeps()
	defer restore()

	var decCount int64
	markNotificationReadFunc = func(ctx context.Context, userID, notificationID int64) (int64, error) {
		return 1, nil
	}
	decrNotificationUnreadFunc = func(ctx context.Context, userID, count int64) error {
		decCount = count
		return nil
	}

	err := MarkNotificationRead(context.Background(), 1001, 2001)
	require.NoError(t, err)
	assert.Equal(t, int64(1), decCount)
}

func TestMarkAllNotificationsRead_DecrementsAffectedUnread(t *testing.T) {
	restore := mockNotificationDeps()
	defer restore()

	var decCount int64
	markAllNotificationsReadFunc = func(ctx context.Context, userID int64) (int64, error) {
		return 5, nil
	}
	decrNotificationUnreadFunc = func(ctx context.Context, userID, count int64) error {
		decCount = count
		return nil
	}

	err := MarkAllNotificationsRead(context.Background(), 1001)
	require.NoError(t, err)
	assert.Equal(t, int64(5), decCount)
}

func TestGetNotificationList(t *testing.T) {
	restore := mockNotificationDeps()
	defer restore()

	listNotificationsFunc = func(ctx context.Context, userID int64, page, size int, status string) ([]*models.NotificationItem, int64, error) {
		assert.Equal(t, int64(1001), userID)
		assert.Equal(t, "unread", status)
		return []*models.NotificationItem{{ActorName: "alice"}}, 1, nil
	}
	getNotificationUnreadFunc = func(ctx context.Context, userID int64) (int64, error) {
		return 3, nil
	}

	data, err := GetNotificationList(context.Background(), 1001, &models.ParamNotificationList{Status: "unread"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), data.Total)
	assert.Equal(t, int64(3), data.UnreadCount)
	assert.Len(t, data.List, 1)
}

func TestProcessNotificationMessage_CreateErrorKeepsPending(t *testing.T) {
	restore := mockNotificationDeps()
	defer restore()

	ackCalled := false
	createNotificationFunc = func(ctx context.Context, n *models.Notification) (bool, error) {
		return false, errors.New("mysql down")
	}
	ackNotificationEventFunc = func(ctx context.Context, messageID string) error {
		ackCalled = true
		return nil
	}

	err := processNotificationMessage(context.Background(), goredis.XMessage{
		ID: "1-0",
		Values: map[string]any{
			"event_id":     "1001",
			"type":         models.NotificationTypeUserFollowed,
			"recipient_id": "2001",
			"actor_id":     "3001",
			"created_at":   "1700000000",
			"dedupe_key":   "user_followed:2001:3001",
		},
	})

	assert.Error(t, err)
	assert.False(t, ackCalled)
}

func mockNotificationDeps() func() {
	origEnsure := ensureNotificationConsumerGroupFunc
	origRead := readNotificationEventsFunc
	origAck := ackNotificationEventFunc
	origIncr := incrNotificationUnreadFunc
	origGet := getNotificationUnreadFunc
	origSet := setNotificationUnreadFunc
	origDecr := decrNotificationUnreadFunc
	origAcquire := acquireNotificationCooldownFunc
	origCreate := createNotificationFunc
	origPublish := publishNotificationEventFunc
	origList := listNotificationsFunc
	origCount := countUnreadNotificationsFunc
	origMark := markNotificationReadFunc
	origMarkAll := markAllNotificationsReadFunc
	origGenID := genNotificationIDFunc

	genNotificationIDFunc = func() int64 { return 9999 }
	acquireNotificationCooldownFunc = func(ctx context.Context, scope string, ttl time.Duration) (bool, error) {
		return true, nil
	}

	return func() {
		ensureNotificationConsumerGroupFunc = origEnsure
		readNotificationEventsFunc = origRead
		ackNotificationEventFunc = origAck
		incrNotificationUnreadFunc = origIncr
		getNotificationUnreadFunc = origGet
		setNotificationUnreadFunc = origSet
		decrNotificationUnreadFunc = origDecr
		acquireNotificationCooldownFunc = origAcquire
		createNotificationFunc = origCreate
		publishNotificationEventFunc = origPublish
		listNotificationsFunc = origList
		countUnreadNotificationsFunc = origCount
		markNotificationReadFunc = origMark
		markAllNotificationsReadFunc = origMarkAll
		genNotificationIDFunc = origGenID
	}
}
