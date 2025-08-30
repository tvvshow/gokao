package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gaokao/recommendation-service/pkg/cppbridge"
)

// HybridHandler 混合推荐处理器
type HybridHandler struct {
	bridge *cppbridge.HybridRecommendationBridge
}

// NewHybridHandler 创建新的混合推荐处理器
func NewHybridHandler(bridge *cppbridge.HybridRecommendationBridge) *HybridHandler {
	return &HybridHandler{
		bridge: bridge,
	}
}

// GenerateHybridPlanRequest 生成混合推荐方案请求
type GenerateHybridPlanRequest struct {
	Student       *cppbridge.Student `json:"student" binding:"required"`
	MaxVolunteers int                `json:"max_volunteers,omitempty"`
	Options       *PlanOptions       `json:"options,omitempty"`
}

// PlanOptions 方案生成选项
type PlanOptions struct {
	EnableDiversity      bool    `json:"enable_diversity"`
	RiskTolerance       float64 `json:"risk_tolerance"`
	PreferenceWeight    float64 `json:"preference_weight"`
	EnableAdaptiveWeights bool   `json:"enable_adaptive_weights"`
}

// UpdateFusionWeightsRequest 更新融合权重请求
type UpdateFusionWeightsRequest struct {
	TraditionalWeight float64 `json:"traditional_weight" binding:"required,min=0,max=1"`
	AIWeight          float64 `json:"ai_weight" binding:"required,min=0,max=1"`
}

// CompareRecommendationsRequest 比较推荐请求
type CompareRecommendationsRequest struct {
	Student          *cppbridge.Student `json:"student" binding:"required"`
	MaxVolunteers    int                `json:"max_volunteers,omitempty"`
	ComparisonModes  []string           `json:"comparison_modes"` // ["traditional", "ai", "hybrid"]
}

// CompareRecommendationsResponse 比较推荐响应
type CompareRecommendationsResponse struct {
	Student            *cppbridge.Student                     `json:"student"`
	TraditionalPlan    *cppbridge.VolunteerPlan              `json:"traditional_plan,omitempty"`
	AIPlan             *cppbridge.VolunteerPlan              `json:"ai_plan,omitempty"`
	HybridPlan         *cppbridge.VolunteerPlan              `json:"hybrid_plan"`
	Comparison         *RecommendationComparison              `json:"comparison"`
	QualityMetrics     map[string]float64                     `json:"quality_metrics"`
}

// RecommendationComparison 推荐比较结果
type RecommendationComparison struct {
	OverlapAnalysis    *OverlapAnalysis    `json:"overlap_analysis"`
	DiversityAnalysis  *DiversityAnalysis  `json:"diversity_analysis"`
	RiskAnalysis       *RiskAnalysis       `json:"risk_analysis"`
	Recommendations    []string            `json:"recommendations"`
}

// OverlapAnalysis 重叠分析
type OverlapAnalysis struct {
	TraditionalAIOverlap     int     `json:"traditional_ai_overlap"`
	TraditionalHybridOverlap int     `json:"traditional_hybrid_overlap"`
	AIHybridOverlap          int     `json:"ai_hybrid_overlap"`
	OverallSimilarity        float64 `json:"overall_similarity"`
}

// DiversityAnalysis 多样性分析
type DiversityAnalysis struct {
	CityDiversity    map[string]int `json:"city_diversity"`
	LevelDiversity   map[string]int `json:"level_diversity"`
	MajorDiversity   map[string]int `json:"major_diversity"`
	DiversityScore   float64        `json:"diversity_score"`
}

// RiskAnalysis 风险分析
type RiskAnalysis struct {
	RushRatio       float64 `json:"rush_ratio"`
	StableRatio     float64 `json:"stable_ratio"`
	SafeRatio       float64 `json:"safe_ratio"`
	OverallRisk     float64 `json:"overall_risk"`
	RiskDistribution map[string]int `json:"risk_distribution"`
}

