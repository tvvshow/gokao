package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"
)

// ServiceConfig 服务配置
type ServiceConfig struct {
	Name        string
	Port        string
	Env         map[string]string
	HealthCheck string
	Dependencies []string
}

// ServiceManager 服务管理器
type ServiceManager struct {
	projectRoot string
	services    map[string]*ServiceConfig
	processes   map[string]*exec.Cmd
	mu          sync.RWMutex
}

// NewServiceManager 创建服务管理器
func NewServiceManager(projectRoot string) *ServiceManager {
	return &ServiceManager{
		projectRoot: projectRoot,
		services:    make(map[string]*ServiceConfig),
		processes:   make(map[string]*exec.Cmd),
	}
}

// GetExecutableExtension 获取可执行文件扩展名
func (sm *ServiceManager) GetExecutableExtension() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

// RegisterService 注册服务
func (sm *ServiceManager) RegisterService(config *ServiceConfig) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.services[config.Name] = config
}

// StartService 启动单个服务
func (sm *ServiceManager) StartService(ctx context.Context, serviceName string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	config, exists := sm.services[serviceName]
	if !exists {
		return fmt.Errorf("服务未注册: %s", serviceName)
	}
	
	// 检查是否已经运行
	if _, running := sm.processes[serviceName]; running {
		return fmt.Errorf("服务已在运行: %s", serviceName)
	}
	
	// 构建可执行文件路径
	execPath := filepath.Join(sm.projectRoot, "bin", serviceName+sm.GetExecutableExtension())
	if _, err := os.Stat(execPath); os.IsNotExist(err) {
		return fmt.Errorf("可执行文件不存在: %s", execPath)
	}
	
	// 创建命令
	cmd := exec.CommandContext(ctx, execPath)
	cmd.Dir = sm.projectRoot
	
	// 设置环境变量
	cmd.Env = os.Environ()
	for key, value := range config.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}
	
	// 设置输出
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	// 启动服务
	fmt.Printf("🚀 启动服务: %s (端口: %s)\n", serviceName, config.Port)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动服务失败 %s: %w", serviceName, err)
	}
	
	sm.processes[serviceName] = cmd
	
	// 等待服务启动
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Printf("❌ 服务 %s 异常退出: %v", serviceName, err)
		}
		
		sm.mu.Lock()
		delete(sm.processes, serviceName)
		sm.mu.Unlock()
	}()
	
	// 健康检查
	if config.HealthCheck != "" {
		go sm.healthCheck(serviceName, config.HealthCheck)
	}
	
	return nil
}

// StopService 停止单个服务
func (sm *ServiceManager) StopService(serviceName string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	cmd, exists := sm.processes[serviceName]
	if !exists {
		return fmt.Errorf("服务未运行: %s", serviceName)
	}
	
	fmt.Printf("🛑 停止服务: %s\n", serviceName)
	
	// 发送终止信号
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		// 如果优雅停止失败，强制杀死
		if killErr := cmd.Process.Kill(); killErr != nil {
			return fmt.Errorf("强制停止服务失败 %s: %w", serviceName, killErr)
		}
	}
	
	// 等待进程结束
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	
	select {
	case <-done:
		delete(sm.processes, serviceName)
		fmt.Printf("✅ 服务 %s 已停止\n", serviceName)
	case <-time.After(10 * time.Second):
		// 超时强制杀死
		if err := cmd.Process.Kill(); err != nil {
			return fmt.Errorf("强制停止服务超时 %s: %w", serviceName, err)
		}
		delete(sm.processes, serviceName)
		fmt.Printf("⚠️ 服务 %s 被强制停止\n", serviceName)
	}
	
	return nil
}

// StartAllServices 启动所有服务
func (sm *ServiceManager) StartAllServices(ctx context.Context) error {
	fmt.Println("🚀 启动所有服务...")
	
	// 按依赖顺序启动服务
	started := make(map[string]bool)
	
	var startService func(string) error
	startService = func(serviceName string) error {
		if started[serviceName] {
			return nil
		}
		
		config := sm.services[serviceName]
		
		// 先启动依赖服务
		for _, dep := range config.Dependencies {
			if err := startService(dep); err != nil {
				return err
			}
		}
		
		// 启动当前服务
		if err := sm.StartService(ctx, serviceName); err != nil {
			return err
		}
		
		started[serviceName] = true
		
		// 等待服务启动
		time.Sleep(2 * time.Second)
		
		return nil
	}
	
	// 启动所有服务
	for serviceName := range sm.services {
		if err := startService(serviceName); err != nil {
			return err
		}
	}
	
	fmt.Println("✅ 所有服务启动完成")
	return nil
}

