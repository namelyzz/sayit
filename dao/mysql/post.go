package mysql

import (
	"github.com/namelyzz/sayit/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

func CreatePost(p *models.Post) (err error) {
	res := db.Omit("UpdateTime").Create(p)
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
		query = query.Where("p.community_id = ?", p.CommunityID)
	}
	if p.UserName != "" {
		query = query.Where("u.username LIKE ?", "%"+p.UserName+"%")
	}
	if p.Keyword != "" {
		query = query.Where("p.title LIKE ?", "%"+p.Keyword+"%")
	}

	if p.StartTime != nil {
		query = query.Where("p.create_time >= ?", time.Unix(*p.StartTime, 0))
	}
	if p.EndTime != nil {
		query = query.Where("p.create_time <= ?", time.Unix(*p.EndTime, 0))
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
		query = query.Order("score " + string(p.Order))
	default:
		query = query.Order("create_time " + string(p.Order))
	}
	return nil
}

func GetPostListByIDs(postIDs []int64) (posts []*models.PostListItem, err error) {
	if len(postIDs) == 0 {
		return nil, nil
	}

	var items []*models.PostListItem
	err = db.Model(&models.PostListItem{}).
		Select(`p.post_id, p.title, p.author_id, p.community_id, p.status, 
                p.create_time, p.update_time, u.user_name, c.community_name,
                CASE 
                    WHEN LENGTH(p.content) > ? THEN CONCAT(SUBSTRING(p.content, 1, ?), ?)
                    ELSE p.content
                END as summary`,
			PostSummaryLength, PostSummaryLength, PostSummarySuffix).
		Table("post p").
		Joins("LEFT JOIN users u ON p.author_id = u.user_id").
		Joins("LEFT JOIN community c ON p.community_id = c.community_id").
		Where("p.post_id IN ?", postIDs).
		Find(&items).Error

	return items, err
}
