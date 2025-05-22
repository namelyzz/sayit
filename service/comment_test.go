package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/namelyzz/sayit/dao/mysql"
	"github.com/namelyzz/sayit/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func restoreCommentTestHooks() func() {
	origGetPostByID := getPostByIDFunc
	origGetCommentByID := getCommentByIDFunc
	origCreateComment := createCommentFunc
	origGetTopLevelComments := getTopLevelCommentsFunc
	origCountTopLevelComments := countTopLevelCommentsFunc
	origGetChildCommentsByParent := getChildCommentsByParentFunc
	origCountChildCommentsByParent := countChildCommentsByParentFunc
	origGenID := genIDFunc

	return func() {
		getPostByIDFunc = origGetPostByID
		getCommentByIDFunc = origGetCommentByID
		createCommentFunc = origCreateComment
		getTopLevelCommentsFunc = origGetTopLevelComments
		countTopLevelCommentsFunc = origCountTopLevelComments
		getChildCommentsByParentFunc = origGetChildCommentsByParent
		countChildCommentsByParentFunc = origCountChildCommentsByParent
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

	getTopLevelCommentsFunc = func(postID int64, page, size int) ([]*mysql.CommentWithAuthor, error) {
		return []*mysql.CommentWithAuthor{}, nil
	}
	countTopLevelCommentsFunc = func(postID int64) (int64, error) {
		return 0, nil
	}

	result, err := GetCommentTree(context.Background(), 100, &models.ParamCommentList{Page: 1, Size: 20}, 0)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.List)
	assert.Equal(t, int64(0), result.Total)
}

func TestGetCommentTree_TopLevelOnly(t *testing.T) {
	defer restoreCommentTestHooks()()

	c1 := newComment(1001, 0, 0)
	c2 := newComment(1002, 0, 0)

	getTopLevelCommentsFunc = func(postID int64, page, size int) ([]*mysql.CommentWithAuthor, error) {
		assert.Equal(t, int64(100), postID)
		assert.Equal(t, 1, page)
		assert.Equal(t, 20, size)
		return []*mysql.CommentWithAuthor{&c1, &c2}, nil
	}
	countTopLevelCommentsFunc = func(postID int64) (int64, error) {
		return 2, nil
	}
	getChildCommentsByParentFunc = func(parentIDs []int64) ([]*mysql.CommentWithAuthor, error) {
		return []*mysql.CommentWithAuthor{}, nil
	}
	countChildCommentsByParentFunc = func(parentIDs []int64) (map[int64]int64, error) {
		return map[int64]int64{}, nil
	}

	result, err := GetCommentTree(context.Background(), 100, &models.ParamCommentList{Page: 1, Size: 20}, 0)

	require.NoError(t, err)
	assert.Len(t, result.List, 2)
	assert.Equal(t, int64(2), result.Total)
	assert.Equal(t, int64(1001), int64(result.List[0].CommentID))
	assert.Equal(t, int64(1002), int64(result.List[1].CommentID))
	assert.Empty(t, result.List[0].Children)
	assert.Empty(t, result.List[1].Children)
}

