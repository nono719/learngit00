package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"nono-system/backend/internal/blockchain"
	"nono-system/backend/internal/config"
	"nono-system/backend/internal/handlers"
	"nono-system/backend/internal/middleware"
)

// Server HTTP服务器
type Server struct {
	config         *config.Config
	db             *gorm.DB
	blockchain     *blockchain.Client
	httpSrv        *http.Server
}

// New 创建新的HTTP服务器
func New(cfg *config.Config, db *gorm.DB) *Server {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// 中间件
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())

	// 初始化区块链客户端
	var bcClient *blockchain.Client
	bcClient, err := blockchain.NewClient(cfg.Blockchain)
	if err != nil {
		log.Printf("Warning: failed to initialize blockchain client: %v (blockchain features will be disabled)", err)
	} else if bcClient != nil {
		log.Printf("Blockchain client initialized successfully")
	}

	srv := &Server{
		config:     cfg,
		db:         db,
		blockchain: bcClient,
	}

	// 注册路由
	srv.registerRoutes(router)

	srv.httpSrv = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	return srv
}

// registerRoutes 注册路由
func (s *Server) registerRoutes(router *gin.Engine) {
	// 健康检查
	router.GET("/health", handlers.HealthCheck)

	// API路由
	api := router.Group("/api/v1")
	{
		// 设备管理
		devices := api.Group("/devices")
		{
			devices.POST("", handlers.RegisterDevice(s.db))
			devices.GET("", handlers.ListDevices(s.db))  // 必须在 /:did 之前
			devices.GET("/:did", handlers.GetDevice(s.db))
			devices.PUT("/:did/status", handlers.UpdateDeviceStatus(s.db))
			devices.DELETE("/:did", handlers.RevokeDevice(s.db))
		}

		// 域管理
		domains := api.Group("/domains")
		{
			domains.POST("", handlers.CreateDomain(s.db))
			domains.GET("", handlers.ListDomains(s.db))
			domains.GET("/:name", handlers.GetDomain(s.db))
		}

		// 跨域认证
		auth := api.Group("/auth")
		{
			auth.POST("/cross-domain", handlers.RequestCrossDomainAuth(s.db, s.blockchain))
			auth.GET("/records/:did", handlers.GetAuthRecords(s.db))
			auth.GET("/logs", handlers.GetAuthLogs(s.db))
			auth.GET("/verify/:txHash", handlers.VerifyTransaction(s.blockchain))
		}
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	return s.httpSrv.ListenAndServe()
}

// Shutdown 优雅关闭服务器
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpSrv.Shutdown(ctx)
}

