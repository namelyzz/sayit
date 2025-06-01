package models

import "time"

// Comment 评论模型，对应数据库 `comment` 表
//
// 嵌套结构设计:
//   - parent_id: 指向直接父评论，顶级评论为 0
//   - root_id: 指向根评论（顶级评论自身），用于快速定位某条评论所属的顶级评论树
//
// 软删除策略:
//   - status=1: 正常状态
//   - status=2: 已删除（保留子评论结构，内容显示为 [已删除]）
//
// LikeCount 是冗余字段，与 Redis 中的点赞记录保持最终一致性:
//   - 点赞时: Redis SADD + MySQL INCR like_count
//   - 取消点赞时: Redis SREM + MySQL DECR like_count
type Comment struct {
	CommentID  SnowflakeID `json:"comment_id" gorm:"column:comment_id"`   // 雪花算法生成的全局唯一评论ID
	PostID     SnowflakeID `json:"post_id" gorm:"column:post_id"`         // 所属帖子ID
	AuthorID   SnowflakeID `json:"author_id" gorm:"column:author_id"`     // 作者用户ID
	ParentID   SnowflakeID `json:"parent_id" gorm:"column:parent_id"`     // 父评论ID（0表示顶级评论）
	RootID     SnowflakeID `json:"root_id" gorm:"column:root_id"`         // 根评论ID（0表示自己是根评论）
	Content    string      `json:"content" gorm:"column:content"`         // 评论内容
	LikeCount  int64       `json:"like_count" gorm:"column:like_count"`   // 点赞数（冗余字段，与 Redis 最终一致）
	Status     int32       `json:"status" gorm:"column:status;default:1"` // 状态: 1=正常, 2=已删除
	CreateTime time.Time   `json:"create_time" gorm:"column:create_time;autoCreateTime"`
	UpdateTime time.Time   `json:"-" gorm:"column:update_time;autoUpdateTime"`
}

func (Comment) TableName() string {
	return "comment"
}

// CommentDetail 评论详情，用于接口返回
// 组合了评论主体、作者信息、子评论树和当前用户的点赞状态
//
// 树形结构:
//   - Children: 递归嵌套的子评论，最多 10 层
//   - ChildCount: 直接子评论总数（用于前端判断是否显示"查看更多"）
//
// 已删除评论处理:
//   - status=2 的评论，Content 会被替换为 [已删除]
//   - 子评论保持不变，继续正常展示
type CommentDetail struct {
	*Comment
	AuthorName string           `json:"author_name"`         // 作者用户名
	Children   []*CommentDetail `json:"children,omitempty"`  // 子评论列表（递归嵌套）
	ChildCount int64            `json:"child_count"`         // 直接子评论总数
	IsLiked    bool             `json:"is_liked"`            // 当前用户是否已点赞（未登录为 false）
}

// ParamCreateComment 创建评论请求参数
type ParamCreateComment struct {
	PostID   string `json:"post_id" binding:"required"`             // 帖子ID
	ParentID string `json:"parent_id"`                              // 父评论ID（0或不传表示顶级评论）
	Content  string `json:"content" binding:"required,max=1024"`    // 评论内容，最长1024字符
}

// CommentListResponse 评论列表响应
type CommentListResponse struct {
	List  []*CommentDetail `json:"list"`  // 评论列表
	Total int64            `json:"total"` // 顶级评论总数
}

// CommentChildrenResponse 子评论列表响应
type CommentChildrenResponse struct {
	List    []*CommentDetail `json:"list"`     // 子评论列表
	Total   int64            `json:"total"`    // 子评论总数
	HasMore bool             `json:"has_more"` // 是否还有更多子评论
}
