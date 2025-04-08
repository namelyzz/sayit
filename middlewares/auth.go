package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/namelyzz/sayit/utils/api"
	"github.com/namelyzz/sayit/utils/jwt"
	"strings"
)

// JWTAuthMiddleware JWT 认证中间件
// 用于保护需要登录才能访问的接口
// 验证流程:
//  1. 从 Authorization 请求头获取 Token（格式: "Bearer <token>"）
//  2. 验证 Token 的格式和签名有效性
//  3. 将解析出的 UserID 存入 gin.Context，供后续 handler 使用
//
// 使用方式: 在路由组中通过 v1.Use(middlewares.JWTAuthMiddleware()) 应用
// 后续 handler 中可通过 api.GetCurrentUserID(c) 获取当前登录用户ID
func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		// 1. 获取 Authorization 请求头
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			api.ResponseError(c, api.CodeNeedLogin)
			c.Abort() // 终止请求链
			return
		}

		// 2. 按空格分割，验证 "Bearer <token>" 格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			api.ResponseError(c, api.CodeInvalidToken)
			c.Abort()
			return
		}

		// 3. 解析并验证 JWT Token
		mc, err := jwt.ParseJWTToken(parts[1])
		if err != nil {
			api.ResponseError(c, api.CodeInvalidToken)
			c.Abort()
			return
		}

		// 4. 将用户ID存入请求上下文，后续 handler 通过 c.Get("userID") 获取
		c.Set(api.CtxUserIDKey, mc.UserID)

		// 5. 放行，继续执行后续的 handler
		c.Next()
	}
}
