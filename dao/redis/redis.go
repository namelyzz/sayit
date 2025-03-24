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
	Prefix             = "sayit:"      // 公共前缀
	KeyPostTimeZset    = "post:time"   // zset;帖子及其发帖时间
	KeyPostScoreZset   = "post:score"  // zset;帖子及其投票的分数
	KeyPostVotedZsetPF = "post:voted:" // zset;记录用户及其投票类型
	KeyCommunitySetPF  = "community:"  // set;保存每个分区下帖子的id
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
