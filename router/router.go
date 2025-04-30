package router

import (
	"github.com/gin-gonic/gin"
	"github.com/namelyzz/sayit/controller"
	"github.com/namelyzz/sayit/middlewares"
	"net/http"
)

func SetupRouter(mode string) *gin.Engine {
	r := gin.New()
	r.Use(middlewares.GinLogger(), middlewares.GinRecovery(true), middlewares.Cors())

	v1 := r.Group("/api/v1")

	// 公开接口（无需 JWT 验证）
	v1.GET("/hot_communities", controller.HotCommunityHandler)        // 热门社区列表
	v1.GET("/random_communities", controller.RandomCommunityHandler) // 随机推荐社区
	v1.GET("/community", controller.CommunityHandler)                // 社区列表
	v1.GET("/community/:id", controller.CommunityDetailHandler)      // 社区详情
	v1.GET("/posts", controller.GetPostListHandler)                  // 帖子列表
	v1.GET("/post_detail/:id", controller.GetPostDetailHandler)      // 帖子详情

	// 用户模块
	v1.POST("/signup", controller.SignupHandler) // 注册
	v1.POST("/login", controller.LoginHandler)   // 登录

	v1.Use(middlewares.JWTAuthMiddleware()) // 应用JWT认证中间件

	{
		v1.POST("/create_post", controller.CreatePostHandler)
		v1.POST("/vote", controller.PostVoteController)

		// 关注社区相关
		v1.POST("/follow", controller.FollowCommunityHandler)
		v1.POST("/unfollow", controller.UnfollowCommunityHandler)
		v1.GET("/is_followed", controller.IsFollowedCommunityHandler)
		v1.GET("/followed_communities", controller.GetFollowedCommunityListHandler)
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "404 Not Found",
		})
	})

	return r
}
