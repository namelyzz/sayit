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
	CreateTime   time.Time   `json:"create_time" gorm:"column:create_time;autoCreateTime"`
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

// SearchSuggestItem 搜索建议项
type SearchSuggestItem struct {
	ID   SnowflakeID `json:"id"`
	Name string      `json:"name"`
}

// CommunityListItem 社区列表项（带详情和帖子数量）
type CommunityListItem struct {
	ID           SnowflakeID `json:"community_id" gorm:"column:community_id"`
	Name         string      `json:"name" gorm:"column:community_name"`
	Introduction string      `json:"introduction" gorm:"column:introduction"`
	PostCount    int64       `json:"post_count" gorm:"column:post_count"`
	CreateTime   time.Time   `json:"create_time" gorm:"column:create_time"`
}

// CommunityListResponse 社区列表响应
type CommunityListResponse struct {
	List  []*CommunityListItem `json:"list"`
	Total int64                `json:"total"`
}

// ParamCreateCommunity 创建社区请求参数
type ParamCreateCommunity struct {
	Name         string `json:"name" binding:"required,min=2,max=128"`
	Introduction string `json:"introduction" binding:"required,min=2,max=256"`
}
