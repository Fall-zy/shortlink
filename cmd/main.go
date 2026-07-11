package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"shortlink/config"
	"shortlink/internal/handler"
	"shortlink/internal/model"
	"shortlink/internal/repository"
	"shortlink/internal/service"
	"shortlink/internal/utils"
	"syscall"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/libtnb/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	//加载配置
	if err := config.LoadConfig("config/config.yaml"); err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}
	cfg := config.AppConfig

	//初始化雪花ID
	utils.InitSnowflake()

	//连接数据库
	var (
		db  *gorm.DB
		err error
	)
	switch cfg.Database.Driver {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.Database.DSN), &gorm.Config{})
	case "mysql":
		log.Fatal("暂未集成")
	default:
		log.Fatalf("不支持的数据库驱动: %s", cfg.Database.Driver)
	}
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	if err := db.AutoMigrate(&model.ShortLink{}); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	//分层初始化
	repo := repository.NewShortLinkRepo(db)
	svc := service.NewShortLinkSvc(repo)
	hdl := handler.NewShortLinkHandler(svc)

	//初始化日志
	utils.InitLogger("debug")
	defer utils.Sync()

	//初始化Redis
	redisConfig := utils.RedisCfg{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	}
	utils.InitRedis(redisConfig)
	defer utils.CloseRedis()

	//设置Gin
	r := gin.New()
	r.Use(ginzap.Ginzap(utils.Logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(utils.Logger, true))
	_ = r.SetTrustedProxies(nil)

	//路由
	r.GET("/", func(c *gin.Context) {
		c.File("web/index.html")
	})

	api := r.Group("/api/v1")
	{
		api.POST("/shorten", hdl.CreateShortLink)
	}
	r.GET("/r/:code", hdl.Redirect)

	Server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}
	go func() {
		if err := Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Logger.Error("listen:", zap.Error(err))
		}
	}()
	utils.Logger.Info("短链服务启动",
		zap.String("addr", Server.Addr),
		zap.String("base_url", cfg.Server.BaseURL),
	)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	utils.Logger.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := Server.Shutdown(ctx); err != nil {
		utils.Logger.Error("Server Shutdown:", zap.Error(err))
	}
	utils.Logger.Info("Server exiting")

}
