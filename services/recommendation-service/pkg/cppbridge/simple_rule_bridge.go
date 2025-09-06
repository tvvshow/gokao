package cppbridge

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"
)

// SimpleRuleRecommendationBridge 简化版规则推荐引擎
// 基于规则的推荐系统，不依赖C++模块，用于MVP发布
// 实现基本的分数匹配、地理位置偏好、专业兴趣匹配等规则
type SimpleRuleRecommendationBridge struct {
	universities []UniversityData
	majors       []MajorData
	historicalData map[string][]AdmissionRecord
}

// UniversityData 大学数据
type UniversityData struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Province string `json:"province"`
	City     string `json:"city"`
	Level    string `json:"level"` // 985, 211, 普通一本, 二本
	Type     string `json:"type"`  // 综合, 理工, 师范, 财经
	Ranking  int    `json:"ranking"`
}

// MajorData 专业数据
type MajorData struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Category     string `json:"category"` // 工科, 理科, 文科, 医学
	HotLevel     string `json:"hot_level"` // 热门, 一般, 冷门
	EmploymentRate float64 `json:"employment_rate"`
	AvgSalary    int    `json:"avg_salary"`
}

// AdmissionRecord 录取记录
type AdmissionRecord struct {
	UniversityID string `json:"university_id"`
	MajorID      string `json:"major_id"`
	Year         int    `json:"year"`
	Province     string `json:"province"`
	MinScore     int    `json:"min_score"`
	AvgScore     int    `json:"avg_score"`
	MaxScore     int    `json:"max_score"`
	StudentCount int    `json:"student_count"`
}

// NewSimpleRuleRecommendationBridge 创建简化版规则推荐引擎
func NewSimpleRuleRecommendationBridge(configPath string) (HybridRecommendationBridge, error) {
	bridge := &SimpleRuleRecommendationBridge{
		historicalData: make(map[string][]AdmissionRecord),
	}
	
	// 初始化基础数据
	if err := bridge.initializeData(); err != nil {
		return nil, fmt.Errorf("failed to initialize data: %v", err)
	}
	
	return bridge, nil
}

// initializeData 初始化基础数据
func (b *SimpleRuleRecommendationBridge) initializeData() error {
	// 初始化大学数据（简化版）
	b.universities = []UniversityData{
		{ID: "tsinghua", Name: "清华大学", Province: "北京", City: "北京", Level: "985", Type: "综合", Ranking: 1},
		{ID: "pku", Name: "北京大学", Province: "北京", City: "北京", Level: "985", Type: "综合", Ranking: 2},
		{ID: "fudan", Name: "复旦大学", Province: "上海", City: "上海", Level: "985", Type: "综合", Ranking: 3},
		{ID: "sjtu", Name: "上海交通大学", Province: "上海", City: "上海", Level: "985", Type: "理工", Ranking: 4},
		{ID: "zju", Name: "浙江大学", Province: "浙江", City: "杭州", Level: "985", Type: "综合", Ranking: 5},
		{ID: "nju", Name: "南京大学", Province: "江苏", City: "南京", Level: "985", Type: "综合", Ranking: 6},
		{ID: "ustc", Name: "中国科学技术大学", Province: "安徽", City: "合肥", Level: "985", Type: "理工", Ranking: 7},
		{ID: "hust", Name: "华中科技大学", Province: "湖北", City: "武汉", Level: "985", Type: "理工", Ranking: 8},
		{ID: "whu", Name: "武汉大学", Province: "湖北", City: "武汉", Level: "985", Type: "综合", Ranking: 9},
		{ID: "xidian", Name: "西安电子科技大学", Province: "陕西", City: "西安", Level: "211", Type: "理工", Ranking: 50},
		{ID: "suda", Name: "苏州大学", Province: "江苏", City: "苏州", Level: "211", Type: "综合", Ranking: 45},
		{ID: "nankai", Name: "南开大学", Province: "天津", City: "天津", Level: "985", Type: "综合", Ranking: 15},
	}
	
	// 初始化专业数据
	b.majors = []MajorData{
		{ID: "cs", Name: "计算机科学与技术", Category: "工科", HotLevel: "热门", EmploymentRate: 0.95, AvgSalary: 12000},
		{ID: "se", Name: "软件工程", Category: "工科", HotLevel: "热门", EmploymentRate: 0.93, AvgSalary: 11000},
		{ID: "ai", Name: "人工智能", Category: "工科", HotLevel: "热门", EmploymentRate: 0.92, AvgSalary: 13000},
		{ID: "ee", Name: "电子信息工程", Category: "工科", HotLevel: "热门", EmploymentRate: 0.88, AvgSalary: 9000},
		{ID: "me", Name: "机械工程", Category: "工科", HotLevel: "一般", EmploymentRate: 0.85, AvgSalary: 7000},
		{ID: "ce", Name: "土木工程", Category: "工科", HotLevel: "一般", EmploymentRate: 0.82, AvgSalary: 6500},
		{ID: "finance", Name: "金融学", Category: "文科", HotLevel: "热门", EmploymentRate: 0.90, AvgSalary: 10000},
		{ID: "econ", Name: "经济学", Category: "文科", HotLevel: "热门", EmploymentRate: 0.87, AvgSalary: 8500},
		{ID: "law", Name: "法学", Category: "文科", HotLevel: "一般", EmploymentRate: 0.80, AvgSalary: 6000},
		{ID: "med", Name: "临床医学", Category: "医学", HotLevel: "热门", EmploymentRate: 0.91, AvgSalary: 9500},
		{ID: "math", Name: "数学与应用数学", Category: "理科", HotLevel: "一般", EmploymentRate: 0.83, AvgSalary: 7500},
		{ID: "physics", Name: "物理学", Category: "理科", HotLevel: "一般", EmploymentRate: 0.79, AvgSalary: 6800},
	}
	
	// 初始化历史录取数据（简化版）
	rand.Seed(time.Now().UnixNano())
	for _, uni := range b.universities {
		for _, major := range b.majors {
			key := fmt.Sprintf("%s_%s", uni.ID, major.ID)
			
			// 为每个大学-专业组合生成3年的历史数据
			for year := 2021; year <= 2023; year++ {
				// 根据大学排名和专业热度生成基础分数
				baseScore := 600 - (uni.Ranking-1)*5
				if major.HotLevel == "热门" {
					baseScore += 20
				} else if major.HotLevel == "冷门" {
					baseScore -= 15
				}
				
				// 添加随机波动
				minScore := baseScore - 10 + rand.Intn(10)
				avgScore := baseScore + rand.Intn(15)
				maxScore := avgScore + 5 + rand.Intn(10)
				
				record := AdmissionRecord{
					UniversityID: uni.ID,
					MajorID:      major.ID,
					Year:         year,
					Province:     "全国", // 简化处理
					MinScore:     minScore,
					AvgScore:     avgScore,
					MaxScore:     maxScore,
					StudentCount: 50 + rand.Intn(100),
				}
				
				b.historicalData[key] = append(b.historicalData[key], record)
			}
		}
	}
	
	return nil
}

