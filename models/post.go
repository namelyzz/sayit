package models

import "time"

type Post struct {
	PostID      int64     `json:"post_id" gorm:"column:post_id"`
	Title       string    `json:"title" gorm:"column:title" binding:"required"`
	Content     string    `json:"content" gorm:"column:content" binding:"required"`
	AuthorID    int64     `json:"author_id" gorm:"column:author_id"`
	CommunityID int64     `json:"community_id" gorm:"column:community_id" binding:"required"`
	Status      int32     `json:"status" gorm:"column:status;default:1"`
	CreateTime  time.Time `json:"create_time" gorm:"column:create_time;autoCreateTime"`
	UpdateTime  time.Time `json:"update_time" gorm:"column:update_time;autoUpdateTime"`
}

func (Post) TableName() string {
	return "post"
}

type PostDetail struct {
	AuthorName string `json:"author_name"`
	*Post
	*CommunityDetail `json:"community"`
}

// PostListItem 帖子列表项 - 用于列表接口
type PostListItem struct {
	PostID        int64     `json:"post_id" gorm:"column:post_id"`
	Title         string    `json:"title" gorm:"column:title"`
	Summary       string    `json:"summary" gorm:"column:summary"`
	AuthorID      int64     `json:"author_id" gorm:"column:author_id"`
	Username      string    `json:"user_name" gorm:"column:user_name"`
	CommunityID   int64     `json:"community_id" gorm:"column:community_id"`
	CommunityName string    `json:"community_name" gorm:"column:community_name"`
	Status        int32     `json:"status" gorm:"column:status"`
	CreateTime    time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime    time.Time `json:"update_time" gorm:"column:update_time"`
	CommentCount  int64     `json:"comment_count" gorm:"column:comment_count"`
	LikeCount     int64     `json:"like_count" gorm:"column:like_count"`
}

func (PostListItem) TableName() string {
	return "post"
}
