package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"sort"
	"strconv"
)

// HotCommunityResult 热门社区结果（Redis 计算后返回）
type HotCommunityResult struct {
	ID          int64   `json:"community_id"`
	Name        string  `json:"community_name"`
	HotScore    float64 `json:"hot_score"`
	PostCount   int     `json:"post_count"`
}

// GetHotCommunities 从 Redis 计算热门社区列表
// 算法: 对每个社区，聚合其所有帖子在 sayit:post:score 中的分数并求和
// 利用已有的 Redis 数据结构，无需额外中间件
//
// Redis 数据结构:
//   sayit:post:score    ZSet - member=postID, score=帖子热度分数(时间戳+投票)
//   sayit:community:<id> Set  - member=postID, 该社区的所有帖子
//
// 流程:
//   1. 获取所有社区 ID 列表（由 caller 传入，避免 Redis 扫描）
//   2. 对每个社区，用 Pipeline 批量获取其所有帖子的热度分数
//   3. 求和得到社区热度，按分数降序排序
func GetHotCommunities(ctx context.Context, communityIDs []int64, limit int) ([]HotCommunityResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	scoreZsetKey := getRedisKey(KeyPostScoreZset)

	// 对每个社区，批量获取其帖子的分数
	type communityScore struct {
		id        int64
		score     float64
		postCount int
	}
	results := make([]communityScore, 0, len(communityIDs))

	for _, commID := range communityIDs {
		communityKey := getRedisKey(KeyCommunitySetPF + strconv.FormatInt(commID, 10))

		// 获取该社区的所有帖子 ID
		postIDs, err := client.SMembers(ctx, communityKey).Result()
		if err != nil || len(postIDs) == 0 {
			continue
		}

		// 批量获取这些帖子在热度排行榜中的分数
		pipe := client.Pipeline()
		getters := make([]*redis.FloatCmd, 0, len(postIDs))
		for _, pid := range postIDs {
			getters = append(getters, pipe.ZScore(ctx, scoreZsetKey, pid))
		}

		_, err = pipe.Exec(ctx)
		if err != nil {
			continue
		}

		// 求和
		var totalScore float64
		count := 0
		for _, g := range getters {
			s, err := g.Result()
			if err == nil && s > 0 {
				totalScore += s
				count++
			}
		}

		if count > 0 {
			results = append(results, communityScore{
				id:        commID,
				score:     totalScore,
				postCount: count,
			})
		}
	}

	// 按热度分数降序排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	// 截取 limit
	if len(results) > limit {
		results = results[:limit]
	}

	// 转换为返回结构
	communities := make([]HotCommunityResult, 0, len(results))
	for _, r := range results {
		communities = append(communities, HotCommunityResult{
			ID:        r.id,
			HotScore:  r.score,
			PostCount: r.postCount,
		})
	}

	return communities, nil
}
