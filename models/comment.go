package models

import "time"

// Comment 评论模型，对应数据库 `comment` 表
// 支持无限嵌套：parent_id 指向父评论，root_id 指向根评论（顶级评论的 root_id = 0）
type Comment struct {
	CommentID  SnowflakeID `json:"comment_id" gorm:"column:comment_id"`   // 雪花算法生成的全局唯一评论ID
	PostID     SnowflakeID `json:"post_id" gorm:"column:post_id"`         // 所属帖子ID
	AuthorID   SnowflakeID `json:"author_id" gorm:"column:author_id"`     // 作者用户ID
	ParentID   SnowflakeID `json:"parent_id" gorm:"column:parent_id"`     // 父评论ID（0表示顶级评论）
	RootID     SnowflakeID `json:"root_id" gorm:"column:root_id"`         // 根评论ID（0表示自己是根评论）
	Content    string      `json:"content" gorm:"column:content"`         // 评论内容
	LikeCount  int64       `json:"like_count" gorm:"column:like_count"`   // 点赞数（冗余字段）
	Status     int32       `json:"status" gorm:"column:status;default:1"` // 状态: 1=正常, 2=已删除
	CreateTime time.Time   `json:"create_time" gorm:"column:create_time;autoCreateTime"`
	UpdateTime time.Time   `json:"-" gorm:"column:update_time;autoUpdateTime"`
}

func (Comment) TableName() string {
	return "comment"
}

// CommentDetail 评论详情，用于接口返回
// 包含评论主体、作者名、子评论列表等
type CommentDetail struct {
	*Comment
	AuthorName string           `json:"author_name"`         // 作者用户名
	Children   []*CommentDetail `json:"children,omitempty"`  // 子评论列表
	ChildCount int64            `json:"child_count"`         // 子评论总数
	IsLiked    bool             `json:"is_liked"`            // 当前用户是否已点赞
}

// ParamCreateComment 创建评论请求参数
type ParamCreateComment struct {
	PostID   string `json:"post_id" binding:"required"`             // 帖子ID
	ParentID string `json:"parent_id"`                              // 父评论ID（0或不传表示顶级评论）
	Content  string `json:"content" binding:"required,max=1024"`    // 评论内容，最长1024字符
}
