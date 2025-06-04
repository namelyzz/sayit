package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func restoreReconcileTestHooks() func() {
	origGetNormalCommentIDs := getNormalCommentIDsFunc
	origGetCommentLikeCounts := getCommentLikeCountsFunc
	origBatchUpdateLikeCount := batchUpdateLikeCountFunc
	origBatchGetCommentLikeCount := batchGetCommentLikeCountFunc

	return func() {
		getNormalCommentIDsFunc = origGetNormalCommentIDs
		getCommentLikeCountsFunc = origGetCommentLikeCounts
		batchUpdateLikeCountFunc = origBatchUpdateLikeCount
		batchGetCommentLikeCountFunc = origBatchGetCommentLikeCount
	}
}

func TestReconcileCommentLikeCount_NoDrift(t *testing.T) {
	defer restoreReconcileTestHooks()()

	getNormalCommentIDsFunc = func(lastID int64, batchSize int) ([]int64, error) {
		return []int64{1001, 1002, 1003}, nil
	}
	batchGetCommentLikeCountFunc = func(ctx context.Context, commentIDs []int64) map[int64]int64 {
		return map[int64]int64{1001: 5, 1002: 3, 1003: 0}
	}
	getCommentLikeCountsFunc = func(commentIDs []int64) (map[int64]int64, error) {
		return map[int64]int64{1001: 5, 1002: 3, 1003: 0}, nil
	}
	batchUpdateLikeCountFunc = func(countMap map[int64]int64) error {
		t.Fatal("should not call update when no drift")
		return nil
	}

	fixed, err := ReconcileCommentLikeCount(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 0, fixed)
}

func TestReconcileCommentLikeCount_WithDrift(t *testing.T) {
	defer restoreReconcileTestHooks()()

	getNormalCommentIDsFunc = func(lastID int64, batchSize int) ([]int64, error) {
		return []int64{1001, 1002, 1003}, nil
	}
	batchGetCommentLikeCountFunc = func(ctx context.Context, commentIDs []int64) map[int64]int64 {
		return map[int64]int64{1001: 5, 1002: 3, 1003: 0}
	}
	getCommentLikeCountsFunc = func(commentIDs []int64) (map[int64]int64, error) {
		return map[int64]int64{1001: 5, 1002: 5, 1003: 2}, nil // 1002 and 1003 have drift
	}
	var updatedMap map[int64]int64
	batchUpdateLikeCountFunc = func(countMap map[int64]int64) error {
		updatedMap = countMap
		return nil
	}

	fixed, err := ReconcileCommentLikeCount(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 2, fixed)
	assert.Equal(t, map[int64]int64{1002: 3, 1003: 0}, updatedMap)
}

func TestReconcileCommentLikeCount_EmptyComments(t *testing.T) {
	defer restoreReconcileTestHooks()()

	getNormalCommentIDsFunc = func(lastID int64, batchSize int) ([]int64, error) {
		return []int64{}, nil
	}

	fixed, err := ReconcileCommentLikeCount(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 0, fixed)
}

