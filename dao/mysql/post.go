package mysql

import (
	"github.com/namelyzz/sayit/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

// CreatePost 将帖子记录插入 MySQL post 表
// 使用 Omit("UpdateTime") 排除 UpdateTime 字段，让数据库使用默认值
// 注意: 此方法是独立的单表写入，CreatePostWithOutbox 才是带事务的完整写入
func CreatePost(p *models.Post) (err error) {
	res := db.Omit("UpdateTime").Create(p)
	if res.Error != nil {
		zap.L().Error("create post failed",
			zap.String("operation", "create_post"),
			zap.Int64("author_id", int64(p.AuthorID)),
			zap.Int64("community_id", p.CommunityID),
			zap.Error(res.Error))
		return res.Error
	}
	return nil
}

// GetPostByID 根据帖子ID查询单个帖子的完整信息
// 用于帖子详情接口，返回帖子的所有字段（不含作者名和社区名）
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

// 帖子列表查询的常量配置
const (
	PostSummaryLength = 30    // 内容摘要截取长度（字符数）
	PostSummarySuffix = "..." // 摘要超长时的后缀
)

// postListSelect 帖子列表查询的 SELECT 子句
// 使用 SQL 函数动态生成内容摘要:
//   - CHAR_LENGTH(p.content) > 30: 判断内容是否超长
//   - SUBSTRING(p.content, 1, 30): 截取前30个字符
//   - CONCAT(..., '...'): 拼接后缀
//
// 同时 JOIN users 和 community 表获取作者名和社区名
const postListSelect = `p.post_id, p.title, p.author_id, p.community_id, p.status,
	p.create_time, p.update_time, u.username AS user_name, c.community_name,
	CASE
		WHEN CHAR_LENGTH(p.content) > ? THEN CONCAT(SUBSTRING(p.content, 1, ?), ?)
		ELSE p.content
	END AS summary`

// GetPostList 帖子列表查询（纯 MySQL 路径）
// 适用于: update_time 排序、有作者名/关键词过滤、或 Redis 异常时的 fallback
// 查询流程: 构建基础查询 -> 应用过滤条件 -> 应用排序 -> 应用分页
func GetPostList(p *models.ParamPostList) (posts []*models.PostListItem, err error) {
	// 1. 构建基础查询（三表 JOIN）
	query := buildPostListQuery()
	// 2. 应用过滤条件（社区、作者、关键词、时间范围、状态）
	query = applyPostListFilters(query, p)
	// 3. 应用排序（create_time 或 update_time，asc 或 desc）
	query = applySorting(query, p)

	// 4. 应用分页
	if p.Page > 0 && p.Size > 0 {
		offset := (p.Page - 1) * p.Size
		query = query.Offset(offset).Limit(p.Size)
	}

	// 5. 执行查询
	var items []*models.PostListItem
	if err = query.Scan(&items).Error; err != nil {
		zap.L().Error("get post list failed",
			zap.Any("params", p),
			zap.Error(err))
		return nil, err
	}

	return items, nil
}

// FilterPostListByIDs 根据帖子 ID 列表查询帖子信息（带过滤条件）
// 使用场景: Redis 返回帖子 ID 后，回查 MySQL 获取完整帖子信息
// 特点:
//   - 使用 WHERE IN 查询指定 ID 的帖子
//   - 仍会应用过滤条件（status 等）
//   - 结果会按照传入的 postIDs 顺序重新排序（保持 Redis 的排序）
func FilterPostListByIDs(postIDs []int64, p *models.ParamPostList) (posts []*models.PostListItem, err error) {
	if len(postIDs) == 0 {
		return []*models.PostListItem{}, nil
	}

	// 构建查询: 基础三表 JOIN + WHERE IN 指定帖子ID
	query := buildPostListQuery().Where("p.post_id IN ?", postIDs)
	// 应用过滤条件（可能过滤掉部分帖子）
	query = applyPostListFilters(query, p)

	var items []*models.PostListItem
	if err = query.Scan(&items).Error; err != nil {
		zap.L().Error("filter post list by ids failed",
			zap.Any("params", p),
			zap.Int("post_id_count", len(postIDs)),
			zap.Error(err))
		return nil, err
	}

	// 保持 Redis 返回的排序顺序（MySQL 查询后顺序可能改变）
	return reorderPostListItems(postIDs, items), nil
}

// GetPostListByIDs 根据帖子 ID 列表查询帖子信息（无过滤条件）
func GetPostListByIDs(postIDs []int64) (posts []*models.PostListItem, err error) {
	return FilterPostListByIDs(postIDs, nil)
}

// CountNormalPostsByAuthor 统计用户正常状态的帖子数量。
func CountNormalPostsByAuthor(authorID int64) (count int64, err error) {
	res := db.Model(&models.Post{}).
		Where("author_id = ? AND status = ?", authorID, 1).
		Count(&count)
	if res.Error != nil {
		zap.L().Error("count normal posts by author failed",
			zap.Int64("author_id", authorID),
			zap.Error(res.Error))
		return 0, res.Error
	}
	return count, nil
}

// GetNormalPostIDsByAuthor 查询用户正常状态帖子ID列表。
func GetNormalPostIDsByAuthor(authorID int64) ([]int64, error) {
	var postIDs []int64
	res := db.Model(&models.Post{}).
		Where("author_id = ? AND status = ?", authorID, 1).
		Pluck("post_id", &postIDs)
	if res.Error != nil {
		zap.L().Error("get normal post ids by author failed",
			zap.Int64("author_id", authorID),
			zap.Error(res.Error))
		return nil, res.Error
	}
	return postIDs, nil
}

// buildPostListQuery 构建帖子列表的基础查询
// 三表 JOIN: post + users + community
// 同时通过 SQL 函数生成内容摘要
func buildPostListQuery() *gorm.DB {
	return db.Table("post p").
		Select(postListSelect, PostSummaryLength, PostSummaryLength, PostSummarySuffix).
		Joins("LEFT JOIN users u ON p.author_id = u.user_id").
		Joins("LEFT JOIN community c ON p.community_id = c.community_id")
}

// applyPostListFilters 应用帖子列表的过滤条件
// 支持的过滤条件:
//   - CommunityID: 精确匹配社区ID
//   - AuthorID: 作者ID精确匹配
//   - UserName: 作者名模糊搜索（LIKE %keyword%）
//   - Keyword: 标题模糊搜索（LIKE %keyword%）
//   - StartTime/EndTime: 创建时间范围筛选
//   - Status: 帖子状态精确匹配
func applyPostListFilters(query *gorm.DB, p *models.ParamPostList) *gorm.DB {
	if p == nil {
		return query
	}

	if p.CommunityID != 0 {
		query = query.Where("p.community_id = ?", p.CommunityID)
	}
	if p.AuthorID != 0 {
		query = query.Where("p.author_id = ?", p.AuthorID)
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

// applySorting 应用排序规则
// 支持的排序字段: create_time, update_time（score 排序走 Redis 路径）
// 支持的排序方向: asc, desc
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

// reorderPostListItems 按照 Redis 返回的 ID 顺序重新排列 MySQL 查询结果
// 原因: MySQL WHERE IN 查询返回的结果顺序不确定，需要保持 Redis 的排序
// 算法: 先建立 ID->Item 的映射，再按 postIDs 顺序提取
func reorderPostListItems(postIDs []int64, items []*models.PostListItem) []*models.PostListItem {
	itemByID := make(map[models.SnowflakeID]*models.PostListItem, len(items))
	for _, item := range items {
		itemByID[item.PostID] = item
	}

	ordered := make([]*models.PostListItem, 0, len(items))
	for _, postID := range postIDs {
		if item, ok := itemByID[models.SnowflakeID(postID)]; ok {
			ordered = append(ordered, item)
		}
	}

	return ordered
}
