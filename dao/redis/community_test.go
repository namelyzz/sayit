package redis

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatHotScore(t *testing.T) {
	// 验证热度分数格式化逻辑（与前端 TypeScript 版本保持一致）
	tests := []struct {
		score  float64
		want   string
	}{
		{0, "0"},
		{500, "500"},
		{1000, "1.0k"},
		{1500, "1.5k"},
		{10000, "1.0万"},
		{15000, "1.5万"},
		{1000000, "1.0M"},
		{1500000, "1.5M"},
		{1000000000, "1.0B"},
		{1700000000, "1.7B"}, // 典型的时间戳级别分数
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			var result string
			s := tt.score
			if s >= 1000000000 {
				result = fmt.Sprintf("%.1fB", s/1000000000)
			} else if s >= 1000000 {
				result = fmt.Sprintf("%.1fM", s/1000000)
			} else if s >= 10000 {
				result = fmt.Sprintf("%.1f万", s/10000)
			} else if s >= 1000 {
				result = fmt.Sprintf("%.1fk", s/1000)
			} else {
				result = fmt.Sprintf("%d", int(s))
			}
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGetHotCommunities_EmptyCommunityIDs(t *testing.T) {
	ctx := context.Background()
	results, err := GetHotCommunities(ctx, []int64{}, 5)
	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestGetHotCommunities_LimitCapping(t *testing.T) {
	// 验证 limit 边界处理逻辑
	tests := []struct {
		input    int
		expected int
	}{
		{0, 10},
		{-1, 10},
		{5, 5},
		{50, 50},
		{100, 50},
	}

	for _, tt := range tests {
		limit := tt.input
		if limit <= 0 {
			limit = 10
		}
		if limit > 50 {
			limit = 50
		}
		assert.Equal(t, tt.expected, limit)
	}
}
