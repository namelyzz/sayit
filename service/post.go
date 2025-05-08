package service

import (
	"context"
	"github.com/namelyzz/sayit/dao/mysql"
	"github.com/namelyzz/sayit/dao/redis"
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/utils/conv"
	"github.com/namelyzz/sayit/utils/snowflake"
	"go.uber.org/zap"
	"time"
)

const redisScanBatchSize = 100

var (
	nowFunc   = time.Now
	genIDFunc = snowflake.GenID
)

// CreatePost 创建帖子的业务逻辑
// 采用 Outbox 模式保证 MySQL 和 Redis 的最终一致性:
//   1. 在同一个事务中写入帖子记录和 outbox_events 记录
//   2. 尝试同步处理 outbox 事件（写入 Redis 排行榜）
//   3. 如果同步失败，后台 OutboxWorker 定时重试
//
// 优点: 即使 Redis 暂时不可用，帖子数据也不会丢失，会由后台任务补偿
func CreatePost(ctx context.Context, p *models.Post) (err error) {
	// 1. 通过雪花算法生成全局唯一帖子 ID
	p.PostID = models.SnowflakeID(genIDFunc())
	now := nowFunc()

	// 2. 设置帖子创建时间
	p.CreateTime = now

	// 3. 构造 Outbox 事件（包含写入 Redis 所需的 payload）
	event, err := newPostCreatedOutboxEvent(p)
	if err != nil {
		return err
	}

	// 4. 事务性写入: 在同一个 MySQL 事务中同时写入 post 表和 outbox_events 表
	// 保证帖子数据和事件记录的原子性
	eventID, err := createPostWithOutboxFunc(ctx, p, event)
	if err != nil {
		return err
	}
	event.ID = eventID

	// 5. 尝试同步处理 Outbox 事件（将帖子写入 Redis 排行榜）
	// 即使失败也不影响帖子创建，后台任务会重试
	if err = processOutboxEvent(ctx, event); err != nil {
		zap.L().Warn("sync post created outbox event failed, will retry",
			zap.Int64("eventID", event.ID),
			zap.Int64("postID", int64(p.PostID)),
			zap.Error(err))
	}

	return nil
}

// GetPostDetailByID 获取帖子详情
// 查询策略: 分三次 MySQL 查询后组装（非 JOIN），返回完整的帖子详情
// 包含: 帖子主体 + 作者用户名 + 社区详情
func GetPostDetailByID(postID int64) (res *models.PostDetail, err error) {
	// 1. 查询帖子主体信息
	post, err := mysql.GetPostByID(postID)
	if err != nil {
		zap.L().Error("mysql.GetPostByID failed",
			zap.Int64("postID", postID),
			zap.Error(err))
		return nil, err
	}

	// 2. 根据帖子的 author_id 查询作者信息
	authorID := post.AuthorID
	user, err := mysql.GetUserByID(authorID)
	if err != nil {
		zap.L().Error("mysql.GetUserByID failed",
			zap.Int64("author_id", authorID),
			zap.Error(err))
		return nil, err
	}

	// 3. 根据帖子的 community_id 查询社区详情
	communityID := post.CommunityID
	detail, err := mysql.GetCommunityDetailByID(communityID)
	if err != nil {
		zap.L().Error("mysql.GetCommunityDetailByID failed",
			zap.Int64("community_id", communityID),
			zap.Error(err))
		return nil, err
	}

	// 4. 组装帖子详情返回
	return &models.PostDetail{
		AuthorName:      user.Username,
		Post:            post,
		CommunityDetail: detail,
	}, nil
}

// GetPostList 获取帖子列表（入口函数，委托给 ListPosts）
func GetPostList(ctx context.Context, p *models.ParamPostList) (posts []*models.PostListItem, err error) {
	return ListPosts(ctx, p)
}

