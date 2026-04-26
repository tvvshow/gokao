package services

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// WeightService 权重配置服务
// 管理6维度权重模型和动态配置
type WeightService struct {
	defaultWeights *WeightConfig
	customWeights  map[string]*WeightConfig // key: userID or sessionID
	mu             sync.RWMutex
	logger         *logrus.Logger
}

// WeightConfig 权重配置
// 6维度权重模型：分数匹配、地理位置、专业兴趣、就业前景、学校排名、竞争程度
type WeightConfig struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	// 6维度权重 (总和应为1.0)
	ScoreMatchWeight     float64 `json:"score_match_weight"`     // 分数匹配权重
	LocationWeight       float64 `json:"location_weight"`        // 地理位置权重
	InterestWeight       float64 `json:"interest_weight"`        // 专业兴趣权重
	EmploymentWeight     float64 `json:"employment_weight"`      // 就业前景权重
	UniversityRankWeight float64 `json:"university_rank_weight"` // 学校排名权重
	CompetitionWeight    float64 `json:"competition_weight"`     // 竞争程度权重

	// 高级配置
	EnableAdaptiveWeights bool    `json:"enable_adaptive_weights"` // 是否启用自适应权重
	MinWeightThreshold    float64 `json:"min_weight_threshold"`    // 最小权重阈值
	MaxWeightThreshold    float64 `json:"max_weight_threshold"`    // 最大权重阈值

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDefault bool      `json:"is_default"`
	IsActive  bool      `json:"is_active"`
}

// NewWeightService 创建权重服务
func NewWeightService(logger *logrus.Logger) *WeightService {
	service := &WeightService{
		customWeights: make(map[string]*WeightConfig),
		logger:        logger,
	}

	// 初始化默认权重配置
	service.initializeDefaultWeights()

	return service
}

// initializeDefaultWeights 初始化默认权重配置
func (s *WeightService) initializeDefaultWeights() {
	s.defaultWeights = &WeightConfig{
		ID:          "default",
		Name:        "默认权重配置",
		Description: "均衡的权重配置，适合大多数考生",

		ScoreMatchWeight:     0.35, // 35%
		LocationWeight:       0.15, // 15%
		InterestWeight:       0.15, // 15%
		EmploymentWeight:     0.15, // 15%
		UniversityRankWeight: 0.10, // 10%
		CompetitionWeight:    0.10, // 10%

		EnableAdaptiveWeights: true,
		MinWeightThreshold:    0.05,
		MaxWeightThreshold:    0.40,

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsDefault: true,
		IsActive:  true,
	}
}

// GetDefaultWeights 获取默认权重配置
func (s *WeightService) GetDefaultWeights() *WeightConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.defaultWeights
}

// GetWeights 获取权重配置
func (s *WeightService) GetWeights(key string) *WeightConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if config, exists := s.customWeights[key]; exists && config.IsActive {
		return config
	}

	return s.defaultWeights
}

// GetWeightMap 以通用 map 形式返回权重，供跨包桥接层使用
func (s *WeightService) GetWeightMap(key string) map[string]float64 {
	config := s.GetWeights(key)
	return map[string]float64{
		"score_match":     config.ScoreMatchWeight,
		"location":        config.LocationWeight,
		"interest":        config.InterestWeight,
		"employment":      config.EmploymentWeight,
		"university_rank": config.UniversityRankWeight,
		"competition":     config.CompetitionWeight,
	}
}

// SetWeights 设置权重配置
func (s *WeightService) SetWeights(key string, config *WeightConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 验证权重总和
	if err := s.validateWeights(config); err != nil {
		return err
	}

	config.UpdatedAt = time.Now()
	config.IsDefault = false

	if config.ID == "" {
		config.ID = fmt.Sprintf("custom_%d", time.Now().UnixNano())
	}

	s.customWeights[key] = config
	s.logger.Infof("权重配置已更新: key=%s, config=%s", key, config.ID)

	return nil
}

// SetWeightMap 以通用 map 形式设置权重，供跨包桥接层使用
func (s *WeightService) SetWeightMap(key string, weights map[string]float64) error {
	config := &WeightConfig{
		ScoreMatchWeight:      weights["score_match"],
		LocationWeight:        weights["location"],
		InterestWeight:        weights["interest"],
		EmploymentWeight:      weights["employment"],
		UniversityRankWeight:  weights["university_rank"],
		CompetitionWeight:     weights["competition"],
		EnableAdaptiveWeights: true,
		MinWeightThreshold:    0.05,
		MaxWeightThreshold:    0.40,
		IsActive:              true,
	}
	return s.SetWeights(key, config)
}

