package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"nono-system/oracle/internal/config"
	"nono-system/oracle/internal/oracle"
	"nono-system/oracle/internal/server"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化预言机服务
	oracleService, err := oracle.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize oracle service: %v", err)
	}

	// 启动HTTP服务器
	httpServer := server.New(cfg, oracleService)
	
	// 启动服务
	go func() {
		if err := httpServer.Start(); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// 启动预言机数据采集任务
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := oracleService.StartDataCollection(ctx); err != nil {
			log.Fatalf("Failed to start data collection: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down oracle service...")
	
	cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	log.Println("Oracle service stopped")
}

