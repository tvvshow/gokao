package services

import (
	"context"
	"data-service/internal/database"
	"data-service/internal/models"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AlgorithmService C++算法引擎服务
type AlgorithmService struct {
	db     *database.DB
	logger *logrus.Logger
}

// NewAlgorithmService 创建算法服务实例
func NewAlgorithmService(db *database.DB, logger *logrus.Logger) *AlgorithmService {
	return &AlgorithmService{
		db:     db,
		logger: logger,
	}
}

// VolunteerMatchRequest 志愿匹配请求
type VolunteerMatchRequest struct {
	UserID   string  `json:"user_id,omitempty"`
	Province string  `json:"province" validate:"required"`
	Category string  `json:"category" validate:"required"` // science, liberal_arts, comprehensive
	Score    float64 `json:"score" validate:"required,min=0,max=1000"`
	Rank     int     `json:"rank,omitempty"`
	
	// 偏好设置
	Preferences VolunteerPreferences `json:"preferences"`
	
	// 风险偏好
	RiskTolerance string `json:"risk_tolerance"` // conservative, moderate, aggressive
	
	// 院校层次偏好
	PreferredLevels []string `json:"preferred_levels,omitempty"` // 985, 211, double_first_class, ordinary
	
	// 地区偏好
	PreferredProvinces []string `json:"preferred_provinces,omitempty"`
	
	// 专业偏好
	PreferredMajors []string `json:"preferred_majors,omitempty"`
}

// VolunteerPreferences 志愿偏好
type VolunteerPreferences struct {
	// 优先级权重 (总和应为1.0)
	SchoolWeight float64 `json:"school_weight"`     // 院校权重
	MajorWeight  float64 `json:"major_weight"`      // 专业权重
	LocationWeight float64 `json:"location_weight"` // 地理位置权重
	EmploymentWeight float64 `json:"employment_weight"` // 就业前景权重
	
	// 具体偏好
	PreferLocation   bool     `json:"prefer_location"`   // 优先考虑地理位置
	PreferMajor      bool     `json:"prefer_major"`      // 优先考虑专业
	PreferEmployment bool     `json:"prefer_employment"` // 优先考虑就业
	AvoidColdMajors  bool     `json:"avoid_cold_majors"` // 避免冷门专业
	
	// 限制条件
	MaxDistance      int      `json:"max_distance,omitempty"`      // 最大距离(km)
	MinEmploymentRate float64 `json:"min_employment_rate,omitempty"` // 最低就业率
	ExcludedProvinces []string `json:"excluded_provinces,omitempty"` // 排除的省份
	ExcludedMajors    []string `json:"excluded_majors,omitempty"`    // 排除的专业
}

// VolunteerMatchResponse 志愿匹配响应
type VolunteerMatchResponse struct {
	RequestID     string                 `json:"request_id"`
	Recommendations []VolunteerRecommendation `json:"recommendations"`
	RiskAnalysis    RiskAnalysis           `json:"risk_analysis"`
	ProcessTime     float64                `json:"process_time"` // 处理时间(毫秒)
	Algorithm       string                 `json:"algorithm"`
	Version         string                 `json:"version"`
	Confidence      float64                `json:"confidence"`
}

// VolunteerRecommendation 志愿推荐
type VolunteerRecommendation struct {
	University      models.University      `json:"university"`
	Major           *models.Major          `json:"major,omitempty"`
	AdmissionChance AdmissionChance        `json:"admission_chance"`
	MatchScore      float64                `json:"match_score"`      // 匹配度分数(0-100)
	RecommendType   string                 `json:"recommend_type"`   // safe, moderate, reach
	Ranking         int                    `json:"ranking"`          // 推荐排名
	Reasons         []string               `json:"reasons"`          // 推荐理由
	Warnings        []string               `json:"warnings,omitempty"` // 风险提示
}

// AdmissionChance 录取概率
type AdmissionChance struct {
	Probability float64 `json:"probability"` // 录取概率(0-1)
	Level       string  `json:"level"`       // very_high, high, medium, low, very_low
	Description string  `json:"description"` // 文字描述
}

// RiskAnalysis 风险分析
type RiskAnalysis struct {
	OverallRisk    string             `json:"overall_risk"`    // low, medium, high
	RiskFactors    []RiskFactor       `json:"risk_factors"`
	Suggestions    []string           `json:"suggestions"`
	PortfolioBalance PortfolioBalance `json:"portfolio_balance"`
}

// RiskFactor 风险因素
type RiskFactor struct {
	Factor      string  `json:"factor"`
	Level       string  `json:"level"`      // low, medium, high
	Impact      float64 `json:"impact"`     // 影响程度(0-1)
	Description string  `json:"description"`
}

// PortfolioBalance 志愿组合平衡
type PortfolioBalance struct {
	SafeCount      int     `json:"safe_count"`      // 保底志愿数量
	ModerateCount  int     `json:"moderate_count"`  // 稳妥志愿数量
	ReachCount     int     `json:"reach_count"`     // 冲刺志愿数量
	RecommendedRatio string `json:"recommended_ratio"` // 推荐比例
}

// MatchVolunteers 志愿匹配
func (s *AlgorithmService) MatchVolunteers(ctx context.Context, req VolunteerMatchRequest) (*VolunteerMatchResponse, error) {
	startTime := time.Now()
	
	// 生成请求ID
	requestID := uuid.New().String()
	
	// 验证请求参数
	if err := s.validateRequest(req); err != nil {
		return nil, fmt.Errorf("请求参数验证失败: %w", err)
	}

	// 如果启用了C++算法引擎，使用算法引擎处理
	if s.db.Config.AlgorithmEngineEnabled {
		response, err := s.matchWithCppEngine(ctx, req, requestID)
		if err != nil {
			s.logger.Warnf("C++算法引擎处理失败，回退到Go实现: %v", err)
			// 回退到Go实现
			return s.matchWithGoImplementation(ctx, req, requestID, startTime)
		}
		return response, nil
	}

	// 使用Go实现
	return s.matchWithGoImplementation(ctx, req, requestID, startTime)
}

// matchWithCppEngine 使用C++算法引擎处理
func (s *AlgorithmService) matchWithCppEngine(ctx context.Context, req VolunteerMatchRequest, requestID string) (*VolunteerMatchResponse, error) {
	// TODO: 集成C++算法引擎
	// 这里应该调用C++模块中的volunteer_matcher
	
	s.logger.Info("调用C++算法引擎进行志愿匹配")
	
	// 暂时返回错误，强制使用Go实现
	return nil, fmt.Errorf("C++算法引擎暂未实现")
}

// matchWithGoImplementation 使用Go实现处理
func (s *AlgorithmService) matchWithGoImplementation(ctx context.Context, req VolunteerMatchRequest, requestID string, startTime time.Time) (*VolunteerMatchResponse, error) {
	s.logger.Info("使用Go实现进行志愿匹配")
	
	// 获取候选院校和专业
	candidates, err := s.getCandidates(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("获取候选数据失败: %w", err)
	}

	// 计算匹配度和录取概率
	recommendations := s.calculateRecommendations(ctx, req, candidates)
	
	// 风险分析
	riskAnalysis := s.analyzeRisk(req, recommendations)
	
	// 计算处理时间
	processTime := float64(time.Since(startTime).Nanoseconds()) / 1e6

	response := &VolunteerMatchResponse{
		RequestID:       requestID,
		Recommendations: recommendations,
		RiskAnalysis:    riskAnalysis,
		ProcessTime:     processTime,
		Algorithm:       "go_implementation",
		Version:         "1.0.0",
		Confidence:      s.calculateConfidence(recommendations),
	}

	// 保存分析结果
	s.saveAnalysisResult(ctx, req, response)

	return response, nil
}

// validateRequest 验证请求参数
func (s *AlgorithmService) validateRequest(req VolunteerMatchRequest) error {
	if req.Province == "" {
		return fmt.Errorf("省份不能为空")
	}
	if req.Category == "" {
		return fmt.Errorf("科类不能为空")
	}
	if req.Score <= 0 || req.Score > 1000 {
		return fmt.Errorf("分数必须在0-1000之间")
	}
	
	// 验证权重总和
	totalWeight := req.Preferences.SchoolWeight + req.Preferences.MajorWeight + 
		req.Preferences.LocationWeight + req.Preferences.EmploymentWeight
	if totalWeight > 0 && (totalWeight < 0.9 || totalWeight > 1.1) {
		return fmt.Errorf("权重总和应该等于1.0")
	}
	
	return nil
}

// getCandidates 获取候选院校和专业
func (s *AlgorithmService) getCandidates(ctx context.Context, req VolunteerMatchRequest) ([]CandidateData, error) {
	var candidates []CandidateData
	
	// 基于分数范围获取可能的院校
	scoreRange := s.calculateScoreRange(req.Score, req.RiskTolerance)
	
	// 查询符合条件的录取数据
	var admissionData []models.AdmissionData
	query := s.db.PostgreSQL.Model(&models.AdmissionData{}).
		Preload("University").
		Preload("Major").
		Where("province = ? AND category = ?", req.Province, req.Category).
		Where("year >= ?", time.Now().Year()-3). // 近3年数据
		Where("avg_score BETWEEN ? AND ?", scoreRange.Min, scoreRange.Max)

	if err := query.Find(&admissionData).Error; err != nil {
		return nil, fmt.Errorf("查询录取数据失败: %w", err)
	}

	// 转换为候选数据
	for _, data := range admissionData {
		if data.University.IsActive && data.University.IsRecruiting {
			candidate := CandidateData{
				University:    data.University,
				Major:         data.Major,
				AdmissionData: data,
			}
			candidates = append(candidates, candidate)
		}
	}

	return candidates, nil
}

// CandidateData 候选数据
type CandidateData struct {
	University    models.University
	Major         *models.Major
	AdmissionData models.AdmissionData
}

// ScoreRange 分数范围
type ScoreRange struct {
	Min float64
	Max float64
}

// calculateScoreRange 计算分数范围
func (s *AlgorithmService) calculateScoreRange(score float64, riskTolerance string) ScoreRange {
	var lowerBound, upperBound float64
	
	switch riskTolerance {
	case "conservative":
		lowerBound = score - 50 // 保守策略，主要考虑低于当前分数的
		upperBound = score + 20
	case "moderate":
		lowerBound = score - 30
		upperBound = score + 30
	case "aggressive":
		lowerBound = score - 20
		upperBound = score + 50 // 激进策略，更多考虑高分院校
	default:
		lowerBound = score - 30
		upperBound = score + 30
	}
	
	// 确保分数在合理范围内
	if lowerBound < 0 {
		lowerBound = 0
	}
	if upperBound > 1000 {
		upperBound = 1000
	}
	
	return ScoreRange{Min: lowerBound, Max: upperBound}
}

// calculateRecommendations 计算推荐结果
func (s *AlgorithmService) calculateRecommendations(ctx context.Context, req VolunteerMatchRequest, candidates []CandidateData) []VolunteerRecommendation {
	var recommendations []VolunteerRecommendation
	
	for _, candidate := range candidates {
		// 计算录取概率
		admissionChance := s.calculateAdmissionChance(req.Score, candidate.AdmissionData)
		
		// 计算匹配度
		matchScore := s.calculateMatchScore(req, candidate)
		
		// 确定推荐类型
		recommendType := s.determineRecommendType(admissionChance.Probability)
		
		// 生成推荐理由
		reasons := s.generateReasons(req, candidate, matchScore)
		
		// 生成风险提示
		warnings := s.generateWarnings(req, candidate, admissionChance)
		
		recommendation := VolunteerRecommendation{
			University:      candidate.University,
			Major:           candidate.Major,
			AdmissionChance: admissionChance,
			MatchScore:      matchScore,
			RecommendType:   recommendType,
			Reasons:         reasons,
			Warnings:        warnings,
		}
		
		recommendations = append(recommendations, recommendation)
	}
	
	// 排序并设置排名
	recommendations = s.sortAndRankRecommendations(recommendations)
	
	// 限制推荐数量
	if len(recommendations) > 30 {
		recommendations = recommendations[:30]
	}
	
	return recommendations
}

// calculateAdmissionChance 计算录取概率
func (s *AlgorithmService) calculateAdmissionChance(score float64, admissionData models.AdmissionData) AdmissionChance {
	var probability float64
	
	if admissionData.AvgScore > 0 {
		scoreDiff := score - admissionData.AvgScore
		
		// 简单的概率计算模型
		if scoreDiff >= 30 {
			probability = 0.9
		} else if scoreDiff >= 15 {
			probability = 0.7
		} else if scoreDiff >= 0 {
			probability = 0.5
		} else if scoreDiff >= -15 {
			probability = 0.3
		} else {
			probability = 0.1
		}
	} else {
		probability = 0.5 // 默认概率
	}
	
	// 确定概率级别
	var level, description string
	if probability >= 0.8 {
		level = "very_high"
		description = "录取概率很高"
	} else if probability >= 0.6 {
		level = "high"
		description = "录取概率较高"
	} else if probability >= 0.4 {
		level = "medium"
		description = "录取概率中等"
	} else if probability >= 0.2 {
		level = "low"
		description = "录取概率较低"
	} else {
		level = "very_low"
		description = "录取概率很低"
	}
	
	return AdmissionChance{
		Probability: probability,
		Level:       level,
		Description: description,
	}
}

// calculateMatchScore 计算匹配度
func (s *AlgorithmService) calculateMatchScore(req VolunteerMatchRequest, candidate CandidateData) float64 {
	score := 0.0
	
	// 院校层次匹配
	if s.isPreferredLevel(candidate.University.Level, req.PreferredLevels) {
		score += 20
	}
	
	// 地理位置匹配
	if s.isPreferredProvince(candidate.University.Province, req.PreferredProvinces) {
		score += 15
	}
	
	// 专业匹配
	if candidate.Major != nil && s.isPreferredMajor(candidate.Major.Name, req.PreferredMajors) {
		score += 25
	}
	
	// 就业前景
	if candidate.Major != nil && candidate.Major.EmploymentRate > req.Preferences.MinEmploymentRate {
		score += 10
	}
	
	// 其他因素...
	
	return score
}

// determineRecommendType 确定推荐类型
func (s *AlgorithmService) determineRecommendType(probability float64) string {
	if probability >= 0.7 {
		return "safe"
	} else if probability >= 0.4 {
		return "moderate"
	} else {
		return "reach"
	}
}

// generateReasons 生成推荐理由
func (s *AlgorithmService) generateReasons(req VolunteerMatchRequest, candidate CandidateData, matchScore float64) []string {
	var reasons []string
	
	if candidate.University.Level == "985" || candidate.University.Level == "211" {
		reasons = append(reasons, "知名高校，教学质量优秀")
	}
	
	if candidate.University.Province == req.Province {
		reasons = append(reasons, "本省院校，学费相对较低")
	}
	
	if candidate.Major != nil && candidate.Major.EmploymentRate > 0.8 {
		reasons = append(reasons, "就业前景良好")
	}
	
	if matchScore > 70 {
		reasons = append(reasons, "与您的偏好高度匹配")
	}
	
	return reasons
}

// generateWarnings 生成风险提示
func (s *AlgorithmService) generateWarnings(req VolunteerMatchRequest, candidate CandidateData, admissionChance AdmissionChance) []string {
	var warnings []string
	
	if admissionChance.Probability < 0.3 {
		warnings = append(warnings, "录取概率较低，建议谨慎填报")
	}
	
	if candidate.Major != nil && candidate.Major.EmploymentRate < 0.6 {
		warnings = append(warnings, "就业率相对较低，需要考虑就业风险")
	}
	
	return warnings
}

// analyzeRisk 风险分析
func (s *AlgorithmService) analyzeRisk(req VolunteerMatchRequest, recommendations []VolunteerRecommendation) RiskAnalysis {
	safeCount := 0
	moderateCount := 0
	reachCount := 0
	
	for _, rec := range recommendations {
		switch rec.RecommendType {
		case "safe":
			safeCount++
		case "moderate":
			moderateCount++
		case "reach":
			reachCount++
		}
	}
	
	// 简单的风险评估
	overallRisk := "medium"
	if safeCount > moderateCount+reachCount {
		overallRisk = "low"
	} else if reachCount > safeCount+moderateCount {
		overallRisk = "high"
	}
	
	return RiskAnalysis{
		OverallRisk: overallRisk,
		RiskFactors: []RiskFactor{
			{
				Factor:      "志愿结构",
				Level:       overallRisk,
				Impact:      0.7,
				Description: "当前志愿结构的风险水平",
			},
		},
		Suggestions: []string{
			"建议保持保底、稳妥、冲刺志愿的合理比例",
			"关注目标院校的历年录取趋势",
		},
		PortfolioBalance: PortfolioBalance{
			SafeCount:        safeCount,
			ModerateCount:    moderateCount,
			ReachCount:       reachCount,
			RecommendedRatio: "3:4:3",
		},
	}
}

// sortAndRankRecommendations 排序并设置排名
func (s *AlgorithmService) sortAndRankRecommendations(recommendations []VolunteerRecommendation) []VolunteerRecommendation {
	// 简单排序：按匹配度降序
	for i := 0; i < len(recommendations)-1; i++ {
		for j := i + 1; j < len(recommendations); j++ {
			if recommendations[i].MatchScore < recommendations[j].MatchScore {
				recommendations[i], recommendations[j] = recommendations[j], recommendations[i]
			}
		}
	}
	
	// 设置排名
	for i := range recommendations {
		recommendations[i].Ranking = i + 1
	}
	
	return recommendations
}

// calculateConfidence 计算置信度
func (s *AlgorithmService) calculateConfidence(recommendations []VolunteerRecommendation) float64 {
	if len(recommendations) == 0 {
		return 0
	}
	
	// 基于推荐数量和匹配度计算置信度
	totalScore := 0.0
	for _, rec := range recommendations {
		totalScore += rec.MatchScore
	}
	
	avgScore := totalScore / float64(len(recommendations))
	return avgScore / 100.0 // 转换为0-1范围
}

// saveAnalysisResult 保存分析结果
func (s *AlgorithmService) saveAnalysisResult(ctx context.Context, req VolunteerMatchRequest, response *VolunteerMatchResponse) {
	// 将结果保存到数据库
	go func() {
		preferences, _ := json.Marshal(req.Preferences)
		results, _ := json.Marshal(response)
		
		analysisResult := models.AnalysisResult{
			RequestID:   response.RequestID,
			Province:    req.Province,
			Score:       req.Score,
			Rank:        req.Rank,
			Category:    req.Category,
			Preferences: string(preferences),
			Results:     string(results),
			Confidence:  response.Confidence,
			ProcessTime: response.ProcessTime,
			Algorithm:   response.Algorithm,
			Version:     response.Version,
		}
		
		// 如果有用户ID，则设置
		if req.UserID != "" {
			if userUUID, err := uuid.Parse(req.UserID); err == nil {
				analysisResult.UserID = &userUUID
			}
		}
		
		if err := s.db.PostgreSQL.Create(&analysisResult).Error; err != nil {
			s.logger.Errorf("保存分析结果失败: %v", err)
		}
	}()
}

// 辅助方法

func (s *AlgorithmService) isPreferredLevel(level string, preferredLevels []string) bool {
	for _, preferred := range preferredLevels {
		if level == preferred {
			return true
		}
	}
	return false
}

func (s *AlgorithmService) isPreferredProvince(province string, preferredProvinces []string) bool {
	for _, preferred := range preferredProvinces {
		if province == preferred {
			return true
		}
	}
	return false
}

func (s *AlgorithmService) isPreferredMajor(majorName string, preferredMajors []string) bool {
	for _, preferred := range preferredMajors {
		if majorName == preferred {
			return true
		}
	}
	return false
}