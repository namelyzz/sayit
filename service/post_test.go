package service

import (
	"testing"

	"github.com/namelyzz/sayit/models"
	"github.com/stretchr/testify/assert"
)

func TestCanUseRedisTimeList(t *testing.T) {
	tests := []struct {
		name string
		p    *models.ParamPostList
		want bool
	}{
		{
			name: "create_time_without_keyword_or_username",
			p: &models.ParamPostList{
				SortBy: models.SortFieldCreateTime,
			},
			want: true,
		},
		{
			name: "score_sort_cannot_use",
			p: &models.ParamPostList{
				SortBy: models.SortFieldScore,
			},
			want: false,
		},
		{
			name: "update_time_sort_cannot_use",
			p: &models.ParamPostList{
				SortBy: models.SortFieldUpdateTime,
			},
			want: false,
		},
		{
			name: "username_filter_disables_redis_time_list",
			p: &models.ParamPostList{
				SortBy:   models.SortFieldCreateTime,
				UserName: "alice",
			},
			want: false,
		},
		{
			name: "keyword_filter_disables_redis_time_list",
			p: &models.ParamPostList{
				SortBy:  models.SortFieldCreateTime,
				Keyword: "redis",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, canUseRedisTimeList(tt.p))
		})
	}
}

func TestHasComplexScoreFilters(t *testing.T) {
	status := 1
	startTime := int64(100)
	endTime := int64(200)

	tests := []struct {
		name string
		p    *models.ParamPostList
		want bool
	}{
		{
			name: "username_is_complex_filter",
			p: &models.ParamPostList{
				UserName: "alice",
			},
			want: true,
		},
		{
			name: "keyword_is_complex_filter",
			p: &models.ParamPostList{
				Keyword: "redis",
			},
			want: true,
		},
		{
			name: "start_time_is_complex_filter",
			p: &models.ParamPostList{
				StartTime: &startTime,
			},
			want: true,
		},
		{
			name: "end_time_is_complex_filter",
			p: &models.ParamPostList{
				EndTime: &endTime,
			},
			want: true,
		},
		{
			name: "basic_paging_and_status_are_not_complex_filters",
			p: &models.ParamPostList{
				CommunityID: 10,
				Page:        2,
				Size:        20,
				Status:      &status,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, hasComplexScoreFilters(tt.p))
		})
	}
}
