package models

import "time"

const EventTypePostCreated = "post_created"

type OutboxEvent struct {
	ID          int64     `gorm:"column:id;primaryKey"`
	EventType   string    `gorm:"column:event_type"`
	AggregateID int64     `gorm:"column:aggregate_id"`
	Payload     []byte    `gorm:"column:payload;type:json"`
	RetryCount  int       `gorm:"column:retry_count"`
	NextRetryAt time.Time `gorm:"column:next_retry_at"`
	LastError   string    `gorm:"column:last_error"`
	CreateTime  time.Time `gorm:"column:create_time;autoCreateTime"`
	UpdateTime  time.Time `gorm:"column:update_time;autoUpdateTime"`
}

func (OutboxEvent) TableName() string {
	return "outbox_events"
}

type PostCreatedPayload struct {
	PostID         int64 `json:"post_id"`
	CommunityID    int64 `json:"community_id"`
	CreateTimeUnix int64 `json:"create_time"`
}
