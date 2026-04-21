package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// 增强版推荐算法结构体
type EnhancedRecommendationEngine struct {
	historicalData map[string][]HistoricalAdmission // 历史录取数据
	userPreferences map[string]float64 // 用户偏好权重
	marketTrends map[string]float64 // 就业市场趋势
}

// 历史录取数据结构
type HistoricalAdmission struct {
	Year int
	MinScore int
	MaxScore int
	AvgScore int
	MinRank int
	MaxRank int
	AvgRank int
	AdmissionCount int
	ApplicationCount int
}

// 增强版推荐请求
type EnhancedRecommendationRequest struct {
	RecommendationRequest // 继承原有请求
	PersonalityType string // 性格类型
	CareerGoals []string // 职业目标
	FamilyIncome int // 家庭收入
	GeographicFlexibility float64 // 地域灵活度 0-1
	MajorFlexibility float64 // 专业灵活度 0-1
	RiskTolerance float64 // 风险承受度 0-1
}

// 增强版推荐结果
type EnhancedRecommendationResult struct {
	RecommendationResult // 继承原有结果
	ConfidenceScore float64 // 推荐置信度
	PersonalizationScore float64 // 个性化匹配度
	MarketOutlook string // 就业市场前景
	AlternativeOptions []RecommendedUniversity // 备选方案
}

// 初始化增强推荐引擎
func NewEnhancedRecommendationEngine() *EnhancedRecommendationEngine {
	return &EnhancedRecommendationEngine{
		historicalData: make(map[string][]HistoricalAdmission),
		userPreferences: make(map[string]float64),
		marketTrends: make(map[string]float64),
	}
}

// 增强版推荐算法主函数
func (e *EnhancedRecommendationEngine) GenerateEnhancedRecommendations(req EnhancedRecommendationRequest) []EnhancedRecommendationResult {
	results := []EnhancedRecommendationResult{}

	// 动态调整风险策略
	adjustedRiskLevel := e.adjustRiskLevel(req)

	// 根据调整后的风险等级生成方案
	switch adjustedRiskLevel {
	case "保守型":
		results = append(results, e.generateEnhancedScheme("冲刺", req, 1))
		results = append(results, e.generateEnhancedScheme("稳妥", req, 5))
		results = append(results, e.generateEnhancedScheme("保底", req, 4))
	case "稳健型":
		results = append(results, e.generateEnhancedScheme("冲刺", req, 2))
		results = append(results, e.generateEnhancedScheme("稳妥", req, 5))
		results = append(results, e.generateEnhancedScheme("保底", req, 3))
	case "激进型":
		results = append(results, e.generateEnhancedScheme("冲刺", req, 4))
		results = append(results, e.generateEnhancedScheme("稳妥", req, 4))
		results = append(results, e.generateEnhancedScheme("保底", req, 2))
	default:
		results = append(results, e.generateEnhancedScheme("冲刺", req, 2))
		results = append(results, e.generateEnhancedScheme("稳妥", req, 5))
		results = append(results, e.generateEnhancedScheme("保底", req, 3))
	}

	return results
}

// 动态调整风险等级
func (e *EnhancedRecommendationEngine) adjustRiskLevel(req EnhancedRecommendationRequest) string {
	// 基于多个因素调整风险等级
	riskScore := req.RiskTolerance

	// 考虑家庭收入因素
	if req.FamilyIncome < 50000 {
		riskScore *= 0.8 // 降低风险承受度
	} else if req.FamilyIncome > 200000 {
		riskScore *= 1.2 // 提高风险承受度
	}

	// 考虑地域灵活度
	riskScore += req.GeographicFlexibility * 0.2

	// 考虑专业灵活度
	riskScore += req.MajorFlexibility * 0.2

	// 根据调整后的风险分数确定等级
	if riskScore <= 0.3 {
		return "保守型"
	} else if riskScore <= 0.7 {
		return "稳健型"
	} else {
		return "激进型"
	}
}