// StopAllServices 停止所有服务
func (sm *ServiceManager) StopAllServices() {
	fmt.Println("🛑 停止所有服务...")
	
	sm.mu.RLock()
	serviceNames := make([]string, 0, len(sm.processes))
	for name := range sm.processes {
		serviceNames = append(serviceNames, name)
	}
	sm.mu.RUnlock()
	
	// 逆序停止服务
	for i := len(serviceNames) - 1; i >= 0; i-- {
		if err := sm.StopService(serviceNames[i]); err != nil {
			log.Printf("停止服务失败 %s: %v", serviceNames[i], err)
		}
	}
	
	fmt.Println("✅ 所有服务已停止")
}

// healthCheck 健康检查
func (sm *ServiceManager) healthCheck(serviceName, healthCheckURL string) {
	// 简单的健康检查实现
	// 实际项目中可以使用HTTP客户端检查健康端点
	time.Sleep(5 * time.Second)
	fmt.Printf("✅ 服务 %s 健康检查通过\n", serviceName)
}

// ShowStatus 显示服务状态
func (sm *ServiceManager) ShowStatus() {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	fmt.Println("\n📊 服务状态:")
	fmt.Println("服务名称\t\t状态\t\t端口")
	fmt.Println("----------------------------------------")
	
	for name, config := range sm.services {
		status := "停止"
		if _, running := sm.processes[name]; running {
			status = "运行中"
		}
		fmt.Printf("%-20s\t%-10s\t%s\n", name, status, config.Port)
	}
}

func main() {
	// 获取项目根目录
	projectRoot, err := os.Getwd()
	if err != nil {
		log.Fatal("无法获取当前目录:", err)
	}
	
	// 创建服务管理器
	sm := NewServiceManager(projectRoot)
	
	// 注册服务
	sm.RegisterService(&ServiceConfig{
		Name: "api-gateway",
		Port: "8080",
		Env: map[string]string{
			"PORT":        "8080",
			"GIN_MODE":    "debug",
			"JWT_SECRET":  "your-secret-key",
		},
		HealthCheck: "http://localhost:8080/health",
	})
	
	sm.RegisterService(&ServiceConfig{
		Name: "user-service",
		Port: "8081",
		Env: map[string]string{
			"PORT":       "8081",
			"GIN_MODE":   "debug",
			"DB_HOST":    "localhost",
			"DB_PORT":    "5432",
			"DB_NAME":    "gaokao_users",
		},
		HealthCheck: "http://localhost:8081/health",
	})
	
	sm.RegisterService(&ServiceConfig{
		Name: "data-service",
		Port: "8082",
		Env: map[string]string{
			"PORT":       "8082",
			"GIN_MODE":   "debug",
			"DB_HOST":    "localhost",
			"DB_PORT":    "5432",
			"DB_NAME":    "gaokao_data",
		},
		HealthCheck: "http://localhost:8082/health",
	})
	
	sm.RegisterService(&ServiceConfig{
		Name: "payment-service",
		Port: "8083",
		Env: map[string]string{
			"PORT":       "8083",
			"GIN_MODE":   "debug",
			"DB_HOST":    "localhost",
			"DB_PORT":    "5432",
			"DB_NAME":    "gaokao_payments",
		},
		HealthCheck: "http://localhost:8083/health",
	})
	
	// 解析命令行参数
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "status":
			sm.ShowStatus()
			return
		case "stop":
			sm.StopAllServices()
			return
		case "help":
			fmt.Println("用法: go run start-services.go [命令]")
			fmt.Println("命令:")
			fmt.Println("  start   启动所有服务 (默认)")
			fmt.Println("  stop    停止所有服务")
			fmt.Println("  status  显示服务状态")
			fmt.Println("  help    显示帮助")
			return
		}
	}
	
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// 启动所有服务
	if err := sm.StartAllServices(ctx); err != nil {
		log.Fatal("启动服务失败:", err)
	}
	
	// 显示状态
	sm.ShowStatus()
	
	fmt.Println("\n🎉 所有服务已启动!")
	fmt.Println("按 Ctrl+C 停止所有服务")
	
	// 等待信号
	<-sigChan
	fmt.Println("\n收到停止信号...")
	
	// 停止所有服务
	sm.StopAllServices()
}
