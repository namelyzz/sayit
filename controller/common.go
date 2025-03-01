package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/namelyzz/sayit/middlewares"
	"github.com/namelyzz/sayit/utils/api"
)

func handleBindError(c *gin.Context, err error) {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}
	api.ResponseErrorWithMsg(
		c,
		api.CodeInvalidParam,
		middlewares.RemoveTopStruct(errs.Translate(middlewares.GetTranslator())),
	)
}