// 生成增强版推荐方案
func (e *EnhancedRecommendationEngine) generateEnhancedScheme(schemeType string, req EnhancedRecommendationRequest, count int) EnhancedRecommendationResult {
	recommendations := []RecommendedUniversity{}

	// 动态计算分数调整范围
	scoreRange := e.calculateDynamicScoreRange(req, schemeType)
	probRange := e.calculateDynamicProbRange(req, schemeType)

	// 筛选候选高校
	candidateUniversities := e.filterUniversitiesEnhanced(req, scoreRange)

	// 计算个性化匹配度并排序
	scoredCandidates := e.calculatePersonalizedScores(candidateUniversities, req)

	// 选择top N个推荐
	for i, candidate := range scoredCandidates {
		if i >= count {
			break
		}

		// 选择最佳专业
		major := e.selectBestMajorEnhanced(candidate.University, req)

		// 计算增强版录取概率
		admissionProb := e.calculateEnhancedAdmissionProbability(req.Score, candidate.University, probRange, req)

		// 计算分数和排名差距
		scoreDiff := req.Score - e.getUniversityMinScoreEnhanced(candidate.University, req.Year)
		rankDiff := req.Rank - e.getUniversityMinRankEnhanced(candidate.University, req.Year)

		// 生成增强版推荐理由
		reason := e.generateEnhancedRecommendReason(candidate.University, major, schemeType, admissionProb, req)

		// 确定增强版风险等级
		riskLevel := e.determineEnhancedRiskLevel(admissionProb, candidate.PersonalizedScore)

		recommendation := RecommendedUniversity{
			University: candidate.University,
			Major: major,
			AdmissionProb: admissionProb,
			ScoreDifference: scoreDiff,
			RankDifference: rankDiff,
			RecommendReason: reason,
			RiskLevel: riskLevel,
		}

		recommendations = append(recommendations, recommendation)
	}

	// 生成增强版风险警告和建议
	riskWarning := e.generateEnhancedRiskWarning(schemeType, recommendations, req)
	suggestions := e.generateEnhancedSuggestions(schemeType, req, recommendations)

	// 计算置信度和个性化匹配度
	confidenceScore := e.calculateConfidenceScore(recommendations, req)
	personalizationScore := e.calculatePersonalizationScore(recommendations, req)

	// 获取就业市场前景
	marketOutlook := e.getMarketOutlook(recommendations)

	// 生成备选方案
	alternativeOptions := e.generateAlternativeOptions(recommendations, req)

	return EnhancedRecommendationResult{
		RecommendationResult: RecommendationResult{
			SchemeType: schemeType,
			Universities: recommendations,
			TotalCount: len(recommendations),
			RiskWarning: riskWarning,
			Suggestions: suggestions,
		},
		ConfidenceScore: confidenceScore,
		PersonalizationScore: personalizationScore,
		MarketOutlook: marketOutlook,
		AlternativeOptions: alternativeOptions,
	}
}

// 动态计算分数调整范围
func (e *EnhancedRecommendationEngine) calculateDynamicScoreRange(req EnhancedRecommendationRequest, schemeType string) [2]int {
	baseAdjustment := map[string]int{
		"冲刺": 20,
		"稳妥": 0,
		"保底": -30,
	}

	// 根据历史数据波动调整
	volatility := e.calculateScoreVolatility(req.Province, req.Year)
	adjustment := baseAdjustment[schemeType]

	// 考虑波动性
	if volatility > 0.2 {
		// 高波动性，扩大范围
		return [2]int{adjustment - 10, adjustment + 10}
	} else {
		// 低波动性，缩小范围
		return [2]int{adjustment - 5, adjustment + 5}
	}
}

// 动态计算概率范围
func (e *EnhancedRecommendationEngine) calculateDynamicProbRange(req EnhancedRecommendationRequest, schemeType string) [2]float64 {
	baseProbRange := map[string][2]float64{
		"冲刺": {0.1, 0.3},
		"稳妥": {0.6, 0.8},
		"保底": {0.9, 0.99},
	}

	baseRange := baseProbRange[schemeType]

	// 根据用户风险承受度调整
	adjustmentFactor := (req.RiskTolerance - 0.5) * 0.2
	return [2]float64{
		math.Max(0.01, baseRange[0]+adjustmentFactor),
		math.Min(0.99, baseRange[1]+adjustmentFactor),
	}
}

