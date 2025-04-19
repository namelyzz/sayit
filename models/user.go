package models

// User 用户模型，对应数据库 `users` 表
// UserID 由雪花算法生成，全局唯一，用于外部引用
// Password 存储的是 SHA256 加盐哈希值，不是明文
// Token 字段不持久化到数据库，仅在登录成功后临时存储用于返回给客户端
type User struct {
	UserID   SnowflakeID `json:"user_id" gorm:"user_id"`
	Username string      `json:"user_name" gorm:"username"`
	Password string      `json:"-" gorm:"password"` // 不返回密码
	Token    string      `json:"token"`             // 登录成功后临时持有的 JWT Token，不入库
}
