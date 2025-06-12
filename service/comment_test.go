package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/namelyzz/sayit/dao/mysql"
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/utils/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func restoreCommentTestHooks() func() {
	origGetPostByID := getPostByIDFunc
	origGetCommentByID := getCommentByIDFunc
	origCreateComment := createCommentFunc
	origSoftDeleteComment := softDeleteCommentFunc
	origIncrCommentLikeCount := incrCommentLikeCountFunc
	origDecrCommentLikeCount := decrCommentLikeCountFunc
	origGetTopLevelComments := getTopLevelCommentsFunc
	origCountTopLevelComments := countTopLevelCommentsFunc
	origCountChildCommentsByParent := countChildCommentsByParentFunc
	origGetChildCommentsByParentID := getChildCommentsByParentIDFunc
	origCountChildCommentsByParentID := countChildCommentsByParentIDFunc
	origGetCommentLikeScore := getCommentLikeScoreFunc
	origCommentLike := commentLikeFunc
	origCommentUnlike := commentUnlikeFunc
	origIsCommentLiked := isCommentLikedFunc
	origIncrCommentCount := incrCommentCountFunc
	origDecrCommentCount := decrCommentCountFunc
	origCheckCommentRateLimit := checkCommentRateLimitFunc
	origGenID := genIDFunc
	origPublishNotificationEvent := publishNotificationEventFunc
	origGenNotificationID := genNotificationIDFunc
	origAcquireNotificationCooldown := acquireNotificationCooldownFunc
	genNotificationIDFunc = func() int64 { return 9999 }
	acquireNotificationCooldownFunc = func(ctx context.Context, scope string, ttl time.Duration) (bool, error) {
		return true, nil
	}

	return func() {
		getPostByIDFunc = origGetPostByID
		getCommentByIDFunc = origGetCommentByID
		createCommentFunc = origCreateComment
		softDeleteCommentFunc = origSoftDeleteComment
		incrCommentLikeCountFunc = origIncrCommentLikeCount
		decrCommentLikeCountFunc = origDecrCommentLikeCount
		getTopLevelCommentsFunc = origGetTopLevelComments
		countTopLevelCommentsFunc = origCountTopLevelComments
		countChildCommentsByParentFunc = origCountChildCommentsByParent
		getChildCommentsByParentIDFunc = origGetChildCommentsByParentID
		countChildCommentsByParentIDFunc = origCountChildCommentsByParentID
		getCommentLikeScoreFunc = origGetCommentLikeScore
		commentLikeFunc = origCommentLike
		commentUnlikeFunc = origCommentUnlike
		isCommentLikedFunc = origIsCommentLiked
		incrCommentCountFunc = origIncrCommentCount
		decrCommentCountFunc = origDecrCommentCount
		checkCommentRateLimitFunc = origCheckCommentRateLimit
		genIDFunc = origGenID
		publishNotificationEventFunc = origPublishNotificationEvent
		genNotificationIDFunc = origGenNotificationID
		acquireNotificationCooldownFunc = origAcquireNotificationCooldown
	}
}

func TestCreateComment_Success_TopLevel(t *testing.T) {
	defer restoreCommentTestHooks()()

	genIDFunc = func() int64 { return 1001 }

	checkCommentRateLimitFunc = func(ctx context.Context, userID int64) (bool, time.Duration, error) {
		return true, 0, nil
	}
	getPostByIDFunc = func(postID int64) (*models.Post, error) {
		assert.Equal(t, int64(100), postID)
		return &models.Post{PostID: 100}, nil
	}

	var createdComment *models.Comment
	createCommentFunc = func(c *models.Comment) error {
		createdComment = c
		return nil
	}
	incrCommentCountFunc = func(ctx context.Context, postID int64) error {
		return nil
	}

	comment, err := CreateComment(context.Background(), 42, &models.ParamCreateComment{
		PostID:  "100",
		Content: "hello world",
	})

	require.NoError(t, err)
	require.NotNil(t, comment)
	assert.Equal(t, int64(1001), int64(comment.CommentID))
	assert.Equal(t, int64(100), int64(comment.PostID))
	assert.Equal(t, int64(42), int64(comment.AuthorID))
	assert.Equal(t, int64(0), int64(comment.ParentID))
	assert.Equal(t, int64(0), int64(comment.RootID))
	assert.Equal(t, "hello world", comment.Content)
	assert.Equal(t, int32(1), comment.Status)
	assert.Equal(t, comment, createdComment)
}

func TestCreateComment_TopLevelPublishesNotificationToPostAuthor(t *testing.T) {
	defer restoreCommentTestHooks()()

	genIDFunc = func() int64 { return 1001 }
	checkCommentRateLimitFunc = func(ctx context.Context, userID int64) (bool, time.Duration, error) {
		return true, 0, nil
	}
	getPostByIDFunc = func(postID int64) (*models.Post, error) {
		return &models.Post{PostID: 100, AuthorID: 77}, nil
	}
	createCommentFunc = func(c *models.Comment) error { return nil }
	incrCommentCountFunc = func(ctx context.Context, postID int64) error { return nil }

	var published *models.NotificationEvent
	publishNotificationEventFunc = func(ctx context.Context, event *models.NotificationEvent) (string, error) {
		published = event
		return "1-0", nil
	}

	_, err := CreateComment(context.Background(), 42, &models.ParamCreateComment{
		PostID:  "100",
		Content: "hello world",
	})

	require.NoError(t, err)
	require.NotNil(t, published)
	assert.Equal(t, models.NotificationTypePostCommented, published.Type)
	assert.Equal(t, int64(77), published.RecipientID)
	assert.Equal(t, int64(42), published.ActorID)
	assert.Equal(t, int64(100), published.PostID)
	assert.Equal(t, int64(1001), published.CommentID)
}

