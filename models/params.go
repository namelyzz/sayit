package models

import (
	"fmt"
	"strings"
)

// ParamSignUp 注册请求参数
// binding tag 用于 gin 的参数验证:
//   - required: 字段必填
//   - eqfield=Password: RePassword 必须与 Password 字段值相等（确认密码校验）
type ParamSignUp struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	RePassword string `json:"re_password" binding:"required,eqfield=Password"`
}

// ParamLogin 登录请求参数
// 仅需用户名和密码，两个字段均为必填
type ParamLogin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ParamUpdateProfile 更新当前用户资料请求参数
type ParamUpdateProfile struct {
	Signature string `json:"signature" binding:"max=120"`
}

/*
定义排序字段和方向的枚举类型
*/

// SortField 排序字段枚举
type SortField string

const (
	SortFieldCreateTime SortField = "create_time" // 按创建时间排序（MySQL 路径）
	SortFieldUpdateTime SortField = "update_time" // 按更新时间排序（MySQL 路径）
	SortFieldScore      SortField = "score"       // 按热度分数排序（Redis 路径优先）
)

// SortDirection 排序方向枚举
type SortDirection string

const (
	SortDirectionDesc SortDirection = "desc" // 倒序（新/热 → 旧/冷）
	SortDirectionAsc  SortDirection = "asc"  // 正序（旧/冷 → 新/热）
)

// SortCondition 若有需要，可以作为支持多字段排序的扩展
type SortCondition struct {
	Field     SortField     `json:"field"`
	Direction SortDirection `json:"direction"`
}

// ParamPostList 获取帖子列表请求参数
// 通过 URL 查询参数传递，例如: /api/v1/posts?community_id=1&page=2&size=10&sort_by=score
// form tag 用于 Gin 的 ShouldBindQuery 绑定查询参数
type ParamPostList struct {
	CommunityID int64  `json:"community_id" form:"community_id"` // 社区ID筛选（精确匹配）
	UserName    string `json:"user_name" form:"user_name"`       // 作者名筛选（模糊搜索）
	Keyword     string `json:"keyword" form:"keyword"`           // 标题关键词筛选（模糊搜索）

	// 按 创建时间 的范围查询（Unix 时间戳，秒级）
	StartTime *int64 `json:"start_time" form:"start_time"` // 起始时间（含）
	EndTime   *int64 `json:"end_time" form:"end_time"`     // 结束时间（含）

	Page   int           `json:"page" form:"page"`       // 页码（从1开始）
	Size   int           `json:"size" form:"size"`       // 每页数量（最大50）
	Status *int          `json:"status" form:"status"`   // 帖子状态筛选（1=正常，默认1）
	SortBy SortField     `json:"sort_by" form:"sort_by"` // 排序字段（create_time/update_time/score）
	Order  SortDirection `json:"order" form:"order"`     // 排序方向（desc/asc）
}

// MaxPageSize 每页最大帖子数量
const (
	MaxPageSize = 50
)

// ValidateAndSetDefaults 校验参数并设置默认值
// 在 controller 层调用，确保参数合法后再传给 service 层
func (p *ParamPostList) ValidateAndSetDefaults() error {
	// 设置默认排序: 按创建时间倒序
	if p.SortBy == "" {
		p.SortBy = SortFieldCreateTime
	}
	if p.Order == "" {
		p.Order = SortDirectionDesc
	}

	// 设置分页默认值: 第1页，每页50条
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Size <= 0 || p.Size > MaxPageSize {
		p.Size = MaxPageSize
	}

	// 去除 keyword 和 username 两边空格，避免空格影响 LIKE 查询
	p.Keyword = strings.TrimSpace(p.Keyword)
	p.UserName = strings.TrimSpace(p.UserName)

	// 时间范围校验: 起始时间不能大于结束时间
	if p.StartTime != nil && p.EndTime != nil {
		if *p.StartTime > *p.EndTime {
			return fmt.Errorf("start_time cannot be greater than end_time")
		}
	}

	// 状态筛选: 默认只返回正常状态的帖子
	if p.Status == nil {
		defaultStatus := 1
		p.Status = &defaultStatus
	}

	// 验证排序字段是否合法
	validSortFields := map[SortField]bool{
		SortFieldCreateTime: true,
		SortFieldUpdateTime: true,
		SortFieldScore:      true,
	}
	if !validSortFields[p.SortBy] {
		return fmt.Errorf("invalid sort_by: %s, supported: create_time, update_time, score", p.SortBy)
	}

	// 验证排序方向是否合法
	validDirections := map[SortDirection]bool{
		SortDirectionDesc: true,
		SortDirectionAsc:  true,
	}
	if !validDirections[p.Order] {
		return fmt.Errorf("invalid order: %s, supported: desc, asc", p.Order)
	}

	return nil
}

// ParamVote 投票请求参数
// UserID 不在请求体中传递，而是从 JWT 上下文中获取（防止伪造）
type ParamVote struct {
	PostID    string `json:"post_id" binding:"required"`               // 帖子ID（雪花算法生成的数字字符串）
	Direction int8   `json:"direction,string" binding:"oneof=1 0 -1" ` // 投票方向: 1=赞成, 0=取消, -1=反对
	// 注意: direction 的 json tag 有 ",string" 修饰，表示 JSON 中可以传字符串 "1" 或数字 1
	// binding:"oneof=1 0 -1" 限制只能是这三个值之一
}
