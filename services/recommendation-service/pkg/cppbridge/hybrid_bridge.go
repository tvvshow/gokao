// +build cgo

package cppbridge

/*
#cgo CFLAGS: -I../../../cpp-modules/volunteer-matcher/include
#cgo LDFLAGS: -L../../../cpp-modules/volunteer-matcher/build -lvolunteer_matcher -lstdc++

#include <stdlib.h>
#include "c_interface.h"

typedef struct {
    char* student_id;
    char* name;
    int total_score;
    int ranking;
    char* province;
    char* subject_combination;
    
    // 单科成绩
    int chinese_score;
    int math_score;
    int english_score;
    int physics_score;
    int chemistry_score;
    int biology_score;
    int politics_score;
    int history_score;
    int geography_score;
    
    // 偏好设置
    char** preferred_cities;
    int preferred_cities_count;
    char** preferred_majors;
    int preferred_majors_count;
    char** avoided_majors;
    int avoided_majors_count;
    
    double city_weight;
    double major_weight;
    double school_ranking_weight;
    
    // 特殊情况
    int is_minority;
    int has_sports_specialty;
    int has_art_specialty;
} C_Student;

typedef struct {
    char* university_id;
    char* university_name;
    char* major_id;
    char* major_name;
    
    double admission_probability;
    char* risk_level;
    int score_gap;
    int ranking_gap;
    
    double match_score;
    char* recommendation_reason;
    char** risk_factors;
    int risk_factors_count;
} C_VolunteerRecommendation;

typedef struct {
    char* student_id;
    C_VolunteerRecommendation* recommendations;
    int recommendations_count;
    
    int total_volunteers;
    int rush_count;
    int stable_count;
    int safe_count;
    
    double overall_risk_score;
    char* plan_quality;
    char** optimization_suggestions;
    int optimization_suggestions_count;
    
    long long generated_time;
} C_VolunteerPlan;

// C++混合推荐引擎接口声明
extern void* CreateHybridRecommendationEngine();
extern void DestroyHybridRecommendationEngine(void* engine);
extern int InitializeHybridEngine(void* engine, void* traditional_matcher, void* ai_engine, const char* config_path);
extern C_VolunteerPlan* GenerateHybridPlan(void* engine, C_Student* student, int max_volunteers);
extern int SetFusionWeights(void* engine, double traditional_weight, double ai_weight);
extern char* GetHybridExplanation(void* engine, C_VolunteerRecommendation* recommendation);
extern char* GetHybridStats(void* engine);
extern void FreeCVolunteerPlan(C_VolunteerPlan* plan);
extern void FreeCString(char* str);
*/
import "C"
import (
	"encoding/json"
	"errors"
	"runtime"
	"time"
	"unsafe"
)

// Student 学生信息结构
type Student struct {
	StudentID          string   `json:"student_id"`
	Name               string   `json:"name"`
	TotalScore         int      `json:"total_score"`
	Ranking            int      `json:"ranking"`
	Province           string   `json:"province"`
	SubjectCombination string   `json:"subject_combination"`
	
	// 单科成绩
	ChineseScore   int `json:"chinese_score"`
	MathScore      int `json:"math_score"`
	EnglishScore   int `json:"english_score"`
	PhysicsScore   int `json:"physics_score"`
	ChemistryScore int `json:"chemistry_score"`
	BiologyScore   int `json:"biology_score"`
	PoliticsScore  int `json:"politics_score"`
	HistoryScore   int `json:"history_score"`
	GeographyScore int `json:"geography_score"`
	
	// 偏好设置
	PreferredCities   []string `json:"preferred_cities"`
	PreferredMajors   []string `json:"preferred_majors"`
	AvoidedMajors     []string `json:"avoided_majors"`
	CityWeight        float64  `json:"city_weight"`
	MajorWeight       float64  `json:"major_weight"`
	SchoolRankingWeight float64 `json:"school_ranking_weight"`
	
	// 特殊情况
	IsMinority        bool `json:"is_minority"`
	HasSportsSpecialty bool `json:"has_sports_specialty"`
	HasArtSpecialty   bool `json:"has_art_specialty"`
}

