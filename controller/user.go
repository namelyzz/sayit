package controller

import (
    "github.com/gin-gonic/gin"
    "github.com/go-playground/validator/v10"
    "github.com/namelyzz/sayit/middlewares"
    "github.com/namelyzz/sayit/models"
    "github.com/namelyzz/sayit/service"
    "github.com/namelyzz/sayit/utils/errors"
    pkgerr "github.com/pkg/errors"
    "go.uber.org/zap"
    "net/http"
)

func SignupHandler(c *gin.Context) {
    p := new(models.ParamSignUp)
    if err := c.ShouldBindJSON(p); err != nil {
        zap.L().Error("SignUp with invalid param", zap.Error(err))
        errs, ok := err.(validator.ValidationErrors)
        if !ok {
            c.JSON(http.StatusOK, gin.H{"msg": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"msg": middlewares.RemoveTopStruct(errs.Translate(middlewares.GetTranslator()))})
        return
    }

    if err := service.SignUp(p); err != nil {
        if pkgerr.Is(err, errors.ErrorUserExist) {
            c.JSON(http.StatusOK, gin.H{"msg": "用户名已存在"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"msg": "注册失败"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"msg": "注册成功"})
}
