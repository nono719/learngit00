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
	"nono-system/backend/internal/models"
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
	// 健康检查（无需认证）
	router.GET("/health", handlers.HealthCheck)

	// API路由
	api := router.Group("/api/v1")
	{
		// 用户认证（无需认证）
		api.POST("/users/register", handlers.RegisterUser(s.db))
		api.POST("/users/login", handlers.Login(s.db))

		// 需要认证的路由组
		authenticated := api.Group("")
		authenticated.Use(middleware.AuthMiddleware(s.db))
		authenticated.Use(middleware.FilterByDataPermission(s.db))
		{
			// 用户信息
			authenticated.GET("/users/me", handlers.GetCurrentUser())
			authenticated.GET("/users", middleware.RequireRole(models.RoleAdmin), handlers.ListUsers(s.db))

			// 设备管理
			devices := authenticated.Group("/devices")
			{
				// 注册设备：管理员全权限，操作人员域级权限
				devices.POST("", 
					middleware.RequirePermission(models.PermDeviceRegister, models.PermDeviceRegisterDomain),
					handlers.RegisterDevice(s.db))
				devices.POST("/batch", 
					middleware.RequirePermission(models.PermDeviceRegister, models.PermDeviceRegisterDomain),
					handlers.BatchRegisterDevices(s.db))
				
				// 查询设备：所有角色都可以查询（受数据权限限制）
				devices.GET("", 
					middleware.RequirePermission(models.PermDeviceQuery),
					handlers.ListDevices(s.db))
				devices.GET("/search", 
					middleware.RequirePermission(models.PermDeviceQuery),
					handlers.SearchDevices(s.db))
				
				// 获取设备状态列表（供预言机使用）- 必须在 /:did 之前，避免路由冲突
				devices.GET("/status", 
					middleware.RequirePermission(models.PermDeviceQuery),
					handlers.GetDeviceStatuses(s.db))
				
				devices.GET("/:did", 
					middleware.RequirePermission(models.PermDeviceQuery),
					handlers.GetDevice(s.db))
				devices.GET("/:did/history", 
					middleware.RequirePermission(models.PermDeviceQuery),
					handlers.GetDeviceHistory(s.db))
				
				// 更新设备状态：管理员和操作人员
				devices.PUT("/:did/status", 
					middleware.RequirePermission(models.PermDeviceUpdate),
					handlers.UpdateDeviceStatus(s.db))
				devices.PUT("/batch/status", 
					middleware.RequirePermission(models.PermDeviceUpdate),
					handlers.BatchUpdateDeviceStatus(s.db))
				
				// 吊销设备：仅管理员
				devices.DELETE("/:did", 
					middleware.RequirePermission(models.PermDeviceRevoke),
					handlers.RevokeDevice(s.db))
			}

			// 域管理（仅管理员）
			domains := authenticated.Group("/domains")
			domains.Use(middleware.RequirePermission(models.PermDomainQuery))
			{
				domains.POST("", 
					middleware.RequirePermission(models.PermDomainCreate),
					handlers.CreateDomain(s.db))
				domains.GET("", handlers.ListDomains(s.db))
				domains.GET("/:name", handlers.GetDomain(s.db))
				domains.PUT("/:name", 
					middleware.RequirePermission(models.PermDomainUpdate),
					handlers.UpdateDomain(s.db))
				domains.DELETE("/:name", 
					middleware.RequirePermission(models.PermDomainDelete),
					handlers.DeleteDomain(s.db))
			}

			// 跨域认证
			auth := authenticated.Group("/auth")
			{
				// 发起跨域认证：管理员和操作人员
				auth.POST("/cross-domain", 
					middleware.RequirePermission(models.PermAuthRequest),
					handlers.RequestCrossDomainAuth(s.db, s.blockchain))
				
				// 同步前端上链的认证记录：管理员和操作人员
				auth.POST("/sync", 
					middleware.RequirePermission(models.PermAuthRequest),
					handlers.SyncAuthRecord(s.db))
				
				// 查询认证记录：所有有查询权限的角色
				auth.GET("/records/:did", 
					middleware.RequirePermission(models.PermAuthQuery, models.PermAuditQuery),
					handlers.GetAuthRecords(s.db))
				auth.GET("/logs", 
					middleware.RequirePermission(models.PermAuthQuery, models.PermAuditQuery),
					handlers.GetAuthLogs(s.db))
				auth.GET("/verify/:txHash", 
					middleware.RequirePermission(models.PermAuthQuery, models.PermAuditQuery),
					handlers.VerifyTransaction(s.blockchain))
			}

			// 统计和仪表板（管理员和审计人员）
			authenticated.GET("/statistics", 
				middleware.RequirePermission(models.PermAuditStats, models.PermSystemView),
				handlers.GetStatistics(s.db))

			// 数据导出（管理员和审计人员）
			export := authenticated.Group("/export")
			export.Use(middleware.RequirePermission(models.PermAuditQuery, models.PermSystemView))
			{
				export.GET("/devices", handlers.ExportDevices(s.db))
				export.GET("/auth-records", handlers.ExportAuthRecords(s.db))
			}
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