// 计算分数波动性
func (e *EnhancedRecommendationEngine) calculateScoreVolatility(province string, year int) float64 {
	// 模拟计算历史分数波动性
	// 实际实现中应该基于真实历史数据
	return 0.15 // 默认15%的波动性
}

// 候选大学评分结构
type ScoredCandidate struct {
	University ExtendedUniversity
	PersonalizedScore float64
}

// 计算个性化匹配度
func (e *EnhancedRecommendationEngine) calculatePersonalizedScores(universities []ExtendedUniversity, req EnhancedRecommendationRequest) []ScoredCandidate {
	scored := []ScoredCandidate{}

	for _, uni := range universities {
		score := 0.0

		// 地理位置偏好 (权重: 0.2)
		if req.PreferredArea != "" && req.PreferredArea != "不限" {
			if uni.Province == req.PreferredArea {
				score += 0.2
			} else {
				score += 0.2 * req.GeographicFlexibility
			}
		} else {
			score += 0.2
		}

		// 学校层次偏好 (权重: 0.3)
		if uni.Level == "985" {
			score += 0.3
		} else if uni.Level == "211" {
			score += 0.25
		} else {
			score += 0.15
		}

		// 专业匹配度 (权重: 0.25)
		majorMatch := e.calculateMajorMatchScore(uni, req)
		score += 0.25 * majorMatch

		// 就业前景 (权重: 0.15)
		employmentScore := e.calculateEmploymentScore(uni, req)
		score += 0.15 * employmentScore

		// 经济因素 (权重: 0.1)
		economicScore := e.calculateEconomicScore(uni, req)
		score += 0.1 * economicScore

		scored = append(scored, ScoredCandidate{
			University: uni,
			PersonalizedScore: score,
		})
	}

	// 按个性化得分排序
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].PersonalizedScore > scored[j].PersonalizedScore
	})

	return scored
}

// 计算专业匹配度
func (e *EnhancedRecommendationEngine) calculateMajorMatchScore(uni ExtendedUniversity, req EnhancedRecommendationRequest) float64 {
	if req.PreferredMajor == "" || req.PreferredMajor == "不限" {
		return 1.0
	}

	// 检查该校是否有匹配的专业
	for _, major := range extendedMajors {
		if major.UniversityID == uni.ID {
			if major.Category == req.PreferredMajor {
				return 1.0
			}
			// 部分匹配
			if strings.Contains(major.Name, req.PreferredMajor) {
				return 0.8
			}
		}
	}

	return req.MajorFlexibility
}

// 计算就业得分
func (e *EnhancedRecommendationEngine) calculateEmploymentScore(uni ExtendedUniversity, req EnhancedRecommendationRequest) float64 {
	// 基于学校层次和地理位置计算就业得分
	score := 0.5

	if uni.Level == "985" {
		score += 0.4
	} else if uni.Level == "211" {
		score += 0.3
	} else {
		score += 0.1
	}

	// 一线城市加分
	firstTierCities := []string{"北京", "上海", "广东", "深圳"}
	for _, city := range firstTierCities {
		if uni.Province == city {
			score += 0.1
			break
		}
	}

	return math.Min(1.0, score)
}

// 计算经济得分
func (e *EnhancedRecommendationEngine) calculateEconomicScore(uni ExtendedUniversity, req EnhancedRecommendationRequest) float64 {
	// 基于家庭收入和学校所在地消费水平计算经济得分
	score := 0.5

	// 本省院校学费较低
	if uni.Province == req.Province {
		score += 0.3
	}

	// 根据家庭收入调整
	if req.FamilyIncome > 100000 {
		score += 0.2
	} else if req.FamilyIncome < 50000 {
		// 低收入家庭，偏向本省和费用较低的学校
		if uni.Province == req.Province {
			score += 0.2
		} else {
			score -= 0.2
		}
	}

	return math.Max(0.0, math.Min(1.0, score))
}

