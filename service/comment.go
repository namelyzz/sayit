package service

import (
	"context"
	"github.com/namelyzz/sayit/dao/mysql"
	"github.com/namelyzz/sayit/dao/redis"
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/utils/api"
	"go.uber.org/zap"
	"strconv"
)

const maxCommentDepth = 10 // 最大嵌套深度

// 以下变量可被测试替换（mock），方便单元测试
var (
	getPostByIDFunc               = mysql.GetPostByID               // 查询帖子
	getCommentByIDFunc            = mysql.GetCommentByID            // 查询评论
	createCommentFunc             = mysql.CreateComment             // 创建评论
	getTopLevelCommentsFunc       = mysql.GetTopLevelComments       // 获取顶级评论
	countTopLevelCommentsFunc     = mysql.CountTopLevelComments     // 统计顶级评论数量
	getChildCommentsByParentFunc  = mysql.GetChildCommentsByParentIDs  // 获取子评论
	countChildCommentsByParentFunc = mysql.CountChildCommentsByParentIDs // 统计子评论数量
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

// GetCommentTree 获取帖子的评论树
// 1. 分页获取顶级评论
// 2. 递归获取子评论，构建树形结构
// 3. 回填点赞状态
func GetCommentTree(ctx context.Context, postID int64, p *models.ParamCommentList, currentUserID int64) (*models.CommentListResponse, error) {
	// 1. 获取顶级评论
	topComments, err := getTopLevelCommentsFunc(postID, p.Page, p.Size)
	if err != nil {
		return nil, err
	}

	// 2. 统计顶级评论总数
	total, err := countTopLevelCommentsFunc(postID)
	if err != nil {
		return nil, err
	}

	// 3. 转换为 CommentDetail 并递归构建子评论树
	details := make([]*models.CommentDetail, 0, len(topComments))
	for _, c := range topComments {
		detail := &models.CommentDetail{
			Comment:    &c.Comment,
			AuthorName: c.AuthorName,
		}
		details = append(details, detail)
	}

	// 4. 递归填充子评论
	if err := fillChildren(details, 1); err != nil {
		return nil, err
	}

	// 5. 回填点赞状态
	fillCommentLiked(ctx, details, currentUserID)

	return &models.CommentListResponse{
		List:  details,
		Total: total,
	}, nil
}

// fillChildren 递归填充子评论
// depth: 当前深度（从1开始）
func fillChildren(parents []*models.CommentDetail, depth int) error {
	if depth >= maxCommentDepth || len(parents) == 0 {
		return nil
	}

	// 收集所有父评论ID
	parentIDs := make([]int64, 0, len(parents))
	for _, p := range parents {
		parentIDs = append(parentIDs, int64(p.CommentID))
	}

	// 批量获取子评论
	children, err := getChildCommentsByParentFunc(parentIDs)
	if err != nil {
		return err
	}

	// 批量统计子评论数量
	childCountMap, err := countChildCommentsByParentFunc(parentIDs)
	if err != nil {
		return err
	}

	// 按 parent_id 分组
	childrenByParent := make(map[int64][]*mysql.CommentWithAuthor, len(parents))
	for _, child := range children {
		pid := int64(child.ParentID)
		childrenByParent[pid] = append(childrenByParent[pid], child)
	}

	// 为每个父评论填充子评论
	for _, parent := range parents {
		pid := int64(parent.CommentID)
		parent.ChildCount = childCountMap[pid]

		childList := childrenByParent[pid]
		if len(childList) == 0 {
			continue
		}

		parent.Children = make([]*models.CommentDetail, 0, len(childList))
		for _, c := range childList {
			// 创建新的 Comment 副本，避免指针共享问题
			commentCopy := c.Comment
			childDetail := &models.CommentDetail{
				Comment:    &commentCopy,
				AuthorName: c.AuthorName,
			}
			parent.Children = append(parent.Children, childDetail)
		}

		// 递归填充下一层
		if err := fillChildren(parent.Children, depth+1); err != nil {
			return err
		}
	}

	return nil
}

// fillCommentLiked 批量回填评论点赞状态（仅登录用户）
func fillCommentLiked(ctx context.Context, details []*models.CommentDetail, currentUserID int64) {
	if currentUserID == 0 {
		return
	}

	// 收集所有评论ID（包括子评论）
	allIDs := collectCommentIDs(details, nil)
	if len(allIDs) == 0 {
		return
	}

	// 批量查询点赞状态
	likedMap := redis.BatchIsCommentLikedByUser(ctx, allIDs, currentUserID)

	// 回填
	setCommentLiked(details, likedMap)
}

// collectCommentIDs 递归收集所有评论ID
func collectCommentIDs(details []*models.CommentDetail, ids []int64) []int64 {
	for _, d := range details {
		ids = append(ids, int64(d.CommentID))
		if len(d.Children) > 0 {
			ids = collectCommentIDs(d.Children, ids)
		}
	}
	return ids
}

// setCommentLiked 递归设置点赞状态
func setCommentLiked(details []*models.CommentDetail, likedMap map[int64]bool) {
	for _, d := range details {
		d.IsLiked = likedMap[int64(d.CommentID)]
		if len(d.Children) > 0 {
			setCommentLiked(d.Children, likedMap)
		}
	}
}
