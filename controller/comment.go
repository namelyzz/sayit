package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/service"
	"github.com/namelyzz/sayit/utils/api"
	"go.uber.org/zap"
	"strconv"
)

// CreateCommentHandler 创建评论接口
//
// 路由: POST /api/v1/comment (需要JWT认证)
// 请求体: JSON { "post_id": "123", "parent_id": "0", "content": "评论内容" }
//
// 参数说明:
//   - post_id: 帖子ID（必填）
//   - parent_id: 父评论ID（可选，不传或为 0 表示顶级评论）
//   - content: 评论内容（必填，最长 1024 字符）
//
// 返回: 创建成功的评论对象（含 comment_id、create_time 等）
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
			zap.String("postID", p.PostID),
			zap.String("content", p.Content),
		)
		// 返回更具体的错误信息，方便调试
		api.ResponseErrorWithMsg(c, api.CodeServerBusy, err.Error())
		return
	}

	// 5. 创建成功，返回评论对象
	api.ResponseSuccess(c, comment)
}

// GetCommentListHandler 获取帖子评论列表接口
//
// 路由: GET /api/v1/post/:id/comments (公开接口，可选JWT)
// 查询参数: page, size
//
// 返回结构: 评论树，每条评论包含:
//   - 评论基本信息（content、author_name、create_time 等）
//   - 子评论列表（children，递归嵌套，最多 10 层）
//   - 子评论总数（child_count）
//   - 当前用户点赞状态（is_liked，未登录为 false）
//
// 已删除评论: content 显示为 [已删除]，子评论保持不变
func GetCommentListHandler(c *gin.Context) {
	// 1. 解析帖子ID
	postIDStr := c.Param("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		zap.L().Error("get comment list with invalid post id",
			zap.String("post_id", postIDStr),
			zap.Error(err))
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	// 2. 绑定查询参数
	p := new(models.ParamCommentList)
	if err := c.ShouldBindQuery(p); err != nil {
		zap.L().Warn("invalid comment list query parameters", zap.Error(err))
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	p.ValidateAndSetDefaults()

	// 3. 获取当前用户ID（可选，未登录为0）
	currentUserID := api.GetOptionalUserID(c)

	// 4. 调用 service 层获取评论树
	data, err := service.GetCommentTree(c.Request.Context(), postID, p, currentUserID)
	if err != nil {
		zap.L().Error("service.GetCommentTree() failed",
			zap.Int64("postID", postID),
			zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	// 5. 返回评论树
	api.ResponseSuccess(c, data)
}

// DeleteCommentHandler 删除评论接口
//
// 路由: DELETE /api/v1/comment/:id (需要JWT认证)
// 路径参数: id - 评论ID
//
// 权限规则:
//   - 帖子作者: 可删除该帖子下的任意评论
//   - 评论作者: 只能删除自己的评论
//
// 删除策略: 软删除（status=2），子评论保持不变
// 幂等设计: 已删除的评论再次删除返回成功
func DeleteCommentHandler(c *gin.Context) {
	// 1. 解析评论ID
	commentIDStr := c.Param("id")
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		zap.L().Error("delete comment with invalid id",
			zap.String("comment_id", commentIDStr),
			zap.Error(err))
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	// 2. 获取当前登录用户ID
	userID, err := api.GetCurrentUserID(c)
	if err != nil {
		api.ResponseError(c, api.CodeNeedLogin)
		return
	}

	// 3. 调用 service 层执行删除逻辑
	if err := service.DeleteComment(c.Request.Context(), userID, commentID); err != nil {
		if err == api.ErrorNoPermission {
			api.ResponseErrorWithMsg(c, api.CodeInvalidParam, err.Error())
			return
		}
		zap.L().Error("service.DeleteComment() failed",
			zap.Int64("commentID", commentID),
			zap.Int64("userID", userID),
			zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	// 4. 删除成功
	api.ResponseSuccess(c, nil)
}

// LikeCommentHandler 点赞评论接口
//
// 路由: POST /api/v1/comment/:id/like (需要JWT认证)
// 路径参数: id - 评论ID
//
// 业务规则:
//   - 不能重复点赞（返回 "重复的点赞"）
//   - 不能点赞已删除的评论
//
// 一致性: Redis SADD + MySQL INCR（乐观重试，MySQL 失败不返回错误）
func LikeCommentHandler(c *gin.Context) {
	// 1. 解析评论ID
	commentIDStr := c.Param("id")
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		zap.L().Error("like comment with invalid id",
			zap.String("comment_id", commentIDStr),
			zap.Error(err))
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	// 2. 获取当前登录用户ID
	userID, err := api.GetCurrentUserID(c)
	if err != nil {
		api.ResponseError(c, api.CodeNeedLogin)
		return
	}

	// 3. 调用 service 层执行点赞逻辑
	if err := service.LikeComment(c.Request.Context(), userID, commentID); err != nil {
		if err == api.ErrorLikeRepeated {
			api.ResponseErrorWithMsg(c, api.CodeInvalidParam, err.Error())
			return
		}
		zap.L().Error("service.LikeComment() failed",
			zap.Int64("commentID", commentID),
			zap.Int64("userID", userID),
			zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	// 4. 点赞成功
	api.ResponseSuccess(c, nil)
}

// UnlikeCommentHandler 取消点赞评论接口
//
// 路由: DELETE /api/v1/comment/:id/like (需要JWT认证)
// 路径参数: id - 评论ID
//
// 业务规则:
//   - 幂等设计: 未点赞时调用取消点赞，直接返回成功（不报错）
//   - 不能取消点赞已删除的评论
//
// 一致性: Redis SREM + MySQL DECR（乐观重试，MySQL 失败不返回错误）
func UnlikeCommentHandler(c *gin.Context) {
	// 1. 解析评论ID
	commentIDStr := c.Param("id")
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		zap.L().Error("unlike comment with invalid id",
			zap.String("comment_id", commentIDStr),
			zap.Error(err))
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	// 2. 获取当前登录用户ID
	userID, err := api.GetCurrentUserID(c)
	if err != nil {
		api.ResponseError(c, api.CodeNeedLogin)
		return
	}

	// 3. 调用 service 层执行取消点赞逻辑
	if err := service.UnlikeComment(c.Request.Context(), userID, commentID); err != nil {
		zap.L().Error("service.UnlikeComment() failed",
			zap.Int64("commentID", commentID),
			zap.Int64("userID", userID),
			zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	// 4. 取消点赞成功
	api.ResponseSuccess(c, nil)
}
