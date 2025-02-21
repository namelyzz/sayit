package mysql

import (
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/utils/api"
	"github.com/namelyzz/sayit/utils/security"
	"gorm.io/gorm"
)

func CheckUserExist(username string) (err error) {
	var count int64
	if err = db.Model(&models.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return api.ErrorUserExist
	}
	return nil
}

func InsertUser(user *models.User) (err error) {
	user.Password = security.HashPassword(user.Password)
	res := db.Create(user)
	return res.Error
}

func Login(user *models.User) (err error) {
	// 这是用户输入时的密码，暂存起来
	userPwd := user.Password

	// 这一步会取出 DB 中的用户信息，将覆盖掉用户输入的数据，此时密码是经过加密的
	err = db.Where("username = ?", user.Username).First(user).Error
	if err == gorm.ErrRecordNotFound {
		return api.ErrorUserNotExist
	}
	if err != nil {
		return err
	}

	// 与暂存的密码进行比对，看是否一致
	if !security.VerifyPassword(userPwd, user.Password) {
		// 密码错误
		return api.ErrorInvalidLogin
	}

	return nil
}
