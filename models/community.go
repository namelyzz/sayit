package models

import "time"

type Community struct {
    ID   int64  `json:"community_id" gorm:"column:community_id"`
    Name string `json:"name" gorm:"column:community_name"`
}

func (Community) TableName() string {
    return "community"
}

type CommunityDetail struct {
    ID           int64     `json:"community_id" gorm:"column:community_id"`
    Name         string    `json:"name" gorm:"column:community_name"`
    Introduction string    `json:"introduction,omitempty" gorm:"introduction"`
    CreateTime   time.Time `json:"create_time" gorm:"create_time"`
}

func (CommunityDetail) TableName() string {
    return "community"
}