func TestCreateComment_TopLevelSelfPostDoesNotPublishNotification(t *testing.T) {
	defer restoreCommentTestHooks()()

	genIDFunc = func() int64 { return 1001 }
	checkCommentRateLimitFunc = func(ctx context.Context, userID int64) (bool, time.Duration, error) {
		return true, 0, nil
	}
	getPostByIDFunc = func(postID int64) (*models.Post, error) {
		return &models.Post{PostID: 100, AuthorID: 42}, nil
	}
	createCommentFunc = func(c *models.Comment) error { return nil }
	incrCommentCountFunc = func(ctx context.Context, postID int64) error { return nil }
	publishNotificationEventFunc = func(ctx context.Context, event *models.NotificationEvent) (string, error) {
		t.Fatal("self post comment should not publish notification")
		return "", nil
	}

	_, err := CreateComment(context.Background(), 42, &models.ParamCreateComment{
		PostID:  "100",
		Content: "hello world",
	})

	require.NoError(t, err)
}

func TestCreateComment_Success_ReplyToTopLevel(t *testing.T) {
	defer restoreCommentTestHooks()()

	genIDFunc = func() int64 { return 1002 }

	checkCommentRateLimitFunc = func(ctx context.Context, userID int64) (bool, time.Duration, error) {
		return true, 0, nil
	}
	getPostByIDFunc = func(postID int64) (*models.Post, error) {
		return &models.Post{PostID: 100}, nil
	}

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		assert.Equal(t, int64(500), commentID)
		return &models.Comment{
			CommentID: 500,
			PostID:    100,
			ParentID:  0, // 顶级评论
			RootID:    0,
		}, nil
	}

	createCommentFunc = func(c *models.Comment) error {
		return nil
	}
	incrCommentCountFunc = func(ctx context.Context, postID int64) error {
		return nil
	}

	comment, err := CreateComment(context.Background(), 42, &models.ParamCreateComment{
		PostID:   "100",
		ParentID: "500",
		Content:  "reply to top level",
	})

	require.NoError(t, err)
	assert.Equal(t, int64(500), int64(comment.ParentID))
	assert.Equal(t, int64(500), int64(comment.RootID)) // 根评论ID = 父评论ID
}

func TestCreateComment_ReplyPublishesNotificationToDirectParent(t *testing.T) {
	defer restoreCommentTestHooks()()

	genIDFunc = func() int64 { return 1002 }
	checkCommentRateLimitFunc = func(ctx context.Context, userID int64) (bool, time.Duration, error) {
		return true, 0, nil
	}
	getPostByIDFunc = func(postID int64) (*models.Post, error) {
		return &models.Post{PostID: 100}, nil
	}
	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{
			CommentID: 500,
			PostID:    100,
			AuthorID:  77,
			ParentID:  0,
		}, nil
	}
	createCommentFunc = func(c *models.Comment) error { return nil }
	incrCommentCountFunc = func(ctx context.Context, postID int64) error { return nil }

	var published *models.NotificationEvent
	publishNotificationEventFunc = func(ctx context.Context, event *models.NotificationEvent) (string, error) {
		published = event
		return "1-0", nil
	}

	_, err := CreateComment(context.Background(), 42, &models.ParamCreateComment{
		PostID:   "100",
		ParentID: "500",
		Content:  "reply",
	})

	require.NoError(t, err)
	require.NotNil(t, published)
	assert.Equal(t, models.NotificationTypeCommentReplied, published.Type)
	assert.Equal(t, int64(77), published.RecipientID)
	assert.Equal(t, int64(42), published.ActorID)
	assert.Equal(t, int64(100), published.PostID)
	assert.Equal(t, int64(1002), published.CommentID)
	assert.Equal(t, int64(500), published.ParentID)
}

func TestCreateComment_ReplySelfDoesNotPublishNotification(t *testing.T) {
	defer restoreCommentTestHooks()()

	genIDFunc = func() int64 { return 1002 }
	checkCommentRateLimitFunc = func(ctx context.Context, userID int64) (bool, time.Duration, error) {
		return true, 0, nil
	}
	getPostByIDFunc = func(postID int64) (*models.Post, error) {
		return &models.Post{PostID: 100}, nil
	}
	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{CommentID: 500, PostID: 100, AuthorID: 42}, nil
	}
	createCommentFunc = func(c *models.Comment) error { return nil }
	incrCommentCountFunc = func(ctx context.Context, postID int64) error { return nil }

	publishNotificationEventFunc = func(ctx context.Context, event *models.NotificationEvent) (string, error) {
		t.Fatal("self reply should not publish notification")
		return "", nil
	}

	_, err := CreateComment(context.Background(), 42, &models.ParamCreateComment{
		PostID:   "100",
		ParentID: "500",
		Content:  "reply",
	})

	require.NoError(t, err)
}

