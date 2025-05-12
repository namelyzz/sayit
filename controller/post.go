package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/service"
	"github.com/namelyzz/sayit/utils/api"
	"go.uber.org/zap"
	"strconv"
)

// CreatePostHandler 创建帖子接口
// 路由: POST /api/v1/create_post (需要JWT认证)
// 请求体: JSON { "title": "xxx", "content": "xxx", "community_id": 1 }
// 流程: 参数绑定 -> 从JWT上下文获取用户ID -> 调用 service.CreatePost -> 返回结果
func CreatePostHandler(c *gin.Context) {
	// 1. 使用请求上下文，支持超时和取消
	ctx := c.Request.Context()

	// 2. 绑定 JSON 请求参数到 Post 模型
	// title、content、community_id 通过 binding:"required" 标签强制必填
	p := new(models.Post)
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Error("create post with invalid param", zap.Error(err))
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	// 3. 从 JWT 中间件设置的上下文中获取当前登录用户的 ID
	// 如果未登录（无 token 或 token 无效），中间件会拦截，这里做双重保护
	userID, err := api.GetCurrentUserID(c)
	if err != nil {
		api.ResponseError(c, api.CodeNeedLogin)
		return
	}

	// 4. 设置帖子作者为当前登录用户（客户端不能篡改作者ID）
	p.AuthorID = userID

	// 5. 调用 service 层执行创建帖子的业务逻辑
	// 包括: 生成帖子ID、写入MySQL、同步写入Redis排行榜
	if err = service.CreatePost(ctx, p); err != nil {
		zap.L().Error("service.CreatePost() failed",
			zap.Error(err),
			zap.Int64("userID", userID),
		)
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	// 6. 创建成功，返回空数据
	api.ResponseSuccess(c, nil)
}

// GetPostDetailHandler 获取帖子详情接口
// 路由: GET /api/v1/post_detail/:id (需要JWT认证)
// 路径参数: id - 帖子ID（雪花算法生成的数字）
// 流程: 解析路径参数 -> 调用 service.GetPostDetailByID -> 返回帖子+作者+社区详情
func GetPostDetailHandler(c *gin.Context) {
	// 1. 从路径参数中获取帖子ID字符串
	postIDStr := c.Param("id")

	// 2. 将字符串转换为 int64（雪花ID是64位整数）
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		zap.L().Error(
			"get post detail with invalid param",
			zap.String("invalid post id", postIDStr),
			zap.Error(err),
		)
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	// 3. 调用 service 层查询帖子详情（包含帖子、作者名、社区详情）
	data, err := service.GetPostDetailByID(c.Request.Context(), postID, api.GetOptionalUserID(c))
	if err != nil {
		zap.L().Error("service.GetPostDetailByID failed", zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	// 4. 返回帖子详情
	api.ResponseSuccess(c, data)
}

// GetPostListHandler 获取帖子列表接口
// 路由: GET /api/v1/posts (需要JWT认证)
// 查询参数: community_id, user_name, keyword, start_time, end_time, page, size, status, sort_by, order
// 流程: 绑定查询参数 -> 校验并设置默认值 -> 调用 service.GetPostList -> 返回列表
//
// 查询策略（按优先级）:
//   - sort_by=score: 优先从 Redis 热度榜取 ID，复杂过滤时批量扫描
//   - sort_by=create_time + 无作者名/关键词过滤: 优先从 Redis 时间榜取 ID
//   - 其他情况或 Redis 异常: 回退到 MySQL 联表查询
func GetPostListHandler(c *gin.Context) {
	p := new(models.ParamPostList)

	// 1. 使用 ShouldBindQuery 自动绑定 URL 查询参数到结构体
	// 例如: /api/v1/posts?community_id=1&page=2&size=10&sort_by=score
	if err := c.ShouldBindQuery(p); err != nil {
		zap.L().Warn("invalid query parameters",
			zap.Error(err),
			zap.Any("params", p))
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	// 2. 设置默认值并验证参数合法性
	// 默认: sort_by=create_time, order=desc, page=1, size=50, status=1
	if err := p.ValidateAndSetDefaults(); err != nil {
		zap.L().Warn("invalid parameters after validation",
			zap.Error(err),
			zap.Any("params", p))
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	// 3. 调用 service 层获取帖子列表（内部会根据排序策略选择 Redis 或 MySQL）
	data, err := service.GetPostListWithViewer(c.Request.Context(), p, api.GetOptionalUserID(c))
	if err != nil {
		zap.L().Error("get post list failed",
			zap.Error(err),
			zap.Any("params", p))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	// 4. 返回帖子列表
	api.ResponseSuccess(c, data)
}
