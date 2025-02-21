package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestCreateJWTToken(t *testing.T) {
	userID := int64(123)
	username := "testUser"

	// 调用创建 Token 的方法
	tokenString, err := CreateJWTToken(userID, username)
	// 确保没有错误
	assert.NoError(t, err, "Error should be nil when generating token")
	// 确保 tokenString 不为空
	assert.NotEmpty(t, tokenString, "Token string should not be empty")

	// 尝试解析生成的 Token
	claims, err := ParseJWTToken(tokenString)
	assert.NoError(t, err, "Error should be nil when parsing token")

	// 确保解析出的 Claims 和传入的值一致
	assert.Equal(t, userID, claims.UserID, "UserID should match")
	assert.Equal(t, username, claims.Username, "Username should match")
}

func TestParseJWTToken_InvalidToken(t *testing.T) {
	// 一个无效的 token（这里随便给一个伪造的 token 字符串）
	invalidToken := "invalid.token.string"

	// 调用解析无效 Token 的方法
	claims, err := ParseJWTToken(invalidToken)
	assert.Error(t, err, "Parsing an invalid token should return an error")
	assert.Nil(t, claims, "Claims should be nil for invalid token")
}

func TestParseJWTToken_ExpiredToken(t *testing.T) {
	// 使用一个过期的 token 进行测试
	// 创建一个过期的 token（过期时间设定为当前时间之前）
	expirationTime := time.Now().Add(-1 * time.Hour) // 1小时前的时间
	claims := UserClaims{
		UserID:   123,
		Username: "testUser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "sayit",
		},
	}

	// 使用密钥签名，生成一个过期的 token 字符串
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	invalidTokenString, err := token.SignedString(jwtSecret)
	assert.NoError(t, err)

	// 调用解析过期 Token 的方法
	_, err = ParseJWTToken(invalidTokenString)

	// 确保返回过期错误
	assert.Error(t, err, "Parsing an expired token should return an error")
	assert.Contains(t, err.Error(), "token is expired", "Error should indicate token is expired")
}

func TestParseJWTToken_ValidToken(t *testing.T) {
	userID := int64(123)
	username := "testUser"

	// 调用创建 Token 的方法
	tokenString, err := CreateJWTToken(userID, username)
	assert.NoError(t, err, "Error should be nil when generating token")

	// 调用解析有效 Token 的方法
	claims, err := ParseJWTToken(tokenString)
	assert.NoError(t, err, "Error should be nil when parsing token")

	// 确保解析出的 Claims 和传入的值一致
	assert.Equal(t, userID, claims.UserID, "UserID should match")
	assert.Equal(t, username, claims.Username, "Username should match")
}

func TestParseJWTToken_SignatureMethodMismatch(t *testing.T) {
	// 手动创建一个错误的 Token，模拟签名密钥不一致的情况
	claims := UserClaims{
		UserID:   123,
		Username: "testUser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "sayit",
		},
	}

	// 创建一个 Token，使用正确的签名方法
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用错误的密钥签名
	invalidTokenString, err := token.SignedString([]byte("wrongSecret"))
	assert.NoError(t, err)

	// 尝试解析这个 Token
	_, err = ParseJWTToken(invalidTokenString)

	// 确保返回签名无效的错误
	assert.Error(t, err, "Parsing a token with signature method mismatch should return an error")
	assert.Contains(t, err.Error(), "signature is invalid", "Error should indicate signature is invalid")
}
