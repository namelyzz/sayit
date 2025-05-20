package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/namelyzz/sayit/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func restoreOutboxTestHooks() func() {
	origCreatePostWithOutbox := createPostWithOutboxFunc
	origGetPendingOutbox := getPendingOutboxFunc
	origMarkOutboxRetry := markOutboxRetryFunc
	origDeleteOutboxEvent := deleteOutboxEventFunc
	origCreatePostRedis := createPostRedisFunc
	origNow := nowFunc
	origGenID := genIDFunc

	return func() {
		createPostWithOutboxFunc = origCreatePostWithOutbox
		getPendingOutboxFunc = origGetPendingOutbox
		markOutboxRetryFunc = origMarkOutboxRetry
		deleteOutboxEventFunc = origDeleteOutboxEvent
		createPostRedisFunc = origCreatePostRedis
		nowFunc = origNow
		genIDFunc = origGenID
	}
}

func TestNewPostCreatedOutboxEventPayload(t *testing.T) {
	defer restoreOutboxTestHooks()()

	now := time.Unix(1700000000, 0)
	nowFunc = func() time.Time { return now }

	event, err := newPostCreatedOutboxEvent(&models.Post{
		PostID:      101,
		CommunityID: 2,
		CreateTime:  now,
	})

	require.NoError(t, err)
	assert.Equal(t, models.EventTypePostCreated, event.EventType)
	assert.Equal(t, int64(101), event.AggregateID)
	assert.Equal(t, now, event.NextRetryAt)

	var payload models.PostCreatedPayload
	require.NoError(t, json.Unmarshal(event.Payload, &payload))
	assert.Equal(t, int64(101), payload.PostID)
	assert.Equal(t, int64(2), payload.CommunityID)
	assert.Equal(t, now.Unix(), payload.CreateTimeUnix)
}

func TestProcessPostCreatedEventSuccessDeletesOutboxEvent(t *testing.T) {
	defer restoreOutboxTestHooks()()

	payloadBytes, err := json.Marshal(models.PostCreatedPayload{
		PostID:         101,
		CommunityID:    2,
		CreateTimeUnix: 1700000000,
	})
	require.NoError(t, err)

	var redisCalled bool
	var deleteCalled bool
	createPostRedisFunc = func(ctx context.Context, postID, communityID int64, score float64) error {
		redisCalled = true
		assert.Equal(t, int64(101), postID)
		assert.Equal(t, int64(2), communityID)
		assert.Equal(t, float64(1700000000), score)
		return nil
	}
	deleteOutboxEventFunc = func(ctx context.Context, id int64) error {
		deleteCalled = true
		assert.Equal(t, int64(9), id)
		return nil
	}
	markOutboxRetryFunc = func(ctx context.Context, id int64, retryCount int, nextRetryAt time.Time, errMsg string) error {
		t.Fatal("mark retry should not be called on success")
		return nil
	}

	err = processOutboxEvent(context.Background(), &models.OutboxEvent{
		ID:        9,
		EventType: models.EventTypePostCreated,
		Payload:   payloadBytes,
	})

	require.NoError(t, err)
	assert.True(t, redisCalled)
	assert.True(t, deleteCalled)
}

func TestProcessPostCreatedEventFailureMarksRetry(t *testing.T) {
	defer restoreOutboxTestHooks()()

	now := time.Unix(1700000000, 0)
	nowFunc = func() time.Time { return now }

	payloadBytes, err := json.Marshal(models.PostCreatedPayload{
		PostID:         101,
		CommunityID:    2,
		CreateTimeUnix: 1700000000,
	})
	require.NoError(t, err)

	redisErr := errors.New("redis down")
	var markRetryCalled bool
	createPostRedisFunc = func(ctx context.Context, postID, communityID int64, score float64) error {
		return redisErr
	}
	markOutboxRetryFunc = func(ctx context.Context, id int64, retryCount int, nextRetryAt time.Time, errMsg string) error {
		markRetryCalled = true
		assert.Equal(t, int64(9), id)
		assert.Equal(t, 3, retryCount)
		assert.Equal(t, now.Add(4*time.Second), nextRetryAt)
		assert.Equal(t, redisErr.Error(), errMsg)
		return nil
	}
	deleteOutboxEventFunc = func(ctx context.Context, id int64) error {
		t.Fatal("delete should not be called on redis failure")
		return nil
	}

	err = processOutboxEvent(context.Background(), &models.OutboxEvent{
		ID:         9,
		EventType:  models.EventTypePostCreated,
		Payload:    payloadBytes,
		RetryCount: 2,
	})

	require.Error(t, err)
	assert.True(t, markRetryCalled)
}

func TestCreatePostReturnsSuccessWhenImmediateRedisSyncFails(t *testing.T) {
	defer restoreOutboxTestHooks()()

	now := time.Unix(1700000000, 0)
	nowFunc = func() time.Time { return now }
	genIDFunc = func() int64 { return 101 }

	var outboxCreated bool
	var markRetryCalled bool
	createPostWithOutboxFunc = func(ctx context.Context, p *models.Post, event *models.OutboxEvent) (int64, error) {
		outboxCreated = true
		assert.Equal(t, models.SnowflakeID(101), p.PostID)
		assert.Equal(t, now, p.CreateTime)
		assert.Equal(t, models.EventTypePostCreated, event.EventType)
		return 9, nil
	}
	createPostRedisFunc = func(ctx context.Context, postID, communityID int64, score float64) error {
		return errors.New("redis down")
	}
	markOutboxRetryFunc = func(ctx context.Context, id int64, retryCount int, nextRetryAt time.Time, errMsg string) error {
		markRetryCalled = true
		assert.Equal(t, int64(9), id)
		assert.Equal(t, 1, retryCount)
		assert.Equal(t, now.Add(time.Second), nextRetryAt)
		assert.Equal(t, "redis down", errMsg)
		return nil
	}

	err := CreatePost(context.Background(), &models.Post{
		Title:       "hello",
		Content:     "world",
		AuthorID:    1,
		CommunityID: 2,
	})

	require.NoError(t, err)
	assert.True(t, outboxCreated)
	assert.True(t, markRetryCalled)
}