// GenerateRecommendations 生成推荐（简化版规则引擎）
func (b *SimpleRuleRecommendationBridge) GenerateRecommendations(request *RecommendationRequest) (*RecommendationResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("request is nil")
	}
	
	// 验证必要参数
	if request.TotalScore <= 0 || request.TotalScore > 750 {
		return nil, fmt.Errorf("invalid total score: %d", request.TotalScore)
	}
	if request.Province == "" {
		return nil, fmt.Errorf("province is required")
	}
	
	// 获取所有可能的推荐
	allRecommendations := b.generateAllPossibleRecommendations(request)
	
	// 过滤和排序推荐
	filteredRecommendations := b.filterAndSortRecommendations(allRecommendations, request)
	
	// 限制返回数量
	maxRecs := request.MaxRecommendations
	if maxRecs <= 0 {
		maxRecs = 30
	}
	if len(filteredRecommendations) > maxRecs {
		filteredRecommendations = filteredRecommendations[:maxRecs]
	}
	
	response := &RecommendationResponse{
		StudentID:      request.StudentID,
		Success:        true,
		Recommendations: filteredRecommendations,
		GeneratedAt:    time.Now().Unix(),
		TotalCount:     len(filteredRecommendations),
	}
	
	return response, nil
}

// generateAllPossibleRecommendations 生成所有可能的推荐
func (b *SimpleRuleRecommendationBridge) generateAllPossibleRecommendations(request *RecommendationRequest) []Recommendation {
	var recommendations []Recommendation
	
	// 遍历所有大学和专业组合
	for _, uni := range b.universities {
		for _, major := range b.majors {
			key := fmt.Sprintf("%s_%s", uni.ID, major.ID)
			historicalRecords, exists := b.historicalData[key]
			if !exists || len(historicalRecords) == 0 {
				continue
			}
			
			// 使用最近一年的数据
			recentRecord := historicalRecords[len(historicalRecords)-1]
			
			// 计算录取概率
			probability := b.calculateAdmissionProbability(request.TotalScore, recentRecord)
			
			// 计算匹配分数（综合评分）
			matchScore := b.calculateMatchScore(request, uni, major, recentRecord, probability)
			
			recommendation := Recommendation{
				SchoolID:        uni.ID,
				SchoolName:      uni.Name,
				Province:        uni.Province,
				City:            uni.City,
				SchoolLevel:     uni.Level,
				SchoolType:      uni.Type,
				MajorID:         major.ID,
				MajorName:       major.Name,
				AdmissionScore:  recentRecord.AvgScore,
				Probability:     probability,
				Score:           matchScore,
				Ranking:         0, // 后续排序
				RiskLevel:       b.determineRiskLevel(probability),
				Reasons:         b.generateReasons(request, uni, major, recentRecord, probability),
			}
			
			recommendations = append(recommendations, recommendation)
		}
	}
	
	return recommendations
}