func TestCreateComment_Success_ReplyToNested(t *testing.T) {
	defer restoreCommentTestHooks()()

	genIDFunc = func() int64 { return 1003 }

	checkCommentRateLimitFunc = func(ctx context.Context, userID int64) (bool, time.Duration, error) {
		return true, 0, nil
	}
	getPostByIDFunc = func(postID int64) (*models.Post, error) {
		return &models.Post{PostID: 100}, nil
	}

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		assert.Equal(t, int64(600), commentID)
		return &models.Comment{
			CommentID: 600,
			PostID:    100,
			ParentID:  500, // 非顶级评论
			RootID:    500, // 根评论ID
		}, nil
	}

	createCommentFunc = func(c *models.Comment) error {
		return nil
	}
	incrCommentCountFunc = func(ctx context.Context, postID int64) error {
		return nil
	}

	comment, err := CreateComment(context.Background(), 42, &models.ParamCreateComment{
		PostID:   "100",
		ParentID: "600",
		Content:  "reply to nested",
	})

	require.NoError(t, err)
	assert.Equal(t, int64(600), int64(comment.ParentID))
	assert.Equal(t, int64(500), int64(comment.RootID)) // 继承父评论的根评论ID
}

func TestCreateComment_InvalidPostID(t *testing.T) {
	defer restoreCommentTestHooks()()

	_, err := CreateComment(context.Background(), 42, &models.ParamCreateComment{
		PostID:  "abc",
		Content: "hello",
	})

	require.Error(t, err)
	assert.Equal(t, "无效的ID", err.Error())
}

func TestCreateComment_InvalidParentID(t *testing.T) {
	defer restoreCommentTestHooks()()

	checkCommentRateLimitFunc = func(ctx context.Context, userID int64) (bool, time.Duration, error) {
		return true, 0, nil
	}
	getPostByIDFunc = func(postID int64) (*models.Post, error) {
		return &models.Post{PostID: 100}, nil
	}

	_, err := CreateComment(context.Background(), 42, &models.ParamCreateComment{
		PostID:   "100",
		ParentID: "abc",
		Content:  "hello",
	})

	require.Error(t, err)
	assert.Equal(t, "无效的ID", err.Error())
}

func TestCreateComment_PostNotFound(t *testing.T) {
	defer restoreCommentTestHooks()()

	checkCommentRateLimitFunc = func(ctx context.Context, userID int64) (bool, time.Duration, error) {
		return true, 0, nil
	}
	getPostByIDFunc = func(postID int64) (*models.Post, error) {
		return nil, errors.New("record not found")
	}

	_, err := CreateComment(context.Background(), 42, &models.ParamCreateComment{
		PostID:  "100",
		Content: "hello",
	})

	require.Error(t, err)
	assert.Equal(t, "record not found", err.Error())
}

func TestCreateComment_ParentCommentNotFound(t *testing.T) {
	defer restoreCommentTestHooks()()

	checkCommentRateLimitFunc = func(ctx context.Context, userID int64) (bool, time.Duration, error) {
		return true, 0, nil
	}
	getPostByIDFunc = func(postID int64) (*models.Post, error) {
		return &models.Post{PostID: 100}, nil
	}

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return nil, errors.New("record not found")
	}

	_, err := CreateComment(context.Background(), 42, &models.ParamCreateComment{
		PostID:   "100",
		ParentID: "500",
		Content:  "hello",
	})

	require.Error(t, err)
	assert.Equal(t, "record not found", err.Error())
}

func TestCreateComment_ParentCommentWrongPost(t *testing.T) {
	defer restoreCommentTestHooks()()

	checkCommentRateLimitFunc = func(ctx context.Context, userID int64) (bool, time.Duration, error) {
		return true, 0, nil
	}
	getPostByIDFunc = func(postID int64) (*models.Post, error) {
		return &models.Post{PostID: 100}, nil
	}

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{
			CommentID: 500,
			PostID:    200, // 不同的帖子
			ParentID:  0,
		}, nil
	}

	_, err := CreateComment(context.Background(), 42, &models.ParamCreateComment{
		PostID:   "100",
		ParentID: "500",
		Content:  "hello",
	})

	require.Error(t, err)
	assert.Equal(t, "无效的参数", err.Error())
}

