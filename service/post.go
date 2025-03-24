package service

import (
	"context"
	"github.com/namelyzz/sayit/dao/mysql"
	"github.com/namelyzz/sayit/dao/redis"
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/utils/snowflake"
	"go.uber.org/zap"
)

/*
CreatePost

这里讨论一个问题：MySQL 中的 CreateTime 和 Redis 中时间排行榜的 Score (时间戳)是否会有差别
是的，会有细微的差别，但这通常不是问题，在绝大多数工程实践中是可以接受的。
简而言之，差别在于 MySQL 依赖数据库的默认行为，而 Redis 的逻辑是在 Go 语言的运行时通过 time.Now().Unix() 获取的
差别有多大？ 通常是 几毫秒到几十毫秒（取决于网络延迟和代码执行速度）。

这个差别有影响吗？几乎没有。
Redis 里的分数是为了计算“热度”。热度算法通常是 Score = 初始时间 + 投票加权。
算法本身就是一种近似模型。对于热度排序来说，帖子 A 是 12:00:00.005 发布的，还是 12:00:00.050 发布的，
根本不影响它在排行榜上的位置。 只要大体顺序对即可。
另一种用途是前端展示，一般都是分钟级，严苛点可能是秒级，用户不可能肉眼分辨出那几十毫秒的误差。

唯一潜在的极端边缘情况：
如果有两个帖子在极短的时间内（比如 1ms 间隔）连续发布，可能会出现：
Redis 里帖子 A 分数比帖子 B 高（排前面），
但在数据库按 CreateTime 排序时，帖子 B 比 帖子 A 晚。
导致“最新列表”的顺序在两个数据源中微调。但对于社区类应用，这完全不是 Bug。
【附修改方案】如果你追求完美的数据一致性，在进入数据库和 Redis 之前，先定格时间，将这个时间传给 dao 层的 redis/mysql 逻辑去写入
*/
func CreatePost(ctx context.Context, p *models.Post) (err error) {
	// 使用雪花算法为帖子生成一个 ID
	p.PostID = snowflake.GenID()
	err = mysql.CreatePost(p)
	if err != nil {
		return err
	}

	err = redis.CreatePost(ctx, p.PostID, p.CommunityID)
	return err
}

func GetPostDetailByID(postID int64) (res *models.PostDetail, err error) {
	post, err := mysql.GetPostByID(postID)
	if err != nil {
		zap.L().Error("mysql.GetPostByID failed",
			zap.Int64("postID", postID),
			zap.Error(err))
		return nil, err
	}

	authorID := post.AuthorID
	user, err := mysql.GetUserByID(authorID)
	if err != nil {
		zap.L().Error("mysql.GetUserByID failed",
			zap.Int64("author_id", authorID),
			zap.Error(err))
		return nil, err
	}

	communityID := post.CommunityID
	detail, err := mysql.GetCommunityDetailByID(communityID)
	if err != nil {
		zap.L().Error("mysql.GetCommunityDetailByID failed",
			zap.Int64("community_id", communityID),
			zap.Error(err))
		return nil, err
	}

	return &models.PostDetail{
		AuthorName:      user.Username,
		Post:            post,
		CommunityDetail: detail,
	}, nil
}

func GetPostList(p *models.ParamPostList) (posts []*models.PostListItem, err error) {
	return mysql.GetPostList(p)
}
