package service

import (
	"context"
	"github.com/namelyzz/sayit/dao/mysql"
	"github.com/namelyzz/sayit/dao/redis"
	"github.com/namelyzz/sayit/models"
)

func GetCommunityList() ([]*models.Community, error) {
	return mysql.GetCommunityList()
}

func GetCommunityDetailByID(id int64) (*models.CommunityDetail, error) {
	return mysql.GetCommunityDetailByID(id)
}

func GetRandomCommunityList(limit int) ([]*models.Community, error) {
	return mysql.GetRandomCommunityList(limit)
}

// GetHotCommunityList 从 Redis 计算热门社区列表
// 流程: 1) MySQL 获取所有社区ID和名称  2) Redis 聚合每个社区的热度分数  3) 合并返回
func GetHotCommunityList(limit int) ([]*models.HotCommunity, error) {
	// Step 1: 从 MySQL 获取所有社区基本信息（轻量查询）
	communities, err := mysql.GetCommunityList()
	if err != nil {
		return nil, err
	}
	if len(communities) == 0 {
		return []*models.HotCommunity{}, nil
	}

	// 提取社区 ID 和建立 ID→名称映射
	commIDs := make([]int64, 0, len(communities))
	nameByID := make(map[int64]string, len(communities))
	for _, c := range communities {
		commIDs = append(commIDs, c.ID.Int64())
		nameByID[c.ID.Int64()] = c.Name
	}

	// Step 2: 从 Redis 计算每个社区的热度分数
	redisResults, err := redis.GetHotCommunities(context.Background(), commIDs, limit)
	if err != nil {
		return nil, err
	}

	// Step 3: 合并 MySQL 的名称和 Redis 的分数
	results := make([]*models.HotCommunity, 0, len(redisResults))
	for _, r := range redisResults {
		results = append(results, &models.HotCommunity{
			ID:       models.SnowflakeID(r.ID),
			Name:     nameByID[r.ID],
			HotScore: r.HotScore,
			PostCount: int64(r.PostCount),
		})
	}

	return results, nil
}