// VolunteerRecommendation 志愿推荐结构
type VolunteerRecommendation struct {
	UniversityID        string   `json:"university_id"`
	UniversityName      string   `json:"university_name"`
	MajorID             string   `json:"major_id"`
	MajorName           string   `json:"major_name"`
	AdmissionProbability float64  `json:"admission_probability"`
	RiskLevel           string   `json:"risk_level"`
	ScoreGap            int      `json:"score_gap"`
	RankingGap          int      `json:"ranking_gap"`
	MatchScore          float64  `json:"match_score"`
	RecommendationReason string   `json:"recommendation_reason"`
	RiskFactors         []string `json:"risk_factors"`
}

// VolunteerPlan 志愿填报方案
type VolunteerPlan struct {
	StudentID               string                    `json:"student_id"`
	Recommendations         []VolunteerRecommendation `json:"recommendations"`
	TotalVolunteers         int                       `json:"total_volunteers"`
	RushCount               int                       `json:"rush_count"`
	StableCount             int                       `json:"stable_count"`
	SafeCount               int                       `json:"safe_count"`
	OverallRiskScore        float64                   `json:"overall_risk_score"`
	PlanQuality             string                    `json:"plan_quality"`
	OptimizationSuggestions []string                  `json:"optimization_suggestions"`
	GeneratedTime           time.Time                 `json:"generated_time"`
}

// FusionWeights 融合权重
type FusionWeights struct {
	TraditionalWeight float64 `json:"traditional_weight"`
	AIWeight          float64 `json:"ai_weight"`
}

// HybridStats 混合引擎统计
type HybridStats struct {
	TotalRequests         int     `json:"total_requests"`
	SuccessfulFusions     int     `json:"successful_fusions"`
	AvgFusionTimeMs       float64 `json:"avg_fusion_time_ms"`
	TraditionalWeightAvg  float64 `json:"traditional_weight_avg"`
	AIWeightAvg           float64 `json:"ai_weight_avg"`
	AdaptiveAdjustments   int     `json:"adaptive_adjustments"`
	CacheSize             int     `json:"cache_size"`
	CacheMaxSize          int     `json:"cache_max_size"`
}

// CppHybridRecommendationBridge C++混合推荐引擎的Go桥接器
type CppHybridRecommendationBridge struct {
	engine            unsafe.Pointer
	traditionalMatcher unsafe.Pointer
	aiEngine          unsafe.Pointer
}

// NewHybridRecommendationBridge 创建新的混合推荐桥接器
func NewHybridRecommendationBridge(configPath string) (HybridRecommendationBridge, error) {
	bridge := &CppHybridRecommendationBridge{}
	
	// 创建C++对象
	bridge.engine = C.CreateHybridRecommendationEngine()
	if bridge.engine == nil {
		return nil, errors.New("failed to create hybrid recommendation engine")
	}
	
	// TODO: 这里需要创建传统匹配器和AI引擎
	// bridge.traditionalMatcher = C.CreateVolunteerMatcher()
	// bridge.aiEngine = C.CreateAIRecommendationEngine()
	
	// 初始化引擎
	configPathC := C.CString(configPath)
	defer C.free(unsafe.Pointer(configPathC))
	
	result := C.InitializeHybridEngine(bridge.engine, bridge.traditionalMatcher, bridge.aiEngine, configPathC)
	if result == 0 {
		bridge.Close()
		return nil, errors.New("failed to initialize hybrid engine")
	}
	
	// 设置finalizer确保资源释放
	runtime.SetFinalizer(bridge, (*HybridRecommendationBridge).Close)
	
	return bridge, nil
}

// Close 关闭桥接器并释放资源
func (b *HybridRecommendationBridge) Close() {
	if b.engine != nil {
		C.DestroyHybridRecommendationEngine(b.engine)
		b.engine = nil
	}
	
	// TODO: 释放其他C++对象
	// if b.traditionalMatcher != nil {
	//     C.DestroyVolunteerMatcher(b.traditionalMatcher)
	//     b.traditionalMatcher = nil
	// }
	// if b.aiEngine != nil {
	//     C.DestroyAIRecommendationEngine(b.aiEngine)
	//     b.aiEngine = nil
	// }
	
	runtime.SetFinalizer(b, nil)
}

