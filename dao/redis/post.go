package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

func CreatePost(ctx context.Context, postID, communityID int64) error {
	pipe := client.TxPipeline()
	now := time.Now().Unix()

	// 按时间入榜：将帖子加入“最新发布”排行榜
	pipe.ZAdd(ctx, getRedisKey(KeyPostTimeZset), redis.Z{
		Score:  float64(now),
		Member: postID,
	})

	// 按分数入榜：将帖子加入“综合热度”排行榜（初始分数设为当前时间戳）。
	// 为什么初始分数是时间戳？ 最终分数 = 基础时间分 + 投票加权分
	// 1.为了保证新帖子有曝光机会：
	// 刚发布的帖子没有投票，如果分数为 0，它会沉底。将分数初始化为当前时间戳，能保证新发布的帖子暂时排在前面（比旧帖子分数高）。
	// 2.随着时间推移，旧帖子的“时间分”虽然小，但如果它的“投票分”很高，总分就会超过这个新帖子。
	pipe.ZAdd(ctx, getRedisKey(KeyPostScoreZset), redis.Z{
		Score:  float64(now),
		Member: postID,
	})

	// 社区关联：将帖子 ID 记录到对应社区的集合中。
	// 用于快速查找某个社区内的帖子列表
	cKey := getRedisKey(KeyCommunitySetPF + strconv.Itoa(int(communityID)))
	pipe.SAdd(ctx, cKey, postID)
	_, err := pipe.Exec(ctx)
	return err
}
