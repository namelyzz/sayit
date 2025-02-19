package security

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/namelyzz/sayit/config"
)

// HashPassword 使用 SHA256 对密码进行加密，使用 salt（来自配置文件）来增加复杂性
func HashPassword(password string) string {
	str := password + config.Conf.Secret
	hash := sha256.New()
	hash.Write([]byte(str))
	return hex.EncodeToString(hash.Sum(nil))
}