// 增强版高校筛选
func (e *EnhancedRecommendationEngine) filterUniversitiesEnhanced(req EnhancedRecommendationRequest, scoreRange [2]int) []ExtendedUniversity {
	filtered := []ExtendedUniversity{}

	for _, uni := range extendedUniversities {
		// 基础筛选条件
		if req.PreferredArea != "" && req.PreferredArea != "不限" {
			if uni.Province != req.PreferredArea && req.GeographicFlexibility < 0.5 {
				continue
			}
		}

		if req.PreferredType != "" && req.PreferredType != "不限" {
			if uni.Type != req.PreferredType {
				continue
			}
		}

		// 增强版分数筛选
		minScore := e.getUniversityMinScoreEnhanced(uni, req.Year)
		scoreDiff := req.Score - minScore

		if scoreDiff >= scoreRange[0] && scoreDiff <= scoreRange[1]+20 {
			filtered = append(filtered, uni)
		}
	}

	return filtered
}

// 增强版最低分数获取
func (e *EnhancedRecommendationEngine) getUniversityMinScoreEnhanced(uni ExtendedUniversity, year int) int {
	// 基于历史数据和趋势预测
	baseScore := 500

	// 根据学校层次调整
	switch uni.Level {
	case "985":
		baseScore = 620
	case "211":
		baseScore = 580
	default:
		baseScore = 520
	}

	// 根据全国排名微调
	if uni.NationalRank <= 10 {
		baseScore += 30
	} else if uni.NationalRank <= 50 {
		baseScore += 15
	} else if uni.NationalRank <= 100 {
		baseScore += 5
	}

	// 考虑年度趋势
	yearTrend := e.calculateYearTrend(year)
	baseScore += int(float64(baseScore) * yearTrend)

	return baseScore
}

// 增强版最低排名获取
func (e *EnhancedRecommendationEngine) getUniversityMinRankEnhanced(uni ExtendedUniversity, year int) int {
	baseRank := 50000

	switch uni.Level {
	case "985":
		baseRank = 5000
	case "211":
		baseRank = 15000
	default:
		baseRank = 30000
	}

	// 根据全国排名调整
	if uni.NationalRank <= 10 {
		baseRank = int(float64(baseRank) * 0.3)
	} else if uni.NationalRank <= 50 {
		baseRank = int(float64(baseRank) * 0.6)
	} else if uni.NationalRank <= 100 {
		baseRank = int(float64(baseRank) * 0.8)
	}

	return baseRank
}

// 计算年度趋势
func (e *EnhancedRecommendationEngine) calculateYearTrend(year int) float64 {
	// 模拟年度分数趋势，实际应基于历史数据
	if year >= 2024 {
		return 0.02 // 2%的增长趋势
	}
	return 0.0
}

// 增强版录取概率计算
func (e *EnhancedRecommendationEngine) calculateEnhancedAdmissionProbability(score int, uni ExtendedUniversity, probRange [2]float64, req EnhancedRecommendationRequest) float64 {
	minScore := e.getUniversityMinScoreEnhanced(uni, req.Year)
	scoreDiff := score - minScore

	// 基础概率计算
	baseProb := 0.5
	if scoreDiff > 30 {
		baseProb = 0.9
	} else if scoreDiff > 15 {
		baseProb = 0.75
	} else if scoreDiff > 0 {
		baseProb = 0.6
	} else if scoreDiff > -15 {
		baseProb = 0.4
	} else {
		baseProb = 0.2
	}

	// 考虑历史波动性
	volatility := e.calculateScoreVolatility(req.Province, req.Year)
	if volatility > 0.2 {
		baseProb *= 0.9 // 高波动性降低确定性
	}

	// 考虑竞争激烈程度
	competitionFactor := e.calculateCompetitionFactor(uni, req)
	baseProb *= competitionFactor

	// 考虑专业热门程度
	majorPopularity := e.calculateMajorPopularity(req.PreferredMajor)
	baseProb *= (1.0 - majorPopularity*0.2)

	// 限制在指定范围内
	return math.Max(probRange[0], math.Min(probRange[1], baseProb))
}

// 计算竞争因子
func (e *EnhancedRecommendationEngine) calculateCompetitionFactor(uni ExtendedUniversity, req EnhancedRecommendationRequest) float64 {
	factor := 1.0

	// 985/211学校竞争更激烈
	if uni.Level == "985" {
		factor *= 0.85
	} else if uni.Level == "211" {
		factor *= 0.9
	}

	// 热门地区竞争激烈
	hotRegions := []string{"北京", "上海", "广东", "江苏", "浙江"}
	for _, region := range hotRegions {
		if uni.Province == region {
			factor *= 0.9
			break
		}
	}

	return factor
}