// ListPosts 获取帖子列表的核心调度函数
// 根据排序字段和过滤条件，自动选择最优的查询策略:
//
// ┌─────────────────────────────────────────────────────────────────────┐
// │ 排序方式        │ 过滤条件              │ 查询策略                  │
// ├─────────────────────────────────────────────────────────────────────┤
// │ score           │ 无复杂过滤            │ Redis 热度榜 → MySQL 详情 │
// │ score           │ 有复杂过滤            │ Redis 批量扫描 → MySQL    │
// │ create_time     │ 无作者名/关键词过滤   │ Redis 时间榜 → MySQL 详情 │
// │ create_time     │ 有作者名/关键词过滤   │ MySQL 联表查询            │
// │ update_time     │ 任意                  │ MySQL 联表查询            │
// └─────────────────────────────────────────────────────────────────────┘
//
// "复杂过滤"指: 作者名模糊搜索、关键词搜索、时间范围筛选
// "Redis 异常"时: 自动 fallback 到 MySQL 查询
func ListPosts(ctx context.Context, p *models.ParamPostList) (posts []*models.PostListItem, err error) {
	switch p.SortBy {
	case models.SortFieldScore:
		// 按热度排序: 优先从 Redis 热度榜获取帖子 ID
		return listPostsByScore(ctx, p)
	case models.SortFieldCreateTime:
		// 按创建时间排序: 如果没有作者名/关键词过滤，优先从 Redis 时间榜获取
		if canUseRedisTimeList(p) {
			offset := (p.Page - 1) * p.Size
			// 从 Redis 获取当前页的帖子 ID 列表
			ids, err := redis.GetPostIDsInOrder(ctx, p, offset, p.Size)
			if err != nil {
				// Redis 查询失败，fallback 到 MySQL
				zap.L().Warn("redis time list failed, fallback to mysql",
					zap.Error(err),
					zap.Any("params", p))
				return mysql.GetPostList(p)
			}

			// 用 Redis 返回的 ID 去 MySQL 查询完整帖子信息
			// FilterPostListByIDs 会保持 Redis 返回的排序顺序
			posts, err := mysql.FilterPostListByIDs(conv.Strings2Int64s(ids), p)
			if err != nil {
				return nil, err
			}
			// 如果 MySQL 返回的帖子数等于 Redis 返回的 ID 数，说明过滤后数量足够
			// 或者 Redis 返回的 ID 数不足一页，说明已经没有更多数据
			if len(posts) == len(ids) || len(ids) < p.Size {
				return posts, nil
			}

			// MySQL 过滤后数量不足（某些帖子被 status 等条件过滤掉了）
			// fallback 到纯 MySQL 查询
			zap.L().Warn("redis time list underfilled after mysql filters, fallback to mysql",
				zap.Any("params", p))
		}
	}

	// 默认: MySQL 联表查询（适用于 update_time 排序、有复杂过滤、或 Redis 异常）
	return mysql.GetPostList(p)
}

// listPostsByScore 按热度分数排序获取帖子列表
// 策略:
//   - 简单查询: Redis 取 ID → MySQL 查详情 → 返回
//   - 复杂查询: Redis 批量扫描 → MySQL 过滤 → 分页截取
//   - Redis 异常: fallback 到 MySQL 按创建时间排序
func listPostsByScore(ctx context.Context, p *models.ParamPostList) ([]*models.PostListItem, error) {
	// 有复杂过滤条件时，使用批量扫描策略
	if hasComplexScoreFilters(p) {
		return listPostsByScoreWithFilters(ctx, p)
	}

	// 简单查询: 直接从 Redis 热度榜获取当前页的帖子 ID
	offset := (p.Page - 1) * p.Size
	ids, err := redis.GetPostIDsInOrder(ctx, p, offset, p.Size)
	if err != nil {
		// Redis 失败，fallback 到 MySQL 按创建时间排序
		zap.L().Warn("redis score list failed, fallback to mysql create_time desc",
			zap.Error(err),
			zap.Any("params", p))
		return fallbackPostListByCreateTime(p)
	}

	// 用 Redis 返回的 ID 去 MySQL 查询完整帖子信息
	posts, err := mysql.FilterPostListByIDs(conv.Strings2Int64s(ids), p)
	if err != nil {
		return nil, err
	}
	// 数量足够或已无更多数据，直接返回
	if len(posts) == len(ids) || len(ids) < p.Size {
		return posts, nil
	}

	// MySQL 过滤后数量不足，使用批量扫描策略重试
	zap.L().Warn("score list underfilled after mysql filters, retry with redis scan",
		zap.Any("params", p))
	return listPostsByScoreWithFilters(ctx, p)
}

