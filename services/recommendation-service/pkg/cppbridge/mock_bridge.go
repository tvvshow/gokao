//go:build !cgo || !linux || !amd64
// +build !cgo !linux !amd64

// 注意：此文件为在非CGO环境下的编译占位符，不应在生产环境中使用。
// 在生产环境中，请确保使用CGO_ENABLED=1进行编译以启用真实的C++推荐引擎。
package cppbridge

import (
	"fmt"
)

// MockHybridRecommendationBridge 是一个编译占位符，用于在非CGO环境下编译
// 它会返回错误，表明需要在CGO环境下运行
type MockHybridRecommendationBridge struct{}

// 确保MockHybridRecommendationBridge实现HybridRecommendationBridge接口
var _ HybridRecommendationBridge = (*MockHybridRecommendationBridge)(nil)

// NewHybridRecommendationBridge 创建新的模拟桥接器
func NewHybridRecommendationBridge(cfg BridgeConfig) (HybridRecommendationBridge, error) {
	return nil, fmt.Errorf("cpp recommendation bridge is not enabled in this build")
}

// Close 关闭桥接器
func (b *MockHybridRecommendationBridge) Close() error {
	return nil
}

// GenerateRecommendations 生成推荐
func (b *MockHybridRecommendationBridge) GenerateRecommendations(request *RecommendationRequest) (*RecommendationResponse, error) {
	return nil, fmt.Errorf("mock bridge is not functional, CGO is required")
}

// GetHybridConfig 获取混合配置
func (b *MockHybridRecommendationBridge) GetHybridConfig() (map[string]interface{}, error) {
	return nil, fmt.Errorf("mock bridge is not functional, CGO is required")
}

// UpdateFusionWeights 更新融合权重
func (b *MockHybridRecommendationBridge) UpdateFusionWeights(weights map[string]float64) error {
	return fmt.Errorf("mock bridge is not functional, CGO is required")
}

// CompareRecommendations 比较推荐结果
func (b *MockHybridRecommendationBridge) CompareRecommendations(request *RecommendationRequest) (map[string]interface{}, error) {
	return nil, fmt.Errorf("mock bridge is not functional, CGO is required")
}

// GetPerformanceMetrics 获取性能指标
func (b *MockHybridRecommendationBridge) GetPerformanceMetrics() (map[string]interface{}, error) {
	return nil, fmt.Errorf("mock bridge is not functional, CGO is required")
}

// GenerateHybridPlan 生成混合方案
func (b *MockHybridRecommendationBridge) GenerateHybridPlan(request *RecommendationRequest) (map[string]interface{}, error) {
	return nil, fmt.Errorf("mock bridge is not functional, CGO is required")
}

// ClearCache 清除缓存
func (b *MockHybridRecommendationBridge) ClearCache() error {
	return fmt.Errorf("mock bridge is not functional, CGO is required")
}

// UpdateModel 更新模型
func (b *MockHybridRecommendationBridge) UpdateModel(modelPath string) error {
	return fmt.Errorf("mock bridge is not functional, CGO is required")
}

// GetSystemStatus 获取系统状态
func (b *MockHybridRecommendationBridge) GetSystemStatus() (map[string]interface{}, error) {
	return nil, fmt.Errorf("mock bridge is not functional, CGO is required")
}
