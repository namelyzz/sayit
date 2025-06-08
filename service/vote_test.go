package service

import (
	"context"
	"testing"

	"github.com/namelyzz/sayit/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func restoreVoteTestHooks() func() {
	origIsWithinOneWeek := isPostCreatedWithinOneWeekFunc
	origGetVoteScore := getPostVoteScoreFunc
	origUpdateVote := updatePostVoteFunc
	origGetPost := getPostForVoteFunc
	origPublish := publishNotificationEventFunc
	origGenNotificationID := genNotificationIDFunc
	genNotificationIDFunc = func() int64 { return 9999 }

	return func() {
		isPostCreatedWithinOneWeekFunc = origIsWithinOneWeek
		getPostVoteScoreFunc = origGetVoteScore
		updatePostVoteFunc = origUpdateVote
		getPostForVoteFunc = origGetPost
		publishNotificationEventFunc = origPublish
		genNotificationIDFunc = origGenNotificationID
	}
}

func TestVoteForPost_FirstVotePublishesNotification(t *testing.T) {
	defer restoreVoteTestHooks()()

	isPostCreatedWithinOneWeekFunc = func(ctx context.Context, postID string) bool { return true }
	getPostVoteScoreFunc = func(ctx context.Context, postID, userID string) float64 { return 0 }
	updatePostVoteFunc = func(ctx context.Context, userID, postID string, voteVal, operate, diff float64) error {
		assert.Equal(t, float64(1), voteVal)
		assert.Equal(t, float64(1), operate)
		assert.Equal(t, float64(1), diff)
		return nil
	}
	getPostForVoteFunc = func(postID int64) (*models.Post, error) {
		return &models.Post{PostID: models.SnowflakeID(postID), AuthorID: 77}, nil
	}

	var published *models.NotificationEvent
	publishNotificationEventFunc = func(ctx context.Context, event *models.NotificationEvent) (string, error) {
		published = event
		return "1-0", nil
	}

	err := VoteForPost(context.Background(), 42, &models.ParamVote{PostID: "100", Direction: 1})

	require.NoError(t, err)
	require.NotNil(t, published)
	assert.Equal(t, models.NotificationTypePostVoted, published.Type)
	assert.Equal(t, int64(77), published.RecipientID)
	assert.Equal(t, int64(42), published.ActorID)
	assert.Equal(t, int64(100), published.PostID)
	assert.Equal(t, int8(1), published.Direction)
}

func TestVoteForPost_CancelDoesNotPublishNotification(t *testing.T) {
	defer restoreVoteTestHooks()()

	isPostCreatedWithinOneWeekFunc = func(ctx context.Context, postID string) bool { return true }
	getPostVoteScoreFunc = func(ctx context.Context, postID, userID string) float64 { return 1 }
	updatePostVoteFunc = func(ctx context.Context, userID, postID string, voteVal, operate, diff float64) error { return nil }
	getPostForVoteFunc = func(postID int64) (*models.Post, error) {
		t.Fatal("cancel vote should not fetch post for notification")
		return nil, nil
	}
	publishNotificationEventFunc = func(ctx context.Context, event *models.NotificationEvent) (string, error) {
		t.Fatal("cancel vote should not publish notification")
		return "", nil
	}

	err := VoteForPost(context.Background(), 42, &models.ParamVote{PostID: "100", Direction: 0})

	require.NoError(t, err)
}

func TestVoteForPost_ChangeVoteDoesNotPublishNotification(t *testing.T) {
	defer restoreVoteTestHooks()()

	isPostCreatedWithinOneWeekFunc = func(ctx context.Context, postID string) bool { return true }
	getPostVoteScoreFunc = func(ctx context.Context, postID, userID string) float64 { return -1 }
	updatePostVoteFunc = func(ctx context.Context, userID, postID string, voteVal, operate, diff float64) error { return nil }
	getPostForVoteFunc = func(postID int64) (*models.Post, error) {
		t.Fatal("changed vote should not fetch post for notification")
		return nil, nil
	}
	publishNotificationEventFunc = func(ctx context.Context, event *models.NotificationEvent) (string, error) {
		t.Fatal("changed vote should not publish notification")
		return "", nil
	}

	err := VoteForPost(context.Background(), 42, &models.ParamVote{PostID: "100", Direction: 1})

	require.NoError(t, err)
}

func TestVoteForPost_SelfVoteDoesNotPublishNotification(t *testing.T) {
	defer restoreVoteTestHooks()()

	isPostCreatedWithinOneWeekFunc = func(ctx context.Context, postID string) bool { return true }
	getPostVoteScoreFunc = func(ctx context.Context, postID, userID string) float64 { return 0 }
	updatePostVoteFunc = func(ctx context.Context, userID, postID string, voteVal, operate, diff float64) error { return nil }
	getPostForVoteFunc = func(postID int64) (*models.Post, error) {
		return &models.Post{PostID: models.SnowflakeID(postID), AuthorID: 42}, nil
	}
	publishNotificationEventFunc = func(ctx context.Context, event *models.NotificationEvent) (string, error) {
		t.Fatal("self vote should not publish notification")
		return "", nil
	}

	err := VoteForPost(context.Background(), 42, &models.ParamVote{PostID: "100", Direction: 1})

	require.NoError(t, err)
}
