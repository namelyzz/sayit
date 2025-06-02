package service

import (
	"context"
	"github.com/namelyzz/sayit/dao/mysql"
	"github.com/namelyzz/sayit/dao/redis"
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/utils/api"
	"go.uber.org/zap"
	"strconv"
)

// 常量定义
const (
	maxCommentDepth           = 10  // 最大嵌套深度，超过此深度不再返回子评论
	deletedContent            = "[已删除]" // 已删除评论的占位内容
	commentLikeScoreMultiplier = 50 // 每个评论点赞的热度分值（帖子投票 432 分/票，评论点赞 50 分/赞，比例约 8.6:1）
	likeRetryCount            = 1  // MySQL 操作失败时的重试次数
)

// mock 变量定义，方便单元测试替换（mock）依赖的函数
// 使用模式: 测试时替换这些变量，验证业务逻辑正确性，无需真实数据库/Redis
var (
	getPostByIDFunc               = mysql.GetPostByID               // 查询帖子
	getCommentByIDFunc            = mysql.GetCommentByID            // 查询评论
	createCommentFunc             = mysql.CreateComment             // 创建评论
	softDeleteCommentFunc         = mysql.SoftDeleteComment         // 软删除评论
	incrCommentLikeCountFunc      = mysql.IncrCommentLikeCount      // 评论点赞数+1
	decrCommentLikeCountFunc      = mysql.DecrCommentLikeCount      // 评论点赞数-1
	getTopLevelCommentsFunc       = mysql.GetTopLevelComments       // 获取顶级评论
	countTopLevelCommentsFunc     = mysql.CountTopLevelComments     // 统计顶级评论数量
	countChildCommentsByParentFunc = mysql.CountChildCommentsByParentIDs // 批量统计子评论数量
	getChildCommentsByParentIDFunc  = mysql.GetChildCommentsByParentID   // 获取单个父评论的子评论（分页）
	countChildCommentsByParentIDFunc = mysql.CountChildCommentsByParentID // 统计单个父评论的子评论数量
	getCommentLikeScoreFunc       = mysql.GetCommentLikeScoreByAuthor  // 获取用户评论点赞总分

	commentLikeFunc   = redis.CommentLikeComment   // Redis 点赞
	commentUnlikeFunc = redis.CommentUnlikeComment  // Redis 取消点赞
	isCommentLikedFunc = redis.IsCommentLikedByUser // Redis 检查点赞
	incrCommentCountFunc = redis.IncrCommentCount   // Redis 帖子评论数+1
	decrCommentCountFunc = redis.DecrCommentCount   // Redis 帖子评论数-1
)

// CreateComment 创建评论的业务逻辑
//
// 业务规则:
//  1. 帖子必须存在且未删除
//  2. 如果是回复评论，父评论必须存在且属于同一帖子
//  3. 根评论ID(root_id)计算规则:
//     - 顶级评论: root_id = 0
//     - 回复顶级评论: root_id = 父评论ID
//     - 回复嵌套评论: root_id = 继承父评论的 root_id
//
// 返回: 创建成功的评论对象（含雪花ID、创建时间等）
func CreateComment(ctx context.Context, userID int64, p *models.ParamCreateComment) (*models.Comment, error) {
	// 1. 解析帖子ID
	postID, err := strconv.ParseInt(p.PostID, 10, 64)
	if err != nil {
		return nil, api.ErrorInvalidID
	}

	// 2. 验证帖子是否存在
	post, err := getPostByIDFunc(postID)
	if err != nil {
		zap.L().Error("post not found",
			zap.Int64("post_id", postID),
			zap.Error(err))
		return nil, err
	}

	// 3. 解析父评论ID
	var parentID int64
	if p.ParentID != "" {
		parentID, err = strconv.ParseInt(p.ParentID, 10, 64)
		if err != nil {
			return nil, api.ErrorInvalidID
		}
	}

	// 4. 如果是回复评论，验证父评论
	var rootID int64
	if parentID != 0 {
		parentComment, err := getCommentByIDFunc(parentID)
		if err != nil {
			zap.L().Error("parent comment not found",
				zap.Int64("parent_id", parentID),
				zap.Error(err))
			return nil, err
		}
		// 父评论必须属于同一帖子
		if int64(parentComment.PostID) != postID {
			return nil, api.ErrorInvalidParam
		}
		// 计算根评论ID：如果父评论是顶级评论，则根评论ID为父评论ID；否则继承父评论的根评论ID
		if int64(parentComment.ParentID) == 0 {
			rootID = parentID
		} else {
			rootID = int64(parentComment.RootID)
		}
	}

	// 5. 生成评论ID
	commentID := genIDFunc()

	// 6. 构造评论对象
	comment := &models.Comment{
		CommentID: models.SnowflakeID(commentID),
		PostID:    models.SnowflakeID(postID),
		AuthorID:  models.SnowflakeID(userID),
		ParentID:  models.SnowflakeID(parentID),
		RootID:    models.SnowflakeID(rootID),
		Content:   p.Content,
		Status:    1,
	}

	// 7. 入库
	if err := createCommentFunc(comment); err != nil {
		zap.L().Error("create comment failed",
			zap.Int64("post_id", postID),
			zap.Int64("author_id", userID),
			zap.Error(err))
		return nil, err
	}

	// 8. 更新 Redis 评论数缓存（失败只记日志，不阻塞）
	if err := incrCommentCountFunc(ctx, postID); err != nil {
		zap.L().Error("incr comment count cache failed",
			zap.Int64("post_id", postID),
			zap.Error(err))
	}

	_ = post // 避免未使用变量警告

	return comment, nil
}

