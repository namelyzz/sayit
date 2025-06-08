package mysql

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestNormalizeNotificationPagination(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		size     int
		wantPage int
		wantSize int
	}{
		{"default_page_and_size", 0, 0, 1, 50},
		{"negative_page", -1, 20, 1, 20},
		{"normal", 2, 30, 2, 30},
		{"cap_size", 1, 100, 1, 50},
		{"negative_size", 1, -10, 1, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, size := normalizeNotificationPagination(tt.page, tt.size)
			assert.Equal(t, tt.wantPage, page)
			assert.Equal(t, tt.wantSize, size)
		})
	}
}

func TestNormalizeNotificationStatus(t *testing.T) {
	assert.Equal(t, notificationStatusUnread, NormalizeNotificationStatus("unread"))
	assert.Equal(t, notificationStatusAll, NormalizeNotificationStatus(""))
	assert.Equal(t, notificationStatusAll, NormalizeNotificationStatus("read"))
	assert.Equal(t, notificationStatusAll, NormalizeNotificationStatus("invalid"))
}

func TestIsDuplicateNotificationError(t *testing.T) {
	assert.True(t, IsDuplicateNotificationError(gorm.ErrDuplicatedKey))
	assert.True(t, IsDuplicateNotificationError(errors.New("Error 1062: Duplicate entry 'x' for key 'uk_dedupe_key'")))
	assert.False(t, IsDuplicateNotificationError(errors.New("connection refused")))
	assert.False(t, IsDuplicateNotificationError(nil))
}
