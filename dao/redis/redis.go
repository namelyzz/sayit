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

const (
	Prefix             = "sayit:"
	KeyPostTimeZset    = "post:time"
	KeyPostScoreZset   = "post:score"
	KeyPostVotedZsetPF = "post:voted:"
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
