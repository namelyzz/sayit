package jwt

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"time"
)

// 在实际应用中，密钥必须保密且足够复杂。而且推荐配置化，不要和代码放在一起
var jwtSecret = []byte("Faiz555WuMingKe")

// UserClaims 必须嵌入 jwt.RegisteredClaims
type UserClaims struct {
	UserID               int64  `json:"user_id"`
	Username             string `json:"username"`
	jwt.RegisteredClaims        // 包含 iss, exp, iat 等标准 Claims
}

func CreateJWTToken(userID int64, username string) (string, error) {
	// 设置 1 小时后过期
	expirationTime := time.Now().Add(1 * time.Hour)

	claims := UserClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),     // 签发时间
			Issuer:    "sayit",                            // 签发人
		},
	}

	// 使用 HS256 算法创建 Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用密钥签名，得到完整的 Token 字符串
	return token.SignedString(jwtSecret)
}

func ParseJWTToken(tokenString string) (*UserClaims, error) {
	var mc = new(UserClaims)

	// V5 库的 ParseWithClaims 保持了相同的签名，但对内部类型处理更严格
	token, err := jwt.ParseWithClaims(tokenString, mc, func(token *jwt.Token) (i any, err error) {
		// 校验签名方法是否是 HS256
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	// 验证错误
	if err != nil {
		// V5 库对过期等错误提供了更细致的区分
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token is expired")
		}
		return nil, err
	}

	// 校验 token.Valid
	if token.Valid {
		// 验证通过后，将 token.Claims 断言回 UserClaims 类型
		if claims, ok := token.Claims.(*UserClaims); ok {
			return claims, nil
		}
	}

	return nil, errors.New("invalid token")
}