// GetCommentTree 获取帖子的顶级评论列表（不含子评论）
//
// 数据获取流程:
//  1. 分页获取顶级评论（parent_id=0，按指定排序方式）
//  2. 统计顶级评论总数（用于分页）
//  3. 批量统计每条评论的子评论数量（child_count）
//  4. 处理已删除评论（content 替换为 [已删除]）
//  5. 回填当前用户的点赞状态（仅登录用户）
//
// 子评论不再递归加载，改为前端按需懒加载（通过 GET /comment/:id/children）
func GetCommentTree(ctx context.Context, postID int64, p *models.ParamCommentList, currentUserID int64) (*models.CommentListResponse, error) {
	// 1. 获取顶级评论
	topComments, err := getTopLevelCommentsFunc(postID, p.Page, p.Size, p.Order)
	if err != nil {
		return nil, err
	}

	// 2. 统计顶级评论总数
	total, err := countTopLevelCommentsFunc(postID)
	if err != nil {
		return nil, err
	}

	// 3. 转换为 CommentDetail
	details := make([]*models.CommentDetail, 0, len(topComments))
	for _, c := range topComments {
		detail := &models.CommentDetail{
			Comment:    &c.Comment,
			AuthorName: c.AuthorName,
		}
		details = append(details, detail)
	}

	// 4. 批量统计子评论数量
	if len(details) > 0 {
		parentIDs := make([]int64, 0, len(details))
		for _, d := range details {
			parentIDs = append(parentIDs, int64(d.CommentID))
		}
		childCountMap, err := countChildCommentsByParentFunc(parentIDs)
		if err != nil {
			return nil, err
		}
		for _, d := range details {
			d.ChildCount = childCountMap[int64(d.CommentID)]
		}
	}

	// 5. 处理已删除评论的内容
	markDeletedContent(details)

	// 6. 回填点赞状态
	fillCommentLiked(ctx, details, currentUserID)

	return &models.CommentListResponse{
		List:  details,
		Total: total,
	}, nil
}

