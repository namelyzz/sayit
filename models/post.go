package models

import "time"

type Post struct {
	ID          int64     `json:"id,string" gorm:"column:id;primaryKey;autoIncrement"`
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
