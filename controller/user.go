package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/namelyzz/sayit/middlewares"
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/service"
	"github.com/namelyzz/sayit/utils/api"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"strconv"
)

// SignupHandler 用户注册接口
// 路由: POST /api/v1/signup
// 请求体: JSON { "username": "xxx", "password": "xxx", "re_password": "xxx" }
// 流程: 参数校验 -> 调用 service.SignUp -> 返回结果
func SignupHandler(c *gin.Context) {
	// 1. 绑定并校验 JSON 请求参数（通过 binding tag 进行参数验证）
	p := new(models.ParamSignUp)
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Error("SignUp with invalid param", zap.Error(err))
		// 将验证错误转换为用户友好的中文提示
		handleBindError(c, err)
		return
	}

	// 2. 调用 service 层执行注册逻辑
	if err := service.SignUp(p); err != nil {
		// 用户名已存在的业务错误，返回特定提示
		if errors.Is(err, api.ErrorUserExist) {
			api.ResponseErrorWithMsg(c, api.CodeUserExist, "用户名已存在")
			return
		}
		// 其他未知错误统一返回服务器繁忙
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	// 3. 注册成功，返回空数据（客户端根据 code 判断成功）
	api.ResponseSuccess(c, nil)
}

// LoginHandler 用户登录接口
// 路由: POST /api/v1/login
// 请求体: JSON { "username": "xxx", "password": "xxx" }
// 流程: 参数校验 -> 调用 service.Login -> 返回 user_id + user_name + JWT token
func LoginHandler(c *gin.Context) {
	// 1. 绑定并校验 JSON 请求参数
	p := new(models.ParamLogin)
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Error("Login with invalid param", zap.Error(err))
		// 判断是否为验证器错误，如果不是则返回通用参数错误
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			api.ResponseError(c, api.CodeInvalidParam)
			return
		}
		// 将验证器错误中的结构体前缀去除，生成更友好的错误信息
		api.ResponseErrorWithMsg(
			c,
			api.CodeInvalidParam,
			middlewares.RemoveTopStruct(errs.Translate(middlewares.GetTranslator())),
		)
		return
	}

	// 2. 调用 service 层执行登录逻辑，返回用户信息（含 JWT token）
	user, err := service.Login(p)
	if err != nil {
		zap.L().Error("login failed", zap.String("username", p.Username), zap.Error(err))
		// 用户不存在
		if errors.Is(err, api.ErrorUserNotExist) {
			api.ResponseError(c, api.CodeUserNotExist)
			return
		}
		// 密码错误
		if errors.Is(err, api.ErrorInvalidLogin) {
			api.ResponseError(c, api.CodeInvalidPassword)
			return
		}
		// 其他未知错误
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	// 3. 登录成功，返回用户ID、用户名和JWT令牌
	api.ResponseSuccess(c, user)
}

// GetUserProfileHandler 获取用户资料页核心公开信息。
// 路由: GET /api/v1/users/:id
func GetUserProfileHandler(c *gin.Context) {
	targetUserID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	profile, err := service.GetUserProfile(0, targetUserID)
	if err != nil {
		if errors.Is(err, api.ErrorUserNotExist) {
			api.ResponseError(c, api.CodeUserNotExist)
			return
		}
		zap.L().Error("get user profile failed",
			zap.Int64("target_user_id", targetUserID),
			zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	api.ResponseSuccess(c, profile)
}

// GetMeHandler 获取当前登录用户资料。
// 路由: GET /api/v1/me (需要JWT认证)
func GetMeHandler(c *gin.Context) {
	currentUserID, err := api.GetCurrentUserID(c)
	if err != nil {
		api.ResponseError(c, api.CodeNeedLogin)
		return
	}

	profile, err := service.GetUserProfile(currentUserID, currentUserID)
	if err != nil {
		if errors.Is(err, api.ErrorUserNotExist) {
			api.ResponseError(c, api.CodeUserNotExist)
			return
		}
		zap.L().Error("get current user profile failed", zap.Int64("user_id", currentUserID), zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	api.ResponseSuccess(c, profile)
}

// UpdateMeHandler 更新当前登录用户资料。
// 路由: PATCH /api/v1/me (需要JWT认证)
func UpdateMeHandler(c *gin.Context) {
	p := new(models.ParamUpdateProfile)
	if err := c.ShouldBindJSON(p); err != nil {
		handleBindError(c, err)
		return
	}

	currentUserID, err := api.GetCurrentUserID(c)
	if err != nil {
		api.ResponseError(c, api.CodeNeedLogin)
		return
	}

	profile, err := service.UpdateUserProfile(currentUserID, p)
	if err != nil {
		if errors.Is(err, api.ErrorUserNotExist) {
			api.ResponseError(c, api.CodeUserNotExist)
			return
		}
		zap.L().Error("update current user profile failed", zap.Int64("user_id", currentUserID), zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	api.ResponseSuccess(c, profile)
}
