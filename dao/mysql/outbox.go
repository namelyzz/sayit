package mysql

import (
	"context"
	"time"

	"github.com/namelyzz/sayit/models"
)

// CreatePostWithOutbox 在同一个事务中同时写入帖子和 Outbox 事件
// 这是 Outbox 模式的核心: 保证帖子数据和事件记录的原子性
// 事务流程:
//   1. 开启事务
//   2. 插入 post 记录
//   3. 插入 outbox_events 记录
//   4. 提交事务（任一步骤失败则回滚）
//
// 返回值: outbox 事件的自增 ID（用于后续处理和重试）
func CreatePostWithOutbox(ctx context.Context, p *models.Post, event *models.OutboxEvent) (int64, error) {
	// 1. 开启事务
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}

	// 2. 插入帖子记录（Omit UpdateTime 让数据库使用默认值）
	if err := tx.Omit("UpdateTime").Create(p).Error; err != nil {
		_ = tx.Rollback().Error
		return 0, err
	}

	// 3. 插入 Outbox 事件记录
	if err := tx.Create(event).Error; err != nil {
		_ = tx.Rollback().Error
		return 0, err
	}

	// 4. 提交事务
	if err := tx.Commit().Error; err != nil {
		return 0, err
	}

	return event.ID, nil
}

// GetPendingOutboxEvents 查询待处理的 Outbox 事件
// 查询条件: next_retry_at <= now（到了重试时间的事件）
// 按 ID 升序排列，保证先进先出
func GetPendingOutboxEvents(ctx context.Context, limit int) ([]*models.OutboxEvent, error) {
	if limit <= 0 {
		return []*models.OutboxEvent{}, nil
	}

	var events []*models.OutboxEvent
	err := db.WithContext(ctx).
		Where("next_retry_at <= ?", time.Now()).
		Order("id ASC").
		Limit(limit).
		Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
}

// MarkOutboxRetry 标记 Outbox 事件为重试状态
// 更新重试次数、下次重试时间和错误信息
// 错误信息会被截断到 512 字符，防止超长错误撑爆数据库字段
func MarkOutboxRetry(ctx context.Context, id int64, retryCount int, nextRetryAt time.Time, errMsg string) error {
	return db.WithContext(ctx).
		Model(&models.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"retry_count":   retryCount,
			"next_retry_at": nextRetryAt,
			"last_error":    truncateOutboxError(errMsg),
		}).Error
}

// DeleteOutboxEvent 删除已成功处理的 Outbox 事件
func DeleteOutboxEvent(ctx context.Context, id int64) error {
	return db.WithContext(ctx).Delete(&models.OutboxEvent{}, id).Error
}

// truncateOutboxError 截断错误消息到最大长度（512 字符）
func truncateOutboxError(msg string) string {
	const maxRunes = 512
	runes := []rune(msg)
	if len(runes) <= maxRunes {
		return msg
	}
	return string(runes[:maxRunes])
}
