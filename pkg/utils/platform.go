package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Platform 平台相关的工具函数
type Platform struct{}

// NewPlatform 创建平台工具实例
func NewPlatform() *Platform {
	return &Platform{}
}

// IsWindows 检查是否为Windows平台
func (p *Platform) IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsLinux 检查是否为Linux平台
func (p *Platform) IsLinux() bool {
	return runtime.GOOS == "linux"
}

// IsMacOS 检查是否为macOS平台
func (p *Platform) IsMacOS() bool {
	return runtime.GOOS == "darwin"
}

// GetExecutableExtension 获取可执行文件扩展名
func (p *Platform) GetExecutableExtension() string {
	if p.IsWindows() {
		return ".exe"
	}
	return ""
}

// GetPathSeparator 获取路径分隔符
func (p *Platform) GetPathSeparator() string {
	return string(filepath.Separator)
}

// JoinPath 跨平台路径连接
func (p *Platform) JoinPath(elements ...string) string {
	return filepath.Join(elements...)
}

// NormalizePath 标准化路径
func (p *Platform) NormalizePath(path string) string {
	// 将所有路径分隔符统一为当前平台的分隔符
	normalized := filepath.Clean(path)
	return normalized
}

// GetConfigDir 获取配置目录
func (p *Platform) GetConfigDir(appName string) (string, error) {
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
	
	if configDir == "" {
		return "", os.ErrNotExist
	}
	
	appConfigDir := filepath.Join(configDir, appName)
	return appConfigDir, nil
}

// GetDataDir 获取数据目录
func (p *Platform) GetDataDir(appName string) (string, error) {
	var dataDir string
	
	switch runtime.GOOS {
	case "windows":
		dataDir = os.Getenv("LOCALAPPDATA")
		if dataDir == "" {
			dataDir = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
		}
	case "darwin":
		dataDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support")
	default: // linux and others
		dataDir = os.Getenv("XDG_DATA_HOME")
		if dataDir == "" {
			dataDir = filepath.Join(os.Getenv("HOME"), ".local", "share")
		}
	}
	
	if dataDir == "" {
		return "", os.ErrNotExist
	}
	
	appDataDir := filepath.Join(dataDir, appName)
	return appDataDir, nil
}

// GetTempDir 获取临时目录
func (p *Platform) GetTempDir(appName string) string {
	tempDir := os.TempDir()
	return filepath.Join(tempDir, appName)
}

// EnsureDir 确保目录存在
func (p *Platform) EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// GetEnvWithDefault 获取环境变量，如果不存在则返回默认值
func (p *Platform) GetEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetExecutablePermission 设置可执行权限（仅在Unix系统上有效）
func (p *Platform) SetExecutablePermission(filePath string) error {
	if p.IsWindows() {
		// Windows上不需要设置执行权限
		return nil
	}
	
	// Unix系统上设置执行权限
	return os.Chmod(filePath, 0755)
}

// GetBinaryName 获取二进制文件名（包含平台特定扩展名）
func (p *Platform) GetBinaryName(baseName string) string {
	return baseName + p.GetExecutableExtension()
}

// ConvertPathForPlatform 将路径转换为当前平台格式
func (p *Platform) ConvertPathForPlatform(path string) string {
	// 替换路径分隔符
	if p.IsWindows() {
		// 将Unix风格的路径转换为Windows风格
		return strings.ReplaceAll(path, "/", "\\")
	} else {
		// 将Windows风格的路径转换为Unix风格
		return strings.ReplaceAll(path, "\\", "/")
	}
}

// GetShellCommand 获取当前平台的shell命令
func (p *Platform) GetShellCommand() (string, []string) {
	if p.IsWindows() {
		return "cmd", []string{"/c"}
	}
	return "sh", []string{"-c"}
}

// GetMakeCommand 获取当前平台的make命令
func (p *Platform) GetMakeCommand() string {
	if p.IsWindows() {
		// Windows上可能使用mingw32-make或nmake
		if _, err := os.LookPath("mingw32-make"); err == nil {
			return "mingw32-make"
		}
		if _, err := os.LookPath("nmake"); err == nil {
			return "nmake"
		}
		return "make" // 回退到make
	}
	return "make"
}

// GetCMakeGenerator 获取当前平台的CMake生成器
func (p *Platform) GetCMakeGenerator() string {
	if p.IsWindows() {
		// Windows上优先使用MinGW Makefiles
		return "MinGW Makefiles"
	}
	return "Unix Makefiles"
}

// GetCPUCount 获取CPU核心数
func (p *Platform) GetCPUCount() int {
	return runtime.NumCPU()
}

// 全局实例
var PlatformUtils = NewPlatform()
