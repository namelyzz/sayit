package api

import (
	"github.com/gin-gonic/gin"
	"github.com/namelyzz/sayit/utils/jwt"
	"strings"
)

const CtxUserIDKey = "userID"

// GetCurrentUserID 获取当前登录的用户ID
func GetCurrentUserID(c *gin.Context) (userID int64, err error) {
	uid, ok := c.Get(CtxUserIDKey)
	if !ok {
		err = ErrorUserNotLogin
		return
	}
	userID, ok = uid.(int64)
	if !ok {
		err = ErrorUserNotLogin
		return
	}
	return
}

// GetOptionalUserID 从 Authorization 头中尝试解析用户ID。
// 公开接口可使用它回填当前用户相关状态；无 token 或 token 无效时按未登录处理。
func GetOptionalUserID(c *gin.Context) int64 {
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		return 0
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		return 0
	}

	claims, err := jwt.ParseJWTToken(parts[1])
	if err != nil {
		return 0
	}
	return claims.UserID
}