// GenerateHybridPlan 生成混合推荐方案
func (b *HybridRecommendationBridge) GenerateHybridPlan(student *Student, maxVolunteers int) (*VolunteerPlan, error) {
	if b.engine == nil {
		return nil, errors.New("hybrid engine not initialized")
	}
	
	// 转换Go结构到C结构
	cStudent := b.studentToC(student)
	defer b.freeCStudent(cStudent)
	
	// 调用C++函数
	cPlan := C.GenerateHybridPlan(b.engine, cStudent, C.int(maxVolunteers))
	if cPlan == nil {
		return nil, errors.New("failed to generate hybrid plan")
	}
	defer C.FreeCVolunteerPlan(cPlan)
	
	// 转换C结构到Go结构
	plan := b.cPlanToGo(cPlan)
	return plan, nil
}

// SetFusionWeights 设置融合权重
func (b *HybridRecommendationBridge) SetFusionWeights(traditionalWeight, aiWeight float64) error {
	if b.engine == nil {
		return errors.New("hybrid engine not initialized")
	}
	
	result := C.SetFusionWeights(b.engine, C.double(traditionalWeight), C.double(aiWeight))
	if result == 0 {
		return errors.New("failed to set fusion weights")
	}
	
	return nil
}

// GetHybridExplanation 获取混合推荐解释
func (b *HybridRecommendationBridge) GetHybridExplanation(recommendation *VolunteerRecommendation) (string, error) {
	if b.engine == nil {
		return "", errors.New("hybrid engine not initialized")
	}
	
	cRec := b.recommendationToC(recommendation)
	defer b.freeCRecommendation(cRec)
	
	cExplanation := C.GetHybridExplanation(b.engine, cRec)
	if cExplanation == nil {
		return "", errors.New("failed to get explanation")
	}
	defer C.FreeCString(cExplanation)
	
	return C.GoString(cExplanation), nil
}

// GetHybridStats 获取混合引擎统计信息
func (b *HybridRecommendationBridge) GetHybridStats() (*HybridStats, error) {
	if b.engine == nil {
		return nil, errors.New("hybrid engine not initialized")
	}
	
	cStats := C.GetHybridStats(b.engine)
	if cStats == nil {
		return nil, errors.New("failed to get stats")
	}
	defer C.FreeCString(cStats)
	
	statsJSON := C.GoString(cStats)
	var stats HybridStats
	err := json.Unmarshal([]byte(statsJSON), &stats)
	if err != nil {
		return nil, err
	}
	
	return &stats, nil
}

// 辅助函数：Go Student转C Student
func (b *HybridRecommendationBridge) studentToC(student *Student) *C.C_Student {
	cStudent := (*C.C_Student)(C.malloc(C.size_t(unsafe.Sizeof(C.C_Student{}))))
	
	cStudent.student_id = C.CString(student.StudentID)
	cStudent.name = C.CString(student.Name)
	cStudent.total_score = C.int(student.TotalScore)
	cStudent.ranking = C.int(student.Ranking)
	cStudent.province = C.CString(student.Province)
	cStudent.subject_combination = C.CString(student.SubjectCombination)
	
	// 单科成绩
	cStudent.chinese_score = C.int(student.ChineseScore)
	cStudent.math_score = C.int(student.MathScore)
	cStudent.english_score = C.int(student.EnglishScore)
	cStudent.physics_score = C.int(student.PhysicsScore)
	cStudent.chemistry_score = C.int(student.ChemistryScore)
	cStudent.biology_score = C.int(student.BiologyScore)
	cStudent.politics_score = C.int(student.PoliticsScore)
	cStudent.history_score = C.int(student.HistoryScore)
	cStudent.geography_score = C.int(student.GeographyScore)
	
	// 偏好城市
	if len(student.PreferredCities) > 0 {
		cStudent.preferred_cities = (**C.char)(C.malloc(C.size_t(len(student.PreferredCities)) * C.size_t(unsafe.Sizeof(uintptr(0)))))
		for i, city := range student.PreferredCities {
			*(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cStudent.preferred_cities)) + uintptr(i)*unsafe.Sizeof(uintptr(0)))) = C.CString(city)
		}
		cStudent.preferred_cities_count = C.int(len(student.PreferredCities))
	}
	
	// 偏好专业
	if len(student.PreferredMajors) > 0 {
		cStudent.preferred_majors = (**C.char)(C.malloc(C.size_t(len(student.PreferredMajors)) * C.size_t(unsafe.Sizeof(uintptr(0)))))
		for i, major := range student.PreferredMajors {
			*(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cStudent.preferred_majors)) + uintptr(i)*unsafe.Sizeof(uintptr(0)))) = C.CString(major)
		}
		cStudent.preferred_majors_count = C.int(len(student.PreferredMajors))
	}
	
	// 避免专业
	if len(student.AvoidedMajors) > 0 {
		cStudent.avoided_majors = (**C.char)(C.malloc(C.size_t(len(student.AvoidedMajors)) * C.size_t(unsafe.Sizeof(uintptr(0)))))
		for i, major := range student.AvoidedMajors {
			*(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cStudent.avoided_majors)) + uintptr(i)*unsafe.Sizeof(uintptr(0)))) = C.CString(major)
		}
		cStudent.avoided_majors_count = C.int(len(student.AvoidedMajors))
	}
	
	// 权重
	cStudent.city_weight = C.double(student.CityWeight)
	cStudent.major_weight = C.double(student.MajorWeight)
	cStudent.school_ranking_weight = C.double(student.SchoolRankingWeight)
	
	// 特殊情况
	if student.IsMinority {
		cStudent.is_minority = 1
	}
	if student.HasSportsSpecialty {
		cStudent.has_sports_specialty = 1
	}
	if student.HasArtSpecialty {
		cStudent.has_art_specialty = 1
	}
	
	return cStudent
}

