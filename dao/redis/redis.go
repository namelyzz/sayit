package redis

import (
	"context"
	"fmt"
	"github.com/namelyzz/sayit/config"
	"github.com/redis/go-redis/v9"
)

var (
	client *redis.Client
)

// Redis Key 常量定义
// 所有 Key 统一使用 "sayit:" 前缀，方便管理和批量操作
const (
	Prefix             = "sayit:"      // 公共前缀
	KeyPostTimeZset    = "post:time"   // ZSet: 帖子发布时间排行榜，score=创建时间戳，member=帖子ID
	KeyPostScoreZset   = "post:score"  // ZSet: 帖子热度排行榜，score=时间戳+投票分，member=帖子ID
	KeyPostVotedZsetPF = "post:voted:" // ZSet前缀: 用户投票记录，sayit:post:voted:<postID>，score=投票方向，member=用户ID
	KeyCommunitySetPF  = "community:"  // Set前缀: 社区帖子集合，sayit:community:<communityID>，member=帖子ID
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

	return nil
}

func Close() {
	_ = client.Close()
}

func getRedisKey(key string) string {
	return Prefix + key
}
