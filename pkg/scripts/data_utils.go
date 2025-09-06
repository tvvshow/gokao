package scripts

import (
	"strings"
)

// DataProcessor 数据处理器
type DataProcessor struct{}

// NewDataProcessor 创建数据处理器实例
func NewDataProcessor() *DataProcessor {
	return &DataProcessor{}
}

// Contains 检查字符串切片是否包含指定字符串
func (dp *DataProcessor) Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ContainsFuzzy 模糊检查字符串切片是否包含指定字符串
func (dp *DataProcessor) ContainsFuzzy(slice []string, item string) bool {
	for _, s := range slice {
		if strings.Contains(item, s) || strings.Contains(s, item) {
			return true
		}
	}
	return false
}

// RemoveDuplicates 移除字符串切片中的重复项
func (dp *DataProcessor) RemoveDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

// IsInString 检查字符串是否包含指定子串（不区分大小写）
func (dp *DataProcessor) IsInString(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// ExtractDomain 从URL中提取域名
func (dp *DataProcessor) ExtractDomain(url string) string {
	// 移除协议部分
	url = strings.Replace(url, "http://", "", 1)
	url = strings.Replace(url, "https://", "", 1)
	
	// 提取域名部分
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		domainParts := strings.Split(parts[0], ":")
		return domainParts[0]
	}
	
	return url
}

// NormalizeString 标准化字符串（去除空格、转换为小写）
func (dp *DataProcessor) NormalizeString(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}

// Truncate 截断字符串到指定长度
func (dp *DataProcessor) Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// CountWords 统计字符串中的单词数
func (dp *DataProcessor) CountWords(s string) int {
	words := strings.Fields(s)
	return len(words)
}

// SplitByComma 按逗号分割字符串并去除空格
func (dp *DataProcessor) SplitByComma(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	
	return result
}