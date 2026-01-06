package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"nono-system/oracle/internal/config"
	"nono-system/oracle/internal/oracle"
)

// Server HTTP服务器
type Server struct {
	config  *config.Config
	oracle  *oracle.Oracle
	httpSrv *http.Server
}

// New 创建新的HTTP服务器
func New(cfg *config.Config, oracleService *oracle.Oracle) *Server {
	router := gin.Default()

	srv := &Server{
		config: cfg,
		oracle: oracleService,
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
	api := router.Group("/api/v1")
	{
		api.GET("/health", s.healthCheck)
		api.GET("/device/:did/status", s.getDeviceStatus)
	}
}

// healthCheck 健康检查
func (s *Server) healthCheck(c *gin.Context) {
	health := s.oracle.HealthCheck()
	c.JSON(http.StatusOK, health)
}

// getDeviceStatus 获取设备状态
func (s *Server) getDeviceStatus(c *gin.Context) {
	did := c.Param("did")
	
	status, err := s.oracle.GetDeviceStatus(did)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// Start 启动服务器
func (s *Server) Start() error {
	return s.httpSrv.ListenAndServe()
}

// Shutdown 优雅关闭服务器
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpSrv.Shutdown(ctx)
}

