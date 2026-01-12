package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// Service 服务结构
type Service struct {
	Name     string
	Command  string
	Args     []string
	Dir      string
	Color    string // ANSI颜色代码
	cmd      *exec.Cmd
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	output   *os.File
}

// 颜色代码
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
)

func main() {
	// 获取项目根目录
	rootDir, err := getRootDir()
	if err != nil {
		log.Fatalf("无法获取项目根目录: %v", err)
	}

	fmt.Println("=========================================")
	fmt.Println("  物联网设备跨域身份认证系统")
	fmt.Println("  统一启动服务")
	fmt.Println("=========================================")
	fmt.Println()

	// 创建服务列表
	services := []*Service{
		{
			Name:    "预言机服务",
			Command: "go",
			Args:    []string{"run", "./cmd/oracle"},
			Dir:     filepath.Join(rootDir, "oracle"),
			Color:   ColorCyan,
		},
		{
			Name:    "后端API服务",
			Command: "go",
			Args:    []string{"run", "./cmd/server"},
			Dir:     filepath.Join(rootDir, "backend"),
			Color:   ColorGreen,
		},
		{
			Name:    "前端开发服务器",
			Command: "npm",
			Args:    []string{"run", "dev"},
			Dir:     filepath.Join(rootDir, "frontend"),
			Color:   ColorYellow,
		},
	}

	// 检查依赖
	if err := checkDependencies(); err != nil {
		log.Fatalf("依赖检查失败: %v", err)
	}

	// 检查前端依赖
	frontendDir := filepath.Join(rootDir, "frontend")
	nodeModulesDir := filepath.Join(frontendDir, "node_modules")
	if _, err := os.Stat(nodeModulesDir); os.IsNotExist(err) {
		fmt.Println("警告: 前端依赖未安装，正在自动安装...")
		fmt.Println("如果安装失败，请手动运行: cd frontend && npm install")
		fmt.Println()
		
		cmd := exec.Command("npm", "install")
		cmd.Dir = frontendDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("前端依赖安装失败: %v\n请手动运行: cd frontend && npm install", err)
		}
		fmt.Println("前端依赖安装完成")
		fmt.Println()
	}

	// 创建上下文用于取消所有服务
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动所有服务
	var wg sync.WaitGroup
	for _, service := range services {
		wg.Add(1)
		go func(svc *Service) {
			defer wg.Done()
			if err := startService(ctx, svc); err != nil {
				log.Printf("%s[%s]%s 启动失败: %v", svc.Color, svc.Name, ColorReset, err)
			}
		}(service)
	}

	// 等待所有服务启动
	time.Sleep(2 * time.Second)

	fmt.Println()
	fmt.Println("=========================================")
	fmt.Println("  所有服务已启动")
	fmt.Println("  按 Ctrl+C 停止所有服务")
	fmt.Println("=========================================")
	fmt.Println()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println()
	fmt.Println("=========================================")
	fmt.Println("  正在停止所有服务...")
	fmt.Println("=========================================")
	fmt.Println()

	// 取消所有服务
	cancel()

	// 停止所有服务
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()

	for _, service := range services {
		if service.cmd != nil && service.cmd.Process != nil {
			fmt.Printf("%s[%s]%s 正在停止...\n", service.Color, service.Name, ColorReset)
			
			// 发送SIGTERM信号
			if err := service.cmd.Process.Signal(syscall.SIGTERM); err != nil {
				fmt.Printf("%s[%s]%s 发送停止信号失败: %v\n", service.Color, service.Name, ColorReset, err)
				// 如果发送信号失败，直接kill
				service.cmd.Process.Kill()
				continue
			}
			
			// 等待进程结束，最多等待5秒
			done := make(chan error, 1)
			go func(cmd *exec.Cmd) {
				done <- cmd.Wait()
			}(service.cmd)
			
			select {
			case <-done:
				fmt.Printf("%s[%s]%s 已停止\n", service.Color, service.Name, ColorReset)
			case <-stopCtx.Done():
				fmt.Printf("%s[%s]%s 超时，强制停止...\n", service.Color, service.Name, ColorReset)
				service.cmd.Process.Kill()
				<-done // 等待kill完成
				fmt.Printf("%s[%s]%s 已强制停止\n", service.Color, service.Name, ColorReset)
			}
		}
	}

	// 等待所有goroutine完成
	wg.Wait()

	fmt.Println()
	fmt.Println("=========================================")
	fmt.Println("  所有服务已停止")
	fmt.Println("=========================================")
}

// getRootDir 获取项目根目录
func getRootDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// 查找包含go.mod的目录
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return wd, nil
}

// checkDependencies 检查依赖
func checkDependencies() error {
	deps := map[string]string{
		"go":   "Go 1.19+",
		"node": "Node.js 16+",
		"npm":  "npm",
	}

	fmt.Println("检查依赖...")
	for cmd, desc := range deps {
		if _, err := exec.LookPath(cmd); err != nil {
			return fmt.Errorf("未找到 %s (%s)，请先安装", cmd, desc)
		}
		fmt.Printf("  ✓ %s 已安装\n", desc)
	}
	fmt.Println()
	return nil
}

// startService 启动服务
func startService(ctx context.Context, service *Service) error {
	// 检查目录是否存在
	if _, err := os.Stat(service.Dir); os.IsNotExist(err) {
		return fmt.Errorf("目录不存在: %s", service.Dir)
	}

	// 创建命令上下文
	cmdCtx, cancel := context.WithCancel(ctx)
	service.cancel = cancel

	// 创建命令
	cmd := exec.CommandContext(cmdCtx, service.Command, service.Args...)
	cmd.Dir = service.Dir
	service.cmd = cmd

	// 设置环境变量（继承当前环境）
	cmd.Env = os.Environ()

	// 创建管道用于捕获输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建stdout管道失败: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("创建stderr管道失败: %v", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动命令失败: %v", err)
	}

	fmt.Printf("%s[%s]%s 已启动 (PID: %d)\n", service.Color, service.Name, ColorReset, cmd.Process.Pid)

	// 启动goroutine读取输出
	service.wg.Add(2)
	go func() {
		defer service.wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Printf("%s[%s]%s %s\n", service.Color, service.Name, ColorReset, line)
		}
	}()

	go func() {
		defer service.wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Printf("%s[%s]%s %s\n", service.Color, service.Name, ColorReset, line)
		}
	}()

	// 等待命令完成（在后台）
	go func() {
		err := cmd.Wait()
		if err != nil && ctx.Err() == nil {
			// 如果不是因为上下文取消而退出，说明服务异常退出
			fmt.Printf("%s[%s]%s 异常退出: %v\n", service.Color, service.Name, ColorReset, err)
		}
	}()

	return nil
}

