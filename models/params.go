package models

import "fmt"

// ParamSignUp 注册请求参数
type ParamSignUp struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	RePassword string `json:"re_password" binding:"required,eqfield=Password"`
}

// ParamLogin 登录请求参数
type ParamLogin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

/*
定义排序字段和方向的枚举类型
*/

type SortField string

const (
	SortFieldCreateTime SortField = "create_time"
	SortFieldUpdateTime SortField = "update_time"
	SortFieldScore      SortField = "score"
)

type SortDirection string

const (
	SortDirectionDesc SortDirection = "desc"
	SortDirectionAsc  SortDirection = "asc"
)

// SortCondition 若有需要，可以作为支持多字段排序的扩展
type SortCondition struct {
	Field     SortField     `json:"field"`
	Direction SortDirection `json:"direction"`
}

type ParamPostList struct {
	CommunityID int64         `json:"community_id" form:"community_id"`
	Page        int           `json:"page" form:"page"`
	Size        int           `json:"size" form:"size"`
	Status      int           `json:"status" form:"status"`
	SortBy      SortField     `json:"sort_by" form:"sort_by"`
	Order       SortDirection `json:"order" form:"order"`
}

func (p *ParamPostList) ValidateAndSetDefaults() error {
	// 设置默认值, 默认按创建时间倒序排序
	if p.SortBy == "" {
		p.SortBy = SortFieldCreateTime
	}
	if p.Order == "" {
		p.Order = SortDirectionDesc
	}

	// 状态筛选（通常只显示正常状态的帖子）
	if p.Status == 0 {
		p.Status = 1
	}

	// 验证排序字段
	validSortFields := map[SortField]bool{
		SortFieldCreateTime: true,
		SortFieldUpdateTime: true,
		SortFieldScore:      true,
	}
	if !validSortFields[p.SortBy] {
		return fmt.Errorf("invalid sort_by: %s, supported: create_time, update_time, score", p.SortBy)
	}

	// 验证排序方向
	validDirections := map[SortDirection]bool{
		SortDirectionDesc: true,
		SortDirectionAsc:  true,
	}
	if !validDirections[p.Order] {
		return fmt.Errorf("invalid order: %s, supported: desc, asc", p.Order)
	}

	return nil
}

type ParamVote struct {
	// UserID 从请求中获取当前的用户
	PostID    string `json:"post_id" binding:"required"`               // 贴子id
	Direction int8   `json:"direction,string" binding:"oneof=1 0 -1" ` // 赞成票(1)还是反对票(-1)取消投票(0)
}