// validateWeights 验证权重配置
func (s *WeightService) validateWeights(config *WeightConfig) error {
	total := config.ScoreMatchWeight + config.LocationWeight +
		config.InterestWeight + config.EmploymentWeight +
		config.UniversityRankWeight + config.CompetitionWeight

	// 允许一定的误差范围
	if total < 0.99 || total > 1.01 {
		return fmt.Errorf("权重总和必须为1.0，当前总和: %.3f", total)
	}

	// 检查单个权重范围
	weights := map[string]float64{
		"score_match_weight":     config.ScoreMatchWeight,
		"location_weight":        config.LocationWeight,
		"interest_weight":        config.InterestWeight,
		"employment_weight":      config.EmploymentWeight,
		"university_rank_weight": config.UniversityRankWeight,
		"competition_weight":     config.CompetitionWeight,
	}

	for name, weight := range weights {
		if weight < 0 {
			return fmt.Errorf("权重不能为负数: %s=%.3f", name, weight)
		}
		if weight > 1.0 {
			return fmt.Errorf("权重不能超过1.0: %s=%.3f", name, weight)
		}
	}

	return nil
}

// DeleteWeights 删除权重配置
func (s *WeightService) DeleteWeights(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.customWeights, key)
	s.logger.Infof("权重配置已删除: key=%s", key)
}

// ListWeights 列出所有权重配置
func (s *WeightService) ListWeights() map[string]*WeightConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*WeightConfig)
	result["default"] = s.defaultWeights

	for key, config := range s.customWeights {
		result[key] = config
	}

	return result
}

// ResetToDefault 重置为默认权重
func (s *WeightService) ResetToDefault(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.customWeights, key)
	s.logger.Infof("权重配置已重置为默认: key=%s", key)
}

// ExportWeights 导出权重配置
func (s *WeightService) ExportWeights(key string) ([]byte, error) {
	config := s.GetWeights(key)
	return json.MarshalIndent(config, "", "  ")
}

// ImportWeights 导入权重配置
func (s *WeightService) ImportWeights(key string, data []byte) error {
	var config WeightConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析权重配置失败: %v", err)
	}

	return s.SetWeights(key, &config)
}

// GetWeightStats 获取权重统计信息
func (s *WeightService) GetWeightStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"default_config_id":   s.defaultWeights.ID,
		"custom_config_count": len(s.customWeights),
		"total_config_count":  len(s.customWeights) + 1,
		"default_weights": map[string]float64{
			"score_match":     s.defaultWeights.ScoreMatchWeight,
			"location":        s.defaultWeights.LocationWeight,
			"interest":        s.defaultWeights.InterestWeight,
			"employment":      s.defaultWeights.EmploymentWeight,
			"university_rank": s.defaultWeights.UniversityRankWeight,
			"competition":     s.defaultWeights.CompetitionWeight,
		},
	}
}

// CreatePresetWeights 创建预设权重配置
func (s *WeightService) CreatePresetWeights() {
	presets := map[string]*WeightConfig{
		"score_focused": {
			ID:                   "score_focused",
			Name:                 "分数优先型",
			Description:          "重点关注分数匹配，适合分数优势明显的考生",
			ScoreMatchWeight:     0.50,
			LocationWeight:       0.10,
			InterestWeight:       0.10,
			EmploymentWeight:     0.10,
			UniversityRankWeight: 0.10,
			CompetitionWeight:    0.10,
			IsActive:             true,
		},
		"location_focused": {
			ID:                   "location_focused",
			Name:                 "地域优先型",
			Description:          "重点关注地理位置，适合有明确地域偏好的考生",
			ScoreMatchWeight:     0.20,
			LocationWeight:       0.35,
			InterestWeight:       0.15,
			EmploymentWeight:     0.10,
			UniversityRankWeight: 0.10,
			CompetitionWeight:    0.10,
			IsActive:             true,
		},
		"employment_focused": {
			ID:                   "employment_focused",
			Name:                 "就业优先型",
			Description:          "重点关注就业前景，适合关注职业发展的考生",
			ScoreMatchWeight:     0.25,
			LocationWeight:       0.10,
			InterestWeight:       0.15,
			EmploymentWeight:     0.30,
			UniversityRankWeight: 0.10,
			CompetitionWeight:    0.10,
			IsActive:             true,
		},
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for key, preset := range presets {
		preset.CreatedAt = time.Now()
		preset.UpdatedAt = time.Now()
		s.customWeights[key] = preset
	}

	s.logger.Info("预设权重配置已创建")
}
