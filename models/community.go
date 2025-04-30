package models

import "time"

type Community struct {
	ID   SnowflakeID `json:"community_id" gorm:"column:community_id"`
	Name string      `json:"name" gorm:"column:community_name"`
}

func (Community) TableName() string {
	return "community"
}

type CommunityDetail struct {
	ID           SnowflakeID `json:"community_id" gorm:"column:community_id"`
	Name         string      `json:"name" gorm:"column:community_name"`
	Introduction string      `json:"introduction,omitempty" gorm:"introduction"`
	CreateTime   time.Time   `json:"create_time" gorm:"create_time"`
}

func (CommunityDetail) TableName() string {
	return "community"
}

// HotCommunity 热门社区（用于接口返回）
type HotCommunity struct {
	ID         SnowflakeID `json:"community_id"`
	Name       string      `json:"community_name"`
	HotScore   float64     `json:"hot_score"`
	PostCount  int64       `json:"post_count"`
}

// CommunityFollow 用户关注社区关联表
// 复合主键 (user_id, community_id) 保证唯一性
type CommunityFollow struct {
	UserID      SnowflakeID `json:"user_id" gorm:"column:user_id;primaryKey"`
	CommunityID SnowflakeID `json:"community_id" gorm:"column:community_id;primaryKey"`
	CreateTime  time.Time   `json:"create_time" gorm:"column:create_time;autoCreateTime"`
}

func (CommunityFollow) TableName() string {
	return "community_follow"
}

// FollowStatus 关注状态（用于接口返回）
type FollowStatus struct {
	IsFollowed bool `json:"is_followed"`
}
