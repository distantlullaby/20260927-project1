package main

import (
	"log"

	"memorylink/config"
	"memorylink/handlers"
	"memorylink/middleware"
	"memorylink/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	models.InitDB()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// 上传的图片与种子图
	r.Static("/uploads", config.UploadDir)

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		// 认证
		api.POST("/auth/register", handlers.Register)
		api.POST("/auth/login", handlers.Login)

		// Feed 流（公开）
		api.GET("/stories", handlers.ListStories)
		api.GET("/stories/:id", handlers.GetStory)
		api.GET("/users/:id/stories", handlers.UserStories)

		// 需要登录
		auth := api.Group("")
		auth.Use(middleware.Auth())
		{
			auth.GET("/me", handlers.Me)
			auth.GET("/profile", handlers.Profile)

			auth.POST("/upload", handlers.Upload)

			auth.POST("/stories", handlers.CreateStory)
			auth.POST("/stories/:id/bounty", handlers.AppendBounty)
			auth.POST("/stories/:id/responses", handlers.CreateResponse)
			auth.POST("/responses/:rid/accept", handlers.AcceptResponse)
		}
	}

	log.Printf("记忆连接后端已启动: http://localhost%s", config.ServerAddr)
	if err := r.Run(config.ServerAddr); err != nil {
		log.Fatal(err)
	}
}
