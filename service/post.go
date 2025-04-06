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

func CreatePost(ctx context.Context, p *models.Post) (err error) {
	p.PostID = genIDFunc()
	now := nowFunc()

	p.CreateTime = now
	event, err := newPostCreatedOutboxEvent(p)
	if err != nil {
		return err
	}

	eventID, err := createPostWithOutboxFunc(ctx, p, event)
	if err != nil {
		return err
	}
	event.ID = eventID

	if err = processOutboxEvent(ctx, event); err != nil {
		zap.L().Warn("sync post created outbox event failed, will retry",
			zap.Int64("eventID", event.ID),
			zap.Int64("postID", p.PostID),
			zap.Error(err))
	}

	return nil
}

func GetPostDetailByID(postID int64) (res *models.PostDetail, err error) {
	post, err := mysql.GetPostByID(postID)
	if err != nil {
		zap.L().Error("mysql.GetPostByID failed",
			zap.Int64("postID", postID),
			zap.Error(err))
		return nil, err
	}

	authorID := post.AuthorID
	user, err := mysql.GetUserByID(authorID)
	if err != nil {
		zap.L().Error("mysql.GetUserByID failed",
			zap.Int64("author_id", authorID),
			zap.Error(err))
		return nil, err
	}

	communityID := post.CommunityID
	detail, err := mysql.GetCommunityDetailByID(communityID)
	if err != nil {
		zap.L().Error("mysql.GetCommunityDetailByID failed",
			zap.Int64("community_id", communityID),
			zap.Error(err))
		return nil, err
	}

	return &models.PostDetail{
		AuthorName:      user.Username,
		Post:            post,
		CommunityDetail: detail,
	}, nil
}

func GetPostList(ctx context.Context, p *models.ParamPostList) (posts []*models.PostListItem, err error) {
	return ListPosts(ctx, p)
}

// ListPosts 获取帖子列表, 支持按分数排序和按创建时间排序
func ListPosts(ctx context.Context, p *models.ParamPostList) (posts []*models.PostListItem, err error) {
	switch p.SortBy {
	case models.SortFieldScore:
		return listPostsByScore(ctx, p)
	case models.SortFieldCreateTime:
		if canUseRedisTimeList(p) {
			offset := (p.Page - 1) * p.Size
			ids, err := redis.GetPostIDsInOrder(ctx, p, offset, p.Size)
			if err != nil {
				zap.L().Warn("redis time list failed, fallback to mysql",
					zap.Error(err),
					zap.Any("params", p))
				return mysql.GetPostList(p)
			}

			posts, err := mysql.FilterPostListByIDs(conv.Strings2Int64s(ids), p)
			if err != nil {
				return nil, err
			}
			if len(posts) == len(ids) || len(ids) < p.Size {
				return posts, nil
			}

			zap.L().Warn("redis time list underfilled after mysql filters, fallback to mysql",
				zap.Any("params", p))
		}
	}

	return mysql.GetPostList(p)
}

// listPostsByScore 按分数排序获取帖子列表
func listPostsByScore(ctx context.Context, p *models.ParamPostList) ([]*models.PostListItem, error) {
	// 复杂的分数查询需要 MySQL 过滤
	if hasComplexScoreFilters(p) {
		return listPostsByScoreWithFilters(ctx, p)
	}

	// 简单的分数查询可以从 Redis 时间列表中获取
	offset := (p.Page - 1) * p.Size
	ids, err := redis.GetPostIDsInOrder(ctx, p, offset, p.Size)
	if err != nil {
		zap.L().Warn("redis score list failed, fallback to mysql create_time desc",
			zap.Error(err),
			zap.Any("params", p))
		return fallbackPostListByCreateTime(p)
	}

	// 从 MySQL 中过滤帖子列表
	posts, err := mysql.FilterPostListByIDs(conv.Strings2Int64s(ids), p)
	if err != nil {
		return nil, err
	}
	// 如果 Redis 中的帖子列表足够，直接返回
	if len(posts) == len(ids) || len(ids) < p.Size {
		return posts, nil
	}

	// 数量不足，记录警告日志并使用带过滤的方法重试
	zap.L().Warn("score list underfilled after mysql filters, retry with redis scan",
		zap.Any("params", p))
	return listPostsByScoreWithFilters(ctx, p)
}

// listPostsByScoreWithFilters 带过滤条件的按分数排序获取帖子列表, 用于复杂的分数查询
func listPostsByScoreWithFilters(ctx context.Context, p *models.ParamPostList) ([]*models.PostListItem, error) {
	targetCount := p.Page * p.Size
	matched := make([]*models.PostListItem, 0, targetCount)

	// 批量获取帖子，直到达到目标数量
	for offset := 0; len(matched) < targetCount; offset += redisScanBatchSize {
		ids, err := redis.GetPostIDsInOrder(ctx, p, offset, redisScanBatchSize)
		if err != nil {
			zap.L().Warn("redis score scan failed, fallback to mysql create_time desc",
				zap.Error(err),
				zap.Any("params", p),
				zap.Int("offset", offset))
			return fallbackPostListByCreateTime(p)
		}
		if len(ids) == 0 {
			break
		}

		// 根据 ID 去 Mysql 批量查询
		items, err := mysql.FilterPostListByIDs(conv.Strings2Int64s(ids), p)
		if err != nil {
			return nil, err
		}
		matched = append(matched, items...)
		if len(ids) < redisScanBatchSize {
			break
		}
	}

	// 计算当前页的起始索引和结束索引, 然后返回当前页的帖子列表
	start := (p.Page - 1) * p.Size
	if start >= len(matched) {
		return []*models.PostListItem{}, nil
	}

	end := start + p.Size
	if end > len(matched) {
		end = len(matched)
	}

	return matched[start:end], nil
}

func canUseRedisTimeList(p *models.ParamPostList) bool {
	return p.SortBy == models.SortFieldCreateTime && p.UserName == "" && p.Keyword == ""
}

func hasComplexScoreFilters(p *models.ParamPostList) bool {
	return p.UserName != "" || p.Keyword != "" || p.StartTime != nil || p.EndTime != nil
}

// fallbackPostListByCreateTime 按创建时间排序获取帖子列表, 用于当 Redis 查询异常时 fallback 到 MySQL
func fallbackPostListByCreateTime(p *models.ParamPostList) ([]*models.PostListItem, error) {
	fallback := *p
	fallback.SortBy = models.SortFieldCreateTime
	fallback.Order = models.SortDirectionDesc
	return mysql.GetPostList(&fallback)
}
