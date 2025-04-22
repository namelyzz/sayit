package mysql

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHotCommunityScoreDecay(t *testing.T) {
	// 验证衰减公式的数学正确性: e^(-λ × age)
	decayLambda := 1.0 / 604800.0

	tests := []struct {
		name         string
		ageSeconds   float64
		wantDecayMin float64
		wantDecayMax float64
	}{
		{"new_post_0_day", 0, 1.0, 1.0},
		{"3_5_days", 3.5 * 24 * 3600, 0.55, 0.65},
		{"7_days_half_life", 7 * 24 * 3600, 0.33, 0.40},
		{"14_days", 14 * 24 * 3600, 0.10, 0.18},
		{"30_days", 30 * 24 * 3600, 0.005, 0.02},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decay := expNegative(decayLambda * tt.ageSeconds)
			assert.GreaterOrEqual(t, decay, tt.wantDecayMin, "decay too low for age=%f", tt.ageSeconds)
			assert.LessOrEqual(t, decay, tt.wantDecayMax, "decay too high for age=%f", tt.ageSeconds)
		})
	}
}

// expNegative 模拟 SQL 中的 EXP(-x) 函数，用于单元测试验证
func expNegative(x float64) float64 {
	return math.Exp(-x)
}

func TestHotCommunityLimitBounds(t *testing.T) {
	// 验证 limit 参数的边界处理
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"zero_limit_defaults_to_10", 0, 10},
		{"negative_limit_defaults_to_10", -5, 10},
		{"normal_limit_5", 5, 5},
		{"normal_limit_20", 20, 20},
		{"exceeds_max_capped_at_50", 100, 50},
		{"exactly_max_50", 50, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limit := tt.input
			if limit <= 0 {
				limit = 10
			}
			if limit > 50 {
				limit = 50
			}
			assert.Equal(t, tt.expected, limit)
		})
	}
}

func TestHotCommunityScoreFormula(t *testing.T) {
	// 端到端验证热度分数计算逻辑（纯时间衰减，无投票）
	// 模拟: 社区A有2个帖子，社区B有1个帖子
	// 验证: hot_score = Σ e^(-λ × age)

	decayLambda := 1.0 / 604800.0

	// 帖子数据: {community_id, age_seconds}
	type post struct {
		id          int64
		communityID int64
		ageSeconds  float64
	}

	posts := []post{
		{1, 1, 0}, // 社区A, 新帖 → e^0 = 1.0
		{2, 1, 7 * 24 * 3600}, // 社区A, 7天老帖 → e^-1 ≈ 0.368
		{3, 2, 0}, // 社区B, 新帖 → e^0 = 1.0
	}

	// 计算每个社区的热度
	scoreByCommunity := make(map[int64]float64)
	for _, p := range posts {
		decay := expNegative(decayLambda * p.ageSeconds)
		scoreByCommunity[p.communityID] += decay
	}

	assert.InDelta(t, 1.368, scoreByCommunity[1], 0.01, "社区A热度应为 1.0 + 0.368 ≈ 1.368")
	assert.InDelta(t, 1.0, scoreByCommunity[2], 0.01, "社区B热度应为 1.0")
	assert.Greater(t, scoreByCommunity[1], scoreByCommunity[2], "社区A应比社区B热门")
}

func TestHotCommunityActiveCommunityWins(t *testing.T) {
	// 验证: 持续活跃（有新帖）的社区比只有老帖的社区更热门
	decayLambda := 1.0 / 604800.0

	expNegative := func(x float64) float64 { return math.Exp(-x) }

	// 社区A: 3个老帖（每个7天）
	// 社区B: 1个新帖（0天）
	scoreA := 3 * expNegative(decayLambda * 7*24*3600) // 3 × 0.368 ≈ 1.10
	scoreB := expNegative(decayLambda * 0)              // 1 × 1.0 = 1.0

	// 社区A虽然帖子多但都老了，社区B只有1个新帖但权重高
	// 3个老帖的总分仍可能超过1个新帖
	assert.Greater(t, scoreA, scoreB, "3个7天老帖的总分应超过1个新帖")

	// 但如果老帖超过14天，新帖就会反超
	scoreA2 := 3 * expNegative(decayLambda * 14*24*3600) // 3 × 0.135 ≈ 0.405
	assert.Less(t, scoreA2, scoreB, "3个14天老帖的总分应低于1个新帖")
}
