package redis

import (
	"context"
	"strconv"
)

func userFollowingKey(userID int64) string {
	return getRedisKey(KeyUserFollowingPF + strconv.FormatInt(userID, 10))
}

func userFollowersKey(userID int64) string {
	return getRedisKey(KeyUserFollowersPF + strconv.FormatInt(userID, 10))
}

// FollowUserCache 写入用户关注关系缓存。
func FollowUserCache(ctx context.Context, followerID, followingID int64) error {
	follower := strconv.FormatInt(followerID, 10)
	following := strconv.FormatInt(followingID, 10)
	pipe := client.TxPipeline()
	pipe.SAdd(ctx, userFollowingKey(followerID), following)
	pipe.SAdd(ctx, userFollowersKey(followingID), follower)
	_, err := pipe.Exec(ctx)
	return err
}

// UnfollowUserCache 删除用户关注关系缓存。
func UnfollowUserCache(ctx context.Context, followerID, followingID int64) error {
	follower := strconv.FormatInt(followerID, 10)
	following := strconv.FormatInt(followingID, 10)
	pipe := client.TxPipeline()
	pipe.SRem(ctx, userFollowingKey(followerID), following)
	pipe.SRem(ctx, userFollowersKey(followingID), follower)
	_, err := pipe.Exec(ctx)
	return err
}

// IsFollowingUserCache 检查用户关注关系缓存。
func IsFollowingUserCache(ctx context.Context, followerID, followingID int64) (bool, error) {
	return client.SIsMember(ctx, userFollowingKey(followerID), strconv.FormatInt(followingID, 10)).Result()
}
