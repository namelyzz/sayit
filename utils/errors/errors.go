package errors

import "github.com/pkg/errors"

var (
	ErrorUserExist    = errors.New("user already exists")
	ErrorUserNotLogin = errors.New("用户未登录")
)
