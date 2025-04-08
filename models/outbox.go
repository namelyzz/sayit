package models

import "time"

// Outbox 事件类型常量
const EventTypePostCreated = "post_created" // 帖子创建事件

// OutboxEvent Outbox 事件模型，对应数据库 `outbox_events` 表
// Outbox 模式用于保证 MySQL 和 Redis 的最终一致性:
//   1. 创建帖子时，帖子数据和事件记录在同一个事务中写入 MySQL
//   2. 后台 Worker 定时扫描待处理事件，将其同步到 Redis
//   3. 同步成功后删除事件记录，失败则更新重试时间（指数退避）
type OutboxEvent struct {
	ID          int64     `gorm:"column:id;primaryKey"`                    // 自增主键
	EventType   string    `gorm:"column:event_type"`                       // 事件类型（如 post_created）
	AggregateID int64     `gorm:"column:aggregate_id"`                     // 关联的业务ID（如帖子ID）
	Payload     []byte    `gorm:"column:payload;type:json"`                // 事件数据（JSON格式）
	RetryCount  int       `gorm:"column:retry_count"`                      // 已重试次数
	NextRetryAt time.Time `gorm:"column:next_retry_at"`                    // 下次重试时间
	LastError   string    `gorm:"column:last_error"`                       // 最近一次错误信息
	CreateTime  time.Time `gorm:"column:create_time;autoCreateTime"`       // 创建时间
	UpdateTime  time.Time `gorm:"column:update_time;autoUpdateTime"`       // 更新时间
}

func (OutboxEvent) TableName() string {
	return "outbox_events"
}

// PostCreatedPayload 帖子创建事件的 payload 结构
// 包含写入 Redis 排行榜所需的最少信息
type PostCreatedPayload struct {
	PostID         int64 `json:"post_id"`      // 帖子ID
	CommunityID    int64 `json:"community_id"` // 社区ID（用于社区帖子集合）
	CreateTimeUnix int64 `json:"create_time"`  // 创建时间的Unix时间戳（作为Redis排行榜的初始分数）
}
