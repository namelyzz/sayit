package mysql

import (
	"github.com/namelyzz/sayit/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CreateComment 将评论记录插入 MySQL comment 表
// 使用 Omit("UpdateTime") 排除 UpdateTime 字段，让数据库使用默认值
// 入库前需确保: PostID、AuthorID、Content 已设置，Status 默认为 1
func CreateComment(c *models.Comment) (err error) {
	res := db.Omit("UpdateTime").Create(c)
	if res.Error != nil {
		zap.L().Error("create comment failed",
			zap.String("operation", "create_comment"),
			zap.Int64("post_id", int64(c.PostID)),
			zap.Int64("author_id", int64(c.AuthorID)),
			zap.Error(res.Error))
		return res.Error
	}
	return nil
}

// GetCommentByID 根据评论ID查询单个评论
func GetCommentByID(commentID int64) (comment *models.Comment, err error) {
	comment = new(models.Comment)
	res := db.Where("comment_id = ?", commentID).First(comment)
	if res.Error != nil {
		return nil, res.Error
	}
	return comment, nil
}

// CommentWithAuthor 包含作者用户名的评论
// 用于评论列表查询，通过 JOIN users 表获取作者名
// 避免 N+1 问题，在 SQL 层面一次性获取
type CommentWithAuthor struct {
	models.Comment
	AuthorName string `json:"author_name" gorm:"column:author_name"`
}

// GetTopLevelComments 获取帖子的顶级评论（parent_id=0），分页
// 使用场景: 评论树的第一层
// 参数: order 为排序方式，"asc" 正序(从旧到新)，"desc" 倒序(从新到旧)
// SQL: SELECT c.*, u.username FROM comment c JOIN users u ON ... WHERE parent_id=0 ORDER BY create_time ASC/DESC
func GetTopLevelComments(postID int64, page, size int, order string) ([]*CommentWithAuthor, error) {
	var comments []*CommentWithAuthor
	offset := (page - 1) * size
	orderClause := "c.create_time DESC"
	if order == "asc" {
		orderClause = "c.create_time ASC"
	}
	res := db.Table("comment c").
		Select("c.*, u.username AS author_name").
		Joins("LEFT JOIN users u ON c.author_id = u.user_id").
		Where("c.post_id = ? AND c.parent_id = 0", postID).
		Order(orderClause).
		Offset(offset).Limit(size).
		Scan(&comments)
	if res.Error != nil {
		zap.L().Error("get top level comments failed",
			zap.Int64("post_id", postID),
			zap.Error(res.Error))
		return nil, res.Error
	}
	return comments, nil
}

// CountTopLevelComments 统计帖子的顶级评论数量
// 用于分页计算，返回该帖子下 parent_id=0 的评论总数
func CountTopLevelComments(postID int64) (int64, error) {
	var count int64
	res := db.Model(&models.Comment{}).
		Where("post_id = ? AND parent_id = 0", postID).
		Count(&count)
	if res.Error != nil {
		return 0, res.Error
	}
	return count, nil
}

// GetChildCommentsByParentIDs 批量获取指定父评论ID列表的子评论
// 使用场景: 递归构建评论树时，批量获取某一层的所有子评论
// 参数: order 为排序方式，"asc" 正序(从旧到新)，"desc" 倒序(从新到旧)
// 注意: 空 parentIDs 时直接返回空切片，避免无意义的 SQL 查询
func GetChildCommentsByParentIDs(parentIDs []int64, order string) ([]*CommentWithAuthor, error) {
	if len(parentIDs) == 0 {
		return []*CommentWithAuthor{}, nil
	}
	var comments []*CommentWithAuthor
	orderClause := "c.create_time ASC"
	if order == "desc" {
		orderClause = "c.create_time DESC"
	}
	res := db.Table("comment c").
		Select("c.*, u.username AS author_name").
		Joins("LEFT JOIN users u ON c.author_id = u.user_id").
		Where("c.parent_id IN ?", parentIDs).
		Order(orderClause).
		Scan(&comments)
	if res.Error != nil {
		zap.L().Error("get child comments failed",
			zap.Int("parent_count", len(parentIDs)),
			zap.Error(res.Error))
		return nil, res.Error
	}
	return comments, nil
}

// SoftDeleteComment 软删除评论，将 status 设为 2
// 幂等设计: 评论不存在或已删除时不返回错误，前端可重复调用
// 子评论保持不变，仅标记当前评论为已删除
func SoftDeleteComment(commentID int64) error {
	res := db.Model(&models.Comment{}).
		Where("comment_id = ? AND status = 1", commentID).
		Update("status", 2)
	if res.Error != nil {
		zap.L().Error("soft delete comment failed",
			zap.Int64("comment_id", commentID),
			zap.Error(res.Error))
		return res.Error
	}
	if res.RowsAffected == 0 {
		return nil // 评论不存在或已删除，不返回错误（幂等）
	}
	return nil
}
// CountChildCommentsByParentIDs 批量统计指定父评论的子评论数量
// 使用场景: 评论树中显示每条评论的子评论总数（ChildCount 字段）
// 使用 GROUP BY parent_id 一次性查询，避免 N+1 问题
// 返回 map[parent_id]count，方便按 parent_id 查找
func CountChildCommentsByParentIDs(parentIDs []int64) (map[int64]int64, error) {
	if len(parentIDs) == 0 {
		return map[int64]int64{}, nil
	}
	type result struct {
		ParentID int64 `gorm:"column:parent_id"`
		Count    int64 `gorm:"column:cnt"`
	}
	var results []result
	res := db.Model(&models.Comment{}).
		Select("parent_id, COUNT(*) AS cnt").
		Where("parent_id IN ?", parentIDs).
		Group("parent_id").
		Scan(&results)
	if res.Error != nil {
		return nil, res.Error
	}
	countMap := make(map[int64]int64, len(results))
	for _, r := range results {
		countMap[r.ParentID] = r.Count
	}
	return countMap, nil
}

// IncrCommentLikeCount 评论点赞数 +1
// 调用时机: Redis SADD 成功后（确认是新点赞）
// 使用 gorm.Expr 确保原子性操作，避免并发问题
// 失败时由 service 层重试（乐观重试策略）
func IncrCommentLikeCount(commentID int64) error {
	res := db.Model(&models.Comment{}).
		Where("comment_id = ?", commentID).
		Update("like_count", gorm.Expr("like_count + 1"))
	if res.Error != nil {
		zap.L().Error("incr comment like_count failed",
			zap.Int64("comment_id", commentID),
			zap.Error(res.Error))
		return res.Error
	}
	return nil
}

// DecrCommentLikeCount 评论点赞数 -1（不会小于 0）
// 调用时机: Redis SREM 成功后（确认取消了点赞）
// WHERE 条件包含 like_count > 0，防止减成负数
// 失败时由 service 层重试（乐观重试策略）
func DecrCommentLikeCount(commentID int64) error {
	res := db.Model(&models.Comment{}).
		Where("comment_id = ? AND like_count > 0", commentID).
		Update("like_count", gorm.Expr("like_count - 1"))
	if res.Error != nil {
		zap.L().Error("decr comment like_count failed",
			zap.Int64("comment_id", commentID),
			zap.Error(res.Error))
		return res.Error
	}
	return nil
}

// GetCommentLikeScoreByAuthor 获取用户所有正常评论的点赞总分
// 使用场景: 计算用户热度值中的评论贡献部分
// 只统计 status=1（正常）的评论，已删除评论不计入
// 使用 COALESCE 处理用户无评论的情况，返回 0 而非 NULL
func GetCommentLikeScoreByAuthor(authorID int64) (int64, error) {
	var total int64
	res := db.Model(&models.Comment{}).
		Where("author_id = ? AND status = 1", authorID).
		Select("COALESCE(SUM(like_count), 0)").
		Scan(&total)
	if res.Error != nil {
		zap.L().Error("get comment like score by author failed",
			zap.Int64("author_id", authorID),
			zap.Error(res.Error))
		return 0, res.Error
	}
	return total, nil
}

// CountCommentsByPostID 统计帖子的评论数量（包含所有层级，只统计正常状态）
// 使用场景: 帖子详情和帖子列表中显示评论总数
func CountCommentsByPostID(postID int64) (int64, error) {
	var count int64
	res := db.Model(&models.Comment{}).
		Where("post_id = ? AND status = 1", postID).
		Count(&count)
	if res.Error != nil {
		zap.L().Error("count comments by post id failed",
			zap.Int64("post_id", postID),
			zap.Error(res.Error))
		return 0, res.Error
	}
	return count, nil
}

// CountCommentsByPostIDs 批量统计多个帖子的评论数量
// 使用场景: 帖子列表中批量获取评论数，避免 N+1 查询
// 返回 map[post_id]count
func CountCommentsByPostIDs(postIDs []int64) (map[int64]int64, error) {
	if len(postIDs) == 0 {
		return map[int64]int64{}, nil
	}
	type result struct {
		PostID int64 `gorm:"column:post_id"`
		Count  int64 `gorm:"column:cnt"`
	}
	var results []result
	res := db.Model(&models.Comment{}).
		Select("post_id, COUNT(*) AS cnt").
		Where("post_id IN ? AND status = 1", postIDs).
		Group("post_id").
		Scan(&results)
	if res.Error != nil {
		zap.L().Error("count comments by post ids failed",
			zap.Int("post_count", len(postIDs)),
			zap.Error(res.Error))
		return nil, res.Error
	}
	countMap := make(map[int64]int64, len(results))
	for _, r := range results {
		countMap[r.PostID] = r.Count
	}
	return countMap, nil
}

// GetChildCommentsByParentID 获取指定父评论的直接子评论（分页）
// 使用场景: 子评论懒加载，点击展开时请求单个父评论的子评论
// 参数: order 为排序方式，"asc" 正序(从旧到新)，"desc" 倒序(从新到旧)
func GetChildCommentsByParentID(parentID int64, page, size int, order string) ([]*CommentWithAuthor, error) {
	var comments []*CommentWithAuthor
	offset := (page - 1) * size
	orderClause := "c.create_time ASC"
	if order == "desc" {
		orderClause = "c.create_time DESC"
	}
	res := db.Table("comment c").
		Select("c.*, u.username AS author_name").
		Joins("LEFT JOIN users u ON c.author_id = u.user_id").
		Where("c.parent_id = ?", parentID).
		Order(orderClause).
		Offset(offset).Limit(size).
		Scan(&comments)
	if res.Error != nil {
		zap.L().Error("get child comments by parent id failed",
			zap.Int64("parent_id", parentID),
			zap.Error(res.Error))
		return nil, res.Error
	}
	return comments, nil
}

// CountChildCommentsByParentID 统计指定父评论的直接子评论数量
// 使用场景: 子评论分页计算和 has_more 判断
func CountChildCommentsByParentID(parentID int64) (int64, error) {
	var count int64
	res := db.Model(&models.Comment{}).
		Where("parent_id = ?", parentID).
		Count(&count)
	if res.Error != nil {
		return 0, res.Error
	}
	return count, nil
}

// GetNormalCommentIDs 分批获取正常状态的评论ID（游标分页）
// 使用场景: 对账任务遍历所有评论
// 参数: lastID 为上一批最后一条记录的 ID，batchSize 为每批数量
// 返回的 ID 按 comment_id 升序排列，方便游标翻页
func GetNormalCommentIDs(lastID int64, batchSize int) ([]int64, error) {
	var ids []int64
	res := db.Model(&models.Comment{}).
		Where("status = 1 AND comment_id > ?", lastID).
		Order("comment_id ASC").
		Limit(batchSize).
		Pluck("comment_id", &ids)
	if res.Error != nil {
		zap.L().Error("get normal comment ids failed",
			zap.Int64("lastID", lastID),
			zap.Error(res.Error))
		return nil, res.Error
	}
	return ids, nil
}

// GetCommentLikeCounts 批量获取评论的 like_count
// 使用场景: 对账任务，获取 MySQL 中的 like_count 与 Redis 对比
// 返回 map[commentID]like_count
func GetCommentLikeCounts(commentIDs []int64) (map[int64]int64, error) {
	if len(commentIDs) == 0 {
		return map[int64]int64{}, nil
	}
	type result struct {
		CommentID int64 `gorm:"column:comment_id"`
		LikeCount int64 `gorm:"column:like_count"`
	}
	var results []result
	res := db.Model(&models.Comment{}).
		Select("comment_id, like_count").
		Where("comment_id IN ?", commentIDs).
		Scan(&results)
	if res.Error != nil {
		zap.L().Error("get comment like counts failed",
			zap.Int("comment_count", len(commentIDs)),
			zap.Error(res.Error))
		return nil, res.Error
	}
	countMap := make(map[int64]int64, len(results))
	for _, r := range results {
		countMap[r.CommentID] = r.LikeCount
	}
	return countMap, nil
}

// BatchUpdateLikeCount 批量更新评论的 like_count
// 使用场景: 对账任务，修复不一致的 like_count
// 使用 CASE WHEN 实现批量不同值更新
func BatchUpdateLikeCount(countMap map[int64]int64) error {
	if len(countMap) == 0 {
		return nil
	}

	// 构建 CASE WHEN SQL
	// UPDATE comment SET like_count = CASE comment_id WHEN ? THEN ? WHEN ? THEN ? ... END WHERE comment_id IN (?, ?, ...)
	ids := make([]int64, 0, len(countMap))
	for id := range countMap {
		ids = append(ids, id)
	}

	// 使用事务逐条更新（简单可靠，对账任务非高频操作）
	// 如果评论数量大，可以改为批量 CASE WHEN，但复杂度较高
	return db.Transaction(func(tx *gorm.DB) error {
		for commentID, likeCount := range countMap {
			res := tx.Model(&models.Comment{}).
				Where("comment_id = ?", commentID).
				Update("like_count", likeCount)
			if res.Error != nil {
				zap.L().Error("batch update like_count failed",
					zap.Int64("comment_id", commentID),
					zap.Int64("like_count", likeCount),
					zap.Error(res.Error))
				return res.Error
			}
		}
		return nil
	})
}