func TestCreateComment_CreateFails(t *testing.T) {
	defer restoreCommentTestHooks()()

	genIDFunc = func() int64 { return 1001 }

	checkCommentRateLimitFunc = func(ctx context.Context, userID int64) (bool, time.Duration, error) {
		return true, 0, nil
	}
	getPostByIDFunc = func(postID int64) (*models.Post, error) {
		return &models.Post{PostID: 100}, nil
	}

	createCommentFunc = func(c *models.Comment) error {
		return errors.New("db error")
	}

	_, err := CreateComment(context.Background(), 42, &models.ParamCreateComment{
		PostID:  "100",
		Content: "hello",
	})

	require.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

func TestCreateComment_RateLimitExceeded(t *testing.T) {
	defer restoreCommentTestHooks()()

	checkCommentRateLimitFunc = func(ctx context.Context, userID int64) (bool, time.Duration, error) {
		return false, 10 * time.Second, nil // 超限，需等待 10 秒
	}

	_, err := CreateComment(context.Background(), 42, &models.ParamCreateComment{
		PostID:  "100",
		Content: "hello",
	})

	require.Error(t, err)
	assert.Equal(t, api.ErrorRateLimitExceeded, err)
}

func TestCreateComment_RateLimitRedisFails(t *testing.T) {
	defer restoreCommentTestHooks()()

	checkCommentRateLimitFunc = func(ctx context.Context, userID int64) (bool, time.Duration, error) {
		return true, 0, errors.New("redis error") // Redis 故障，放行
	}
	genIDFunc = func() int64 { return 1001 }
	getPostByIDFunc = func(postID int64) (*models.Post, error) {
		return &models.Post{PostID: 100}, nil
	}
	createCommentFunc = func(c *models.Comment) error {
		return nil
	}
	incrCommentCountFunc = func(ctx context.Context, postID int64) error {
		return nil
	}

	comment, err := CreateComment(context.Background(), 42, &models.ParamCreateComment{
		PostID:  "100",
		Content: "hello",
	})

	require.NoError(t, err) // Redis 故障时放行，创建成功
	require.NotNil(t, comment)
}

// ========== GetCommentTree 测试 ==========

func newComment(id, parentID, rootID int64) mysql.CommentWithAuthor {
	return mysql.CommentWithAuthor{
		Comment: models.Comment{
			CommentID:  models.SnowflakeID(id),
			PostID:     100,
			AuthorID:   42,
			ParentID:   models.SnowflakeID(parentID),
			RootID:     models.SnowflakeID(rootID),
			Content:    "comment",
			Status:     1,
			CreateTime: time.Now(),
		},
		AuthorName: "testuser",
	}
}

func newCommentPtr(id, parentID, rootID int64) *mysql.CommentWithAuthor {
	c := newComment(id, parentID, rootID)
	return &c
}

func TestGetCommentTree_EmptyComments(t *testing.T) {
	defer restoreCommentTestHooks()()

	getTopLevelCommentsFunc = func(postID int64, page, size int, order string) ([]*mysql.CommentWithAuthor, error) {
		return []*mysql.CommentWithAuthor{}, nil
	}
	countTopLevelCommentsFunc = func(postID int64) (int64, error) {
		return 0, nil
	}

	result, err := GetCommentTree(context.Background(), 100, &models.ParamCommentList{Page: 1, Size: 20, Order: "desc"}, 0)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.List)
	assert.Equal(t, int64(0), result.Total)
}

func TestGetCommentTree_TopLevelOnly(t *testing.T) {
	defer restoreCommentTestHooks()()

	c1 := newComment(1001, 0, 0)
	c2 := newComment(1002, 0, 0)

	getTopLevelCommentsFunc = func(postID int64, page, size int, order string) ([]*mysql.CommentWithAuthor, error) {
		assert.Equal(t, int64(100), postID)
		assert.Equal(t, 1, page)
		assert.Equal(t, 20, size)
		return []*mysql.CommentWithAuthor{&c1, &c2}, nil
	}
	countTopLevelCommentsFunc = func(postID int64) (int64, error) {
		return 2, nil
	}
	countChildCommentsByParentFunc = func(parentIDs []int64) (map[int64]int64, error) {
		return map[int64]int64{1001: 3, 1002: 0}, nil
	}

	result, err := GetCommentTree(context.Background(), 100, &models.ParamCommentList{Page: 1, Size: 20, Order: "desc"}, 0)

	require.NoError(t, err)
	assert.Len(t, result.List, 2)
	assert.Equal(t, int64(2), result.Total)
	assert.Equal(t, int64(1001), int64(result.List[0].CommentID))
	assert.Equal(t, int64(1002), int64(result.List[1].CommentID))
	assert.Nil(t, result.List[0].Children) // 不再递归加载子评论
	assert.Equal(t, int64(3), result.List[0].ChildCount)
	assert.Nil(t, result.List[1].Children)
	assert.Equal(t, int64(0), result.List[1].ChildCount)
}

func TestGetCommentTree_WithChildCount(t *testing.T) {
	defer restoreCommentTestHooks()()

	getTopLevelCommentsFunc = func(postID int64, page, size int, order string) ([]*mysql.CommentWithAuthor, error) {
		return []*mysql.CommentWithAuthor{newCommentPtr(1001, 0, 0)}, nil
	}
	countTopLevelCommentsFunc = func(postID int64) (int64, error) {
		return 1, nil
	}
	countChildCommentsByParentFunc = func(parentIDs []int64) (map[int64]int64, error) {
		assert.Equal(t, []int64{1001}, parentIDs)
		return map[int64]int64{1001: 5}, nil
	}

	result, err := GetCommentTree(context.Background(), 100, &models.ParamCommentList{Page: 1, Size: 20, Order: "desc"}, 0)

	require.NoError(t, err)
	require.Len(t, result.List, 1)

	top := result.List[0]
	assert.Equal(t, int64(1001), int64(top.CommentID))
	assert.Equal(t, int64(5), top.ChildCount)
	assert.Nil(t, top.Children) // 不再递归加载
}

