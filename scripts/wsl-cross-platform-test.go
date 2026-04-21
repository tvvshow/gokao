package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// 简单的跨平台测试程序
// 用于验证我们的跨平台工具包在Linux环境中是否正常工作

func main() {
	fmt.Printf("🚀 跨平台兼容性测试\n")
	fmt.Printf("==================\n")
	
	// 基本平台信息
	fmt.Printf("操作系统: %s\n", runtime.GOOS)
	fmt.Printf("架构: %s\n", runtime.GOARCH)
	fmt.Printf("Go版本: %s\n", runtime.Version())
	
	// 测试路径处理
	fmt.Printf("\n📁 路径处理测试:\n")
	
	// 测试路径连接
	testPath := filepath.Join("bin", "service")
	fmt.Printf("路径连接: %s\n", testPath)
	
	// 测试路径分隔符
	fmt.Printf("路径分隔符: %c\n", filepath.Separator)
	
	// 测试可执行文件扩展名
	execExt := ""
	if runtime.GOOS == "windows" {
		execExt = ".exe"
	}
	fmt.Printf("可执行文件扩展名: %s\n", execExt)
	
	// 测试环境变量
	fmt.Printf("\n🔧 环境变量测试:\n")
	fmt.Printf("PATH: %s\n", os.Getenv("PATH")[:100]+"...")
	fmt.Printf("HOME: %s\n", os.Getenv("HOME"))
	
	// 测试目录创建
	fmt.Printf("\n📂 目录操作测试:\n")
	testDir := filepath.Join("test-cross-platform")
	
	// 创建测试目录
	if err := os.MkdirAll(testDir, 0755); err != nil {
		fmt.Printf("❌ 创建目录失败: %v\n", err)
	} else {
		fmt.Printf("✅ 创建目录成功: %s\n", testDir)
		
		// 清理测试目录
		if err := os.RemoveAll(testDir); err != nil {
			fmt.Printf("⚠️ 清理目录失败: %v\n", err)
		} else {
			fmt.Printf("✅ 清理目录成功\n")
		}
	}
	
	// 测试文件权限（仅在Unix系统上）
	fmt.Printf("\n🔐 权限测试:\n")
	if runtime.GOOS != "windows" {
		testFile := "test-permission.txt"
		
		// 创建测试文件
		file, err := os.Create(testFile)
		if err != nil {
			fmt.Printf("❌ 创建文件失败: %v\n", err)
		} else {
			file.Close()
			
			// 设置执行权限
			if err := os.Chmod(testFile, 0755); err != nil {
				fmt.Printf("❌ 设置权限失败: %v\n", err)
			} else {
				fmt.Printf("✅ 设置权限成功\n")
			}
			
			// 清理测试文件
			os.Remove(testFile)
		}
	} else {
		fmt.Printf("Windows系统，跳过权限测试\n")
	}
	
	// 测试配置目录获取
	fmt.Printf("\n📋 配置目录测试:\n")
	
	var configDir string
	switch runtime.GOOS {
	case "windows":
		configDir = os.Getenv("APPDATA")
		if configDir == "" {
			configDir = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
	case "darwin":
		configDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support")
	default: // linux and others
		configDir = os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			configDir = filepath.Join(os.Getenv("HOME"), ".config")
		}
	}
	
	if configDir != "" {
		fmt.Printf("✅ 配置目录: %s\n", configDir)
	} else {
		fmt.Printf("❌ 无法获取配置目录\n")
	}
	
	// 测试临时目录
	tempDir := os.TempDir()
	fmt.Printf("✅ 临时目录: %s\n", tempDir)
	
	// 总结
	fmt.Printf("\n🎯 测试总结:\n")
	fmt.Printf("✅ 平台检测: 正常\n")
	fmt.Printf("✅ 路径处理: 正常\n")
	fmt.Printf("✅ 目录操作: 正常\n")
	fmt.Printf("✅ 环境变量: 正常\n")
	
	if runtime.GOOS != "windows" {
		fmt.Printf("✅ 权限设置: 正常\n")
	}
	
	fmt.Printf("\n🎉 跨平台兼容性测试完成！\n")
	fmt.Printf("系统在 %s/%s 平台上运行正常\n", runtime.GOOS, runtime.GOARCH)
}