func TestReconcileCommentLikeCount_GetIDsFails(t *testing.T) {
	defer restoreReconcileTestHooks()()

	getNormalCommentIDsFunc = func(lastID int64, batchSize int) ([]int64, error) {
		return nil, errors.New("db error")
	}

	_, err := ReconcileCommentLikeCount(context.Background())

	require.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

func TestReconcileCommentLikeCount_GetLikeCountsFails(t *testing.T) {
	defer restoreReconcileTestHooks()()

	getNormalCommentIDsFunc = func(lastID int64, batchSize int) ([]int64, error) {
		return []int64{1001}, nil
	}
	batchGetCommentLikeCountFunc = func(ctx context.Context, commentIDs []int64) map[int64]int64 {
		return map[int64]int64{1001: 5}
	}
	getCommentLikeCountsFunc = func(commentIDs []int64) (map[int64]int64, error) {
		return nil, errors.New("db error")
	}

	_, err := ReconcileCommentLikeCount(context.Background())

	require.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

func TestReconcileCommentLikeCount_UpdateFails(t *testing.T) {
	defer restoreReconcileTestHooks()()

	getNormalCommentIDsFunc = func(lastID int64, batchSize int) ([]int64, error) {
		return []int64{1001}, nil
	}
	batchGetCommentLikeCountFunc = func(ctx context.Context, commentIDs []int64) map[int64]int64 {
		return map[int64]int64{1001: 5}
	}
	getCommentLikeCountsFunc = func(commentIDs []int64) (map[int64]int64, error) {
		return map[int64]int64{1001: 3}, nil // drift
	}
	batchUpdateLikeCountFunc = func(countMap map[int64]int64) error {
		return errors.New("update error")
	}

	_, err := ReconcileCommentLikeCount(context.Background())

	require.Error(t, err)
	assert.Equal(t, "update error", err.Error())
}

func TestReconcileCommentLikeCount_MultipleBatches(t *testing.T) {
	defer restoreReconcileTestHooks()()

	callCount := 0
	getNormalCommentIDsFunc = func(lastID int64, batchSize int) ([]int64, error) {
		callCount++
		switch callCount {
		case 1:
			assert.Equal(t, int64(0), lastID)
			return []int64{1001, 1002}, nil
		case 2:
			assert.Equal(t, int64(1002), lastID)
			return []int64{1003}, nil
		default:
			return []int64{}, nil
		}
	}
	batchGetCommentLikeCountFunc = func(ctx context.Context, commentIDs []int64) map[int64]int64 {
		return map[int64]int64{1001: 5, 1002: 3, 1003: 0}
	}
	getCommentLikeCountsFunc = func(commentIDs []int64) (map[int64]int64, error) {
		return map[int64]int64{1001: 5, 1002: 5, 1003: 0}, nil // 1002 has drift
	}
	var updatedMap map[int64]int64
	batchUpdateLikeCountFunc = func(countMap map[int64]int64) error {
		updatedMap = countMap
		return nil
	}

	fixed, err := ReconcileCommentLikeCount(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 1, fixed)
	assert.Equal(t, map[int64]int64{1002: 3}, updatedMap)
}

func TestReconcileCommentLikeCount_RedisKeyMissing(t *testing.T) {
	defer restoreReconcileTestHooks()()

	getNormalCommentIDsFunc = func(lastID int64, batchSize int) ([]int64, error) {
		return []int64{1001, 1002}, nil
	}
	batchGetCommentLikeCountFunc = func(ctx context.Context, commentIDs []int64) map[int64]int64 {
		// 1001 has no Redis key (not in map), 1002 has 3 likes
		return map[int64]int64{1002: 3}
	}
	getCommentLikeCountsFunc = func(commentIDs []int64) (map[int64]int64, error) {
		return map[int64]int64{1001: 5, 1002: 3}, nil // 1001 has drift (MySQL=5, Redis=0)
	}
	var updatedMap map[int64]int64
	batchUpdateLikeCountFunc = func(countMap map[int64]int64) error {
		updatedMap = countMap
		return nil
	}

	fixed, err := ReconcileCommentLikeCount(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 1, fixed)
	assert.Equal(t, map[int64]int64{1001: 0}, updatedMap) // Redis key missing → fix to 0
}

func TestSplitIntoSubBatches(t *testing.T) {
	tests := []struct {
		name      string
		ids       []int64
		batchSize int
		expected  [][]int64
	}{
		{
			name:      "empty",
			ids:       []int64{},
			batchSize: 3,
			expected:  nil,
		},
		{
			name:      "exact batch",
			ids:       []int64{1, 2, 3},
			batchSize: 3,
			expected:  [][]int64{{1, 2, 3}},
		},
		{
			name:      "multiple batches",
			ids:       []int64{1, 2, 3, 4, 5},
			batchSize: 2,
			expected:  [][]int64{{1, 2}, {3, 4}, {5}},
		},
		{
			name:      "single element",
			ids:       []int64{1},
			batchSize: 5,
			expected:  [][]int64{{1}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitIntoSubBatches(tt.ids, tt.batchSize)
			assert.Equal(t, tt.expected, result)
		})
	}
}