// 辅助函数：释放C Student
func (b *HybridRecommendationBridge) freeCStudent(cStudent *C.C_Student) {
	if cStudent == nil {
		return
	}
	
	C.free(unsafe.Pointer(cStudent.student_id))
	C.free(unsafe.Pointer(cStudent.name))
	C.free(unsafe.Pointer(cStudent.province))
	C.free(unsafe.Pointer(cStudent.subject_combination))
	
	// 释放数组
	if cStudent.preferred_cities != nil {
		for i := 0; i < int(cStudent.preferred_cities_count); i++ {
			ptr := *(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cStudent.preferred_cities)) + uintptr(i)*unsafe.Sizeof(uintptr(0))))
			C.free(unsafe.Pointer(ptr))
		}
		C.free(unsafe.Pointer(cStudent.preferred_cities))
	}
	
	if cStudent.preferred_majors != nil {
		for i := 0; i < int(cStudent.preferred_majors_count); i++ {
			ptr := *(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cStudent.preferred_majors)) + uintptr(i)*unsafe.Sizeof(uintptr(0))))
			C.free(unsafe.Pointer(ptr))
		}
		C.free(unsafe.Pointer(cStudent.preferred_majors))
	}
	
	if cStudent.avoided_majors != nil {
		for i := 0; i < int(cStudent.avoided_majors_count); i++ {
			ptr := *(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cStudent.avoided_majors)) + uintptr(i)*unsafe.Sizeof(uintptr(0))))
			C.free(unsafe.Pointer(ptr))
		}
		C.free(unsafe.Pointer(cStudent.avoided_majors))
	}
	
	C.free(unsafe.Pointer(cStudent))
}

// 辅助函数：Go Recommendation转C Recommendation
func (b *HybridRecommendationBridge) recommendationToC(rec *VolunteerRecommendation) *C.C_VolunteerRecommendation {
	cRec := (*C.C_VolunteerRecommendation)(C.malloc(C.size_t(unsafe.Sizeof(C.C_VolunteerRecommendation{}))))
	
	cRec.university_id = C.CString(rec.UniversityID)
	cRec.university_name = C.CString(rec.UniversityName)
	cRec.major_id = C.CString(rec.MajorID)
	cRec.major_name = C.CString(rec.MajorName)
	cRec.admission_probability = C.double(rec.AdmissionProbability)
	cRec.risk_level = C.CString(rec.RiskLevel)
	cRec.score_gap = C.int(rec.ScoreGap)
	cRec.ranking_gap = C.int(rec.RankingGap)
	cRec.match_score = C.double(rec.MatchScore)
	cRec.recommendation_reason = C.CString(rec.RecommendationReason)
	
	// 风险因素
	if len(rec.RiskFactors) > 0 {
		cRec.risk_factors = (**C.char)(C.malloc(C.size_t(len(rec.RiskFactors)) * C.size_t(unsafe.Sizeof(uintptr(0)))))
		for i, factor := range rec.RiskFactors {
			*(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cRec.risk_factors)) + uintptr(i)*unsafe.Sizeof(uintptr(0)))) = C.CString(factor)
		}
		cRec.risk_factors_count = C.int(len(rec.RiskFactors))
	}
	
	return cRec
}

