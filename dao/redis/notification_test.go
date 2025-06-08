package redis

import (
	"errors"
	"testing"

	"github.com/namelyzz/sayit/models"
	"github.com/stretchr/testify/assert"
)

func TestNotificationEventToStreamValues(t *testing.T) {
	event := &models.NotificationEvent{
		EventID:     1001,
		Type:        models.NotificationTypeCommentReplied,
		RecipientID: 2001,
		ActorID:     3001,
		PostID:      4001,
		CommentID:   5001,
		ParentID:    6001,
		Direction:   -1,
		CreatedAt:   1700000000,
		DedupeKey:   "comment_replied:5001",
	}

	values := notificationEventToStreamValues(event)

	assert.Equal(t, "1001", values["event_id"])
	assert.Equal(t, models.NotificationTypeCommentReplied, values["type"])
	assert.Equal(t, "2001", values["recipient_id"])
	assert.Equal(t, "3001", values["actor_id"])
	assert.Equal(t, "4001", values["post_id"])
	assert.Equal(t, "5001", values["comment_id"])
	assert.Equal(t, "6001", values["parent_id"])
	assert.Equal(t, "-1", values["direction"])
	assert.Equal(t, "1700000000", values["created_at"])
	assert.Equal(t, "comment_replied:5001", values["dedupe_key"])
}

func TestNotificationEventToStreamValues_OmitZeroOptionalFields(t *testing.T) {
	event := &models.NotificationEvent{
		EventID:     1001,
		Type:        models.NotificationTypeUserFollowed,
		RecipientID: 2001,
		ActorID:     3001,
		CreatedAt:   1700000000,
		DedupeKey:   "user_followed:2001:3001",
	}

	values := notificationEventToStreamValues(event)

	assert.NotContains(t, values, "post_id")
	assert.NotContains(t, values, "comment_id")
	assert.NotContains(t, values, "parent_id")
	assert.NotContains(t, values, "direction")
}

func TestNotificationKeys(t *testing.T) {
	assert.Equal(t, "sayit:notification:stream", notificationStreamKey())
	assert.Equal(t, "sayit:notification:unread:123", notificationUnreadKey(123))
}

func TestIsBusyGroupError(t *testing.T) {
	assert.True(t, isBusyGroupError(errors.New("BUSYGROUP Consumer Group name already exists")))
	assert.False(t, isBusyGroupError(errors.New("ERR invalid stream")))
	assert.False(t, isBusyGroupError(nil))
}
