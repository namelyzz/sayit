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
	v1.GET("/hot_communities", controller.HotCommunityHandler)       // 热门社区列表
	v1.GET("/random_communities", controller.RandomCommunityHandler) // 随机推荐社区
	v1.GET("/community", controller.CommunityHandler)                // 社区列表
	v1.GET("/community/:id", controller.CommunityDetailHandler)      // 社区详情
	v1.GET("/posts", controller.GetPostListHandler)                  // 帖子列表
	v1.GET("/post_detail/:id", controller.GetPostDetailHandler)      // 帖子详情
	v1.GET("/post/:id/comments", controller.GetCommentListHandler)   // 帖子评论列表
	v1.GET("/comment/:id/children", controller.GetCommentChildrenHandler) // 评论的子评论列表
	v1.GET("/users/:id/posts", controller.GetUserPostsHandler)       // 用户发布的帖子
	v1.GET("/users/:id/followers", controller.GetUserFollowersHandler)
	v1.GET("/users/:id/following", controller.GetUserFollowingHandler)
	v1.GET("/users/:id", controller.GetUserProfileHandler) // 用户公开资料

	// 用户模块
	v1.POST("/signup", controller.SignupHandler) // 注册
	v1.POST("/login", controller.LoginHandler)   // 登录

	v1.Use(middlewares.JWTAuthMiddleware()) // 应用JWT认证中间件

	{
		v1.GET("/me", controller.GetMeHandler)
		v1.PATCH("/me", controller.UpdateMeHandler)
		v1.POST("/users/:id/follow", controller.FollowUserHandler)
		v1.DELETE("/users/:id/follow", controller.UnfollowUserHandler)
		v1.GET("/users/:id/follow_status", controller.GetUserFollowStatusHandler)

		v1.POST("/create_post", controller.CreatePostHandler)
		v1.POST("/vote", controller.PostVoteController)

		// 评论相关
		v1.POST("/comment", controller.CreateCommentHandler)
		v1.DELETE("/comment/:id", controller.DeleteCommentHandler)
		v1.POST("/comment/:id/like", controller.LikeCommentHandler)
		v1.DELETE("/comment/:id/like", controller.UnlikeCommentHandler)

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
