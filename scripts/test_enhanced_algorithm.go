package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

// 简化的大学结构
type SimpleUniversity struct {
	ID           int
	Name         string
	Province     string
	Level        string
	NationalRank int
}

// 简化的专业结构
type SimpleMajor struct {
	ID       int
	Name     string
	Category string
}

// 简化的推荐请求
type SimpleRecommendationRequest struct {
	Score                 int
	Province              string
	RiskTolerance         float64
	GeographicFlexibility float64
	MajorFlexibility      float64
}

// 推荐结果
type SimpleRecommendationResult struct {
	University       SimpleUniversity
	Major           SimpleMajor
	AdmissionProb   float64
	ScoreDifference int
	RiskLevel       string
	RecommendReason string
}

// 简化的推荐引擎
type SimpleEnhancedEngine struct {
	universities []SimpleUniversity
	majors       []SimpleMajor
}

// 创建简化引擎
func NewSimpleEnhancedEngine() *SimpleEnhancedEngine {
	return &SimpleEnhancedEngine{
		universities: generateSampleUniversities(),
		majors:       generateSampleMajors(),
	}
}

// 生成示例大学数据
func generateSampleUniversities() []SimpleUniversity {
	return []SimpleUniversity{
		{ID: 1, Name: "清华大学", Province: "北京", Level: "985", NationalRank: 1},
		{ID: 2, Name: "北京大学", Province: "北京", Level: "985", NationalRank: 2},
		{ID: 3, Name: "浙江大学", Province: "浙江", Level: "985", NationalRank: 3},
		{ID: 4, Name: "上海交通大学", Province: "上海", Level: "985", NationalRank: 4},
		{ID: 5, Name: "华中科技大学", Province: "湖北", Level: "985", NationalRank: 10},
		{ID: 6, Name: "北京理工大学", Province: "北京", Level: "211", NationalRank: 25},
		{ID: 7, Name: "华南理工大学", Province: "广东", Level: "211", NationalRank: 30},
		{ID: 8, Name: "大连理工大学", Province: "辽宁", Level: "211", NationalRank: 35},
	}
}

// 生成示例专业数据
func generateSampleMajors() []SimpleMajor {
	return []SimpleMajor{
		{ID: 1, Name: "计算机科学与技术", Category: "工学"},
		{ID: 2, Name: "软件工程", Category: "工学"},
		{ID: 3, Name: "电子信息工程", Category: "工学"},
		{ID: 4, Name: "机械工程", Category: "工学"},
		{ID: 5, Name: "土木工程", Category: "工学"},
	}
}

// 生成增强推荐
func (e *SimpleEnhancedEngine) GenerateRecommendations(req SimpleRecommendationRequest) []SimpleRecommendationResult {
	results := []SimpleRecommendationResult{}
	rand.Seed(time.Now().UnixNano())

	for _, uni := range e.universities {
		// 计算基础分数差异
		estimatedMinScore := e.estimateMinScore(uni)
		scoreDiff := req.Score - estimatedMinScore

		// 地域过滤
		if req.GeographicFlexibility < 0.5 && uni.Province != req.Province {
			continue
		}

		// 分数范围过滤
		if scoreDiff < -50 || scoreDiff > 100 {
			continue
		}

		// 计算录取概率
		admissionProb := e.calculateAdmissionProbability(scoreDiff, uni)

		// 风险过滤
		if req.RiskTolerance < 0.3 && admissionProb < 0.7 {
			continue
		}
		if req.RiskTolerance > 0.7 && admissionProb > 0.9 {
			continue
		}

		// 选择专业
		major := e.selectBestMajor(uni, req)

		// 生成推荐结果
		result := SimpleRecommendationResult{
			University:       uni,
			Major:           major,
			AdmissionProb:   admissionProb,
			ScoreDifference: scoreDiff,
			RiskLevel:       e.determineRiskLevel(admissionProb),
			RecommendReason: e.generateReason(uni, major, admissionProb),
		}

		results = append(results, result)
	}

	// 排序结果
	sort.Slice(results, func(i, j int) bool {
		return results[i].AdmissionProb > results[j].AdmissionProb
	})

	// 限制结果数量
	if len(results) > 10 {
		results = results[:10]
	}

	return results
}

