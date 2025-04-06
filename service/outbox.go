package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/namelyzz/sayit/dao/mysql"
	"github.com/namelyzz/sayit/dao/redis"
	"github.com/namelyzz/sayit/models"
	"go.uber.org/zap"
)

const (
	outboxBatchSize      = 50
	outboxWorkerInterval = 3 * time.Second
	outboxMaxRetryDelay  = 60 * time.Second
)

var (
	createPostWithOutboxFunc = mysql.CreatePostWithOutbox
	getPendingOutboxFunc     = mysql.GetPendingOutboxEvents
	markOutboxRetryFunc      = mysql.MarkOutboxRetry
	deleteOutboxEventFunc    = mysql.DeleteOutboxEvent
	createPostRedisFunc      = redis.CreatePost
)

func StartOutboxWorker(ctx context.Context) {
	ticker := time.NewTicker(outboxWorkerInterval)
	defer ticker.Stop()

	for {
		if err := ConsumePendingOutboxEvents(ctx); err != nil {
			zap.L().Warn("consume pending outbox events failed", zap.Error(err))
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func ConsumePendingOutboxEvents(ctx context.Context) error {
	events, err := getPendingOutboxFunc(ctx, outboxBatchSize)
	if err != nil {
		return err
	}

	for _, event := range events {
		if err := processOutboxEvent(ctx, event); err != nil {
			zap.L().Warn("process outbox event failed",
				zap.Int64("eventID", event.ID),
				zap.String("eventType", event.EventType),
				zap.Error(err))
		}
	}
	return nil
}

func processOutboxEvent(ctx context.Context, event *models.OutboxEvent) error {
	switch event.EventType {
	case models.EventTypePostCreated:
		return processPostCreatedEvent(ctx, event)
	default:
		err := fmt.Errorf("unsupported outbox event type: %s", event.EventType)
		if retryErr := markOutboxEventRetry(ctx, event, err); retryErr != nil {
			return retryErr
		}
		return err
	}
}

func processPostCreatedEvent(ctx context.Context, event *models.OutboxEvent) error {
	var payload models.PostCreatedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		if retryErr := markOutboxEventRetry(ctx, event, err); retryErr != nil {
			return retryErr
		}
		return err
	}

	if err := createPostRedisFunc(ctx, payload.PostID, payload.CommunityID, float64(payload.CreateTimeUnix)); err != nil {
		if retryErr := markOutboxEventRetry(ctx, event, err); retryErr != nil {
			return retryErr
		}
		return err
	}

	return deleteOutboxEventFunc(ctx, event.ID)
}

func markOutboxEventRetry(ctx context.Context, event *models.OutboxEvent, cause error) error {
	nextRetryCount := event.RetryCount + 1
	nextRetryAt := nowFunc().Add(outboxRetryDelay(event.RetryCount))
	return markOutboxRetryFunc(ctx, event.ID, nextRetryCount, nextRetryAt, cause.Error())
}

func outboxRetryDelay(retryCount int) time.Duration {
	if retryCount < 0 {
		retryCount = 0
	}
	delay := time.Second
	for i := 0; i < retryCount && delay < outboxMaxRetryDelay; i++ {
		delay *= 2
	}
	if delay > outboxMaxRetryDelay {
		return outboxMaxRetryDelay
	}
	return delay
}

func newPostCreatedOutboxEvent(p *models.Post) (*models.OutboxEvent, error) {
	payload := models.PostCreatedPayload{
		PostID:         p.PostID,
		CommunityID:    p.CommunityID,
		CreateTimeUnix: p.CreateTime.Unix(),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &models.OutboxEvent{
		EventType:   models.EventTypePostCreated,
		AggregateID: p.PostID,
		Payload:     payloadBytes,
		NextRetryAt: nowFunc(),
	}, nil
}
