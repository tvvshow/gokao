package types

import (
	"context"
	"time"
)

// AnalyticsService 分析服务接口
type AnalyticsService interface {
	GetRecommendationStats(userID string, startTime, endTime time.Time) (*RecommendationStats, error)
	GetSystemMetrics() (*SystemMetrics, error)
	GetUserBehaviorAnalysis(userID string) (*UserBehaviorAnalysis, error)
	GetAlgorithmPerformance() (*AlgorithmPerformance, error)
	GetPerformanceMetrics(ctx context.Context) (map[string]interface{}, error)
	GetFusionStatistics(ctx context.Context) (map[string]interface{}, error)
	GenerateQualityReport(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error)
	GetRecommendationTrends(ctx context.Context, timeRange string) (map[string]interface{}, error)
}

// RecommendationStats 推荐统计数据
type RecommendationStats struct {
	UserID             string                 `json:"user_id"`
	TotalRequests      int                    `json:"total_requests"`
	SuccessRate        float64                `json:"success_rate"`
	AvgResponseTime    float64                `json:"avg_response_time_ms"`
	AlgorithmBreakdown map[string]int         `json:"algorithm_breakdown"`
	TimeRange          TimeRange              `json:"time_range"`
	TopRecommendations []RecommendationItem   `json:"top_recommendations"`
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	Timestamp      time.Time `json:"timestamp"`
	CPUUsage       float64   `json:"cpu_usage"`
	MemoryUsage    float64   `json:"memory_usage"`
	DiskUsage      float64   `json:"disk_usage"`
	ActiveRequests int       `json:"active_requests"`
	CacheHitRate   float64   `json:"cache_hit_rate"`
	QPS            float64   `json:"qps"`
	AvgLatency     float64   `json:"avg_latency_ms"`
	ErrorRate      float64   `json:"error_rate"`
	ServiceHealth  string    `json:"service_health"`
}

// UserBehaviorAnalysis 用户行为分析
type UserBehaviorAnalysis struct {
	UserID                 string           `json:"user_id"`
	RequestPatterns        []RequestPattern `json:"request_patterns"`
	PreferenceProfile      PreferenceProfile `json:"preference_profile"`
	EngagementMetrics      EngagementMetrics `json:"engagement_metrics"`
	RecommendationFeedback []FeedbackItem   `json:"recommendation_feedback"`
	LastAnalysisTime       time.Time        `json:"last_analysis_time"`
}

// AlgorithmPerformance 算法性能分析
type AlgorithmPerformance struct {
	TraditionalAlgorithm AnalyticsPerformanceMetrics `json:"traditional_algorithm"`
	AIAlgorithm          AnalyticsPerformanceMetrics `json:"ai_algorithm"`
	HybridAlgorithm      AnalyticsPerformanceMetrics `json:"hybrid_algorithm"`
	ComparisonMetrics    ComparisonMetrics           `json:"comparison_metrics"`
	LastUpdated          time.Time                   `json:"last_updated"`
}

// AnalyticsPerformanceMetrics 分析性能指标
type AnalyticsPerformanceMetrics struct {
	AvgResponseTime  float64 `json:"avg_response_time_ms"`
	SuccessRate      float64 `json:"success_rate"`
	Accuracy         float64 `json:"accuracy"`
	Precision        float64 `json:"precision"`
	Recall           float64 `json:"recall"`
	F1Score          float64 `json:"f1_score"`
	UserSatisfaction float64 `json:"user_satisfaction"`
	ThroughputQPS    float64 `json:"throughput_qps"`
}

// ComparisonMetrics 算法比较指标
type ComparisonMetrics struct {
	BestPerforming     string             `json:"best_performing"`
	PerformanceGain    float64            `json:"performance_gain_percent"`
	RecommendationDiff float64            `json:"recommendation_diff_percent"`
	OptimalWeights     map[string]float64 `json:"optimal_weights"`
}

// TimeRange 时间范围
type TimeRange struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// RequestPattern 请求模式
type RequestPattern struct {
	TimeOfDay   string  `json:"time_of_day"`
	Frequency   int     `json:"frequency"`
	AvgDuration float64 `json:"avg_duration_ms"`
	RequestType string  `json:"request_type"`
}

// PreferenceProfile 偏好档案
type PreferenceProfile struct {
	PreferredSchools   []string           `json:"preferred_schools"`
	PreferredMajors    []string           `json:"preferred_majors"`
	PreferredLocations []string           `json:"preferred_locations"`
	ScoreRange         ScoreRange         `json:"score_range"`
	Priorities         map[string]float64 `json:"priorities"`
}

// EngagementMetrics 参与度指标
type EngagementMetrics struct {
	ClickThroughRate float64 `json:"click_through_rate"`
	ViewTime         float64 `json:"avg_view_time_seconds"`
	FavoriteRate     float64 `json:"favorite_rate"`
	ShareRate        float64 `json:"share_rate"`
	ConversionRate   float64 `json:"conversion_rate"`
}

// FeedbackItem 反馈项
type FeedbackItem struct {
	RecommendationID string    `json:"recommendation_id"`
	Rating           int       `json:"rating"`
	Feedback         string    `json:"feedback"`
	Timestamp        time.Time `json:"timestamp"`
}

// ScoreRange 分数范围
type ScoreRange struct {
	MinScore int `json:"min_score"`
	MaxScore int `json:"max_score"`
}

// RecommendationItem 推荐项
type RecommendationItem struct {
	SchoolID    string  `json:"school_id"`
	SchoolName  string  `json:"school_name"`
	MajorID     string  `json:"major_id"`
	MajorName   string  `json:"major_name"`
	Score       float64 `json:"score"`
	Count       int     `json:"count"`
	SuccessRate float64 `json:"success_rate"`
}