// listPostsByScoreWithFilters 带过滤条件的按分数排序获取帖子列表
// 使用场景: 有作者名、关键词、时间范围等复杂过滤条件
// 策略: 从 Redis 批量获取帖子 ID（每批100个），用 MySQL 过滤后累积，
//       直到收集够当前页所需的帖子数量，再进行分页截取
func listPostsByScoreWithFilters(ctx context.Context, p *models.ParamPostList) ([]*models.PostListItem, error) {
	// 计算目标数量: 需要收集 page*size 个帖子才能正确分页
	targetCount := p.Page * p.Size
	matched := make([]*models.PostListItem, 0, targetCount)

	// 批量从 Redis 获取帖子 ID，每批 redisScanBatchSize(100) 个
	for offset := 0; len(matched) < targetCount; offset += redisScanBatchSize {
		ids, err := redis.GetPostIDsInOrder(ctx, p, offset, redisScanBatchSize)
		if err != nil {
			// Redis 异常，fallback 到 MySQL
			zap.L().Warn("redis score scan failed, fallback to mysql create_time desc",
				zap.Error(err),
				zap.Any("params", p),
				zap.Int("offset", offset))
			return fallbackPostListByCreateTime(p)
		}
		if len(ids) == 0 {
			break // Redis 中没有更多数据
		}

		// 用这批 ID 去 MySQL 查询并过滤
		items, err := mysql.FilterPostListByIDs(conv.Strings2Int64s(ids), p)
		if err != nil {
			return nil, err
		}
		matched = append(matched, items...)
		// 如果这批 ID 数量不足 batchSize，说明 Redis 中没有更多数据
		if len(ids) < redisScanBatchSize {
			break
		}
	}

	// 从累积的匹配结果中截取当前页的数据
	start := (p.Page - 1) * p.Size
	if start >= len(matched) {
		return []*models.PostListItem{}, nil // 超出范围，返回空列表
	}

	end := start + p.Size
	if end > len(matched) {
		end = len(matched)
	}

	return matched[start:end], nil
}

// canUseRedisTimeList 判断是否可以使用 Redis 时间榜查询
// 条件: 按创建时间排序 + 无作者名过滤 + 无关键词过滤
// 原因: Redis 时间榜只存储帖子ID和时间戳，不支持作者名和关键词的模糊搜索
func canUseRedisTimeList(p *models.ParamPostList) bool {
	return p.SortBy == models.SortFieldCreateTime && p.AuthorID == 0 && p.UserName == "" && p.Keyword == ""
}

// hasComplexScoreFilters 判断是否有复杂的过滤条件
// 复杂过滤包括: 作者名、关键词、时间范围
// 这些条件无法在 Redis 中直接过滤，需要回查 MySQL
func hasComplexScoreFilters(p *models.ParamPostList) bool {
	return p.AuthorID != 0 || p.UserName != "" || p.Keyword != "" || p.StartTime != nil || p.EndTime != nil
}

// fallbackPostListByCreateTime Redis 查询异常时的降级方案
// 强制按创建时间倒序查询，忽略客户端指定的排序方式
func fallbackPostListByCreateTime(p *models.ParamPostList) ([]*models.PostListItem, error) {
	fallback := *p
	fallback.SortBy = models.SortFieldCreateTime
	fallback.Order = models.SortDirectionDesc
	return mysql.GetPostList(&fallback)
}
