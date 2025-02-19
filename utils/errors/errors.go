package errors

import "github.com/pkg/errors"

var (
	ErrorUserExist = errors.New("user already exists")
)