// 计算专业热门程度
func (e *EnhancedRecommendationEngine) calculateMajorPopularity(major string) float64 {
	// 模拟专业热门程度
	hotMajors := map[string]float64{
		"计算机": 0.8,
		"金融": 0.7,
		"医学": 0.6,
		"法学": 0.5,
	}

	if popularity, exists := hotMajors[major]; exists {
		return popularity
	}
	return 0.3 // 默认热门程度
}

// 增强版专业选择
func (e *EnhancedRecommendationEngine) selectBestMajorEnhanced(uni ExtendedUniversity, req EnhancedRecommendationRequest) ExtendedMajor {
	if req.PreferredMajor == "" || req.PreferredMajor == "不限" {
		// 根据就业前景和个人特质推荐专业
		return e.recommendMajorByCareerGoals(uni, req)
	}

	// 寻找匹配的专业
	for _, major := range extendedMajors {
		if major.UniversityID == uni.ID {
			if major.Category == req.PreferredMajor || strings.Contains(major.Name, req.PreferredMajor) {
				return major
			}
		}
	}

	// 如果没有完全匹配，返回该校的热门专业
	return e.getPopularMajor(uni)
}

// 根据职业目标推荐专业
func (e *EnhancedRecommendationEngine) recommendMajorByCareerGoals(uni ExtendedUniversity, req EnhancedRecommendationRequest) ExtendedMajor {
	// 基于职业目标匹配专业
	for _, goal := range req.CareerGoals {
		for _, major := range extendedMajors {
			if major.UniversityID == uni.ID {
				if e.isCareerMajorMatch(goal, major.Category) {
					return major
				}
			}
		}
	}

	return e.getPopularMajor(uni)
}

// 职业专业匹配
func (e *EnhancedRecommendationEngine) isCareerMajorMatch(career, major string) bool {
	careerMajorMap := map[string][]string{
		"软件工程师": {"计算机", "软件工程"},
		"金融分析师": {"金融", "经济学"},
		"医生": {"医学", "临床医学"},
		"律师": {"法学"},
	}

	if majors, exists := careerMajorMap[career]; exists {
		for _, m := range majors {
			if strings.Contains(major, m) {
				return true
			}
		}
	}
	return false
}

// 获取热门专业
func (e *EnhancedRecommendationEngine) getPopularMajor(uni ExtendedUniversity) ExtendedMajor {
	for _, major := range extendedMajors {
		if major.UniversityID == uni.ID {
			return major
		}
	}
	return ExtendedMajor{Name: "通用专业", Category: "工学"}
}

// 增强版推荐理由生成
func (e *EnhancedRecommendationEngine) generateEnhancedRecommendReason(uni ExtendedUniversity, major ExtendedMajor, schemeType string, prob float64, req EnhancedRecommendationRequest) string {
	reason := []string{}

	// 学校层次理由
	if uni.Level == "985" {
		reason = append(reason, "985工程重点大学，学术声誉卓越")
	} else if uni.Level == "211" {
		reason = append(reason, "211工程重点大学，教学质量优秀")
	} else {
		reason = append(reason, "综合实力较强的本科院校")
	}

	// 专业匹配理由
	if req.PreferredMajor != "" && req.PreferredMajor != "不限" {
		if strings.Contains(major.Name, req.PreferredMajor) {
			reason = append(reason, fmt.Sprintf("专业'%s'与您的偏好高度匹配", major.Name))
		}
	}

	// 地理位置理由
	if uni.Province == req.Province {
		reason = append(reason, "本省院校，学费相对较低")
	} else if req.GeographicFlexibility > 0.7 {
		reason = append(reason, "外省优质教育资源，拓宽视野")
	}

	// 录取概率理由
	if prob > 0.8 {
		reason = append(reason, "录取概率很高，建议重点考虑")
	} else if prob > 0.6 {
		reason = append(reason, "录取概率较高，值得填报")
	} else if prob > 0.3 {
		reason = append(reason, "有一定录取机会，可作为冲刺目标")
	} else {
		reason = append(reason, "录取难度较大，需谨慎考虑")
	}

	// 方案类型特定理由
	switch schemeType {
	case "冲刺":
		reason = append(reason, "冲刺方案：挑战更高层次院校")
	case "稳妥":
		reason = append(reason, "稳妥方案：录取把握较大")
	case "保底":
		reason = append(reason, "保底方案：确保有学可上")
	}

	return strings.Join(reason, "；")
}