func TestGetCommentTree_WithNestedChildren(t *testing.T) {
	defer restoreCommentTestHooks()()

	getTopLevelCommentsFunc = func(postID int64, page, size int) ([]*mysql.CommentWithAuthor, error) {
		return []*mysql.CommentWithAuthor{newCommentPtr(1001, 0, 0)}, nil
	}
	countTopLevelCommentsFunc = func(postID int64) (int64, error) {
		return 1, nil
	}

	callCount := 0
	getChildCommentsByParentFunc = func(parentIDs []int64) ([]*mysql.CommentWithAuthor, error) {
		callCount++
		switch callCount {
		case 1:
			// 第一次调用：获取 c1 的子评论
			assert.Equal(t, []int64{1001}, parentIDs)
			return []*mysql.CommentWithAuthor{newCommentPtr(1002, 1001, 1001)}, nil
		case 2:
			// 第二次调用：获取 c2 的子评论
			assert.Equal(t, []int64{1002}, parentIDs)
			return []*mysql.CommentWithAuthor{newCommentPtr(1003, 1002, 1001)}, nil
		default:
			// 第三次调用：获取 c3 的子评论（无子评论）
			assert.Equal(t, []int64{1003}, parentIDs)
			return []*mysql.CommentWithAuthor{}, nil
		}
	}
	countChildCommentsByParentFunc = func(parentIDs []int64) (map[int64]int64, error) {
		return map[int64]int64{parentIDs[0]: 1}, nil
	}

	result, err := GetCommentTree(context.Background(), 100, &models.ParamCommentList{Page: 1, Size: 20}, 0)

	require.NoError(t, err)
	require.Len(t, result.List, 1)

	// 验证第一层
	top := result.List[0]
	assert.Equal(t, int64(1001), int64(top.CommentID))
	assert.Equal(t, int64(1), top.ChildCount)
	require.Len(t, top.Children, 1)

	// 验证第二层
	child := top.Children[0]
	assert.Equal(t, int64(1002), int64(child.CommentID))
	assert.Equal(t, int64(1), child.ChildCount)
	require.Len(t, child.Children, 1)

	// 验证第三层
	grandchild := child.Children[0]
	assert.Equal(t, int64(1003), int64(grandchild.CommentID))
}

func TestGetCommentTree_GetTopLevelFails(t *testing.T) {
	defer restoreCommentTestHooks()()

	getTopLevelCommentsFunc = func(postID int64, page, size int) ([]*mysql.CommentWithAuthor, error) {
		return nil, errors.New("db error")
	}

	_, err := GetCommentTree(context.Background(), 100, &models.ParamCommentList{Page: 1, Size: 20}, 0)

	require.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

func TestGetCommentTree_CountFails(t *testing.T) {
	defer restoreCommentTestHooks()()

	getTopLevelCommentsFunc = func(postID int64, page, size int) ([]*mysql.CommentWithAuthor, error) {
		return []*mysql.CommentWithAuthor{}, nil
	}
	countTopLevelCommentsFunc = func(postID int64) (int64, error) {
		return 0, errors.New("count error")
	}

	_, err := GetCommentTree(context.Background(), 100, &models.ParamCommentList{Page: 1, Size: 20}, 0)

	require.Error(t, err)
	assert.Equal(t, "count error", err.Error())
}

func TestGetCommentTree_GetChildrenFails(t *testing.T) {
	defer restoreCommentTestHooks()()

	c1 := newComment(1001, 0, 0)

	getTopLevelCommentsFunc = func(postID int64, page, size int) ([]*mysql.CommentWithAuthor, error) {
		return []*mysql.CommentWithAuthor{&c1}, nil
	}
	countTopLevelCommentsFunc = func(postID int64) (int64, error) {
		return 1, nil
	}
	getChildCommentsByParentFunc = func(parentIDs []int64) ([]*mysql.CommentWithAuthor, error) {
		return nil, errors.New("children error")
	}

	_, err := GetCommentTree(context.Background(), 100, &models.ParamCommentList{Page: 1, Size: 20}, 0)

	require.Error(t, err)
	assert.Equal(t, "children error", err.Error())
}

func TestGetCommentTree_Pagination(t *testing.T) {
	defer restoreCommentTestHooks()()

	c1 := newComment(1001, 0, 0)

	getTopLevelCommentsFunc = func(postID int64, page, size int) ([]*mysql.CommentWithAuthor, error) {
		assert.Equal(t, 2, page)
		assert.Equal(t, 5, size)
		return []*mysql.CommentWithAuthor{&c1}, nil
	}
	countTopLevelCommentsFunc = func(postID int64) (int64, error) {
		return 10, nil
	}
	getChildCommentsByParentFunc = func(parentIDs []int64) ([]*mysql.CommentWithAuthor, error) {
		return []*mysql.CommentWithAuthor{}, nil
	}
	countChildCommentsByParentFunc = func(parentIDs []int64) (map[int64]int64, error) {
		return map[int64]int64{}, nil
	}

	result, err := GetCommentTree(context.Background(), 100, &models.ParamCommentList{Page: 2, Size: 5}, 0)

	require.NoError(t, err)
	assert.Equal(t, int64(10), result.Total)
	assert.Len(t, result.List, 1)
}