// GenerateHybridPlan 生成混合推荐方案
// @Summary 生成混合推荐方案
// @Description 使用混合推荐引擎为学生生成志愿填报方案
// @Tags hybrid
// @Accept json
// @Produce json
// @Param request body GenerateHybridPlanRequest true "生成混合推荐方案请求"
// @Success 200 {object} APIResponse{data=cppbridge.VolunteerPlan}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /hybrid/plan [post]
func (h *HybridHandler) GenerateHybridPlan(c *gin.Context) {
	var req GenerateHybridPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request parameters",
			Error:   err.Error(),
		})
		return
	}

	// 设置默认值
	if req.MaxVolunteers == 0 {
		req.MaxVolunteers = 96
	}

	// 验证学生信息
	if err := validateStudent(req.Student); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid student information",
			Error:   err.Error(),
		})
		return
	}

	// 生成混合推荐方案
	plan, err := h.bridge.GenerateHybridPlan(req.Student, req.MaxVolunteers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to generate hybrid plan",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Hybrid plan generated successfully",
		Data:    plan,
	})
}

// UpdateFusionWeights 更新融合权重
// @Summary 更新融合权重
// @Description 动态调整传统算法和AI推荐的融合权重
// @Tags hybrid
// @Accept json
// @Produce json
// @Param request body UpdateFusionWeightsRequest true "更新融合权重请求"
// @Success 200 {object} APIResponse{data=cppbridge.FusionWeights}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /hybrid/weights [put]
func (h *HybridHandler) UpdateFusionWeights(c *gin.Context) {
	var req UpdateFusionWeightsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request parameters",
			Error:   err.Error(),
		})
		return
	}

	// 验证权重总和
	if abs(req.TraditionalWeight+req.AIWeight-1.0) > 0.001 {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Weights must sum to 1.0",
		})
		return
	}

	// 更新权重
	err := h.bridge.SetFusionWeights(req.TraditionalWeight, req.AIWeight)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to update fusion weights",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Fusion weights updated successfully",
		Data: cppbridge.FusionWeights{
			TraditionalWeight: req.TraditionalWeight,
			AIWeight:          req.AIWeight,
		},
	})
}

// GetHybridConfig 获取混合引擎配置
// @Summary 获取混合引擎配置
// @Description 获取当前混合推荐引擎的配置信息
// @Tags hybrid
// @Produce json
// @Success 200 {object} APIResponse{data=cppbridge.HybridStats}
// @Failure 500 {object} APIResponse
// @Router /hybrid/config [get]
func (h *HybridHandler) GetHybridConfig(c *gin.Context) {
	stats, err := h.bridge.GetHybridStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to get hybrid config",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Hybrid config retrieved successfully",
		Data:    stats,
	})
}

// CompareRecommendations 比较不同推荐算法的结果
// @Summary 比较推荐算法
// @Description 比较传统算法、AI推荐和混合推荐的结果
// @Tags hybrid
// @Accept json
// @Produce json
// @Param request body CompareRecommendationsRequest true "比较推荐请求"
// @Success 200 {object} APIResponse{data=CompareRecommendationsResponse}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /hybrid/compare [post]
func (h *HybridHandler) CompareRecommendations(c *gin.Context) {
	var req CompareRecommendationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request parameters",
			Error:   err.Error(),
		})
		return
	}

	// 设置默认值
	if req.MaxVolunteers == 0 {
		req.MaxVolunteers = 96
	}
	if len(req.ComparisonModes) == 0 {
		req.ComparisonModes = []string{"traditional", "ai", "hybrid"}
	}

	// 验证学生信息
	if err := validateStudent(req.Student); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid student information",
			Error:   err.Error(),
		})
		return
	}

	response := &CompareRecommendationsResponse{
		Student:        req.Student,
		QualityMetrics: make(map[string]float64),
	}

	// 生成混合推荐（必须）
	hybridPlan, err := h.bridge.GenerateHybridPlan(req.Student, req.MaxVolunteers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to generate hybrid plan",
			Error:   err.Error(),
		})
		return
	}
	response.HybridPlan = hybridPlan

	// TODO: 这里需要集成传统算法和AI推荐的单独调用
	// if contains(req.ComparisonModes, "traditional") {
	//     traditionalPlan, err := h.traditionalMatcher.GenerateVolunteerPlan(req.Student, req.MaxVolunteers)
	//     if err == nil {
	//         response.TraditionalPlan = traditionalPlan
	//     }
	// }

	// if contains(req.ComparisonModes, "ai") {
	//     aiPlan, err := h.aiEngine.GenerateRecommendations(req.Student, req.MaxVolunteers)
	//     if err == nil {
	//         response.AIPlan = convertAIToVolunteerPlan(aiPlan)
	//     }
	// }

	// 进行比较分析
	response.Comparison = h.analyzeRecommendations(response)
	response.QualityMetrics = h.calculateQualityMetrics(response)

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Recommendations compared successfully",
		Data:    response,
	})
}

