package mysql

import (
	"context"
	"time"

	"github.com/namelyzz/sayit/models"
)

func CreatePostWithOutbox(ctx context.Context, p *models.Post, event *models.OutboxEvent) (int64, error) {
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}

	if err := tx.Omit("UpdateTime").Create(p).Error; err != nil {
		_ = tx.Rollback().Error
		return 0, err
	}

	if err := tx.Create(event).Error; err != nil {
		_ = tx.Rollback().Error
		return 0, err
	}

	if err := tx.Commit().Error; err != nil {
		return 0, err
	}

	return event.ID, nil
}

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

func DeleteOutboxEvent(ctx context.Context, id int64) error {
	return db.WithContext(ctx).Delete(&models.OutboxEvent{}, id).Error
}

func truncateOutboxError(msg string) string {
	const maxRunes = 512
	runes := []rune(msg)
	if len(runes) <= maxRunes {
		return msg
	}
	return string(runes[:maxRunes])
}