// 增强版风险等级确定
func (e *EnhancedRecommendationEngine) determineEnhancedRiskLevel(prob float64, personalizedScore float64) string {
	// 综合考虑录取概率和个性化匹配度
	combinedScore := prob*0.7 + personalizedScore*0.3

	if combinedScore > 0.8 {
		return "低风险"
	} else if combinedScore > 0.5 {
		return "中等风险"
	} else {
		return "高风险"
	}
}

// 增强版风险警告生成
func (e *EnhancedRecommendationEngine) generateEnhancedRiskWarning(schemeType string, recommendations []RecommendedUniversity, req EnhancedRecommendationRequest) string {
	warnings := []string{}

	// 基于方案类型的警告
	switch schemeType {
	case "冲刺":
		warnings = append(warnings, "冲刺方案风险较高，建议合理搭配稳妥和保底方案")
	case "稳妥":
		warnings = append(warnings, "稳妥方案相对安全，但仍需关注专业录取分数线")
	case "保底":
		warnings = append(warnings, "保底方案录取概率高，建议优先选择心仪专业")
	}

	// 基于推荐结果的警告
	highRiskCount := 0
	for _, rec := range recommendations {
		if rec.RiskLevel == "高风险" {
			highRiskCount++
		}
	}

	if highRiskCount > len(recommendations)/2 {
		warnings = append(warnings, "当前推荐中高风险院校较多，建议适当降低期望")
	}

	// 基于用户特征的警告
	if req.GeographicFlexibility < 0.3 {
		warnings = append(warnings, "地域限制较严格，可能错过优质外省院校")
	}

	if req.MajorFlexibility < 0.3 {
		warnings = append(warnings, "专业选择较固定，建议考虑相关专业以增加录取机会")
	}

	return strings.Join(warnings, "；")
}

// 增强版建议生成
func (e *EnhancedRecommendationEngine) generateEnhancedSuggestions(schemeType string, req EnhancedRecommendationRequest, recommendations []RecommendedUniversity) []string {
	suggestions := []string{}

	// 通用建议
	suggestions = append(suggestions, "建议合理搭配不同层次的院校，形成梯度填报")
	suggestions = append(suggestions, "关注各院校历年录取分数线变化趋势")
	suggestions = append(suggestions, "考虑专业的就业前景和个人兴趣")

	// 基于方案类型的建议
	switch schemeType {
	case "冲刺":
		suggestions = append(suggestions, "冲刺院校建议选择相对冷门但实力强的专业")
		suggestions = append(suggestions, "关注征集志愿机会")
	case "稳妥":
		suggestions = append(suggestions, "稳妥方案可适当考虑热门专业")
		suggestions = append(suggestions, "注意专业录取规则，避免被调剂")
	case "保底":
		suggestions = append(suggestions, "保底院校建议优先选择优势专业")
		suggestions = append(suggestions, "可考虑地理位置较好的院校")
	}

	// 个性化建议
	if req.FamilyIncome < 50000 {
		suggestions = append(suggestions, "建议关注国家助学政策和奖学金机会")
		suggestions = append(suggestions, "优先考虑本省院校以降低生活成本")
	}

	if len(req.CareerGoals) > 0 {
		suggestions = append(suggestions, "建议选择与职业目标匹配度高的专业")
	}

	return suggestions
}

// 计算置信度
func (e *EnhancedRecommendationEngine) calculateConfidenceScore(recommendations []RecommendedUniversity, req EnhancedRecommendationRequest) float64 {
	if len(recommendations) == 0 {
		return 0.0
	}

	totalProb := 0.0
	for _, rec := range recommendations {
		totalProb += rec.AdmissionProb
	}

	avgProb := totalProb / float64(len(recommendations))

	// 基于数据完整性调整置信度
	dataCompleteness := 0.8 // 假设数据完整性为80%

	return avgProb * dataCompleteness
}

