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
