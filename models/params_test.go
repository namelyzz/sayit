package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParamPostListValidateAndSetDefaults_Defaults(t *testing.T) {
	p := &ParamPostList{}

	err := p.ValidateAndSetDefaults()

	require.NoError(t, err)
	assert.Equal(t, SortFieldCreateTime, p.SortBy)
	assert.Equal(t, SortDirectionDesc, p.Order)
	assert.Equal(t, 1, p.Page)
	assert.Equal(t, MaxPageSize, p.Size)
	require.NotNil(t, p.Status)
	assert.Equal(t, 1, *p.Status)
}

func TestParamPostListValidateAndSetDefaults_TrimFields(t *testing.T) {
	p := &ParamPostList{
		UserName: "  alice  ",
		Keyword:  "  hello world  ",
	}

	err := p.ValidateAndSetDefaults()

	require.NoError(t, err)
	assert.Equal(t, "alice", p.UserName)
	assert.Equal(t, "hello world", p.Keyword)
}

func TestParamPostListValidateAndSetDefaults_InvalidTimeRange(t *testing.T) {
	startTime := int64(200)
	endTime := int64(100)
	p := &ParamPostList{
		StartTime: &startTime,
		EndTime:   &endTime,
	}

	err := p.ValidateAndSetDefaults()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "start_time cannot be greater than end_time")
}

func TestParamPostListValidateAndSetDefaults_InvalidSortBy(t *testing.T) {
	p := &ParamPostList{
		SortBy: SortField("hot"),
	}

	err := p.ValidateAndSetDefaults()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sort_by")
}

func TestParamPostListValidateAndSetDefaults_InvalidOrder(t *testing.T) {
	p := &ParamPostList{
		Order: SortDirection("up"),
	}

	err := p.ValidateAndSetDefaults()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid order")
}

func TestParamPostListValidateAndSetDefaults_PreserveExplicitValues(t *testing.T) {
	status := 2
	p := &ParamPostList{
		Page:   3,
		Size:   20,
		Status: &status,
		SortBy: SortFieldScore,
		Order:  SortDirectionAsc,
	}

	err := p.ValidateAndSetDefaults()

	require.NoError(t, err)
	assert.Equal(t, 3, p.Page)
	assert.Equal(t, 20, p.Size)
	assert.Equal(t, SortFieldScore, p.SortBy)
	assert.Equal(t, SortDirectionAsc, p.Order)
	require.NotNil(t, p.Status)
	assert.Equal(t, 2, *p.Status)
}

func TestParamPostListValidateAndSetDefaults_SizeNormalization(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{name: "non_positive_size", size: 0},
		{name: "too_large_size", size: MaxPageSize + 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ParamPostList{Size: tt.size}

			err := p.ValidateAndSetDefaults()

			require.NoError(t, err)
			assert.Equal(t, MaxPageSize, p.Size)
		})
	}
}
