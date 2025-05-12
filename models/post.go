package models

import "time"

// Post 帖子模型，对应数据库 `post` 表
// 创建时必须填写: Title, Content, CommunityID
// AuthorID 由 controller 从 JWT 上下文中获取，客户端不可篡改
// PostID 由雪花算法生成，CreateTime 由 service 层设置
type Post struct {
	PostID      SnowflakeID `json:"post_id" gorm:"column:post_id"`                              // 雪花算法生成的全局唯一帖子ID
	Title       string      `json:"title" gorm:"column:title" binding:"required"`               // 帖子标题，必填
	Content     string      `json:"content" gorm:"column:content" binding:"required"`           // 帖子内容，必填
	AuthorID    int64       `json:"author_id" gorm:"column:author_id"`                          // 作者用户ID（JWT上下文自动设置）
	CommunityID int64       `json:"community_id" gorm:"column:community_id" binding:"required"` // 所属社区ID，必填
	Status      int32       `json:"status" gorm:"column:status;default:1"`                      // 帖子状态: 1=正常
	CreateTime  time.Time   `json:"create_time" gorm:"column:create_time;autoCreateTime"`       // 创建时间
	UpdateTime  time.Time   `json:"update_time" gorm:"column:update_time;autoUpdateTime"`       // 更新时间
}

func (Post) TableName() string {
	return "post"
}

// PostDetail 帖子详情，用于详情接口返回
// 组合了帖子主体、作者名和社区详情（分三次查询后组装）
type PostDetail struct {
	AuthorName       string             `json:"author_name"`       // 作者用户名
	LikeCount        int64              `json:"like_count"`        // 帖子热度分值
	VoteCount        int64              `json:"vote_count"`        // 赞成票与反对票的净值
	CommentCount     int64              `json:"comment_count"`     // 预留字段，当前未实现
	CurrentUserVote  int8               `json:"current_user_vote"` // 当前用户投票状态: 1=赞成, 0=未投/未登录, -1=反对
	*Post                               // 嵌入帖子主体
	*CommunityDetail `json:"community"` // 社区详情
}

// PostListItem 帖子列表项 - 用于列表接口
// 通过 SQL JOIN 查询组装，包含作者名、社区名和内容摘要
type PostListItem struct {
	PostID          SnowflakeID `json:"post_id" gorm:"column:post_id"`
	Title           string      `json:"title" gorm:"column:title"`
	Summary         string      `json:"summary" gorm:"column:summary"` // 内容摘要（SQL中截取前30字符）
	AuthorID        int64       `json:"author_id" gorm:"column:author_id"`
	Username        string      `json:"user_name" gorm:"column:user_name"` // JOIN users 表获取
	CommunityID     int64       `json:"community_id" gorm:"column:community_id"`
	CommunityName   string      `json:"community_name" gorm:"column:community_name"` // JOIN community 表获取
	Status          int32       `json:"status" gorm:"column:status"`
	CreateTime      time.Time   `json:"create_time" gorm:"column:create_time"`
	UpdateTime      time.Time   `json:"update_time" gorm:"column:update_time"`
	CommentCount    int64       `json:"comment_count" gorm:"column:comment_count"` // 预留字段，当前未实现
	LikeCount       int64       `json:"like_count" gorm:"column:like_count"`       // 帖子热度分值
	VoteCount       int64       `json:"vote_count" gorm:"-"`                       // 赞成票与反对票的净值
	CurrentUserVote int8        `json:"current_user_vote" gorm:"-"`                // 当前用户投票状态
}

func (PostListItem) TableName() string {
	return "post"
}
