package api

import "github.com/pkg/errors"

var (
	ErrorUserExist    = errors.New("用户已存在")
	ErrorUserNotExist = errors.New("用户不存在")
	ErrorUserNotLogin = errors.New("用户未登录")
	ErrorInvalidLogin = errors.New("用户名或密码错误")
)
