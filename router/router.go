package router

import (
	"github.com/gin-gonic/gin"
	"github.com/namelyzz/sayit/controller"
	"github.com/namelyzz/sayit/middlewares"
	"net/http"
)

func SetupRouter(mode string) *gin.Engine {
	r := gin.New()
	r.Use(middlewares.GinLogger(), middlewares.GinRecovery(true))

	// 用户模块
	r.POST("/signup", controller.SignupHandler) // 注册
	r.POST("/login", controller.LoginHandler)   // 登录

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "404 Not Found",
		})
	})

	return r
}
