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

	v1 := r.Group("/api/v1")

	// 用户模块
	v1.POST("/signup", controller.SignupHandler) // 注册
	v1.POST("/login", controller.LoginHandler)   // 登录

	v1.Use(middlewares.JWTAuthMiddleware()) // 应用JWT认证中间件

	{
		v1.GET("/community", controller.CommunityHandler)
		v1.GET("/community/:id", controller.CommunityDetailHandler)

		v1.POST("/post", controller.CreatePostHandler)
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "404 Not Found",
		})
	})

	return r
}
