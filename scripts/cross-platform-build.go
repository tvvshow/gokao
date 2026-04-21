package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// BuildConfig 构建配置
type BuildConfig struct {
	ProjectRoot string
	Services    []string
	BuildDir    string
	CGOEnabled  bool
	Release     bool
	SkipTests   bool
	SkipDocker  bool
	Verbose     bool
}

// PlatformBuilder 跨平台构建器
type PlatformBuilder struct {
	config *BuildConfig
}

// NewPlatformBuilder 创建构建器
func NewPlatformBuilder(config *BuildConfig) *PlatformBuilder {
	return &PlatformBuilder{config: config}
}

// GetExecutableExtension 获取可执行文件扩展名
func (pb *PlatformBuilder) GetExecutableExtension() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

// GetShellCommand 获取shell命令
func (pb *PlatformBuilder) GetShellCommand() (string, []string) {
	if runtime.GOOS == "windows" {
		return "cmd", []string{"/c"}
	}
	return "sh", []string{"-c"}
}

// RunCommand 运行命令
func (pb *PlatformBuilder) RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = pb.config.ProjectRoot

	if pb.config.Verbose {
		fmt.Printf("Running: %s %s\n", name, strings.Join(args, " "))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}

// CheckDependencies 检查依赖
func (pb *PlatformBuilder) CheckDependencies() error {
	fmt.Println("🔍 检查依赖...")

	// 检查Go
	if err := pb.RunCommand("go", "version"); err != nil {
		return fmt.Errorf("Go未安装或不可用: %w", err)
	}

	// 检查Git
	if err := pb.RunCommand("git", "--version"); err != nil {
		log.Println("⚠️ Git未安装，版本信息将使用默认值")
	}

	// 检查Docker（可选）
	if !pb.config.SkipDocker {
		if err := pb.RunCommand("docker", "--version"); err != nil {
			log.Println("⚠️ Docker未安装，将跳过Docker构建")
			pb.config.SkipDocker = true
		}
	}

	fmt.Println("✅ 依赖检查完成")
	return nil
}

// CreateBuildDir 创建构建目录
func (pb *PlatformBuilder) CreateBuildDir() error {
	buildDir := filepath.Join(pb.config.ProjectRoot, pb.config.BuildDir)
	return os.MkdirAll(buildDir, 0755)
}

// BuildService 构建单个服务
func (pb *PlatformBuilder) BuildService(serviceName string) error {
	fmt.Printf("🔨 构建服务: %s\n", serviceName)

	serviceDir := filepath.Join(pb.config.ProjectRoot, "services", serviceName)
	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		return fmt.Errorf("服务目录不存在: %s", serviceDir)
	}

	// 检查main.go是否存在
	mainFile := filepath.Join(serviceDir, "main.go")
	if _, err := os.Stat(mainFile); os.IsNotExist(err) {
		// 尝试cmd目录
		mainFile = filepath.Join(serviceDir, "cmd", serviceName, "main.go")
		if _, err := os.Stat(mainFile); os.IsNotExist(err) {
			return fmt.Errorf("找不到main.go文件: %s", serviceName)
		}
	}

	// 设置环境变量
	env := os.Environ()

	// 在Linux环境下，对于user-service禁用CGO以使用存根实现
	if runtime.GOOS == "linux" && serviceName == "user-service" {
		env = append(env, "CGO_ENABLED=0")
	} else if pb.config.CGOEnabled {
		env = append(env, "CGO_ENABLED=1")
	} else {
		env = append(env, "CGO_ENABLED=0")
	}

	// 构建参数
	outputPath := filepath.Join(pb.config.ProjectRoot, pb.config.BuildDir, serviceName+pb.GetExecutableExtension())

	args := []string{"build"}
	if pb.config.Release {
		args = append(args, "-ldflags", "-s -w")
	}
	if pb.config.Verbose {
		args = append(args, "-v")
	}
	args = append(args, "-o", outputPath, ".")

	// 执行构建
	cmd := exec.Command("go", args...)
	cmd.Dir = serviceDir
	cmd.Env = env

	if pb.config.Verbose {
		fmt.Printf("  命令: go %s\n", strings.Join(args, " "))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("构建失败 %s: %w", serviceName, err)
	}

	fmt.Printf("  ✅ %s 构建完成\n", serviceName)
	return nil
}

