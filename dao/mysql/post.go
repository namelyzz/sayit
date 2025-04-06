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
	PostSummaryLength = 30    // 帖子摘要长度
	PostSummarySuffix = "..." // 帖子摘要细节
)

const postListSelect = `p.post_id, p.title, p.author_id, p.community_id, p.status,
	p.create_time, p.update_time, u.username AS user_name, c.community_name,
	CASE
		WHEN CHAR_LENGTH(p.content) > ? THEN CONCAT(SUBSTRING(p.content, 1, ?), ?)
		ELSE p.content
	END AS summary`

func GetPostList(p *models.ParamPostList) (posts []*models.PostListItem, err error) {
	query := buildPostListQuery()
	query = applyPostListFilters(query, p)
	query = applySorting(query, p)

	if p.Page > 0 && p.Size > 0 {
		offset := (p.Page - 1) * p.Size
		query = query.Offset(offset).Limit(p.Size)
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

func FilterPostListByIDs(postIDs []int64, p *models.ParamPostList) (posts []*models.PostListItem, err error) {
	if len(postIDs) == 0 {
		return []*models.PostListItem{}, nil
	}

	query := buildPostListQuery().Where("p.post_id IN ?", postIDs)
	query = applyPostListFilters(query, p)

	var items []*models.PostListItem
	if err = query.Scan(&items).Error; err != nil {
		zap.L().Error("filter post list by ids failed",
			zap.Any("params", p),
			zap.Int("post_id_count", len(postIDs)),
			zap.Error(err))
		return nil, err
	}

	return reorderPostListItems(postIDs, items), nil
}

func GetPostListByIDs(postIDs []int64) (posts []*models.PostListItem, err error) {
	return FilterPostListByIDs(postIDs, nil)
}

func buildPostListQuery() *gorm.DB {
	return db.Table("post p").
		Select(postListSelect, PostSummaryLength, PostSummaryLength, PostSummarySuffix).
		Joins("LEFT JOIN users u ON p.author_id = u.user_id").
		Joins("LEFT JOIN community c ON p.community_id = c.community_id")
}

func applyPostListFilters(query *gorm.DB, p *models.ParamPostList) *gorm.DB {
	if p == nil {
		return query
	}

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
	if p.Status != nil {
		query = query.Where("p.status = ?", *p.Status)
	}

	return query
}

func applySorting(query *gorm.DB, p *models.ParamPostList) *gorm.DB {
	orderBy := "p.create_time"
	switch p.SortBy {
	case models.SortFieldCreateTime:
		orderBy = "p.create_time"
	case models.SortFieldUpdateTime:
		orderBy = "p.update_time"
	default:
		orderBy = "p.create_time"
	}

	order := "desc"
	if p.Order == models.SortDirectionAsc {
		order = "asc"
	}

	return query.Order(orderBy + " " + order)
}

func reorderPostListItems(postIDs []int64, items []*models.PostListItem) []*models.PostListItem {
	itemByID := make(map[int64]*models.PostListItem, len(items))
	for _, item := range items {
		itemByID[item.PostID] = item
	}

	ordered := make([]*models.PostListItem, 0, len(items))
	for _, postID := range postIDs {
		if item, ok := itemByID[postID]; ok {
			ordered = append(ordered, item)
		}
	}

	return ordered
}
