package service

import (
	"context"
	"errors"
	"testing"

	"github.com/namelyzz/sayit/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func restoreCommentTestHooks() func() {
	origGetPostByID := getPostByIDFunc
	origGetCommentByID := getCommentByIDFunc
	origCreateComment := createCommentFunc
	origGenID := genIDFunc

	return func() {
		getPostByIDFunc = origGetPostByID
		getCommentByIDFunc = origGetCommentByID
		createCommentFunc = origCreateComment
		genIDFunc = origGenID
	}
}

func TestCreateComment_Success_TopLevel(t *testing.T) {
	defer restoreCommentTestHooks()()

	genIDFunc = func() int64 { return 1001 }

	getPostByIDFunc = func(postID int64) (*models.Post, error) {
		assert.Equal(t, int64(100), postID)
		return &models.Post{PostID: 100}, nil
	}

	var createdComment *models.Comment
	createCommentFunc = func(c *models.Comment) error {
		createdComment = c
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

func TestCreateComment_Success_ReplyToTopLevel(t *testing.T) {
	defer restoreCommentTestHooks()()

	genIDFunc = func() int64 { return 1002 }

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

	comment, err := CreateComment(context.Background(), 42, &models.ParamCreateComment{
		PostID:   "100",
		ParentID: "500",
		Content:  "reply to top level",
	})

	require.NoError(t, err)
	assert.Equal(t, int64(500), int64(comment.ParentID))
	assert.Equal(t, int64(500), int64(comment.RootID)) // 根评论ID = 父评论ID
}

func TestCreateComment_Success_ReplyToNested(t *testing.T) {
	defer restoreCommentTestHooks()()

	genIDFunc = func() int64 { return 1003 }

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
