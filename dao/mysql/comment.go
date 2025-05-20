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
