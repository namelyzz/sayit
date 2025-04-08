package security

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/namelyzz/sayit/config"
)

// HashPassword 使用 SHA256 对密码进行加盐哈希
// 算法: SHA256(password + salt)
// salt 来自配置文件 config.Conf.Secret，防止彩虹表攻击
// 返回值: 64字符的十六进制哈希字符串
func HashPassword(password string) string {
	str := password + config.Conf.Secret
	hash := sha256.New()
	hash.Write([]byte(str))
	return hex.EncodeToString(hash.Sum(nil))
}

// VerifyPassword 验证用户输入的密码是否与存储的哈希值匹配
// 原理: 将输入密码用同样的方式哈希后，与数据库中存储的哈希值进行字符串比较
func VerifyPassword(inputPassword, storedHash string) bool {
	hashedInput := HashPassword(inputPassword)
	return hashedInput == storedHash
}
