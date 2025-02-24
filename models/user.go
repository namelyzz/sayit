package models

type User struct {
	UserID   int64  `gorm:"user_id"`
	Username string `gorm:"username"`
	Password string `gorm:"password"`

	Token string
}
