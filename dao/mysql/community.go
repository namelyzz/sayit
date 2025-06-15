package mysql

import (
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/utils/api"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func GetCommunityList() (communities []*models.Community, err error) {
	res := db.Model(&models.Community{}).
		Select("community_id", "community_name").
		Find(&communities)

	if res.Error != nil {
		zap.L().Error("get community list failed", zap.Error(res.Error))
		return nil, res.Error
	}

	if len(communities) == 0 {
		zap.L().Warn("there is no community in db")
	}

	return communities, nil
}

func GetCommunityDetailByID(id int64) (detail *models.CommunityDetail, err error) {
	detail = new(models.CommunityDetail)
	res := db.Model(&models.CommunityDetail{}).
		Select("community_id", "community_name", "introduction", "create_time").
		Where("community_id = ?", id).
		First(detail)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, api.ErrorInvalidID
		}

		zap.L().Error("get community detail failed", zap.Error(res.Error))
		return nil, res.Error
	}

	return detail, nil
}

// GetRandomCommunityList 随机获取指定数量的社区
// 使用 ORDER BY RAND() 实现随机抽样，适用于社区数量中等的场景
func GetRandomCommunityList(limit int) ([]*models.Community, error) {
	if limit <= 0 {
		limit = 5
	}
	if limit > 50 {
		limit = 50
	}

	communities := make([]*models.Community, 0, limit)
	res := db.Model(&models.Community{}).
		Select("community_id", "community_name").
		Order("RAND()").
		Limit(limit).
		Find(&communities)

	if res.Error != nil {
		zap.L().Error("get random community list failed", zap.Error(res.Error))
		return nil, res.Error
	}

	return communities, nil
}

// GetHotCommunityList 获取热门社区列表
//
// 热度算法: 加权时间衰减 (Weighted Time Decay)，纯 MySQL 实现，不依赖 Redis
//
// 公式:
//   hot_score(community) = Σ e^(-λ × age_seconds)
//
// 其中:
//   - age_seconds: 帖子创建至今的秒数 = UNIX_TIMESTAMP(NOW()) - UNIX_TIMESTAMP(create_time)
//   - λ (decay constant): 衰减系数 = 1 / (7 × 24 × 3600) = 1/604800
//
// 七天半衰期说明:
//   - 半衰期 T½ = 7天 = 604800秒，意味着一个帖子在7天后，其热度衰减到原来的 50%
//   - 衰减公式: decay_factor = e^(-λ × age)
//   - 推导: 为了让 e^(-λ × T½) = 0.5，取 λ = ln(2) / T½ ≈ 0.693 / 604800
//           为简化计算，取 λ = 1/T½ = 1/604800 ≈ 1.653e-6
//           此时 e^(-1) ≈ 0.368，即7天后衰减到 36.8%
//   - 实际效果:
//       0天:  e^(-0)            = 1.00  (100% 权重)
//       3.5天: e^(-0.5)          = 0.61  (61% 权重)
//       7天:  e^(-1.0)           = 0.37  (37% 权重)
//       14天: e^(-2.0)           = 0.14  (14% 权重)
//       30天: e^(-4.3)           = 0.01  (1% 权重，基本沉底)
//
// 设计意图:
//   - 新帖子即使刚发布，因为衰减因子接近1，可排在前面
//   - 老帖子如果没有新帖子，热度会自然衰减
//   - 活跃社区（持续有新帖）会保持高热度
//   - 纯 SQL 实现，无需 Redis 等中间件
//   - 注: 投票数据存储在 Redis 中，MySQL 无 vote 表，本算法暂不纳入投票因素
func GetHotCommunityList(limit int) (communities []*models.HotCommunity, err error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	// SQL 逻辑:
	// 1. 对每个帖子计算衰减得分 = e^(-λ × age_seconds)
	// 2. GROUP BY community，求和得到社区热度
	// 3. 同时统计帖子数量
	const query = `
		SELECT
			c.community_id,
			c.community_name,
			SUM(EXP(- ? * (UNIX_TIMESTAMP(NOW()) - UNIX_TIMESTAMP(p.create_time)))) AS hot_score,
			COUNT(p.post_id) AS post_count
		FROM community c
		INNER JOIN post p ON p.community_id = c.community_id AND p.status = 1
		GROUP BY c.community_id, c.community_name
		ORDER BY hot_score DESC
		LIMIT ?
	`

	decayLambda := 1.0 / 604800.0 // λ = 1/604800，对应7天半衰期

	res := db.Raw(query, decayLambda, limit).Scan(&communities)
	if res.Error != nil {
		zap.L().Error("get hot community list failed", zap.Error(res.Error))
		return nil, res.Error
	}

	if len(communities) == 0 {
		zap.L().Warn("there is no hot community")
	}

	return communities, nil
}

// SearchCommunitiesByName 根据社区名搜索社区（右模糊，走索引）
func SearchCommunitiesByName(keyword string, limit int) ([]*models.SearchSuggestItem, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 20 {
		limit = 20
	}

	var communities []*models.SearchSuggestItem
	res := db.Model(&models.Community{}).
		Select("community_id AS id, community_name AS name").
		Where("community_name LIKE ?", keyword+"%").
		Order("community_name ASC").
		Limit(limit).
		Find(&communities)

	if res.Error != nil {
		zap.L().Error("search communities by name failed", zap.Error(res.Error))
		return nil, res.Error
	}

	return communities, nil
}
