// +build !cgo

package cppbridge

import (
	"fmt"
	"math/rand"
	"time"
)

// MockHybridRecommendationBridge 模拟的混合推荐桥接器
type MockHybridRecommendationBridge struct {
	configPath string
	initialized bool
}

// 确保MockHybridRecommendationBridge实现HybridRecommendationBridge接口
var _ HybridRecommendationBridge = (*MockHybridRecommendationBridge)(nil)

// NewHybridRecommendationBridge 创建新的模拟桥接器
func NewHybridRecommendationBridge(configPath string) (HybridRecommendationBridge, error) {
	return &MockHybridRecommendationBridge{
		configPath: configPath,
		initialized: true,
	}, nil
}

// Close 关闭桥接器
func (b *MockHybridRecommendationBridge) Close() error {
	b.initialized = false
	return nil
}

// GenerateRecommendations 生成推荐
func (b *MockHybridRecommendationBridge) GenerateRecommendations(request *RecommendationRequest) (*RecommendationResponse, error) {
	if !b.initialized {
		return nil, fmt.Errorf("bridge not initialized")
	}

	// 模拟推荐生成
	rand.Seed(time.Now().UnixNano())
	
	var recommendations []Recommendation
	for i := 0; i < 10; i++ {
		recommendations = append(recommendations, Recommendation{
			SchoolID:      fmt.Sprintf("school_%03d", i+1),
			SchoolName:    fmt.Sprintf("大学%d", i+1),
			MajorID:       fmt.Sprintf("major_%03d", i+1),
			MajorName:     fmt.Sprintf("专业%d", i+1),
			AdmissionScore: 580 + rand.Intn(100),
			Probability:   0.5 + rand.Float64()*0.4,
			RiskLevel:     []string{"low", "medium", "high"}[rand.Intn(3)],
			Ranking:       i + 1,
			Algorithm:     "mock",
		})
	}

	response := &RecommendationResponse{
		StudentID:       request.StudentID,
		Recommendations: recommendations,
		TotalCount:      len(recommendations),
		GeneratedAt:     time.Now().Unix(),
		Algorithm:       "hybrid_mock",
		Success:         true,
	}

	return response, nil
}

// GetHybridConfig 获取混合配置
func (b *MockHybridRecommendationBridge) GetHybridConfig() (map[string]interface{}, error) {
	config := map[string]interface{}{
		"traditional_weight": 0.6,
		"ai_weight":         0.4,
		"diversity_factor":  0.15,
		"max_candidates":    500,
		"enabled":          true,
	}
	return config, nil
}

// UpdateFusionWeights 更新融合权重
func (b *MockHybridRecommendationBridge) UpdateFusionWeights(weights map[string]float64) error {
	// 模拟权重更新
	fmt.Printf("Mock: Updated fusion weights: %+v\n", weights)
	return nil
}

// CompareRecommendations 比较推荐结果
func (b *MockHybridRecommendationBridge) CompareRecommendations(request *RecommendationRequest) (map[string]interface{}, error) {
	traditional, _ := b.GenerateRecommendations(request)
	ai, _ := b.GenerateRecommendations(request)
	hybrid, _ := b.GenerateRecommendations(request)

	comparison := map[string]interface{}{
		"traditional": traditional,
		"ai":         ai,
		"hybrid":     hybrid,
		"metrics": map[string]interface{}{
			"diversity_score":    0.85,
			"accuracy_score":     0.92,
			"performance_score":  0.88,
		},
	}

	return comparison, nil
}

// GetPerformanceMetrics 获取性能指标
func (b *MockHybridRecommendationBridge) GetPerformanceMetrics() (map[string]interface{}, error) {
	metrics := map[string]interface{}{
		"avg_response_time": 95.5,
		"success_rate":      0.95,
		"cache_hit_rate":    0.85,
		"qps":              150.2,
		"memory_usage":     78.5,
		"cpu_usage":        65.2,
	}
	return metrics, nil
}

// GenerateHybridPlan 生成混合方案
func (b *MockHybridRecommendationBridge) GenerateHybridPlan(request *RecommendationRequest) (map[string]interface{}, error) {
	plan := map[string]interface{}{
		"student_id": request.StudentID,
		"rush_tier": []interface{}{
			map[string]interface{}{
				"school_name": "清华大学",
				"major_name":  "计算机科学与技术",
				"probability": 0.3,
			},
		},
		"stable_tier": []interface{}{
			map[string]interface{}{
				"school_name": "北京理工大学",
				"major_name":  "软件工程",
				"probability": 0.7,
			},
		},
		"safe_tier": []interface{}{
			map[string]interface{}{
				"school_name": "北京工业大学",
				"major_name":  "信息工程",
				"probability": 0.9,
			},
		},
		"strategy": "balanced",
		"confidence": 0.85,
	}
	return plan, nil
}

// ClearCache 清空缓存
func (b *MockHybridRecommendationBridge) ClearCache() error {
	fmt.Println("Mock: Cache cleared")
	return nil
}

// UpdateModel 更新模型
func (b *MockHybridRecommendationBridge) UpdateModel(modelPath string) error {
	fmt.Printf("Mock: Model updated with path: %s\n", modelPath)
	return nil
}

// GetSystemStatus 获取系统状态
func (b *MockHybridRecommendationBridge) GetSystemStatus() (map[string]interface{}, error) {
	status := map[string]interface{}{
		"status":        "healthy",
		"uptime":        time.Now().Unix() - 3600, // 1小时前启动
		"version":       "1.0.0-mock",
		"memory_usage":  "128MB",
		"cache_size":    1500,
		"active_requests": 5,
	}
	return status, nil
}