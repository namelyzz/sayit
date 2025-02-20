package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/namelyzz/sayit/middlewares"
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/service"
	"github.com/namelyzz/sayit/utils/errors"
	pkgerr "github.com/pkg/errors"
	"go.uber.org/zap"
)

func SignupHandler(c *gin.Context) {
	p := new(models.ParamSignUp)
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Error("SignUp with invalid param", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}

		ResponseErrorWithMsg(
			c,
			CodeInvalidParam,
			middlewares.RemoveTopStruct(errs.Translate(middlewares.GetTranslator())),
		)
		return
	}

	if err := service.SignUp(p); err != nil {
		if pkgerr.Is(err, errors.ErrorUserExist) {
			ResponseErrorWithMsg(c, CodeUserExist, "用户名已存在")
			return
		}

		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccess(c, nil)
}

func LoginHandler(c *gin.Context) {
	p := new(models.ParamLogin)
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Error("Login with invalid param", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}

		ResponseErrorWithMsg(
			c,
			CodeInvalidParam,
			middlewares.RemoveTopStruct(errs.Translate(middlewares.GetTranslator())),
		)
		return
	}

	user, err := service.Login(p)
	if err != nil {
		zap.L().Error("login failed", zap.String("username", p.Username), zap.Error(err))
		if pkgerr.Is(err, errors.ErrorUserNotExist) {
			ResponseError(c, CodeUserNotExist)
			return
		}
		if pkgerr.Is(err, errors.ErrorInvalidLogin) {
			ResponseError(c, CodeInvalidPassword)
			return
		}

		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccess(c, gin.H{
		"user_id":   fmt.Sprintf("%d", user.UserID),
		"user_name": user.Username,
	})
}
