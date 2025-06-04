package service

import (
	"context"
	"github.com/namelyzz/sayit/dao/mysql"
	"github.com/namelyzz/sayit/dao/redis"
	"go.uber.org/zap"
	"sync"
	"time"
)

// 对账任务参数
const (
	reconcileBatchSize    = 1000 // 每批查询的评论数量
	reconcileConcurrency  = 5    // 并发处理的 goroutine 数量
)

// mock 变量定义，方便单元测试替换
var (
	getNormalCommentIDsFunc  = mysql.GetNormalCommentIDs       // 分批获取评论ID
	getCommentLikeCountsFunc = mysql.GetCommentLikeCounts      // 批量获取 MySQL like_count
	batchUpdateLikeCountFunc = mysql.BatchUpdateLikeCount      // 批量更新 like_count
	batchGetCommentLikeCountFunc = redis.BatchGetCommentLikeCount // 批量获取 Redis SCARD
)

// reconcileJob 对账批次任务
type reconcileJob struct {
	commentIDs []int64
}

// reconcileResult 对账批次结果
type reconcileResult struct {
	fixed int
	err   error
}

// StartLikeCountReconciler 启动点赞数对账定时任务
// 在 main.go 中以 goroutine 方式启动: go service.StartLikeCountReconciler(ctx, time.Hour)
// 使用 context 控制生命周期，服务关闭时自动退出
func StartLikeCountReconciler(ctx context.Context, interval time.Duration) {
	zap.L().Info("like count reconciler started",
		zap.Duration("interval", interval))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 启动后立即执行一次对账
	runReconciliation(ctx)

	for {
		select {
		case <-ctx.Done():
			zap.L().Info("like count reconciler stopped")
			return
		case <-ticker.C:
			runReconciliation(ctx)
		}
	}
}

// runReconciliation 执行一次对账任务
func runReconciliation(ctx context.Context) {
	start := time.Now()
	fixed, err := ReconcileCommentLikeCount(ctx)
	duration := time.Since(start)

	if err != nil {
		zap.L().Error("like count reconciliation failed",
			zap.Duration("duration", duration),
			zap.Error(err))
	} else if fixed > 0 {
		zap.L().Info("like count reconciliation completed",
			zap.Int("fixed", fixed),
			zap.Duration("duration", duration))
	} else {
		zap.L().Debug("like count reconciliation completed, no drift found",
			zap.Duration("duration", duration))
	}
}

// ReconcileCommentLikeCount 对账修复评论点赞数
//
// 算法:
//  1. 使用游标分批获取所有正常状态的评论 ID
//  2. 每批并发查询 Redis SCARD 和 MySQL like_count
//  3. 对比找出不一致的记录
//  4. 批量更新 MySQL like_count 为 Redis 的真实值
//
// 返回值: 修复的评论数量和错误
func ReconcileCommentLikeCount(ctx context.Context) (int, error) {
	var lastID int64
	var totalFixed int

	for {
		// 1. 游标分批获取评论 ID
		ids, err := getNormalCommentIDsFunc(lastID, reconcileBatchSize)
		if err != nil {
			return totalFixed, err
		}
		if len(ids) == 0 {
			break
		}

		// 2. 处理当前批次
		fixed, err := reconcileBatch(ctx, ids)
		if err != nil {
			return totalFixed, err
		}
		totalFixed += fixed

		// 3. 更新游标
		lastID = ids[len(ids)-1]

		// 4. 如果不足一批，说明已处理完所有数据
		if len(ids) < reconcileBatchSize {
			break
		}
	}

	return totalFixed, nil
}

// reconcileBatch 处理单批对账任务
// 将评论 ID 分成多个子批次，并发处理
func reconcileBatch(ctx context.Context, ids []int64) (int, error) {
	// 将大批次拆分为多个子批次
	subBatches := splitIntoSubBatches(ids, reconcileBatchSize/reconcileConcurrency+1)

	// 并发处理子批次
	jobs := make(chan reconcileJob, len(subBatches))
	results := make(chan reconcileResult, len(subBatches))

	// 启动 worker goroutine
	var wg sync.WaitGroup
	for i := 0; i < reconcileConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				fixed, err := reconcileSubBatch(ctx, job.commentIDs)
				results <- reconcileResult{fixed: fixed, err: err}
			}
		}()
	}

	// 分发任务
	for _, batch := range subBatches {
		jobs <- reconcileJob{commentIDs: batch}
	}
	close(jobs)

	// 等待所有 worker 完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	var totalFixed int
	for res := range results {
		if res.err != nil {
			return totalFixed, res.err
		}
		totalFixed += res.fixed
	}

	return totalFixed, nil
}

// reconcileSubBatch 处理单个子批次
func reconcileSubBatch(ctx context.Context, ids []int64) (int, error) {
	// 1. 批量获取 Redis SCARD（真实点赞数）
	redisCounts := batchGetCommentLikeCountFunc(ctx, ids)

	// 2. 批量获取 MySQL like_count
	mysqlCounts, err := getCommentLikeCountsFunc(ids)
	if err != nil {
		return 0, err
	}

	// 3. 对比找出不一致的
	toFix := make(map[int64]int64)
	for _, id := range ids {
		rc := redisCounts[id]  // Redis 中的真实值（Key 不存在时为 0）
		mc := mysqlCounts[id]  // MySQL 中的 like_count
		if rc != mc {
			toFix[id] = rc
		}
	}

	// 4. 批量修正
	if len(toFix) > 0 {
		if err := batchUpdateLikeCountFunc(toFix); err != nil {
			return 0, err
		}
	}

	return len(toFix), nil
}

// splitIntoSubBatches 将切片拆分为多个子批次
func splitIntoSubBatches(ids []int64, batchSize int) [][]int64 {
	if len(ids) == 0 {
		return nil
	}

	var batches [][]int64
	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		batches = append(batches, ids[i:end])
	}
	return batches
}
