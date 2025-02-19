package mysql

import (
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/utils/errors"
	"github.com/namelyzz/sayit/utils/security"
)

func CheckUserExist(username string) (err error) {
	var count int64
	if err = db.Model(&models.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return errors.ErrorUserExist
	}
	return nil
}

func InsertUser(user *models.User) (err error) {
	user.Password = security.HashPassword(user.Password)
	res := db.Create(user)
	return res.Error
}
