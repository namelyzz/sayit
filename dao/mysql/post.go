package mysql

import (
	"github.com/namelyzz/sayit/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func CreatePost(p *models.Post) (err error) {
	res := db.Omit("CreateTime", "UpdateTime").Create(p)
	if res.Error != nil {
		zap.L().Error("create post failed",
			zap.String("operation", "create_post"),
			zap.Int64("author_id", p.AuthorID),
			zap.Int64("community_id", p.CommunityID),
			zap.Error(res.Error))
		return res.Error
	}
	return nil
}

func GetPostByID(postID int64) (post *models.Post, err error) {
	post = new(models.Post)
	res := db.Model(&models.Post{}).
		Select("post_id", "title", "content", "author_id", "community_id", "status", "create_time", "update_time").
		Where("post_id = ?", postID).First(post)

	if res.Error != nil {
		return nil, res.Error
	}
	return post, nil
}

const (
	// PostSummaryLength 帖子摘要长度（字符数）
	PostSummaryLength = 30
	// PostSummarySuffix 摘要后缀
	PostSummarySuffix = "..."
)

func GetPostList(p *models.ParamPostList) (posts []*models.PostListItem, err error) {
	query := db.Model(&models.PostListItem{}).
		Select(`post_id, title, author_id, community_id, status, 
                create_time, update_time,
                CASE 
                    WHEN LENGTH(content) > ? THEN CONCAT(SUBSTRING(content, 1, ?), ?)
                    ELSE content
                END as summary`,
			PostSummaryLength,
			PostSummaryLength,
			PostSummarySuffix)

	if p.CommunityID != 0 {
		query.Where("community_id = ?", p.CommunityID)
	}

	query = query.Where("status = ?", p.Status)

	if err = applySorting(query, p); err != nil {
		return nil, err
	}

	if p.Page > 0 && p.Size > 0 {
		offset := (p.Page - 1) * p.Size
		query = query.Offset(int(offset)).Limit(int(p.Size))
	}

	var items []*models.PostListItem
	if err = query.Scan(&items).Error; err != nil {
		zap.L().Error("get post list failed",
			zap.Any("params", p),
			zap.Error(err))
		return nil, err
	}

	return items, nil
}

func applySorting(query *gorm.DB, p *models.ParamPostList) error {
	switch p.SortBy {
	case models.SortFieldCreateTime:
		query = query.Order("create_time " + string(p.Order))
	case models.SortFieldUpdateTime:
		query = query.Order("update_time " + string(p.Order))
	case models.SortFieldScore:
		// 按热度排序需要特殊处理，当前版本先记录警告并使用默认排序
		zap.L().Warn("score sorting requires Redis integration, using default sorting")
		query = query.Order("create_time " + string(p.Order))
	default:
		query = query.Order("create_time " + string(p.Order))
	}
	return nil
}