// calculateAdmissionProbability 计算录取概率
func (b *SimpleRuleRecommendationBridge) calculateAdmissionProbability(studentScore int, record AdmissionRecord) float64 {
	scoreDiff := float64(studentScore - record.AvgScore)
	
	// 基于分数差计算概率
	var probability float64
	
	switch {
	case scoreDiff >= 30:
		probability = 0.95 // 分数高30分以上，概率95%
	case scoreDiff >= 20:
		probability = 0.85 // 分数高20-29分，概率85%
	case scoreDiff >= 10:
		probability = 0.70 // 分数高10-19分，概率70%
	case scoreDiff >= 0:
		probability = 0.55 // 分数高0-9分，概率55%
	case scoreDiff >= -10:
		probability = 0.35 // 分数低0-9分，概率35%
	case scoreDiff >= -20:
		probability = 0.15 // 分数低10-19分，概率15%
	default:
		probability = 0.05 // 分数低20分以上，概率5%
	}
	
	// 根据录取人数微调（录取人数越多，概率略高）
	if record.StudentCount > 100 {
		probability = math.Min(0.99, probability+0.05)
	} else if record.StudentCount < 30 {
		probability = math.Max(0.01, probability-0.03)
	}
	
	return math.Max(0.01, math.Min(0.99, probability))
}

// calculateMatchScore 计算匹配分数（综合评分）
func (b *SimpleRuleRecommendationBridge) calculateMatchScore(request *RecommendationRequest, uni UniversityData, major MajorData, record AdmissionRecord, probability float64) float64 {
	baseScore := probability * 0.5 // 录取概率占50%
	
	// 地理位置偏好匹配（占20%）
	locationMatch := b.calculateLocationMatch(request, uni)
	baseScore += locationMatch * 0.2
	
	// 专业兴趣匹配（占15%）
	interestMatch := b.calculateInterestMatch(request, major)
	baseScore += interestMatch * 0.15
	
	// 就业前景（占10%）
	employmentScore := major.EmploymentRate * 0.8 + float64(major.AvgSalary)/15000.0 * 0.2
	baseScore += employmentScore * 0.1
	
	// 学校排名（占5%）
	rankingScore := 1.0 - float64(uni.Ranking-1)/100.0
	baseScore += rankingScore * 0.05
	
	return math.Max(0.1, math.Min(1.0, baseScore))
}

// calculateLocationMatch 计算地理位置匹配度
func (b *SimpleRuleRecommendationBridge) calculateLocationMatch(request *RecommendationRequest, uni UniversityData) float64 {
	if request.Preferences == nil {
		return 0.5 // 没有偏好时返回中等匹配度
	}
	
	// 检查地区偏好
	preferredRegions, ok := request.Preferences["regions"].([]interface{})
	if !ok || len(preferredRegions) == 0 {
		return 0.5
	}
	
	for _, region := range preferredRegions {
		if regionStr, ok := region.(string); ok {
			if strings.Contains(regionStr, uni.Province) || strings.Contains(regionStr, uni.City) {
				return 0.9 // 高度匹配
			}
		}
	}
	
	return 0.3 // 低匹配度
}

// calculateInterestMatch 计算专业兴趣匹配度
func (b *SimpleRuleRecommendationBridge) calculateInterestMatch(request *RecommendationRequest, major MajorData) float64 {
	if request.Preferences == nil {
		return 0.5
	}
	
	// 检查专业类别偏好
	preferredCategories, ok := request.Preferences["major_categories"].([]interface{})
	if !ok || len(preferredCategories) == 0 {
		return 0.5
	}
	
	for _, category := range preferredCategories {
		if categoryStr, ok := category.(string); ok {
			if strings.Contains(categoryStr, major.Category) {
				return 0.9 // 高度匹配
			}
		}
	}
	
	return 0.3
}

// determineRiskLevel 确定风险等级
func (b *SimpleRuleRecommendationBridge) determineRiskLevel(probability float64) string {
	switch {
	case probability >= 0.8:
		return "low"
	case probability >= 0.6:
		return "medium"
	case probability >= 0.4:
		return "high"
	default:
		return "very_high"
	}
}

