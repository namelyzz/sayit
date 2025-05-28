package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"strconv"
)

// CommentLikeComment 为评论点赞
// 使用 Redis Set 存储点赞用户，利用 Set 的去重特性实现幂等
// Redis Key: sayit:comment:liked:<commentID>，member=userID
//
// 返回值设计:
//   - added=true: 新点赞成功，service 层需调用 MySQL INCR like_count
//   - added=false: 已点赞过，service 层返回 ErrorLikeRepeated
func CommentLikeComment(ctx context.Context, commentID, userID int64) (bool, error) {
	key := getRedisKey(KeyCommentLikedPF + strconv.FormatInt(commentID, 10))
	added, err := client.SAdd(ctx, key, userID).Result()
	if err != nil {
		return false, err
	}
	return added == 1, nil
}

// CommentUnlikeComment 取消评论点赞
// 使用 Redis Set 的 SREM 操作，利用 Set 的去重特性实现幂等
// Redis Key: sayit:comment:liked:<commentID>，member=userID
//
// 返回值设计:
//   - removed=true: 取消成功，service 层需调用 MySQL DECR like_count
//   - removed=false: 未点赞过，service 层直接返回成功（幂等）
func CommentUnlikeComment(ctx context.Context, commentID, userID int64) (bool, error) {
	key := getRedisKey(KeyCommentLikedPF + strconv.FormatInt(commentID, 10))
	removed, err := client.SRem(ctx, key, userID).Result()
	if err != nil {
		return false, err
	}
	return removed == 1, nil
}

// IsCommentLikedByUser 检查用户是否已点赞某评论
// 使用场景: 单条评论的点赞状态查询
// Redis Key: sayit:comment:liked:<commentID>
// 注意: 未登录用户（userID=0）直接返回 false，不查询 Redis
func IsCommentLikedByUser(ctx context.Context, commentID, userID int64) bool {
	key := getRedisKey(KeyCommentLikedPF + strconv.FormatInt(commentID, 10))
	exists, _ := client.SIsMember(ctx, key, userID).Result()
	return exists
}

// BatchIsCommentLikedByUser 批量检查用户是否已点赞多条评论
// 使用场景: 评论树加载时，批量回填当前用户的点赞状态（is_liked 字段）
// 使用 Pipeline 批量执行 SISMEMBER，减少网络往返次数
// 返回 map[commentID]bool，方便按 commentID 查找
//
// 注意:
//   - commentIDs 为空或 userID=0 时直接返回 nil，避免无意义查询
//   - Pipeline 是非事务性的，但 SISMEMBER 是只读操作，不影响一致性
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