func TestGetCommentTree_GetTopLevelFails(t *testing.T) {
	defer restoreCommentTestHooks()()

	getTopLevelCommentsFunc = func(postID int64, page, size int, order string) ([]*mysql.CommentWithAuthor, error) {
		return nil, errors.New("db error")
	}

	_, err := GetCommentTree(context.Background(), 100, &models.ParamCommentList{Page: 1, Size: 20, Order: "desc"}, 0)

	require.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

func TestGetCommentTree_CountFails(t *testing.T) {
	defer restoreCommentTestHooks()()

	getTopLevelCommentsFunc = func(postID int64, page, size int, order string) ([]*mysql.CommentWithAuthor, error) {
		return []*mysql.CommentWithAuthor{}, nil
	}
	countTopLevelCommentsFunc = func(postID int64) (int64, error) {
		return 0, errors.New("count error")
	}

	_, err := GetCommentTree(context.Background(), 100, &models.ParamCommentList{Page: 1, Size: 20, Order: "desc"}, 0)

	require.Error(t, err)
	assert.Equal(t, "count error", err.Error())
}

func TestGetCommentTree_CountChildrenFails(t *testing.T) {
	defer restoreCommentTestHooks()()

	c1 := newComment(1001, 0, 0)

	getTopLevelCommentsFunc = func(postID int64, page, size int, order string) ([]*mysql.CommentWithAuthor, error) {
		return []*mysql.CommentWithAuthor{&c1}, nil
	}
	countTopLevelCommentsFunc = func(postID int64) (int64, error) {
		return 1, nil
	}
	countChildCommentsByParentFunc = func(parentIDs []int64) (map[int64]int64, error) {
		return nil, errors.New("count children error")
	}

	_, err := GetCommentTree(context.Background(), 100, &models.ParamCommentList{Page: 1, Size: 20, Order: "desc"}, 0)

	require.Error(t, err)
	assert.Equal(t, "count children error", err.Error())
}

func TestGetCommentTree_Pagination(t *testing.T) {
	defer restoreCommentTestHooks()()

	c1 := newComment(1001, 0, 0)

	getTopLevelCommentsFunc = func(postID int64, page, size int, order string) ([]*mysql.CommentWithAuthor, error) {
		assert.Equal(t, 2, page)
		assert.Equal(t, 5, size)
		return []*mysql.CommentWithAuthor{&c1}, nil
	}
	countTopLevelCommentsFunc = func(postID int64) (int64, error) {
		return 10, nil
	}
	countChildCommentsByParentFunc = func(parentIDs []int64) (map[int64]int64, error) {
		return map[int64]int64{1001: 0}, nil
	}

	result, err := GetCommentTree(context.Background(), 100, &models.ParamCommentList{Page: 2, Size: 5, Order: "desc"}, 0)

	require.NoError(t, err)
	assert.Equal(t, int64(10), result.Total)
	assert.Len(t, result.List, 1)
}

// ========== GetCommentChildren 测试 ==========

func TestGetCommentChildren_Success(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		assert.Equal(t, int64(1001), commentID)
		return &models.Comment{CommentID: 1001, PostID: 100, Status: 1}, nil
	}
	getChildCommentsByParentIDFunc = func(parentID int64, page, size int, order string) ([]*mysql.CommentWithAuthor, error) {
		assert.Equal(t, int64(1001), parentID)
		assert.Equal(t, 1, page)
		assert.Equal(t, 5, size)
		assert.Equal(t, "desc", order)
		return []*mysql.CommentWithAuthor{
			newCommentPtr(2001, 1001, 1001),
			newCommentPtr(2002, 1001, 1001),
		}, nil
	}
	countChildCommentsByParentIDFunc = func(parentID int64) (int64, error) {
		assert.Equal(t, int64(1001), parentID)
		return 7, nil
	}
	countChildCommentsByParentFunc = func(parentIDs []int64) (map[int64]int64, error) {
		return map[int64]int64{2001: 2, 2002: 0}, nil
	}

	result, err := GetCommentChildren(context.Background(), 1001, &models.ParamCommentChildren{Page: 1, Size: 5, Order: "desc"}, 0)

	require.NoError(t, err)
	assert.Len(t, result.List, 2)
	assert.Equal(t, int64(7), result.Total)
	assert.True(t, result.HasMore) // 1*5 < 7
	assert.Equal(t, int64(2001), int64(result.List[0].CommentID))
	assert.Equal(t, int64(2), result.List[0].ChildCount)
	assert.Equal(t, int64(2002), int64(result.List[1].CommentID))
	assert.Equal(t, int64(0), result.List[1].ChildCount)
}

func TestGetCommentChildren_NoMore(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{CommentID: 1001, PostID: 100, Status: 1}, nil
	}
	getChildCommentsByParentIDFunc = func(parentID int64, page, size int, order string) ([]*mysql.CommentWithAuthor, error) {
		return []*mysql.CommentWithAuthor{newCommentPtr(2001, 1001, 1001)}, nil
	}
	countChildCommentsByParentIDFunc = func(parentID int64) (int64, error) {
		return 1, nil
	}
	countChildCommentsByParentFunc = func(parentIDs []int64) (map[int64]int64, error) {
		return map[int64]int64{2001: 0}, nil
	}

	result, err := GetCommentChildren(context.Background(), 1001, &models.ParamCommentChildren{Page: 1, Size: 5, Order: "desc"}, 0)

	require.NoError(t, err)
	assert.Len(t, result.List, 1)
	assert.Equal(t, int64(1), result.Total)
	assert.False(t, result.HasMore) // 1*5 >= 1
}

