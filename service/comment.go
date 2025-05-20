package service

import (
	"context"
	"github.com/namelyzz/sayit/dao/mysql"
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/utils/api"
	"go.uber.org/zap"
	"strconv"
)

// 以下变量可被测试替换（mock），方便单元测试
var (
	getPostByIDFunc    = mysql.GetPostByID    // 查询帖子
	getCommentByIDFunc = mysql.GetCommentByID // 查询评论
	createCommentFunc  = mysql.CreateComment  // 创建评论
)

// CreateComment 创建评论的业务逻辑
// 1. 验证帖子是否存在
// 2. 如果是回复评论，验证父评论是否存在且属于同一帖子
// 3. 生成评论ID并入库
func CreateComment(ctx context.Context, userID int64, p *models.ParamCreateComment) (*models.Comment, error) {
	// 1. 解析帖子ID
	postID, err := strconv.ParseInt(p.PostID, 10, 64)
	if err != nil {
		return nil, api.ErrorInvalidID
	}

	// 2. 验证帖子是否存在
	post, err := getPostByIDFunc(postID)
	if err != nil {
		zap.L().Error("post not found",
			zap.Int64("post_id", postID),
			zap.Error(err))
		return nil, err
	}

	// 3. 解析父评论ID
	var parentID int64
	if p.ParentID != "" {
		parentID, err = strconv.ParseInt(p.ParentID, 10, 64)
		if err != nil {
			return nil, api.ErrorInvalidID
		}
	}

	// 4. 如果是回复评论，验证父评论
	var rootID int64
	if parentID != 0 {
		parentComment, err := getCommentByIDFunc(parentID)
		if err != nil {
			zap.L().Error("parent comment not found",
				zap.Int64("parent_id", parentID),
				zap.Error(err))
			return nil, err
		}
		// 父评论必须属于同一帖子
		if int64(parentComment.PostID) != postID {
			return nil, api.ErrorInvalidParam
		}
		// 计算根评论ID：如果父评论是顶级评论，则根评论ID为父评论ID；否则继承父评论的根评论ID
		if int64(parentComment.ParentID) == 0 {
			rootID = parentID
		} else {
			rootID = int64(parentComment.RootID)
		}
	}

	// 5. 生成评论ID
	commentID := genIDFunc()

	// 6. 构造评论对象
	comment := &models.Comment{
		CommentID: models.SnowflakeID(commentID),
		PostID:    models.SnowflakeID(postID),
		AuthorID:  models.SnowflakeID(userID),
		ParentID:  models.SnowflakeID(parentID),
		RootID:    models.SnowflakeID(rootID),
		Content:   p.Content,
		Status:    1,
	}

	// 7. 入库
	if err := createCommentFunc(comment); err != nil {
		zap.L().Error("create comment failed",
			zap.Int64("post_id", postID),
			zap.Int64("author_id", userID),
			zap.Error(err))
		return nil, err
	}

	_ = post // 避免未使用变量警告

	return comment, nil
}
