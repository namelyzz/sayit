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

// Outbox 模式配置常量
const (
	outboxBatchSize      = 50               // 每次批量处理的事件数量
	outboxWorkerInterval = 10 * time.Second // Worker 轮询间隔
	outboxMaxRetryDelay  = 60 * time.Second // 最大重试延迟（指数退避上限）
)

// 以下变量可被测试替换（mock），方便单元测试
var (
	createPostWithOutboxFunc = mysql.CreatePostWithOutbox   // 事务性写入帖子+outbox
	getPendingOutboxFunc     = mysql.GetPendingOutboxEvents // 获取待处理事件
	markOutboxRetryFunc      = mysql.MarkOutboxRetry        // 标记事件重试
	deleteOutboxEventFunc    = mysql.DeleteOutboxEvent      // 删除已处理事件
	createPostRedisFunc      = redis.CreatePost             // 写入Redis排行榜
)

// StartOutboxWorker 启动 Outbox 后台补偿 worker
// 在 main.go 中通过 go service.StartOutboxWorker(context.Background()) 启动
// 功能: 定时轮询 outbox_events 表，处理失败的事件（重试写入 Redis）
// 退出条件: ctx 被取消（程序关闭时）
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

// ConsumePendingOutboxEvents 批量消费待处理的 Outbox 事件
// 查询条件: next_retry_at <= now（到了重试时间的事件）
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

// processOutboxEvent 根据事件类型分发处理
// 当前支持的事件类型: post_created
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

// processPostCreatedEvent 处理帖子创建事件
// 步骤: 反序列化 payload -> 调用 redis.CreatePost 写入排行榜 -> 删除已处理事件
func processPostCreatedEvent(ctx context.Context, event *models.OutboxEvent) error {
	// 1. 从事件 payload 中解析帖子信息
	var payload models.PostCreatedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		if retryErr := markOutboxEventRetry(ctx, event, err); retryErr != nil {
			return retryErr
		}
		return err
	}

	// 2. 将帖子写入 Redis 排行榜（时间榜 + 热度榜 + 社区集合）
	if err := createPostRedisFunc(ctx, payload.PostID, payload.CommunityID, float64(payload.CreateTimeUnix)); err != nil {
		if retryErr := markOutboxEventRetry(ctx, event, err); retryErr != nil {
			return retryErr
		}
		return err
	}

	// 3. 处理成功，删除 outbox 事件记录
	return deleteOutboxEventFunc(ctx, event.ID)
}

// markOutboxEventRetry 标记事件为重试状态
// 更新重试次数、下次重试时间和错误信息
func markOutboxEventRetry(ctx context.Context, event *models.OutboxEvent, cause error) error {
	nextRetryCount := event.RetryCount + 1
	nextRetryAt := nowFunc().Add(outboxRetryDelay(event.RetryCount))
	return markOutboxRetryFunc(ctx, event.ID, nextRetryCount, nextRetryAt, cause.Error())
}

// outboxRetryDelay 计算指数退避延迟时间
// 策略: 1s -> 2s -> 4s -> 8s -> 16s -> 32s -> 60s(上限)
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

// newPostCreatedOutboxEvent 构造帖子创建的 Outbox 事件
// payload 包含写入 Redis 所需的信息: postID、communityID、createTimeUnix
func newPostCreatedOutboxEvent(p *models.Post) (*models.OutboxEvent, error) {
	payload := models.PostCreatedPayload{
		PostID:         int64(p.PostID),
		CommunityID:    int64(p.CommunityID),
		CreateTimeUnix: p.CreateTime.Unix(),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &models.OutboxEvent{
		EventType:   models.EventTypePostCreated,
		AggregateID: int64(p.PostID),
		Payload:     payloadBytes,
		NextRetryAt: nowFunc(), // 立即可重试（用于同步处理尝试）
	}, nil
}