func TestGetCommentChildren_ParentNotFound(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return nil, errors.New("record not found")
	}

	_, err := GetCommentChildren(context.Background(), 9999, &models.ParamCommentChildren{Page: 1, Size: 5, Order: "desc"}, 0)

	require.Error(t, err)
	assert.Equal(t, "record not found", err.Error())
}

func TestGetCommentChildren_GetChildrenFails(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{CommentID: 1001, PostID: 100, Status: 1}, nil
	}
	getChildCommentsByParentIDFunc = func(parentID int64, page, size int, order string) ([]*mysql.CommentWithAuthor, error) {
		return nil, errors.New("db error")
	}

	_, err := GetCommentChildren(context.Background(), 1001, &models.ParamCommentChildren{Page: 1, Size: 5, Order: "desc"}, 0)

	require.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

func TestGetCommentChildren_CountFails(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{CommentID: 1001, PostID: 100, Status: 1}, nil
	}
	getChildCommentsByParentIDFunc = func(parentID int64, page, size int, order string) ([]*mysql.CommentWithAuthor, error) {
		return []*mysql.CommentWithAuthor{}, nil
	}
	countChildCommentsByParentIDFunc = func(parentID int64) (int64, error) {
		return 0, errors.New("count error")
	}

	_, err := GetCommentChildren(context.Background(), 1001, &models.ParamCommentChildren{Page: 1, Size: 5, Order: "desc"}, 0)

	require.Error(t, err)
	assert.Equal(t, "count error", err.Error())
}

func TestGetCommentChildren_EmptyChildren(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{CommentID: 1001, PostID: 100, Status: 1}, nil
	}
	getChildCommentsByParentIDFunc = func(parentID int64, page, size int, order string) ([]*mysql.CommentWithAuthor, error) {
		return []*mysql.CommentWithAuthor{}, nil
	}
	countChildCommentsByParentIDFunc = func(parentID int64) (int64, error) {
		return 0, nil
	}

	result, err := GetCommentChildren(context.Background(), 1001, &models.ParamCommentChildren{Page: 1, Size: 5, Order: "desc"}, 0)

	require.NoError(t, err)
	assert.Empty(t, result.List)
	assert.Equal(t, int64(0), result.Total)
	assert.False(t, result.HasMore)
}

// ========== DeleteComment 测试 ==========

func TestDeleteComment_Success_ByCommentAuthor(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{
			CommentID: 1001,
			PostID:    100,
			AuthorID:  42, // 评论作者
			Status:    1,
		}, nil
	}

	var deletedID int64
	softDeleteCommentFunc = func(commentID int64) error {
		deletedID = commentID
		return nil
	}
	decrCommentCountFunc = func(ctx context.Context, postID int64) error {
		return nil
	}

	err := DeleteComment(context.Background(), 42, 1001) // userID = 评论作者

	require.NoError(t, err)
	assert.Equal(t, int64(1001), deletedID)
}

func TestDeleteComment_Success_ByPostAuthor(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{
			CommentID: 1001,
			PostID:    100,
			AuthorID:  42, // 评论作者
			Status:    1,
		}, nil
	}
	getPostByIDFunc = func(postID int64) (*models.Post, error) {
		return &models.Post{
			PostID:   100,
			AuthorID: 99, // 帖子作者
		}, nil
	}

	var deletedID int64
	softDeleteCommentFunc = func(commentID int64) error {
		deletedID = commentID
		return nil
	}
	decrCommentCountFunc = func(ctx context.Context, postID int64) error {
		return nil
	}

	err := DeleteComment(context.Background(), 99, 1001) // userID = 帖子作者

	require.NoError(t, err)
	assert.Equal(t, int64(1001), deletedID)
}

func TestDeleteComment_CommentNotFound(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return nil, errors.New("record not found")
	}

	err := DeleteComment(context.Background(), 42, 1001)

	require.Error(t, err)
	assert.Equal(t, "record not found", err.Error())
}

func TestDeleteComment_AlreadyDeleted(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{
			CommentID: 1001,
			PostID:    100,
			AuthorID:  42,
			Status:    2, // 已删除
		}, nil
	}

	softDeleteCommentFunc = func(commentID int64) error {
		t.Fatal("should not call soft delete for already deleted comment")
		return nil
	}

	err := DeleteComment(context.Background(), 42, 1001)

	require.NoError(t, err) // 幂等返回成功
}

func TestDeleteComment_NoPermission(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{
			CommentID: 1001,
			PostID:    100,
			AuthorID:  42, // 评论作者
			Status:    1,
		}, nil
	}
	getPostByIDFunc = func(postID int64) (*models.Post, error) {
		return &models.Post{
			PostID:   100,
			AuthorID: 99, // 帖子作者
		}, nil
	}

	err := DeleteComment(context.Background(), 77, 1001) // userID = 既不是评论作者也不是帖子作者

	require.Error(t, err)
	assert.Equal(t, "没有权限执行此操作", err.Error())
}

