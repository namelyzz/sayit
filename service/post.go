package service

import (
	"github.com/namelyzz/sayit/dao/mysql"
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/utils/snowflake"
	"go.uber.org/zap"
)

func CreatePost(p *models.Post) (err error) {
	// 使用雪花算法为帖子生成一个 ID
	p.PostID = snowflake.GenID()
	err = mysql.CreatePost(p)
	if err != nil {
		return err
	}
	return
}

func GetPostDetailByID(postID int64) (res *models.PostDetail, err error) {
	post, err := mysql.GetPostByID(postID)
	if err != nil {
		zap.L().Error("mysql.GetPostByID failed",
			zap.Int64("postID", postID),
			zap.Error(err))
		return nil, err
	}

	authorID := post.AuthorID
	user, err := mysql.GetUserByID(authorID)
	if err != nil {
		zap.L().Error("mysql.GetUserByID failed",
			zap.Int64("author_id", authorID),
			zap.Error(err))
		return nil, err
	}

	communityID := post.CommunityID
	detail, err := mysql.GetCommunityDetailByID(communityID)
	if err != nil {
		zap.L().Error("mysql.GetCommunityDetailByID failed",
			zap.Int64("community_id", communityID),
			zap.Error(err))
		return nil, err
	}

	return &models.PostDetail{
		AuthorName:      user.Username,
		Post:            post,
		CommunityDetail: detail,
	}, nil
}

func GetPostList(p *models.ParamPostList) (posts []*models.PostListItem, err error) {
	return mysql.GetPostList(p)
}
