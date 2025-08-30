package cppbridge

import "time"

// HybridRecommendationBridge 混合推荐桥接器接口
type HybridRecommendationBridge interface {
	Close() error
	GenerateRecommendations(request *RecommendationRequest) (*RecommendationResponse, error)
	GetHybridConfig() (map[string]interface{}, error)
	UpdateFusionWeights(weights map[string]float64) error
	CompareRecommendations(request *RecommendationRequest) (map[string]interface{}, error)
	GetPerformanceMetrics() (map[string]interface{}, error)
	GenerateHybridPlan(request *RecommendationRequest) (map[string]interface{}, error)
	ClearCache() error
	UpdateModel(modelPath string) error
	GetSystemStatus() (map[string]interface{}, error)
}

// RecommendationRequest 推荐请求
type RecommendationRequest struct {
	StudentID          string                 `json:"student_id"`
	Name               string                 `json:"name"`
	TotalScore         int                    `json:"total_score"`
	Ranking            int                    `json:"ranking"`
	Province           string                 `json:"province"`
	SubjectCombination string                 `json:"subject_combination"`
	ChineseScore       int                    `json:"chinese_score"`
	MathScore          int                    `json:"math_score"`
	EnglishScore       int                    `json:"english_score"`
	Physics            int                    `json:"physics,omitempty"`
	Chemistry          int                    `json:"chemistry,omitempty"`
	Biology            int                    `json:"biology,omitempty"`
	History            int                    `json:"history,omitempty"`
	Geography          int                    `json:"geography,omitempty"`
	Politics           int                    `json:"politics,omitempty"`
	Preferences        map[string]interface{} `json:"preferences,omitempty"`
	Filters            map[string]interface{} `json:"filters,omitempty"`
	MaxRecommendations int                    `json:"max_recommendations"`
	Algorithm          string                 `json:"algorithm"`
}

// RecommendationResponse 推荐响应
type RecommendationResponse struct {
	StudentID       string           `json:"student_id"`
	Recommendations []Recommendation `json:"recommendations"`
	TotalCount      int              `json:"total_count"`
	GeneratedAt     int64            `json:"generated_at"`
	Algorithm       string           `json:"algorithm"`
	Success         bool             `json:"success"`
	ErrorMessage    string           `json:"error_message,omitempty"`
}

// Recommendation 推荐项
type Recommendation struct {
	SchoolID       string  `json:"school_id"`
	SchoolName     string  `json:"school_name"`
	MajorID        string  `json:"major_id"`
	MajorName      string  `json:"major_name"`
	AdmissionScore int     `json:"admission_score"`
	Probability    float64 `json:"probability"`
	RiskLevel      string  `json:"risk_level"`
	Ranking        int     `json:"ranking"`
	Algorithm      string  `json:"algorithm"`
	Reasons        []string `json:"reasons,omitempty"`
	Score          float64 `json:"score"`
}

// BatchRecommendationRequest 批量推荐请求
type BatchRecommendationRequest struct {
	Requests   []RecommendationRequest `json:"requests"`
	BatchSize  int                     `json:"batch_size"`
	Timeout    int                     `json:"timeout"`
	Algorithm  string                  `json:"algorithm"`
}

// BatchRecommendationResponse 批量推荐响应
type BatchRecommendationResponse struct {
	Responses   []RecommendationResponse `json:"responses"`
	TotalCount  int                      `json:"total_count"`
	SuccessCount int                     `json:"success_count"`
	FailedCount int                      `json:"failed_count"`
	ProcessedAt int64                    `json:"processed_at"`
}

// FusionWeights 融合权重
type FusionWeights struct {
	TraditionalWeight float64 `json:"traditional_weight"`
	AIWeight          float64 `json:"ai_weight"`
	DiversityFactor   float64 `json:"diversity_factor"`
	UpdatedAt         int64   `json:"updated_at"`
}

// HybridConfig 混合配置
type HybridConfig struct {
	TraditionalWeight     float64 `json:"traditional_weight"`
	AIWeight              float64 `json:"ai_weight"`
	DiversityFactor       float64 `json:"diversity_factor"`
	MaxSameCityRatio      int     `json:"max_same_city_ratio"`
	MaxSameLevelRatio     int     `json:"max_same_level_ratio"`
	RushRatio             float64 `json:"rush_ratio"`
	StableRatio           float64 `json:"stable_ratio"`
	SafeRatio             float64 `json:"safe_ratio"`
	MaxCandidates         int     `json:"max_candidates"`
	ScoreThreshold        float64 `json:"score_threshold"`
	EnableParallel        bool    `json:"enable_parallel"`
	EnableAdaptiveWeights bool    `json:"enable_adaptive_weights"`
	ConfidenceThreshold   float64 `json:"confidence_threshold"`
}

// ComparisonResult 比较结果
type ComparisonResult struct {
	Traditional  *RecommendationResponse `json:"traditional"`
	AI           *RecommendationResponse `json:"ai"`
	Hybrid       *RecommendationResponse `json:"hybrid"`
	Metrics      ComparisonMetrics       `json:"metrics"`
	GeneratedAt  int64                   `json:"generated_at"`
}

// ComparisonMetrics 比较指标
type ComparisonMetrics struct {
	DiversityScore   float64 `json:"diversity_score"`
	AccuracyScore    float64 `json:"accuracy_score"`
	PerformanceScore float64 `json:"performance_score"`
	CoverageScore    float64 `json:"coverage_score"`
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	AvgResponseTime float64 `json:"avg_response_time"`
	SuccessRate     float64 `json:"success_rate"`
	CacheHitRate    float64 `json:"cache_hit_rate"`
	QPS             float64 `json:"qps"`
	MemoryUsage     float64 `json:"memory_usage"`
	CPUUsage        float64 `json:"cpu_usage"`
	ActiveRequests  int     `json:"active_requests"`
	TotalRequests   int64   `json:"total_requests"`
	ErrorCount      int64   `json:"error_count"`
	LastUpdated     int64   `json:"last_updated"`
}

// SystemStatus 系统状态
type SystemStatus struct {
	Status         string    `json:"status"`
	Uptime         int64     `json:"uptime"`
	Version        string    `json:"version"`
	MemoryUsage    string    `json:"memory_usage"`
	CacheSize      int       `json:"cache_size"`
	ActiveRequests int       `json:"active_requests"`
	Health         string    `json:"health"`
	LastCheck      time.Time `json:"last_check"`
}

// HybridPlan 混合方案
type HybridPlan struct {
	StudentID   string             `json:"student_id"`
	RushTier    []RecommendationTier `json:"rush_tier"`
	StableTier  []RecommendationTier `json:"stable_tier"`
	SafeTier    []RecommendationTier `json:"safe_tier"`
	Strategy    string             `json:"strategy"`
	Confidence  float64            `json:"confidence"`
	GeneratedAt int64              `json:"generated_at"`
}

// RecommendationTier 推荐层级
type RecommendationTier struct {
	SchoolName  string  `json:"school_name"`
	MajorName   string  `json:"major_name"`
	Probability float64 `json:"probability"`
	Ranking     int     `json:"ranking"`
	ScoreRange  string  `json:"score_range"`
}