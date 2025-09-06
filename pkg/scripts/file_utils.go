package scripts

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FileWriter 文件写入器
type FileWriter struct{}

// NewFileWriter 创建文件写入器实例
func NewFileWriter() *FileWriter {
	return &FileWriter{}
}

// SaveJSON 保存数据为JSON格式
func (fw *FileWriter) SaveJSON(data interface{}, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("编码JSON失败: %w", err)
	}
	
	return nil
}

// SaveToFile 保存数据到文件
func (fw *FileWriter) SaveToFile(data []byte, filename string) error {
	return os.WriteFile(filename, data, 0644)
}

// LoadJSON 从JSON文件加载数据
func (fw *FileWriter) LoadJSON(filename string, v interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(v)
}

// EnsureDir 确保目录存在
func (fw *FileWriter) EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// CopyFile 复制文件
func (fw *FileWriter) CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("复制文件失败: %w", err)
	}

	return nil
}

// FileExists 检查文件是否存在
func (fw *FileWriter) FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// GetFileSize 获取文件大小
func (fw *FileWriter) GetFileSize(filename string) (int64, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// ReadFile 读取文件内容
func (fw *FileWriter) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// JoinPath 连接路径
func (fw *FileWriter) JoinPath(elements ...string) string {
	return filepath.Join(elements...)
}