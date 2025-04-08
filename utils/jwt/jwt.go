package jwt

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"time"
)

// jwtSecret JWT 签名密钥
// 生产环境中应从配置文件读取，且定期轮换
var jwtSecret = []byte("Faiz555WuMingKe")

// UserClaims 自定义 JWT 声明（Payload 部分）
// 嵌入 jwt.RegisteredClaims 以包含标准字段: iss(签发者), exp(过期时间), iat(签发时间)
type UserClaims struct {
	UserID               int64  `json:"user_id"`
	Username             string `json:"username"`
	jwt.RegisteredClaims        // 标准 Claims: iss, exp, iat 等
}

// CreateJWTToken 创建 JWT Token
// 算法: HS256 (HMAC-SHA256)
// 有效期: 1 小时
// 返回值: 签名后的完整 JWT 字符串（Header.Payload.Signature）
func CreateJWTToken(userID int64, username string) (string, error) {
	// 设置过期时间为当前时间 + 1 小时
	expirationTime := time.Now().Add(1 * time.Hour)

	// 构造 Claims（包含自定义字段和标准字段）
	claims := UserClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),     // 签发时间
			Issuer:    "sayit",                            // 签发者标识
		},
	}

	// 使用 HS256 算法创建 Token 对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 用密钥对 Token 进行签名，返回完整的 JWT 字符串
	return token.SignedString(jwtSecret)
}

// ParseJWTToken 解析并验证 JWT Token
// 验证步骤: 签名算法校验 -> 签名有效性验证 -> 过期时间检查
// 返回值: 解析出的 UserClaims 或错误信息
func ParseJWTToken(tokenString string) (*UserClaims, error) {
	var mc = new(UserClaims)

	// 解析 Token 并验证签名
	token, err := jwt.ParseWithClaims(tokenString, mc, func(token *jwt.Token) (i any, err error) {
		// 校验签名算法是否为预期的 HMAC 系列（防止算法替换攻击）
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		// 区分过期错误和其他错误
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token is expired")
		}
		return nil, err
	}

	// Token 解析成功且签名有效
	if token.Valid {
		if claims, ok := token.Claims.(*UserClaims); ok {
			return claims, nil
		}
	}

	return nil, errors.New("invalid token")
}