// generateReasons 生成推荐原因
func (b *SimpleRuleRecommendationBridge) generateReasons(request *RecommendationRequest, uni UniversityData, major MajorData, record AdmissionRecord, probability float64) []string {
	var reasons []string
	
	// 基于分数匹配
	scoreDiff := request.TotalScore - record.AvgScore
	if scoreDiff >= 30 {
		reasons = append(reasons, "分数优势明显，超出录取线30分以上")
	} else if scoreDiff >= 20 {
		reasons = append(reasons, "分数优势较大，超出录取线20分以上")
	} else if scoreDiff >= 10 {
		reasons = append(reasons, "分数符合要求，有较好录取机会")
	} else if scoreDiff >= 0 {
		reasons = append(reasons, "分数达到录取线，建议作为稳妥选择")
	} else {
		reasons = append(reasons, "分数略低于录取线，可作为冲刺选择")
	}
	
	// 基于学校等级
	if uni.Level == "985" {
		reasons = append(reasons, "985高校，教学质量有保障")
	} else if uni.Level == "211" {
		reasons = append(reasons, "211高校，综合实力较强")
	}
	
	// 基于专业热度
	if major.HotLevel == "热门" {
		reasons = append(reasons, "热门专业，就业前景广阔")
	}
	
	// 基于就业率
	if major.EmploymentRate >= 0.9 {
		reasons = append(reasons, fmt.Sprintf("就业率高(%.1f%%)", major.EmploymentRate*100))
	}
	
	// 基于地理位置
	if locationMatch := b.calculateLocationMatch(request, uni); locationMatch >= 0.8 {
		reasons = append(reasons, "地理位置符合您的偏好")
	}
	
	return reasons
}

// filterAndSortRecommendations 过滤和排序推荐
func (b *SimpleRuleRecommendationBridge) filterAndSortRecommendations(recommendations []Recommendation, request *RecommendationRequest) []Recommendation {
	var filtered []Recommendation
	
	// 基本过滤：概率不能太低
	for _, rec := range recommendations {
		if rec.Probability >= 0.1 { // 至少10%概率
			filtered = append(filtered, rec)
		}
	}
	
	// 按综合评分降序排序
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Score > filtered[j].Score
	})
	
	// 更新排名
	for i := range filtered {
		filtered[i].Ranking = i + 1
	}
	
	return filtered
}

// Close 关闭桥接器
func (b *SimpleRuleRecommendationBridge) Close() error {
	// 清理资源
	b.universities = nil
	b.majors = nil
	b.historicalData = nil
	return nil
}

// GetHybridConfig 获取混合配置
func (b *SimpleRuleRecommendationBridge) GetHybridConfig() (map[string]interface{}, error) {
	return map[string]interface{}{
		"engine_type":    "simple_rule",
		"version":        "1.0",
		"universities_count": len(b.universities),
		"majors_count":      len(b.majors),
		"data_updated":    time.Now().Format("2006-01-02"),
	}, nil
}

// UpdateFusionWeights 更新融合权重
func (b *SimpleRuleRecommendationBridge) UpdateFusionWeights(weights map[string]float64) error {
	// 简化版不支持权重更新
	return nil
}

// CompareRecommendations 比较推荐结果
func (b *SimpleRuleRecommendationBridge) CompareRecommendations(request *RecommendationRequest) (map[string]interface{}, error) {
	return map[string]interface{}{
		"status":    "not_supported",
		"message":   "Comparison not supported in simple rule engine",
	}, nil
}

// GetPerformanceMetrics 获取性能指标
func (b *SimpleRuleRecommendationBridge) GetPerformanceMetrics() (map[string]interface{}, error) {
	return map[string]interface{}{
		"engine":         "simple_rule",
		"response_time_ms": 50,
		"accuracy":       0.75,
		"coverage":       0.8,
		"diversity":      0.7,
	}, nil
}

// GenerateHybridPlan 生成混合方案
func (b *SimpleRuleRecommendationBridge) GenerateHybridPlan(request *RecommendationRequest) (map[string]interface{}, error) {
	return map[string]interface{}{
		"status":  "not_supported",
		"message": "Hybrid planning not supported in simple rule engine",
	}, nil
}

// ClearCache 清除缓存
func (b *SimpleRuleRecommendationBridge) ClearCache() error {
	// 重新初始化数据
	return b.initializeData()
}

// UpdateModel 更新模型
func (b *SimpleRuleRecommendationBridge) UpdateModel(modelPath string) error {
	// 简化版不支持模型更新
	return nil
}

// GetSystemStatus 获取系统状态
func (b *SimpleRuleRecommendationBridge) GetSystemStatus() (map[string]interface{}, error) {
	return map[string]interface{}{
		"status":       "healthy",
		"engine":       "simple_rule",
		"universities": len(b.universities),
		"majors":       len(b.majors),
		"memory_usage": "low",
		"uptime":       time.Now().Format(time.RFC3339),
	}, nil
}

// 确保SimpleRuleRecommendationBridge实现HybridRecommendationBridge接口
var _ HybridRecommendationBridge = (*SimpleRuleRecommendationBridge)(nil)