// GetCommentChildren 获取单条评论的子评论（分页）
//
// 数据获取流程:
//  1. 验证父评论存在
//  2. 分页获取直接子评论
//  3. 统计子评论总数（用于分页和 has_more 判断）
//  4. 批量统计每条子评论的 child_count（支持递归展开）
//  5. 处理已删除评论
//  6. 回填点赞状态
func GetCommentChildren(ctx context.Context, parentID int64, p *models.ParamCommentChildren, currentUserID int64) (*models.CommentChildrenResponse, error) {
	// 1. 验证父评论存在
	_, err := getCommentByIDFunc(parentID)
	if err != nil {
		return nil, err
	}

	// 2. 分页获取子评论
	children, err := getChildCommentsByParentIDFunc(parentID, p.Page, p.Size, p.Order)
	if err != nil {
		return nil, err
	}

	// 3. 统计子评论总数
	total, err := countChildCommentsByParentIDFunc(parentID)
	if err != nil {
		return nil, err
	}

	// 4. 转换为 CommentDetail
	details := make([]*models.CommentDetail, 0, len(children))
	for _, c := range children {
		detail := &models.CommentDetail{
			Comment:    &c.Comment,
			AuthorName: c.AuthorName,
		}
		details = append(details, detail)
	}

	// 5. 批量统计子评论的 child_count
	if len(details) > 0 {
		childIDs := make([]int64, 0, len(details))
		for _, d := range details {
			childIDs = append(childIDs, int64(d.CommentID))
		}
		childCountMap, err := countChildCommentsByParentFunc(childIDs)
		if err != nil {
			return nil, err
		}
		for _, d := range details {
			d.ChildCount = childCountMap[int64(d.CommentID)]
		}
	}

	// 6. 处理已删除评论
	markDeletedContent(details)

	// 7. 回填点赞状态
	fillCommentLiked(ctx, details, currentUserID)

	// 8. 计算 has_more
	hasMore := int64(p.Page*p.Size) < total

	return &models.CommentChildrenResponse{
		List:    details,
		Total:   total,
		HasMore: hasMore,
	}, nil
}

// fillCommentLiked 批量回填评论点赞状态（仅登录用户）
//
// 流程:
//  1. 未登录用户（userID=0）直接返回，不查询 Redis
//  2. 递归收集所有评论ID（包括子评论）
//  3. 使用 Pipeline 批量查询 Redis SISMEMBER
//  4. 递归设置每条评论的 IsLiked 字段
func fillCommentLiked(ctx context.Context, details []*models.CommentDetail, currentUserID int64) {
	if currentUserID == 0 {
		return
	}

	// 收集所有评论ID（包括子评论）
	allIDs := collectCommentIDs(details, nil)
	if len(allIDs) == 0 {
		return
	}

	// 批量查询点赞状态
	likedMap := redis.BatchIsCommentLikedByUser(ctx, allIDs, currentUserID)

	// 回填
	setCommentLiked(details, likedMap)
}

// collectCommentIDs 递归收集所有评论ID
func collectCommentIDs(details []*models.CommentDetail, ids []int64) []int64 {
	for _, d := range details {
		ids = append(ids, int64(d.CommentID))
		if len(d.Children) > 0 {
			ids = collectCommentIDs(d.Children, ids)
		}
	}
	return ids
}

// setCommentLiked 递归设置点赞状态
func setCommentLiked(details []*models.CommentDetail, likedMap map[int64]bool) {
	for _, d := range details {
		d.IsLiked = likedMap[int64(d.CommentID)]
		if len(d.Children) > 0 {
			setCommentLiked(d.Children, likedMap)
		}
	}
}

// DeleteComment 删除评论的业务逻辑
//
// 权限规则:
//   - 帖子作者: 可删除该帖子下的任意评论
//   - 评论作者: 只能删除自己的评论
//   - 其他用户: 无权删除
//
// 策略:
//   - 软删除: status 设为 2，不物理删除
//   - 子评论: 保持不变，继续正常展示
//   - 幂等: 已删除的评论再次删除返回成功
func DeleteComment(ctx context.Context, userID, commentID int64) error {
	// 1. 查询评论是否存在
	comment, err := getCommentByIDFunc(commentID)
	if err != nil {
		return err
	}

	// 2. 已删除的评论，幂等返回成功
	if comment.Status == 2 {
		return nil
	}

	// 3. 权限校验：评论作者 或 帖子作者
	isCommentAuthor := int64(comment.AuthorID) == userID
	isPostAuthor := false
	if !isCommentAuthor {
		post, err := getPostByIDFunc(int64(comment.PostID))
		if err != nil {
			return err
		}
		isPostAuthor = int64(post.AuthorID) == userID
	}

	if !isCommentAuthor && !isPostAuthor {
		return api.ErrorNoPermission
	}

	// 4. 软删除
	if err := softDeleteCommentFunc(commentID); err != nil {
		return err
	}

	// 5. 更新 Redis 评论数缓存（失败只记日志，不阻塞）
	if err := decrCommentCountFunc(ctx, int64(comment.PostID)); err != nil {
		zap.L().Error("decr comment count cache failed",
			zap.Int64("post_id", int64(comment.PostID)),
			zap.Error(err))
	}

	return nil
}