// 计算个性化匹配度
func (e *EnhancedRecommendationEngine) calculatePersonalizationScore(recommendations []RecommendedUniversity, req EnhancedRecommendationRequest) float64 {
	if len(recommendations) == 0 {
		return 0.0
	}

	// 基于用户偏好的匹配程度
	matchScore := 0.0

	for _, rec := range recommendations {
		score := 0.0

		// 地域匹配
		if req.PreferredArea != "" && req.PreferredArea != "不限" {
			if rec.University.Province == req.PreferredArea {
				score += 0.3
			}
		} else {
			score += 0.3
		}

		// 专业匹配
		if req.PreferredMajor != "" && req.PreferredMajor != "不限" {
			if strings.Contains(rec.Major.Name, req.PreferredMajor) {
				score += 0.4
			}
		} else {
			score += 0.4
		}

		// 学校类型匹配
		if req.PreferredType != "" && req.PreferredType != "不限" {
			if rec.University.Type == req.PreferredType {
				score += 0.3
			}
		} else {
			score += 0.3
		}

		matchScore += score
	}

	return matchScore / float64(len(recommendations))
}

// 获取就业市场前景
func (e *EnhancedRecommendationEngine) getMarketOutlook(recommendations []RecommendedUniversity) string {
	if len(recommendations) == 0 {
		return "数据不足"
	}

	// 分析推荐院校的就业前景
	top985Count := 0
	top211Count := 0

	for _, rec := range recommendations {
		if rec.University.Level == "985" {
			top985Count++
		} else if rec.University.Level == "211" {
			top211Count++
		}
	}

	if top985Count > 0 {
		return "就业前景优秀，985院校毕业生在就业市场具有显著优势"
	} else if top211Count > 0 {
		return "就业前景良好，211院校毕业生就业竞争力较强"
	} else {
		return "就业前景一般，建议关注专业实力和个人能力提升"
	}
}

// 生成备选方案
func (e *EnhancedRecommendationEngine) generateAlternativeOptions(recommendations []RecommendedUniversity, req EnhancedRecommendationRequest) []RecommendedUniversity {
	alternatives := []RecommendedUniversity{}

	// 生成地域扩展的备选方案
	if req.GeographicFlexibility < 0.5 {
		// 推荐一些外省的优质院校
		for _, uni := range extendedUniversities {
			if uni.Province != req.Province && uni.Level != "" {
				// 简化的备选推荐
				alternative := RecommendedUniversity{
					University: uni,
					Major: e.getPopularMajor(uni),
					AdmissionProb: 0.6,
					RecommendReason: "地域扩展选择，优质教育资源",
					RiskLevel: "中等风险",
				}
				alternatives = append(alternatives, alternative)
				if len(alternatives) >= 3 {
					break
				}
			}
		}
	}

	return alternatives
}

func main() {
	fmt.Println("Enhanced Recommendation Algorithm initialized")
	// 初始化增强推荐引擎
	engine := NewEnhancedRecommendationEngine()

	// 示例测试
	testReq := EnhancedRecommendationRequest{
		RecommendationRequest: RecommendationRequest{
			Score: 580,
			Rank: 15000,
			Province: "北京",
			Year: 2024,
			RiskLevel: "稳健型",
			PreferredArea: "不限",
			PreferredType: "不限",
			PreferredMajor: "计算机",
		},
		PersonalityType: "理性型",
		CareerGoals: []string{"软件工程师"},
		FamilyIncome: 80000,
		GeographicFlexibility: 0.7,
		MajorFlexibility: 0.6,
		RiskTolerance: 0.5,
	}

	results := engine.GenerateEnhancedRecommendations(testReq)
	fmt.Printf("Generated %d enhanced recommendation schemes\n", len(results))

	for _, result := range results {
		fmt.Printf("Scheme: %s, Universities: %d, Confidence: %.2f\n",
			result.SchemeType, result.TotalCount, result.ConfidenceScore)
	}
}