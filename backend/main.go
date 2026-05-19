// LiteWiki — 轻量级内部知识库系统
package main

import (
	"litewiki/config"
	"litewiki/database"
	"litewiki/handlers"
	bffHandlers "litewiki/handlers/bff"
	adminHandlers "litewiki/handlers/admin"
	"litewiki/middleware"
	"litewiki/services"
	"litewiki/utils"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// 日志配置
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	if err := database.Init(cfg.DBPath); err != nil {
		log.Fatal().Err(err).Msg("数据库初始化失败")
	}
	defer database.Close()

	// 执行迁移
	if err := database.Migrate(); err != nil {
		log.Fatal().Err(err).Msg("数据库迁移失败")
	}

	// 重建 FTS 索引（中文分词）
	if err := database.RebuildFTSIndex(utils.Tokenize); err != nil {
		log.Fatal().Err(err).Msg("FTS 索引重建失败")
	}

	// 清理已导入文档中的语雀链接
	if err := database.CleanExistingYuqueLinks(utils.CleanYuqueLinks, utils.Tokenize); err != nil {
		log.Warn().Err(err).Msg("清理语雀链接失败（非致命）")
	}

	// 创建初始管理员
	if err := database.Seed(cfg.AdminInitPassword); err != nil {
		log.Fatal().Err(err).Msg("初始化管理员账号失败")
	}

	// 启动限流器清理
	go middleware.CleanupVisitors()

	// 初始化导入任务队列
	services.InitImportQueue()
	defer services.GlobalQueue.Shutdown()

	// Gin 配置
	gin.SetMode(cfg.GinMode)
	r := gin.Default()
	r.SetTrustedProxies([]string{"172.16.0.0/12", "192.168.0.0/16", "10.0.0.0/8"})
	r.MaxMultipartMemory = 500 << 20 // 500MB

	// 全局中间件
	r.Use(middleware.CORS())

	// 上传文件访问（需要鉴权）
	uploadsHandler := &handlers.UploadsHandler{UploadDir: cfg.UploadDir}
	r.GET("/uploads/*filepath", middleware.Auth(cfg.JWTSecret), uploadsHandler.ServeFile)

	// ===== BFF 路由组（前端专用）=====
	bff := r.Group("/bff")
	{
		authHandler := &bffHandlers.BFFAuth{JWTSecret: cfg.JWTSecret}
		treeHandler := &bffHandlers.BFFTree{}
		docHandler := &bffHandlers.BFFDocument{}
		searchHandler := &bffHandlers.BFFSearch{}

		homeHandler := &bffHandlers.BFFHome{}
		bff.GET("/home-stats", middleware.Auth(cfg.JWTSecret), homeHandler.GetHomeStats)

		s := &bffHandlers.BFFSettings{}
		bff.GET("/settings", s.GetPublicSettings)

		// 验证码
		bff.GET("/captcha", bffHandlers.GenerateCaptcha)

		auth := bff.Group("/auth")
		{
			auth.POST("/register", middleware.RegisterRateLimit(), authHandler.Register)
			auth.POST("/login", middleware.LoginRateLimit(), authHandler.Login)
			auth.POST("/logout", middleware.Auth(cfg.JWTSecret), authHandler.Logout)
			auth.GET("/me", middleware.Auth(cfg.JWTSecret), authHandler.GetMe)
		}

		tree := bff.Group("/tree", middleware.Auth(cfg.JWTSecret))
		{
			tree.GET("", treeHandler.GetTree)
		}

		docs := bff.Group("/docs", middleware.Auth(cfg.JWTSecret))
		{
			docs.GET("", docHandler.ListAllPublished)
			docs.GET("/:id", docHandler.GetDocument)
			docs.GET("/category/:id", docHandler.ListDocsByCategory)
		}

		bff.GET("/search", middleware.Auth(cfg.JWTSecret), middleware.SearchRateLimit(), searchHandler.Search)

		creditsHandler := &bffHandlers.BFFCredits{}
		bff.GET("/credits", middleware.Auth(cfg.JWTSecret), creditsHandler.List)

		user := bff.Group("/user", middleware.Auth(cfg.JWTSecret))
		{
			user.PUT("/profile", authHandler.UpdateProfile)
			user.PUT("/password", authHandler.ChangePassword)
		}
	}

	// ===== Admin 路由组（管理员专用）=====
	admin := r.Group("/api/admin", middleware.Auth(cfg.JWTSecret), middleware.RequireRole("admin"))
	{
		userHandler := &adminHandlers.AdminUser{}
		inviteHandler := &adminHandlers.AdminInviteCode{}
		catHandler := &adminHandlers.AdminCategory{}
		docHandler := &adminHandlers.AdminDocument{}
		auditHandler := &adminHandlers.AdminAuditLog{}
		uploadHandler := &adminHandlers.AdminUpload{UploadDir: cfg.UploadDir}
		importHandler := &adminHandlers.AdminImport{UploadDir: cfg.UploadDir}
		githubImportHandler := &adminHandlers.GitHubImport{}
		importTaskHandler := &adminHandlers.AdminImportTask{UploadDir: cfg.UploadDir}
		settingsHandler := &adminHandlers.AdminSetting{}
		admin.GET("/settings", settingsHandler.GetSettings)
		admin.PUT("/settings", settingsHandler.UpdateSettings)

		creditHandler := &adminHandlers.AdminCredit{}

		// 用户管理
		admin.GET("/users", userHandler.List)
		admin.PUT("/users/:id/role", userHandler.UpdateRole)
		admin.PUT("/users/:id/status", userHandler.UpdateStatus)
		admin.DELETE("/users/:id", userHandler.Delete)

		// 邀请码管理
		admin.GET("/invites", inviteHandler.List)
		admin.POST("/invites", inviteHandler.Create)
		admin.POST("/invites/batch", inviteHandler.BatchCreate)
		admin.DELETE("/invites/:id", inviteHandler.Delete)

		// 分类管理
		admin.GET("/categories", catHandler.List)
		admin.POST("/categories", catHandler.Create)
		admin.PUT("/categories/:id", catHandler.Update)
		admin.DELETE("/categories/:id", catHandler.Delete)
		admin.POST("/categories/batch-delete", catHandler.BatchDelete)
		admin.POST("/categories/cascade-delete", catHandler.CascadeBatchDelete)
		admin.PUT("/categories/sort", catHandler.Sort)

		// 文档管理
		admin.GET("/documents", docHandler.List)
		admin.POST("/documents", docHandler.Create)
		admin.GET("/documents/:id", docHandler.Get)
		admin.PUT("/documents/:id", docHandler.Update)
		admin.DELETE("/documents/:id", docHandler.Delete)
		admin.PUT("/documents/:id/publish", docHandler.Publish)
		admin.GET("/documents/:id/versions", docHandler.GetVersions)
		admin.POST("/documents/:id/rollback/:vid", docHandler.Rollback)
		admin.POST("/documents/batch-delete", docHandler.BatchDelete)

		// 审计日志
		admin.GET("/audit-log", auditHandler.List)
		admin.GET("/audit/login-stats", auditHandler.LoginStats)

		// 文件上传
		admin.POST("/upload", uploadHandler.Upload)

		// ZIP 导入（旧版同步）
		admin.POST("/import", importHandler.Import)

		// GitHub 仓库导入（旧版同步）
		admin.POST("/import/github", githubImportHandler.Import)

		// 异步导入任务
		admin.POST("/import/async", importTaskHandler.ImportAsync)
		admin.POST("/import/batch", importTaskHandler.ImportBatch)
		admin.POST("/import/github/async", importTaskHandler.GitHubImportAsync)
		admin.GET("/import/tasks", importTaskHandler.List)
		admin.GET("/import/tasks/:id", importTaskHandler.Get)
		admin.POST("/import/tasks/:id/retry", importTaskHandler.Retry)
		admin.DELETE("/import/tasks/:id", importTaskHandler.Delete)
		admin.POST("/import/tasks/:id/cancel", importTaskHandler.Cancel)

		// 致谢管理
		admin.GET("/credits", creditHandler.List)
		admin.POST("/credits", creditHandler.Create)
		admin.PUT("/credits/:id", creditHandler.Update)
		admin.DELETE("/credits/:id", creditHandler.Delete)
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 启动服务
	addr := ":" + cfg.Port
	log.Info().Str("addr", addr).Msg("LiteWiki 启动中...")
	if err := r.Run(addr); err != nil {
		log.Fatal().Err(err).Msg("服务启动失败")
	}
}
