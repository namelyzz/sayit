package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/service"
	"github.com/namelyzz/sayit/utils/api"
	"go.uber.org/zap"
)

// CreateCommentHandler 创建评论接口
// 路由: POST /api/v1/comment (需要JWT认证)
// 请求体: JSON { "post_id": "123", "parent_id": "0", "content": "评论内容" }
// 流程: 参数绑定 -> 从JWT上下文获取用户ID -> 调用 service.CreateComment -> 返回结果
func CreateCommentHandler(c *gin.Context) {
	// 1. 使用请求上下文，支持超时和取消
	ctx := c.Request.Context()

	// 2. 绑定 JSON 请求参数
	p := new(models.ParamCreateComment)
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Error("create comment with invalid param", zap.Error(err))
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	// 3. 从 JWT 中间件设置的上下文中获取当前登录用户的 ID
	userID, err := api.GetCurrentUserID(c)
	if err != nil {
		api.ResponseError(c, api.CodeNeedLogin)
		return
	}

	// 4. 调用 service 层执行创建评论的业务逻辑
	comment, err := service.CreateComment(ctx, userID, p)
	if err != nil {
		zap.L().Error("service.CreateComment() failed",
			zap.Error(err),
			zap.Int64("userID", userID),
		)
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	// 5. 创建成功，返回评论对象
	api.ResponseSuccess(c, comment)
}
