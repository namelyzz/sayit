package mysql

import (
	"github.com/namelyzz/sayit/models"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"
)

// FollowUser 关注用户，重复关注按成功处理。
func FollowUser(followerID, followingID int64) error {
	follow := &models.UserFollow{
		FollowerID:  models.SnowflakeID(followerID),
		FollowingID: models.SnowflakeID(followingID),
	}
	res := db.Clauses(clause.OnConflict{DoNothing: true}).Create(follow)
	if res.Error != nil {
		zap.L().Error("follow user failed",
			zap.Int64("follower_id", followerID),
			zap.Int64("following_id", followingID),
			zap.Error(res.Error))
		return res.Error
	}
	return nil
}

// UnfollowUser 取消关注用户，关系不存在时也按成功处理。
func UnfollowUser(followerID, followingID int64) error {
	res := db.Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Delete(&models.UserFollow{})
	if res.Error != nil {
		zap.L().Error("unfollow user failed",
			zap.Int64("follower_id", followerID),
			zap.Int64("following_id", followingID),
			zap.Error(res.Error))
		return res.Error
	}
	return nil
}

// IsFollowingUser 检查 followerID 是否关注 followingID。
func IsFollowingUser(followerID, followingID int64) (bool, error) {
	var count int64
	res := db.Model(&models.UserFollow{}).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Count(&count)
	if res.Error != nil {
		zap.L().Error("check following user failed",
			zap.Int64("follower_id", followerID),
			zap.Int64("following_id", followingID),
			zap.Error(res.Error))
		return false, res.Error
	}
	return count > 0, nil
}

// CountFollowers 统计关注 userID 的用户数。
func CountFollowers(userID int64) (int64, error) {
	var count int64
	res := db.Model(&models.UserFollow{}).
		Where("following_id = ?", userID).
		Count(&count)
	if res.Error != nil {
		zap.L().Error("count user followers failed", zap.Int64("user_id", userID), zap.Error(res.Error))
		return 0, res.Error
	}
	return count, nil
}

// CountFollowing 统计 userID 关注的用户数。
func CountFollowing(userID int64) (int64, error) {
	var count int64
	res := db.Model(&models.UserFollow{}).
		Where("follower_id = ?", userID).
		Count(&count)
	if res.Error != nil {
		zap.L().Error("count user following failed", zap.Int64("user_id", userID), zap.Error(res.Error))
		return 0, res.Error
	}
	return count, nil
}
