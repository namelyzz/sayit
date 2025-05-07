package models

import "time"

// User 用户模型，对应数据库 `users` 表
// UserID 由雪花算法生成，全局唯一，用于外部引用
// Password 存储的是 SHA256 加盐哈希值，不是明文
// Token 字段不持久化到数据库，仅在登录成功后临时存储用于返回给客户端
type User struct {
	UserID     SnowflakeID `json:"user_id" gorm:"column:user_id"`
	Username   string      `json:"user_name" gorm:"column:username"`
	Password   string      `json:"-" gorm:"column:password"` // 不返回密码
	Signature  string      `json:"signature" gorm:"column:signature"`
	CreateTime time.Time   `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time   `json:"-" gorm:"column:update_time"`
	Token      string      `json:"token" gorm:"-"` // 登录成功后临时持有的 JWT Token，不入库
}

func (User) TableName() string {
	return "users"
}

type UserProfile struct {
	UserID     SnowflakeID `json:"user_id"`
	Username   string      `json:"user_name"`
	Signature  string      `json:"signature"`
	CreateTime time.Time   `json:"create_time"`
	PostCount  int64       `json:"post_count"`
	PostScore  int64       `json:"post_score"`
	IsSelf     bool        `json:"is_self"`
}
