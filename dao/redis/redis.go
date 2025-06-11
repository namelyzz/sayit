package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis_rate/v10"
	"github.com/namelyzz/sayit/config"
	"github.com/redis/go-redis/v9"
)

var (
	client  *redis.Client
	limiter *redis_rate.Limiter
)

// Redis Key 常量定义
// 所有 Key 统一使用 "sayit:" 前缀，方便管理和批量操作
const (
	Prefix             = "sayit:"          // 公共前缀
	KeyPostTimeZset    = "post:time"       // ZSet: 帖子发布时间排行榜，score=创建时间戳，member=帖子ID
	KeyPostScoreZset   = "post:score"      // ZSet: 帖子热度排行榜，score=时间戳+投票分，member=帖子ID
	KeyPostVotedZsetPF = "post:voted:"     // ZSet前缀: 用户投票记录，sayit:post:voted:<postID>，score=投票方向，member=用户ID
	KeyCommunitySetPF  = "community:"      // Set前缀: 社区帖子集合，sayit:community:<communityID>，member=帖子ID
	KeyUserFollowingPF = "user:following:" // Set前缀: 用户关注的人，sayit:user:following:<userID>，member=被关注用户ID
	KeyUserFollowersPF = "user:followers:" // Set前缀: 用户的粉丝，sayit:user:followers:<userID>，member=粉丝用户ID
	KeyCommentLikedPF       = "comment:liked:"       // Set前缀: 评论点赞用户集合，sayit:comment:liked:<commentID>，member=用户ID
	KeyCommentCountPF       = "comment:count:"       // String前缀: 帖子评论计数缓存，sayit:comment:count:<postID>，value=评论数
	KeyRateLimitCommentPF   = "ratelimit:comment:"   // Sorted Set前缀: 评论创建频率限制，sayit:ratelimit:comment:<userID>
	KeyNotificationStream   = "notification:stream"  // Stream: 通知事件队列，sayit:notification:stream
	KeyNotificationUnreadPF = "notification:unread:" // String前缀: 用户未读通知数，sayit:notification:unread:<userID>
	KeyNotificationCooldown = "notification:cooldown:" // String前缀: 通知冷却门禁，sayit:notification:cooldown:<type>:<actor>:<recipient>:<target>
)

func Init(cfg *config.RedisConfig) (err error) {
	client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	ctx := context.Background()
	if _, err = client.Ping(ctx).Result(); err != nil {
		return err
	}

	// 初始化频率限制器
	limiter = redis_rate.NewLimiter(client)

	return nil
}

func Close() {
	_ = client.Close()
}

func getRedisKey(key string) string {
	return Prefix + key
}
