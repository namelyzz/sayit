package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/namelyzz/sayit/utils/api"
	"github.com/namelyzz/sayit/utils/jwt"
	"strings"
)

func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			api.ResponseError(c, api.CodeNeedLogin)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			api.ResponseError(c, api.CodeInvalidToken)
			c.Abort()
			return
		}

		mc, err := jwt.ParseJWTToken(parts[1])
		if err != nil {
			api.ResponseError(c, api.CodeInvalidToken)
			c.Abort()
			return
		}

		// 将当前请求的userID信息保存到请求的上下文c上
		c.Set(api.CtxUserIDKey, mc.UserID)

		c.Next() // 后续的处理请求的函数中 可以用过c.Get(CtxUserIDKey) 来获取当前请求的用户信息
	}
}
