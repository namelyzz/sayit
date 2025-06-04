package redis

import (
	"context"
	"github.com/go-redis/redis_rate/v10"
	"strconv"
	"time"
)

// 评论创建频率限制参数：严格每 10 秒最多 1 条评论，不允许突发
var commentRateLimit = redis_rate.Limit{
	Rate:   1,
	Burst:  1,
	Period: 10 * time.Second,
}

// CheckCommentRateLimit 检查用户评论创建频率限制
// 使用 go-redis/redis_rate 库实现 GCRA（Generic Cell Rate Algorithm）
//
// 返回值:
//   - allowed: true=允许创建，false=超出频率限制
//   - retryAfter: 超限时需要等待的时间（用于提示用户）
//   - err: Redis 错误，调用方应放行请求（容错）
func CheckCommentRateLimit(ctx context.Context, userID int64) (bool, time.Duration, error) {
	key := getRedisKey(KeyRateLimitCommentPF + strconv.FormatInt(userID, 10))
	res, err := limiter.Allow(ctx, key, commentRateLimit)
	if err != nil {
		// Redis 故障时放行，不影响正常用户
		return true, 0, err
	}
	if res.Allowed == 0 {
		// 超限，返回需要等待的时间
		return false, res.RetryAfter, nil
	}
	return true, 0, nil
}