// 估算最低分数
func (e *SimpleEnhancedEngine) estimateMinScore(uni SimpleUniversity) int {
	baseScore := 500
	if uni.Level == "985" {
		baseScore = 650 - uni.NationalRank
	} else if uni.Level == "211" {
		baseScore = 600 - uni.NationalRank/2
	} else {
		baseScore = 550 - uni.NationalRank/5
	}
	return baseScore
}

// 计算录取概率
func (e *SimpleEnhancedEngine) calculateAdmissionProbability(scoreDiff int, uni SimpleUniversity) float64 {
	baseProb := 0.5

	// 分数差异影响
	if scoreDiff > 30 {
		baseProb += 0.3
	} else if scoreDiff > 10 {
		baseProb += 0.2
	} else if scoreDiff < -10 {
		baseProb -= 0.2
	} else if scoreDiff < -30 {
		baseProb -= 0.3
	}

	// 学校层次影响
	if uni.Level == "985" {
		baseProb -= 0.1
	} else if uni.Level == "211" {
		baseProb -= 0.05
	}

	// 确保概率在合理范围内
	if baseProb > 0.95 {
		baseProb = 0.95
	}
	if baseProb < 0.05 {
		baseProb = 0.05
	}

	return math.Round(baseProb*100) / 100
}

// 选择最佳专业
func (e *SimpleEnhancedEngine) selectBestMajor(uni SimpleUniversity, req SimpleRecommendationRequest) SimpleMajor {
	// 简化：随机选择一个专业
	if len(e.majors) > 0 {
		return e.majors[rand.Intn(len(e.majors))]
	}
	return SimpleMajor{ID: 1, Name: "计算机科学与技术", Category: "工学"}
}

// 确定风险等级
func (e *SimpleEnhancedEngine) determineRiskLevel(prob float64) string {
	if prob >= 0.8 {
		return "低风险"
	} else if prob >= 0.6 {
		return "中等风险"
	} else {
		return "高风险"
	}
}

// 生成推荐理由
func (e *SimpleEnhancedEngine) generateReason(uni SimpleUniversity, major SimpleMajor, prob float64) string {
	reason := fmt.Sprintf("%s是%s层次院校", uni.Name, uni.Level)
	if prob >= 0.8 {
		reason += "，录取把握较大"
	} else if prob >= 0.6 {
		reason += "，录取概率适中"
	} else {
		reason += "，需要谨慎考虑"
	}
	reason += fmt.Sprintf("，%s专业就业前景良好", major.Name)
	return reason
}

// 测试函数
func main() {
	fmt.Println("=== 增强推荐算法测试 ===")

	// 创建引擎
	engine := NewSimpleEnhancedEngine()

	// 测试请求
	testReq := SimpleRecommendationRequest{
		Score:                 620,
		Province:              "北京",
		RiskTolerance:         0.6,
		GeographicFlexibility: 0.7,
		MajorFlexibility:      0.8,
	}

	fmt.Printf("测试参数：分数=%d, 省份=%s, 风险承受度=%.1f\n",
		testReq.Score, testReq.Province, testReq.RiskTolerance)
	fmt.Println("\n推荐结果：")

	// 生成推荐
	results := engine.GenerateRecommendations(testReq)

	if len(results) == 0 {
		fmt.Println("未找到合适的推荐")
		return
	}

	// 显示结果
	for i, result := range results {
		fmt.Printf("%d. %s - %s\n", i+1, result.University.Name, result.Major.Name)
		fmt.Printf("   录取概率: %.0f%%, 分数差异: %+d, 风险等级: %s\n",
			result.AdmissionProb*100, result.ScoreDifference, result.RiskLevel)
		fmt.Printf("   推荐理由: %s\n\n", result.RecommendReason)
	}

	fmt.Printf("总共生成 %d 个推荐方案\n", len(results))
	fmt.Println("\n=== 测试完成 ===")
}