// ExplainHybridRecommendation 解释混合推荐
// @Summary 解释混合推荐
// @Description 获取特定混合推荐的详细解释
// @Tags hybrid
// @Accept json
// @Produce json
// @Param recommendation body cppbridge.VolunteerRecommendation true "推荐信息"
// @Success 200 {object} APIResponse{data=string}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /hybrid/explain [post]
func (h *HybridHandler) ExplainHybridRecommendation(c *gin.Context) {
	var recommendation cppbridge.VolunteerRecommendation
	if err := c.ShouldBindJSON(&recommendation); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid recommendation data",
			Error:   err.Error(),
		})
		return
	}

	explanation, err := h.bridge.GetHybridExplanation(&recommendation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to get explanation",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Explanation generated successfully",
		Data:    explanation,
	})
}

// 辅助函数：验证学生信息
func validateStudent(student *cppbridge.Student) error {
	if student == nil {
		return errors.New("student cannot be nil")
	}
	if student.StudentID == "" {
		return errors.New("student ID is required")
	}
	if student.TotalScore <= 0 {
		return errors.New("total score must be positive")
	}
	if student.Ranking <= 0 {
		return errors.New("ranking must be positive")
	}
	if student.Province == "" {
		return errors.New("province is required")
	}
	return nil
}

// 辅助函数：分析推荐结果
func (h *HybridHandler) analyzeRecommendations(response *CompareRecommendationsResponse) *RecommendationComparison {
	comparison := &RecommendationComparison{
		OverlapAnalysis:   &OverlapAnalysis{},
		DiversityAnalysis: &DiversityAnalysis{},
		RiskAnalysis:      &RiskAnalysis{},
		Recommendations:   []string{},
	}

	if response.HybridPlan == nil {
		return comparison
	}

	// 分析多样性
	comparison.DiversityAnalysis = h.analyzeDiversity(response.HybridPlan.Recommendations)

	// 分析风险分布
	comparison.RiskAnalysis = h.analyzeRisk(response.HybridPlan)

	// 生成建议
	comparison.Recommendations = h.generateComparisonRecommendations(response)

	// TODO: 如果有其他推荐结果，进行重叠分析
	// if response.TraditionalPlan != nil && response.AIPlan != nil {
	//     comparison.OverlapAnalysis = h.analyzeOverlap(
	//         response.TraditionalPlan.Recommendations,
	//         response.AIPlan.Recommendations,
	//         response.HybridPlan.Recommendations)
	// }

	return comparison
}

// 辅助函数：分析多样性
func (h *HybridHandler) analyzeDiversity(recommendations []cppbridge.VolunteerRecommendation) *DiversityAnalysis {
	cityCount := make(map[string]int)
	levelCount := make(map[string]int)
	majorCount := make(map[string]int)

	for _, rec := range recommendations {
		// 简化的分类方法
		city := rec.UniversityID[:2] // 假设前两位代表城市
		level := rec.UniversityID[2:3] // 假设第三位代表层次
		majorCategory := rec.MajorID[:2] // 假设前两位代表专业类别

		cityCount[city]++
		levelCount[level]++
		majorCount[majorCategory]++
	}

	// 计算多样性分数
	totalRecs := float64(len(recommendations))
	diversityScore := 1.0
	if totalRecs > 0 {
		cityDiversity := float64(len(cityCount)) / totalRecs
		levelDiversity := float64(len(levelCount)) / totalRecs
		majorDiversity := float64(len(majorCount)) / totalRecs
		diversityScore = (cityDiversity + levelDiversity + majorDiversity) / 3.0
	}

	return &DiversityAnalysis{
		CityDiversity:  cityCount,
		LevelDiversity: levelCount,
		MajorDiversity: majorCount,
		DiversityScore: diversityScore,
	}
}

