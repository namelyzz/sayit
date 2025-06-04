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

// IncrCommentCount 帖子评论数 +1（缓存）
// 调用时机: 评论创建成功后
// Redis Key: sayit:comment:count:<postID>，String 类型
// 如果 Key 不存在，INCR 会自动创建并设为 1
func IncrCommentCount(ctx context.Context, postID int64) error {
	key := getRedisKey(KeyCommentCountPF + strconv.FormatInt(postID, 10))
	return client.Incr(ctx, key).Err()
}

// DecrCommentCount 帖子评论数 -1（缓存）
// 调用时机: 评论删除成功后
// 使用 Lua 脚本确保不会减成负数
func DecrCommentCount(ctx context.Context, postID int64) error {
	key := getRedisKey(KeyCommentCountPF + strconv.FormatInt(postID, 10))
	// Lua 脚本: 只有当前值 > 0 时才减 1
	script := redis.NewScript(`
		local val = tonumber(redis.call('GET', KEYS[1]))
		if val == nil then return nil end
		if val > 0 then return redis.call('DECR', KEYS[1]) end
		return val
	`)
	_, err := script.Run(ctx, client, []string{key}).Result()
	// key 不存在时返回 redis.Nil，属于正常情况
	if err == redis.Nil {
		return nil
	}
	return err
}

// GetCommentCount 获取帖子评论数缓存
// 返回值: 评论数和 error
// 缓存未命中时返回 (0, redis.Nil)，service 层据此判断是否需要回填
func GetCommentCount(ctx context.Context, postID int64) (int64, error) {
	key := getRedisKey(KeyCommentCountPF + strconv.FormatInt(postID, 10))
	val, err := client.Get(ctx, key).Int64()
	if err != nil {
		return 0, err
	}
	return val, nil
}

// SetCommentCount 设置帖子评论数缓存（回填用）
// 调用时机: 缓存未命中时，从 MySQL 查询后回填
func SetCommentCount(ctx context.Context, postID, count int64) error {
	key := getRedisKey(KeyCommentCountPF + strconv.FormatInt(postID, 10))
	return client.Set(ctx, key, count, 0).Err() // 不设 TTL
}

// BatchGetCommentCount 批量获取帖子评论数缓存
// 使用 Pipeline 批量执行 GET，减少网络往返次数
// 返回 map[postID]count，仅包含缓存命中的条目
func BatchGetCommentCount(ctx context.Context, postIDs []int64) map[int64]int64 {
	if len(postIDs) == 0 {
		return nil
	}
	pipe := client.Pipeline()
	cmds := make(map[int64]*redis.StringCmd, len(postIDs))
	for _, pid := range postIDs {
		key := getRedisKey(KeyCommentCountPF + strconv.FormatInt(pid, 10))
		cmds[pid] = pipe.Get(ctx, key)
	}
	_, _ = pipe.Exec(ctx)
	result := make(map[int64]int64, len(postIDs))
	for pid, cmd := range cmds {
		val, err := cmd.Int64()
		if err == nil {
			result[pid] = val
		}
	}
	return result
}

// BatchSetCommentCount 批量设置帖子评论数缓存（批量回填）
// 使用 Pipeline 批量执行 SET，减少网络往返次数
func BatchSetCommentCount(ctx context.Context, countMap map[int64]int64) error {
	if len(countMap) == 0 {
		return nil
	}
	pipe := client.Pipeline()
	for postID, count := range countMap {
		key := getRedisKey(KeyCommentCountPF + strconv.FormatInt(postID, 10))
		pipe.Set(ctx, key, count, 0)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// BatchGetCommentLikeCount 批量获取评论的真实点赞数（Redis SCARD）
// 使用场景: 对账任务，批量获取 Redis 中的真实点赞数
// 使用 Pipeline 批量执行 SCARD，减少网络往返次数
// 返回 map[commentID]count，Key 不存在时 count 为 0
func BatchGetCommentLikeCount(ctx context.Context, commentIDs []int64) map[int64]int64 {
	if len(commentIDs) == 0 {
		return nil
	}
	pipe := client.Pipeline()
	cmds := make(map[int64]*redis.IntCmd, len(commentIDs))
	for _, cid := range commentIDs {
		key := getRedisKey(KeyCommentLikedPF + strconv.FormatInt(cid, 10))
		cmds[cid] = pipe.SCard(ctx, key)
	}
	_, _ = pipe.Exec(ctx)
	result := make(map[int64]int64, len(commentIDs))
	for cid, cmd := range cmds {
		val, err := cmd.Result()
		if err == nil {
			result[cid] = val
		}
		// Key 不存在时 SCARD 返回 0，无需特殊处理
	}
	return result
}
