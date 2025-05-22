package mysql

import (
	"github.com/namelyzz/sayit/models"
	"go.uber.org/zap"
)

// CreateComment 将评论记录插入 MySQL comment 表
func CreateComment(c *models.Comment) (err error) {
	res := db.Omit("UpdateTime").Create(c)
	if res.Error != nil {
		zap.L().Error("create comment failed",
			zap.String("operation", "create_comment"),
			zap.Int64("post_id", int64(c.PostID)),
			zap.Int64("author_id", int64(c.AuthorID)),
			zap.Error(res.Error))
		return res.Error
	}
	return nil
}

// GetCommentByID 根据评论ID查询单个评论
func GetCommentByID(commentID int64) (comment *models.Comment, err error) {
	comment = new(models.Comment)
	res := db.Where("comment_id = ?", commentID).First(comment)
	if res.Error != nil {
		return nil, res.Error
	}
	return comment, nil
}

// CommentWithAuthor 包含作者用户名的评论
type CommentWithAuthor struct {
	models.Comment
	AuthorName string `json:"author_name" gorm:"column:author_name"`
}

// GetTopLevelComments 获取帖子的顶级评论（parent_id=0），按创建时间倒序，分页
func GetTopLevelComments(postID int64, page, size int) ([]*CommentWithAuthor, error) {
	var comments []*CommentWithAuthor
	offset := (page - 1) * size
	res := db.Table("comment c").
		Select("c.*, u.username AS author_name").
		Joins("LEFT JOIN users u ON c.author_id = u.user_id").
		Where("c.post_id = ? AND c.parent_id = 0", postID).
		Order("c.create_time DESC").
		Offset(offset).Limit(size).
		Scan(&comments)
	if res.Error != nil {
		zap.L().Error("get top level comments failed",
			zap.Int64("post_id", postID),
			zap.Error(res.Error))
		return nil, res.Error
	}
	return comments, nil
}

// CountTopLevelComments 统计帖子的顶级评论数量
func CountTopLevelComments(postID int64) (int64, error) {
	var count int64
	res := db.Model(&models.Comment{}).
		Where("post_id = ? AND parent_id = 0", postID).
		Count(&count)
	if res.Error != nil {
		return 0, res.Error
	}
	return count, nil
}

// GetChildCommentsByParentIDs 批量获取指定父评论ID列表的子评论，按创建时间正序
func GetChildCommentsByParentIDs(parentIDs []int64) ([]*CommentWithAuthor, error) {
	if len(parentIDs) == 0 {
		return []*CommentWithAuthor{}, nil
	}
	var comments []*CommentWithAuthor
	res := db.Table("comment c").
		Select("c.*, u.username AS author_name").
		Joins("LEFT JOIN users u ON c.author_id = u.user_id").
		Where("c.parent_id IN ?", parentIDs).
		Order("c.create_time ASC").
		Scan(&comments)
	if res.Error != nil {
		zap.L().Error("get child comments failed",
			zap.Int("parent_count", len(parentIDs)),
			zap.Error(res.Error))
		return nil, res.Error
	}
	return comments, nil
}

// CountChildCommentsByParentIDs 批量统计指定父评论的子评论数量
func CountChildCommentsByParentIDs(parentIDs []int64) (map[int64]int64, error) {
	if len(parentIDs) == 0 {
		return map[int64]int64{}, nil
	}
	type result struct {
		ParentID int64 `gorm:"column:parent_id"`
		Count    int64 `gorm:"column:cnt"`
	}
	var results []result
	res := db.Model(&models.Comment{}).
		Select("parent_id, COUNT(*) AS cnt").
		Where("parent_id IN ?", parentIDs).
		Group("parent_id").
		Scan(&results)
	if res.Error != nil {
		return nil, res.Error
	}
	countMap := make(map[int64]int64, len(results))
	for _, r := range results {
		countMap[r.ParentID] = r.Count
	}
	return countMap, nil
}
