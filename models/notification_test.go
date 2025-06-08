package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationTableName(t *testing.T) {
	assert.Equal(t, "notifications", (Notification{}).TableName())
}

func TestNotificationJSONUsesStringIDs(t *testing.T) {
	postID := SnowflakeID(3001)
	direction := int8(1)
	n := Notification{
		NotificationID: SnowflakeID(1001),
		RecipientID:    SnowflakeID(2001),
		ActorID:        SnowflakeID(2002),
		Type:           NotificationTypePostVoted,
		PostID:         &postID,
		Direction:      &direction,
		Title:          "Alice 赞成了你的帖子",
		Content:        "点击查看帖子",
		Link:           "/post/3001",
	}

	data, err := json.Marshal(n)
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, "1001", decoded["notification_id"])
	assert.Equal(t, "2001", decoded["recipient_id"])
	assert.Equal(t, "2002", decoded["actor_id"])
	assert.Equal(t, "3001", decoded["post_id"])
	assert.Equal(t, float64(1), decoded["direction"])
	assert.NotContains(t, decoded, "dedupe_key")
}