func TestDeleteComment_SoftDeleteFails(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{
			CommentID: 1001,
			PostID:    100,
			AuthorID:  42,
			Status:    1,
		}, nil
	}
	softDeleteCommentFunc = func(commentID int64) error {
		return errors.New("db error")
	}

	err := DeleteComment(context.Background(), 42, 1001)

	require.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

// ========== LikeComment 测试 ==========

func TestLikeComment_Success(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{
			CommentID: 1001,
			PostID:    100,
			AuthorID:  42,
			Status:    1,
		}, nil
	}
	commentLikeFunc = func(ctx context.Context, commentID, userID int64) (bool, error) {
		assert.Equal(t, int64(1001), commentID)
		assert.Equal(t, int64(42), userID)
		return true, nil // 新点赞
	}
	var incrCalled bool
	incrCommentLikeCountFunc = func(commentID int64) error {
		incrCalled = true
		assert.Equal(t, int64(1001), commentID)
		return nil
	}

	err := LikeComment(context.Background(), 42, 1001)

	require.NoError(t, err)
	assert.True(t, incrCalled)
}

func TestLikeComment_PublishesNotificationForOtherAuthor(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{CommentID: 1001, PostID: 100, AuthorID: 77, Status: 1}, nil
	}
	commentLikeFunc = func(ctx context.Context, commentID, userID int64) (bool, error) { return true, nil }
	incrCommentLikeCountFunc = func(commentID int64) error { return nil }

	var published *models.NotificationEvent
	publishNotificationEventFunc = func(ctx context.Context, event *models.NotificationEvent) (string, error) {
		published = event
		return "1-0", nil
	}

	err := LikeComment(context.Background(), 42, 1001)

	require.NoError(t, err)
	require.NotNil(t, published)
	assert.Equal(t, models.NotificationTypeCommentLiked, published.Type)
	assert.Equal(t, int64(77), published.RecipientID)
	assert.Equal(t, int64(42), published.ActorID)
	assert.Equal(t, int64(100), published.PostID)
	assert.Equal(t, int64(1001), published.CommentID)
}

func TestLikeComment_SelfLikeDoesNotPublishNotification(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{CommentID: 1001, PostID: 100, AuthorID: 42, Status: 1}, nil
	}
	commentLikeFunc = func(ctx context.Context, commentID, userID int64) (bool, error) { return true, nil }
	incrCommentLikeCountFunc = func(commentID int64) error { return nil }
	publishNotificationEventFunc = func(ctx context.Context, event *models.NotificationEvent) (string, error) {
		t.Fatal("self like should not publish notification")
		return "", nil
	}

	err := LikeComment(context.Background(), 42, 1001)

	require.NoError(t, err)
}

func TestLikeComment_CommentNotFound(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return nil, errors.New("record not found")
	}

	err := LikeComment(context.Background(), 42, 1001)

	require.Error(t, err)
	assert.Equal(t, "record not found", err.Error())
}

func TestLikeComment_CommentDeleted(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{
			CommentID: 1001,
			PostID:    100,
			AuthorID:  42,
			Status:    2, // 已删除
		}, nil
	}

	err := LikeComment(context.Background(), 42, 1001)

	require.Error(t, err)
	assert.Equal(t, "无效的参数", err.Error())
}

func TestLikeComment_AlreadyLiked(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{
			CommentID: 1001,
			PostID:    100,
			AuthorID:  42,
			Status:    1,
		}, nil
	}
	commentLikeFunc = func(ctx context.Context, commentID, userID int64) (bool, error) {
		return false, nil // 已点赞过
	}
	incrCommentLikeCountFunc = func(commentID int64) error {
		t.Fatal("should not call incr for already liked comment")
		return nil
	}

	err := LikeComment(context.Background(), 42, 1001)

	require.Error(t, err)
	assert.Equal(t, "重复的点赞", err.Error())
}

func TestLikeComment_RedisFails(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{
			CommentID: 1001,
			PostID:    100,
			AuthorID:  42,
			Status:    1,
		}, nil
	}
	commentLikeFunc = func(ctx context.Context, commentID, userID int64) (bool, error) {
		return false, errors.New("redis error")
	}

	err := LikeComment(context.Background(), 42, 1001)

	require.Error(t, err)
	assert.Equal(t, "redis error", err.Error())
}

func TestLikeComment_MySQLFailsButNoError(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{
			CommentID: 1001,
			PostID:    100,
			AuthorID:  42,
			Status:    1,
		}, nil
	}
	commentLikeFunc = func(ctx context.Context, commentID, userID int64) (bool, error) {
		return true, nil // 新点赞
	}
	incrCommentLikeCountFunc = func(commentID int64) error {
		return errors.New("db error") // MySQL 失败
	}

	// 即使 MySQL 失败，点赞仍然成功（Redis 已记录）
	err := LikeComment(context.Background(), 42, 1001)

	require.NoError(t, err)
}

func TestLikeComment_MySQLFailsFirstThenRetrySucceeds(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{
			CommentID: 1001,
			PostID:    100,
			AuthorID:  42,
			Status:    1,
		}, nil
	}
	commentLikeFunc = func(ctx context.Context, commentID, userID int64) (bool, error) {
		return true, nil
	}
	callCount := 0
	incrCommentLikeCountFunc = func(commentID int64) error {
		callCount++
		if callCount == 1 {
			return errors.New("db error") // 第一次失败
		}
		return nil // 重试成功
	}

	err := LikeComment(context.Background(), 42, 1001)

	require.NoError(t, err)
	assert.Equal(t, 2, callCount) // 调用了 2 次
}

