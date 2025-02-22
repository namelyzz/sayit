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
