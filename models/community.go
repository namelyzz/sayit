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
	ID        SnowflakeID `json:"community_id"`
	Name      string      `json:"community_name"`
	HotScore  float64     `json:"hot_score"` // 热度分数
	PostCount int64       `json:"post_count"`
}