// 辅助函数：分析风险
func (h *HybridHandler) analyzeRisk(plan *cppbridge.VolunteerPlan) *RiskAnalysis {
	riskDistribution := make(map[string]int)
	for _, rec := range plan.Recommendations {
		riskDistribution[rec.RiskLevel]++
	}

	total := float64(plan.TotalVolunteers)
	rushRatio := float64(plan.RushCount) / total
	stableRatio := float64(plan.StableCount) / total
	safeRatio := float64(plan.SafeCount) / total

	return &RiskAnalysis{
		RushRatio:        rushRatio,
		StableRatio:      stableRatio,
		SafeRatio:        safeRatio,
		OverallRisk:      plan.OverallRiskScore,
		RiskDistribution: riskDistribution,
	}
}

// 辅助函数：生成比较建议
func (h *HybridHandler) generateComparisonRecommendations(response *CompareRecommendationsResponse) []string {
	recommendations := []string{}

	if response.HybridPlan != nil {
		plan := response.HybridPlan

		// 基于风险分布的建议
		rushRatio := float64(plan.RushCount) / float64(plan.TotalVolunteers)
		safeRatio := float64(plan.SafeCount) / float64(plan.TotalVolunteers)

		if rushRatio > 0.4 {
			recommendations = append(recommendations, "冲刺志愿比例较高，建议适当增加稳妥志愿")
		}
		if safeRatio < 0.15 {
			recommendations = append(recommendations, "保底志愿偏少，建议增加保底选择以降低风险")
		}

		// 基于质量评分的建议
		if plan.OverallRiskScore < 0.5 {
			recommendations = append(recommendations, "整体方案偏保守，可考虑适当增加冲刺机会")
		}

		// 基于多样性的建议
		recommendations = append(recommendations, "混合推荐已优化多样性，建议关注地域和专业分布平衡")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "推荐方案整体平衡，建议根据个人偏好微调")
	}

	return recommendations
}

// 辅助函数：计算质量指标
func (h *HybridHandler) calculateQualityMetrics(response *CompareRecommendationsResponse) map[string]float64 {
	metrics := make(map[string]float64)

	if response.HybridPlan != nil {
		plan := response.HybridPlan

		// 风险平衡度
		idealRushRatio := 0.3
		idealStableRatio := 0.5
		idealSafeRatio := 0.2

		actualRushRatio := float64(plan.RushCount) / float64(plan.TotalVolunteers)
		actualStableRatio := float64(plan.StableCount) / float64(plan.TotalVolunteers)
		actualSafeRatio := float64(plan.SafeCount) / float64(plan.TotalVolunteers)

		riskBalance := 1.0 - (abs(actualRushRatio-idealRushRatio)+
			abs(actualStableRatio-idealStableRatio)+
			abs(actualSafeRatio-idealSafeRatio))/3.0

		metrics["risk_balance"] = riskBalance
		metrics["overall_risk"] = plan.OverallRiskScore

		// 多样性分数
		if response.Comparison != nil && response.Comparison.DiversityAnalysis != nil {
			metrics["diversity_score"] = response.Comparison.DiversityAnalysis.DiversityScore
		}

		// 平均匹配分数
		totalMatchScore := 0.0
		for _, rec := range plan.Recommendations {
			totalMatchScore += rec.MatchScore
		}
		if len(plan.Recommendations) > 0 {
			metrics["avg_match_score"] = totalMatchScore / float64(len(plan.Recommendations))
		}
	}

	return metrics
}

// 辅助函数：计算绝对值
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// 辅助函数：检查切片是否包含元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}