// BuildAllServices 构建所有服务
func (pb *PlatformBuilder) BuildAllServices() error {
	fmt.Println("🚀 开始构建所有服务...")

	for _, service := range pb.config.Services {
		if err := pb.BuildService(service); err != nil {
			return err
		}
	}

	fmt.Println("✅ 所有服务构建完成")
	return nil
}

// RunTests 运行测试
func (pb *PlatformBuilder) RunTests() error {
	if pb.config.SkipTests {
		return nil
	}

	fmt.Println("🧪 运行测试...")

	for _, service := range pb.config.Services {
		serviceDir := filepath.Join(pb.config.ProjectRoot, "services", service)
		if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
			continue
		}

		fmt.Printf("  测试服务: %s\n", service)

		cmd := exec.Command("go", "test", "-v", "./...")
		cmd.Dir = serviceDir

		if pb.config.Verbose {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}

		if err := cmd.Run(); err != nil {
			log.Printf("⚠️ 服务 %s 测试失败: %v", service, err)
		}
	}

	fmt.Println("✅ 测试完成")
	return nil
}

// ShowResults 显示构建结果
func (pb *PlatformBuilder) ShowResults() {
	fmt.Println("\n📋 构建结果:")

	buildDir := filepath.Join(pb.config.ProjectRoot, pb.config.BuildDir)
	files, err := os.ReadDir(buildDir)
	if err != nil {
		log.Printf("无法读取构建目录: %v", err)
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			fmt.Printf("  - %s\n", file.Name())
		}
	}
}

// Build 执行完整构建流程
func (pb *PlatformBuilder) Build() error {
	fmt.Printf("🚀 高考志愿填报系统 - 跨平台构建\n")
	fmt.Printf("平台: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("模式: %s\n", map[bool]string{true: "Release", false: "Debug"}[pb.config.Release])

	// 检查依赖
	if err := pb.CheckDependencies(); err != nil {
		return err
	}

	// 创建构建目录
	if err := pb.CreateBuildDir(); err != nil {
		return fmt.Errorf("创建构建目录失败: %w", err)
	}

	// 构建服务
	if err := pb.BuildAllServices(); err != nil {
		return err
	}

	// 运行测试
	if err := pb.RunTests(); err != nil {
		return err
	}

	// 显示结果
	pb.ShowResults()

	fmt.Println("\n🎉 构建完成!")
	return nil
}

func main() {
	// 获取项目根目录
	projectRoot, err := os.Getwd()
	if err != nil {
		log.Fatal("无法获取当前目录:", err)
	}

	// 默认配置
	config := &BuildConfig{
		ProjectRoot: projectRoot,
		Services: []string{
			"api-gateway",
			"user-service",
			"data-service",
			"payment-service",
		},
		BuildDir:   "bin",
		CGOEnabled: true,
		Release:    false,
		SkipTests:  false,
		SkipDocker: true,
		Verbose:    false,
	}

	// 解析命令行参数
	for i, arg := range os.Args[1:] {
		switch arg {
		case "--release":
			config.Release = true
		case "--skip-tests":
			config.SkipTests = true
		case "--skip-docker":
			config.SkipDocker = true
		case "--verbose", "-v":
			config.Verbose = true
		case "--disable-cgo":
			config.CGOEnabled = false
		case "--help", "-h":
			fmt.Println("用法: go run cross-platform-build.go [选项]")
			fmt.Println("选项:")
			fmt.Println("  --release      发布模式构建")
			fmt.Println("  --skip-tests   跳过测试")
			fmt.Println("  --skip-docker  跳过Docker构建")
			fmt.Println("  --disable-cgo  禁用CGO")
			fmt.Println("  --verbose, -v  详细输出")
			fmt.Println("  --help, -h     显示帮助")
			return
		default:
			if i == 0 && !strings.HasPrefix(arg, "--") {
				// 第一个参数可能是服务名
				config.Services = []string{arg}
			}
		}
	}

	// 创建构建器并执行构建
	builder := NewPlatformBuilder(config)
	if err := builder.Build(); err != nil {
		log.Fatal("构建失败:", err)
	}
}
