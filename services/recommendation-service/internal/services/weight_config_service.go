package services

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// WeightConfig 权重配置
type WeightConfig struct {
	ScoreWeight        float64 `json:"score_weight"`
	RankWeight         float64 `json:"rank_weight"`
	LocationWeight     float64 `json:"location_weight"`
	MajorWeight        float64 `json:"major_weight"`
	EmploymentWeight   float64 `json:"employment_weight"`
	ReputationWeight   float64 `json:"reputation_weight"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// DefaultWeightConfig 默认权重配置
var DefaultWeightConfig = WeightConfig{
	ScoreWeight:      0.35,
	RankWeight:       0.25,
	LocationWeight:   0.15,
	MajorWeight:      0.15,
	EmploymentWeight: 0.05,
	ReputationWeight: 0.05,
}

// WeightConfigService 权重配置服务
type WeightConfigService struct {
	config     WeightConfig
	configPath string
	mu         sync.RWMutex
	logger     *logrus.Logger
}

// NewWeightConfigService 创建新的权重配置服务
func NewWeightConfigService(configPath string, logger *logrus.Logger) *WeightConfigService {
	service := &WeightConfigService{
		config:     DefaultWeightConfig,
		configPath: configPath,
		logger:     logger,
	}
	// 尝试从文件加载配置
	if err := service.LoadFromFile(); err != nil {
		logger.WithError(err).Warn("Failed to load weight config from file, using defaults")
	}
	return service
}

// GetConfig 获取当前配置
func (s *WeightConfigService) GetConfig() WeightConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// UpdateConfig 更新配置
func (s *WeightConfigService) UpdateConfig(config WeightConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 验证权重总和
	total := config.ScoreWeight + config.RankWeight + config.LocationWeight +
		config.MajorWeight + config.EmploymentWeight + config.ReputationWeight
	if total < 0.99 || total > 1.01 {
		return fmt.Errorf("weight sum must equal 1.0, got %.2f", total)
	}

	config.UpdatedAt = time.Now()
	s.config = config

	// 保存到文件
	if err := s.saveToFile(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// LoadFromFile 从文件加载配置
func (s *WeightConfigService) LoadFromFile() error {
	if s.configPath == "" {
		return nil
	}

	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var config WeightConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	s.mu.Lock()
	s.config = config
	s.mu.Unlock()

	return nil
}

// saveToFile 保存配置到文件
func (s *WeightConfigService) saveToFile() error {
	if s.configPath == "" {
		return nil
	}

	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.configPath, data, 0644)
}

// ResetToDefault 重置为默认配置
func (s *WeightConfigService) ResetToDefault() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = DefaultWeightConfig
	s.config.UpdatedAt = time.Now()
}
