package mysql

import (
	"testing"

	"github.com/namelyzz/sayit/models"
	"github.com/stretchr/testify/assert"
)

func TestReorderPostListItems_ReordersByRequestedIDs(t *testing.T) {
	postIDs := []int64{3, 1, 2}
	items := []*models.PostListItem{
		{PostID: 1},
		{PostID: 2},
		{PostID: 3},
	}

	ordered := reorderPostListItems(postIDs, items)

	assert.Len(t, ordered, 3)
	assert.Equal(t, int64(3), ordered[0].PostID)
	assert.Equal(t, int64(1), ordered[1].PostID)
	assert.Equal(t, int64(2), ordered[2].PostID)
}

func TestReorderPostListItems_SkipsMissingIDs(t *testing.T) {
	postIDs := []int64{4, 2, 1}
	items := []*models.PostListItem{
		{PostID: 1},
		{PostID: 2},
	}

	ordered := reorderPostListItems(postIDs, items)

	assert.Len(t, ordered, 2)
	assert.Equal(t, int64(2), ordered[0].PostID)
	assert.Equal(t, int64(1), ordered[1].PostID)
}

func TestReorderPostListItems_IgnoresUnrequestedItems(t *testing.T) {
	postIDs := []int64{2, 1}
	items := []*models.PostListItem{
		{PostID: 1},
		{PostID: 2},
		{PostID: 99},
	}

	ordered := reorderPostListItems(postIDs, items)

	assert.Len(t, ordered, 2)
	assert.Equal(t, int64(2), ordered[0].PostID)
	assert.Equal(t, int64(1), ordered[1].PostID)
}

func TestReorderPostListItems_EmptyInput(t *testing.T) {
	tests := []struct {
		name    string
		postIDs []int64
		items   []*models.PostListItem
	}{
		{
			name:    "empty_post_ids",
			postIDs: []int64{},
			items:   []*models.PostListItem{{PostID: 1}},
		},
		{
			name:    "empty_items",
			postIDs: []int64{1, 2},
			items:   []*models.PostListItem{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ordered := reorderPostListItems(tt.postIDs, tt.items)

			assert.Empty(t, ordered)
		})
	}
}