// 辅助函数：释放C Recommendation
func (b *HybridRecommendationBridge) freeCRecommendation(cRec *C.C_VolunteerRecommendation) {
	if cRec == nil {
		return
	}
	
	C.free(unsafe.Pointer(cRec.university_id))
	C.free(unsafe.Pointer(cRec.university_name))
	C.free(unsafe.Pointer(cRec.major_id))
	C.free(unsafe.Pointer(cRec.major_name))
	C.free(unsafe.Pointer(cRec.risk_level))
	C.free(unsafe.Pointer(cRec.recommendation_reason))
	
	if cRec.risk_factors != nil {
		for i := 0; i < int(cRec.risk_factors_count); i++ {
			ptr := *(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cRec.risk_factors)) + uintptr(i)*unsafe.Sizeof(uintptr(0))))
			C.free(unsafe.Pointer(ptr))
		}
		C.free(unsafe.Pointer(cRec.risk_factors))
	}
	
	C.free(unsafe.Pointer(cRec))
}

// 辅助函数：C Plan转Go Plan
func (b *HybridRecommendationBridge) cPlanToGo(cPlan *C.C_VolunteerPlan) *VolunteerPlan {
	plan := &VolunteerPlan{
		StudentID:        C.GoString(cPlan.student_id),
		TotalVolunteers:  int(cPlan.total_volunteers),
		RushCount:        int(cPlan.rush_count),
		StableCount:      int(cPlan.stable_count),
		SafeCount:        int(cPlan.safe_count),
		OverallRiskScore: float64(cPlan.overall_risk_score),
		PlanQuality:      C.GoString(cPlan.plan_quality),
		GeneratedTime:    time.Unix(int64(cPlan.generated_time), 0),
	}
	
	// 转换推荐列表
	recommendationCount := int(cPlan.recommendations_count)
	plan.Recommendations = make([]VolunteerRecommendation, recommendationCount)
	
	for i := 0; i < recommendationCount; i++ {
		cRec := (*C.C_VolunteerRecommendation)(unsafe.Pointer(uintptr(unsafe.Pointer(cPlan.recommendations)) + 
			uintptr(i)*unsafe.Sizeof(C.C_VolunteerRecommendation{})))
		
		rec := VolunteerRecommendation{
			UniversityID:         C.GoString(cRec.university_id),
			UniversityName:       C.GoString(cRec.university_name),
			MajorID:              C.GoString(cRec.major_id),
			MajorName:            C.GoString(cRec.major_name),
			AdmissionProbability: float64(cRec.admission_probability),
			RiskLevel:            C.GoString(cRec.risk_level),
			ScoreGap:             int(cRec.score_gap),
			RankingGap:           int(cRec.ranking_gap),
			MatchScore:           float64(cRec.match_score),
			RecommendationReason: C.GoString(cRec.recommendation_reason),
		}
		
		// 转换风险因素
		riskFactorCount := int(cRec.risk_factors_count)
		rec.RiskFactors = make([]string, riskFactorCount)
		for j := 0; j < riskFactorCount; j++ {
			factor := *(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cRec.risk_factors)) + 
				uintptr(j)*unsafe.Sizeof(uintptr(0))))
			rec.RiskFactors[j] = C.GoString(factor)
		}
		
		plan.Recommendations[i] = rec
	}
	
	// 转换优化建议
	suggestionCount := int(cPlan.optimization_suggestions_count)
	plan.OptimizationSuggestions = make([]string, suggestionCount)
	for i := 0; i < suggestionCount; i++ {
		suggestion := *(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cPlan.optimization_suggestions)) + 
			uintptr(i)*unsafe.Sizeof(uintptr(0))))
		plan.OptimizationSuggestions[i] = C.GoString(suggestion)
	}
	
	return plan
}