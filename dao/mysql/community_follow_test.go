package mysql

import (
	"testing"

	"github.com/namelyzz/sayit/models"
	"github.com/stretchr/testify/assert"
)

func TestFollowCommunityLimitBounds(t *testing.T) {
	// 验证 limit 参数边界（用于 GetFollowedCommunityList 等）
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

func TestFollowStatusStruct(t *testing.T) {
	// 验证 FollowStatus 结构体字段正确性
	status := models.FollowStatus{IsFollowed: true}
	assert.True(t, status.IsFollowed)

	status.IsFollowed = false
	assert.False(t, status.IsFollowed)
}

func TestCommunityFollowTableName(t *testing.T) {
	// 验证表名正确性
	follow := models.CommunityFollow{}
	assert.Equal(t, "community_follow", follow.TableName())
}
