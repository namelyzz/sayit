package mysql

import (
	"github.com/namelyzz/sayit/models"
	"go.uber.org/zap"
)

// FollowCommunity 关注社区
func FollowCommunity(userID, communityID int64) error {
	follow := &models.CommunityFollow{
		UserID:      models.SnowflakeID(userID),
		CommunityID: models.SnowflakeID(communityID),
	}
	res := db.Create(follow)
	if res.Error != nil {
		zap.L().Error("follow community failed",
			zap.Int64("user_id", userID),
			zap.Int64("community_id", communityID),
			zap.Error(res.Error))
		return res.Error
	}
	return nil
}

// UnfollowCommunity 取消关注社区
func UnfollowCommunity(userID, communityID int64) error {
	res := db.Where("user_id = ? AND community_id = ?", userID, communityID).
		Delete(&models.CommunityFollow{})
	if res.Error != nil {
		zap.L().Error("unfollow community failed",
			zap.Int64("user_id", userID),
			zap.Int64("community_id", communityID),
			zap.Error(res.Error))
		return res.Error
	}
	return nil
}

// IsFollowingCommunity 检查用户是否已关注某社区
func IsFollowingCommunity(userID, communityID int64) (bool, error) {
	var count int64
	res := db.Model(&models.CommunityFollow{}).
		Where("user_id = ? AND community_id = ?", userID, communityID).
		Count(&count)
	if res.Error != nil {
		zap.L().Error("check following community failed",
			zap.Int64("user_id", userID),
			zap.Int64("community_id", communityID),
			zap.Error(res.Error))
		return false, res.Error
	}
	return count > 0, nil
}

// GetFollowedCommunityList 获取用户已关注的社区列表
func GetFollowedCommunityList(userID int64) ([]*models.Community, error) {
	var communities []*models.Community
	res := db.Model(&models.Community{}).
		Select("community.community_id", "community.community_name").
		Joins("JOIN community_follow cf ON cf.community_id = community.community_id").
		Where("cf.user_id = ?", userID).
		Find(&communities)
	if res.Error != nil {
		zap.L().Error("get followed community list failed",
			zap.Int64("user_id", userID),
			zap.Error(res.Error))
		return nil, res.Error
	}
	return communities, nil
}
