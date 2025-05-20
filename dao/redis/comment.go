package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"strconv"
)

// CommentLikeComment 为评论点赞
// 使用 Redis Set 存储点赞用户，同时使用 INCRBY 更新帖子评论计数
func CommentLikeComment(ctx context.Context, commentID, userID int64) error {
	pipe := client.TxPipeline()
	pipe.SAdd(ctx, getRedisKey(KeyCommentLikedPF+strconv.FormatInt(commentID, 10)), userID)
	_, err := pipe.Exec(ctx)
	return err
}

// CommentUnlikeComment 取消评论点赞
func CommentUnlikeComment(ctx context.Context, commentID, userID int64) error {
	pipe := client.TxPipeline()
	pipe.SRem(ctx, getRedisKey(KeyCommentLikedPF+strconv.FormatInt(commentID, 10)), userID)
	_, err := pipe.Exec(ctx)
	return err
}

// IsCommentLikedByUser 检查用户是否已点赞某评论
func IsCommentLikedByUser(ctx context.Context, commentID, userID int64) bool {
	key := getRedisKey(KeyCommentLikedPF + strconv.FormatInt(commentID, 10))
	exists, _ := client.SIsMember(ctx, key, userID).Result()
	return exists
}

// BatchIsCommentLikedByUser 批量检查用户是否已点赞多条评论
func BatchIsCommentLikedByUser(ctx context.Context, commentIDs []int64, userID int64) map[int64]bool {
	if len(commentIDs) == 0 || userID == 0 {
		return nil
	}
	pipe := client.Pipeline()
	cmds := make(map[int64]*redis.BoolCmd, len(commentIDs))
	for _, cid := range commentIDs {
		key := getRedisKey(KeyCommentLikedPF + strconv.FormatInt(cid, 10))
		cmds[cid] = pipe.SIsMember(ctx, key, userID)
	}
	_, _ = pipe.Exec(ctx)
	result := make(map[int64]bool, len(commentIDs))
	for cid, cmd := range cmds {
		result[cid], _ = cmd.Result()
	}
	return result
}