// ========== UnlikeComment 测试 ==========

func TestUnlikeComment_Success(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{
			CommentID: 1001,
			PostID:    100,
			AuthorID:  42,
			Status:    1,
		}, nil
	}
	commentUnlikeFunc = func(ctx context.Context, commentID, userID int64) (bool, error) {
		assert.Equal(t, int64(1001), commentID)
		assert.Equal(t, int64(42), userID)
		return true, nil // 取消成功
	}
	var decrCalled bool
	decrCommentLikeCountFunc = func(commentID int64) error {
		decrCalled = true
		assert.Equal(t, int64(1001), commentID)
		return nil
	}

	err := UnlikeComment(context.Background(), 42, 1001)

	require.NoError(t, err)
	assert.True(t, decrCalled)
}

func TestUnlikeComment_CommentNotFound(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return nil, errors.New("record not found")
	}

	err := UnlikeComment(context.Background(), 42, 1001)

	require.Error(t, err)
	assert.Equal(t, "record not found", err.Error())
}

func TestUnlikeComment_CommentDeleted(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{
			CommentID: 1001,
			PostID:    100,
			AuthorID:  42,
			Status:    2, // 已删除
		}, nil
	}

	err := UnlikeComment(context.Background(), 42, 1001)

	require.Error(t, err)
	assert.Equal(t, "无效的参数", err.Error())
}

func TestUnlikeComment_NotLiked_Idempotent(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{
			CommentID: 1001,
			PostID:    100,
			AuthorID:  42,
			Status:    1,
		}, nil
	}
	commentUnlikeFunc = func(ctx context.Context, commentID, userID int64) (bool, error) {
		return false, nil // 未点赞过
	}
	decrCommentLikeCountFunc = func(commentID int64) error {
		t.Fatal("should not call decr for not liked comment")
		return nil
	}

	err := UnlikeComment(context.Background(), 42, 1001)

	require.NoError(t, err) // 幂等返回成功
}

func TestUnlikeComment_RedisFails(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{
			CommentID: 1001,
			PostID:    100,
			AuthorID:  42,
			Status:    1,
		}, nil
	}
	commentUnlikeFunc = func(ctx context.Context, commentID, userID int64) (bool, error) {
		return false, errors.New("redis error")
	}

	err := UnlikeComment(context.Background(), 42, 1001)

	require.Error(t, err)
	assert.Equal(t, "redis error", err.Error())
}

func TestUnlikeComment_MySQLFailsButNoError(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{
			CommentID: 1001,
			PostID:    100,
			AuthorID:  42,
			Status:    1,
		}, nil
	}
	commentUnlikeFunc = func(ctx context.Context, commentID, userID int64) (bool, error) {
		return true, nil // 取消成功
	}
	decrCommentLikeCountFunc = func(commentID int64) error {
		return errors.New("db error") // MySQL 失败
	}

	// 即使 MySQL 失败，取消点赞仍然成功（Redis 已记录）
	err := UnlikeComment(context.Background(), 42, 1001)

	require.NoError(t, err)
}

func TestUnlikeComment_MySQLFailsFirstThenRetrySucceeds(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentByIDFunc = func(commentID int64) (*models.Comment, error) {
		return &models.Comment{
			CommentID: 1001,
			PostID:    100,
			AuthorID:  42,
			Status:    1,
		}, nil
	}
	commentUnlikeFunc = func(ctx context.Context, commentID, userID int64) (bool, error) {
		return true, nil
	}
	callCount := 0
	decrCommentLikeCountFunc = func(commentID int64) error {
		callCount++
		if callCount == 1 {
			return errors.New("db error") // 第一次失败
		}
		return nil // 重试成功
	}

	err := UnlikeComment(context.Background(), 42, 1001)

	require.NoError(t, err)
	assert.Equal(t, 2, callCount) // 调用了 2 次
}

// ========== GetUserCommentScore 测试 ==========

func TestGetUserCommentScore_WithLikes(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentLikeScoreFunc = func(authorID int64) (int64, error) {
		assert.Equal(t, int64(42), authorID)
		return 10, nil // 10 个点赞
	}

	score, err := GetUserCommentScore(42)

	require.NoError(t, err)
	assert.Equal(t, int64(500), score) // 10 × 50 = 500
}

func TestGetUserCommentScore_NoComments(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentLikeScoreFunc = func(authorID int64) (int64, error) {
		return 0, nil // 无评论
	}

	score, err := GetUserCommentScore(42)

	require.NoError(t, err)
	assert.Equal(t, int64(0), score)
}

func TestGetUserCommentScore_NoLikes(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentLikeScoreFunc = func(authorID int64) (int64, error) {
		return 0, nil // 有评论但无点赞
	}

	score, err := GetUserCommentScore(42)

	require.NoError(t, err)
	assert.Equal(t, int64(0), score)
}

func TestGetUserCommentScore_QueryFails(t *testing.T) {
	defer restoreCommentTestHooks()()

	getCommentLikeScoreFunc = func(authorID int64) (int64, error) {
		return 0, errors.New("db error")
	}

	_, err := GetUserCommentScore(42)

	require.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}
