package main

import (
	"log"
	"net/http"
	"os"

	"memoryconnect/config"
	"memoryconnect/database"
	"memoryconnect/handlers"
	"memoryconnect/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		log.Fatalf("创建上传目录失败：%v", err)
	}

	db := database.Init(cfg)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(middleware.CORS())

	// 静态图片：/uploads/xxx -> 本地 uploads 目录
	r.Static("/uploads", cfg.UploadDir)

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true, "slogan": "你帮我再看一眼，我把记忆还给你"})
	})

	authH := &handlers.AuthHandler{DB: db, Cfg: cfg}
	storyH := &handlers.StoryHandler{DB: db, Cfg: cfg}
	userH := &handlers.UserHandler{DB: db}
	uploadH := &handlers.UploadHandler{Cfg: cfg}

	api := r.Group("/api")
	{
		// 认证
		api.POST("/auth/register", authH.Register)
		api.POST("/auth/login", authH.Login)

		// Feed 流（游客可看）
		api.GET("/stories", middleware.OptionalAuth(db), storyH.List)

		// 登录后操作
		authed := api.Group("", middleware.RequiredAuth(db))
		{
			authed.POST("/stories", storyH.Create)
			authed.POST("/stories/:id/reward", storyH.AppendReward)
			authed.POST("/stories/:id/responses", storyH.CreateResponse)
			authed.POST("/stories/:id/responses/:rid/accept", storyH.Accept)

			authed.GET("/me", userH.Me)
			authed.GET("/me/ledgers", userH.Ledger)
			authed.GET("/me/stories", userH.MyStories)
			authed.GET("/me/responses", userH.MyResponses)

			authed.POST("/uploads", uploadH.Upload)
		}
	}

	srv := &http.Server{Addr: ":" + cfg.ServerPort, Handler: r}
	log.Printf("记忆连接后端已启动：http://localhost:%s （演示账号 alin/xiaoman/oldchen，密码 123456）", cfg.ServerPort)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
