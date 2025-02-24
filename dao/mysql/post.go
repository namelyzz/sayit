package mysql

import (
	"github.com/namelyzz/sayit/models"
	"go.uber.org/zap"
)

func CreatePost(p *models.Post) (err error) {
	res := db.Omit("CreateTime", "UpdateTime").Create(p)
	if res.Error != nil {
		zap.L().Error("create post failed",
			zap.String("operation", "create_post"),
			zap.Int64("author_id", p.AuthorID),
			zap.Int64("community_id", p.CommunityID),
			zap.Error(res.Error))
		return res.Error
	}
	return nil
}

func GetPostByID(postID int64) (post *models.Post, err error) {
	post = new(models.Post)
	res := db.Model(&models.Post{}).
		Select("post_id", "title", "content", "author_id", "community_id", "status", "create_time", "update_time").
		Where("post_id = ?", postID).First(post)

	if res.Error != nil {
		return nil, res.Error
	}
	return post, nil
}
