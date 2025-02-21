package service

import (
	"github.com/namelyzz/sayit/dao/mysql"
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/utils/jwt"
	"github.com/namelyzz/sayit/utils/snowflake"
)

func SignUp(p *models.ParamSignUp) (err error) {
	// 1. 先判断用户是否存在
	if err = mysql.CheckUserExist(p.Username); err != nil {
		return err
	}

	// 2. 通过雪花算法生成用户 ID, 然后构造用户数据
	userID := snowflake.GenID()
	user := &models.User{
		UserID:   userID,
		Username: p.Username,
		Password: p.Password,
	}

	// 3. 入库
	return mysql.InsertUser(user)
}

func Login(p *models.ParamLogin) (user *models.User, err error) {
	user = &models.User{Username: p.Username, Password: p.Password}
	if err = mysql.Login(user); err != nil {
		return nil, err
	}

	token, err := jwt.CreateJWTToken(user.UserID, user.Username)
	if err != nil {
		return nil, err
	}
	user.Token = token
	return user, nil
}
