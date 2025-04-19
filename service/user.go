package service

import (
	"github.com/namelyzz/sayit/dao/mysql"
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/utils/jwt"
	"github.com/namelyzz/sayit/utils/snowflake"
)

// SignUp 用户注册业务逻辑
// 步骤: 检查用户名是否已存在 -> 生成雪花ID -> 构造用户对象 -> 密码加密后入库
func SignUp(p *models.ParamSignUp) (err error) {
	// 1. 先判断用户是否存在（用户名唯一性校验）
	if err = mysql.CheckUserExist(p.Username); err != nil {
		return err
	}

	// 2. 通过雪花算法生成全局唯一用户 ID，然后构造用户数据
	// 雪花ID相比自增ID的优势: 全局唯一、趋势递增、不暴露用户数量、支持分布式
	userID := snowflake.GenID()
	user := &models.User{
		UserID:   models.SnowflakeID(userID),
		Username: p.Username,
		Password: p.Password,
	}

	// 3. 入库（dao 层会对密码进行 SHA256 加密后存储）
	return mysql.InsertUser(user)
}

// Login 用户登录业务逻辑
// 步骤: 从数据库查询用户并验证密码 -> 生成 JWT Token -> 返回用户信息
func Login(p *models.ParamLogin) (user *models.User, err error) {
	// 1. 构造用户对象，调用 dao 层验证用户名和密码
	user = &models.User{Username: p.Username, Password: p.Password}
	if err = mysql.Login(user); err != nil {
		return nil, err
	}

	// 2. 密码验证通过，生成 JWT Token（有效期1小时）
	token, err := jwt.CreateJWTToken(int64(user.UserID), user.Username)
	if err != nil {
		return nil, err
	}
	// 3. 将 token 附加到用户对象返回给 controller 层
	user.Token = token
	return user, nil
}