// markDeletedContent 递归处理已删除评论的内容
// 将 status=2 的评论的 content 替换为 [已删除]
func markDeletedContent(details []*models.CommentDetail) {
	for _, d := range details {
		if d.Status == 2 {
			d.Content = deletedContent
		}
		if len(d.Children) > 0 {
			markDeletedContent(d.Children)
		}
	}
}

// LikeComment 点赞评论的业务逻辑
//
// 一致性策略: 幂等设计 + 乐观重试
//  - Redis 先操作: SADD 记录点赞状态（幂等，已存在返回 0）
//  - MySQL 后操作: INCR like_count（失败重试 1 次）
//  - 容错: MySQL 失败时只记日志，不返回错误（Redis 已记录，可后续补偿）
//
// 返回错误:
//   - ErrorLikeRepeated: 已点赞过
//   - ErrorInvalidParam: 评论已删除
//   - 其他: 评论不存在、Redis 失败等
func LikeComment(ctx context.Context, userID, commentID int64) error {
	// 1. 查询评论是否存在
	comment, err := getCommentByIDFunc(commentID)
	if err != nil {
		return err
	}

	// 2. 检查评论是否已删除
	if comment.Status == 2 {
		return api.ErrorInvalidParam
	}

	// 3. Redis SADD（幂等）
	added, err := commentLikeFunc(ctx, commentID, userID)
	if err != nil {
		return err
	}

	// 4. 如果返回 0，说明已点赞
	if !added {
		return api.ErrorLikeRepeated
	}

	// 5. MySQL INCR like_count（乐观重试）
	for i := 0; i <= likeRetryCount; i++ {
		if err := incrCommentLikeCountFunc(commentID); err == nil {
			return nil
		}
		// 最后一次重试仍失败，记录日志但不返回错误
		if i == likeRetryCount {
			zap.L().Error("incr comment like_count failed after retry",
				zap.Int64("comment_id", commentID),
				zap.Int64("user_id", userID))
		}
	}

	return nil
}

// UnlikeComment 取消点赞评论的业务逻辑
//
// 一致性策略: 幂等设计 + 乐观重试
//  - Redis 先操作: SREM 移除点赞状态（幂等，未点赞返回 0）
//  - MySQL 后操作: DECR like_count（失败重试 1 次）
//  - 容错: MySQL 失败时只记日志，不返回错误（Redis 已移除，可后续补偿）
//
// 幂等设计:
//   - 未点赞时调用取消点赞，直接返回成功（不报错）
//   - 前端无需区分"取消成功"和"未点赞"两种状态
func UnlikeComment(ctx context.Context, userID, commentID int64) error {
	// 1. 查询评论是否存在
	comment, err := getCommentByIDFunc(commentID)
	if err != nil {
		return err
	}

	// 2. 检查评论是否已删除
	if comment.Status == 2 {
		return api.ErrorInvalidParam
	}

	// 3. Redis SREM（幂等）
	removed, err := commentUnlikeFunc(ctx, commentID, userID)
	if err != nil {
		return err
	}

	// 4. 如果返回 0，说明未点赞，幂等返回成功
	if !removed {
		return nil
	}

	// 5. MySQL DECR like_count（乐观重试）
	for i := 0; i <= likeRetryCount; i++ {
		if err := decrCommentLikeCountFunc(commentID); err == nil {
			return nil
		}
		// 最后一次重试仍失败，记录日志但不返回错误
		if i == likeRetryCount {
			zap.L().Error("decr comment like_count failed after retry",
				zap.Int64("comment_id", commentID),
				zap.Int64("user_id", userID))
		}
	}

	return nil
}

// GetUserCommentScore 获取用户评论点赞的热度分数
//
// 计算公式: 用户所有正常评论的 like_count 总和 × 50
// 与帖子投票分值对比:
//   - 帖子投票: 1 票 = 432 分
//   - 评论点赞: 1 赞 = 50 分
//   - 比例约 8.6:1
//
// 使用场景: 后续热度接口调用此函数，与帖子投票分合并计算用户总热度
func GetUserCommentScore(userID int64) (int64, error) {
	likeCount, err := getCommentLikeScoreFunc(userID)
	if err != nil {
		return 0, err
	}
	return likeCount * commentLikeScoreMultiplier, nil
}
