package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/oktetopython/gaokao/services/recommendation-service/internal/cache"
	"github.com/oktetopython/gaokao/services/recommendation-service/internal/llm"
	"github.com/oktetopython/gaokao/services/recommendation-service/pkg/cppbridge"
)

type stubAnalyzer struct {
	report string
	err    error
}

func (s stubAnalyzer) AnalyzeRecommendation(ctx context.Context, input llm.RecommendationAnalysisInput) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.report, nil
}

type stubBridge struct {
	status        map[string]interface{}
	mu            sync.Mutex
	calls         int
	failProvince  map[string]error
	responseCount int
}

func (b *stubBridge) Close() error { return nil }
func (b *stubBridge) GenerateRecommendations(request *cppbridge.RecommendationRequest) (*cppbridge.RecommendationResponse, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.calls++
	if err := b.failProvince[request.Province]; err != nil {
		return nil, err
	}
	b.responseCount++
	return &cppbridge.RecommendationResponse{
		StudentID:   request.StudentID,
		TotalCount:  1,
		GeneratedAt: 1710000000,
		Algorithm:   request.Algorithm,
		Success:     true,
		Recommendations: []cppbridge.Recommendation{
			{
				SchoolID:    "u1",
				SchoolName:  "测试大学",
				MajorID:     "m1",
				MajorName:   "计算机科学与技术",
				Probability: 0.82,
				RiskLevel:   "medium",
				Province:    request.Province,
				SchoolType:  "综合",
			},
		},
	}, nil
}
func (b *stubBridge) GetHybridConfig() (map[string]interface{}, error)     { return nil, nil }
func (b *stubBridge) UpdateFusionWeights(weights map[string]float64) error { return nil }
func (b *stubBridge) CompareRecommendations(request *cppbridge.RecommendationRequest) (map[string]interface{}, error) {
	return nil, nil
}
func (b *stubBridge) GetPerformanceMetrics() (map[string]interface{}, error) { return nil, nil }
func (b *stubBridge) GenerateHybridPlan(request *cppbridge.RecommendationRequest) (map[string]interface{}, error) {
	return nil, nil
}
func (b *stubBridge) ClearCache() error                  { return nil }
func (b *stubBridge) UpdateModel(modelPath string) error { return nil }
func (b *stubBridge) GetSystemStatus() (map[string]interface{}, error) {
	if b.status == nil {
		return map[string]interface{}{"status": "healthy", "engine": "test"}, nil
	}
	return b.status, nil
}

func TestGenerateAnalysisReportUsesAnalyzer(t *testing.T) {
	handler := NewSimpleRecommendationHandler(nil, cache.NewMemoryCache(), stubAnalyzer{report: "来自模型的分析"})
	student := &StudentInfo{
		Score:       intPtr(640),
		Province:    "广东",
		ScienceType: "物理",
		Preferences: StudentPreferences{RiskTolerance: "medium"},
	}
	request, err := handler.convertStudentInfoToRequest(student)
	if err != nil {
		t.Fatalf("convertStudentInfoToRequest failed: %v", err)
	}

	report := handler.generateAnalysisReport(context.Background(), student, request, sampleFrontendRecommendations())
	if report != "来自模型的分析" {
		t.Fatalf("expected analyzer report, got %q", report)
	}
}

func TestGenerateAnalysisReportFallsBackLocally(t *testing.T) {
	handler := NewSimpleRecommendationHandler(nil, cache.NewMemoryCache(), stubAnalyzer{err: errors.New("boom")})
	student := &StudentInfo{
		Score:       intPtr(640),
		Province:    "广东",
		ScienceType: "物理",
		Preferences: StudentPreferences{RiskTolerance: "medium"},
	}
	request, err := handler.convertStudentInfoToRequest(student)
	if err != nil {
		t.Fatalf("convertStudentInfoToRequest failed: %v", err)
	}

	report := handler.generateAnalysisReport(context.Background(), student, request, sampleFrontendRecommendations())
	expected := "根据您的分数和偏好，为您推荐了3所院校。其中稳妥选择1个，适中选择1个，冲刺选择1个。建议合理搭配，确保志愿填报的科学性和安全性。"
	if report != expected {
		t.Fatalf("expected fallback report %q, got %q", expected, report)
	}
}

func TestBuildRecommendationAnalysisInput(t *testing.T) {
	handler := NewSimpleRecommendationHandler(nil, cache.NewMemoryCache(), nil)
	rank := 12345
	student := &StudentInfo{
		Score:       intPtr(640),
		Province:    "广东",
		ScienceType: "物理",
		Rank:        &rank,
		Preferences: StudentPreferences{
			Regions:             []string{"广东", "上海"},
			MajorCategories:     []string{"计算机"},
			UniversityTypes:     []string{"综合"},
			RiskTolerance:       "low",
			SpecialRequirements: "希望一线城市",
		},
	}
	request, err := handler.convertStudentInfoToRequest(student)
	if err != nil {
		t.Fatalf("convertStudentInfoToRequest failed: %v", err)
	}
	request.Name = "Alice"

	input := handler.buildRecommendationAnalysisInput(student, request, sampleFrontendRecommendations())
	if input.StudentName != "Alice" || input.Score != 640 || input.Province != "广东" {
		t.Fatalf("unexpected input core fields: %+v", input)
	}
	if input.Rank == nil || *input.Rank != 12345 {
		t.Fatalf("unexpected rank: %+v", input.Rank)
	}
	if len(input.Recommendations) != 3 {
		t.Fatalf("expected 3 recommendations, got %d", len(input.Recommendations))
	}
	if input.Recommendations[0].SchoolName != "清华大学" || input.Recommendations[0].MajorName != "计算机科学与技术" {
		t.Fatalf("unexpected first recommendation: %+v", input.Recommendations[0])
	}
}

func TestGetSystemStatusIncludesAnalysisAndCache(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewSimpleRecommendationHandler(&stubBridge{status: map[string]interface{}{"status": "healthy", "engine": "test"}}, cache.NewMemoryCache(), llm.NewLocalFallbackAnalyzer())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/status", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.GetSystemStatus(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	analysis, ok := payload["analysis"].(map[string]interface{})
	if !ok {
		t.Fatalf("analysis status missing: %+v", payload)
	}
	if analysis["provider"] != "local-fallback" {
		t.Fatalf("unexpected analysis provider: %+v", analysis)
	}
	if analysis["status"] != "degraded" {
		t.Fatalf("unexpected analysis status: %+v", analysis)
	}
	if analysis["fallback_mode"] != "local_rules" {
		t.Fatalf("unexpected fallback mode: %+v", analysis)
	}
	cacheStatus, ok := payload["cache"].(map[string]interface{})
	if !ok || cacheStatus["healthy"] != true {
		t.Fatalf("unexpected cache status: %+v", payload["cache"])
	}
}

func sampleFrontendRecommendations() []FrontendRecommendation {
	return []FrontendRecommendation{
		{
			University:           FrontendUniversity{Name: "清华大学", Province: "北京", City: "北京", Type: "综合"},
			AdmissionProbability: 85,
			MatchScore:           92,
			RecommendReason:      "分数匹配度高，综合实力强",
			RiskLevel:            "low",
			SuggestedMajors:      []FrontendMajor{{Name: "计算机科学与技术"}},
		},
		{
			University:           FrontendUniversity{Name: "上海交通大学", Province: "上海", City: "上海", Type: "综合"},
			AdmissionProbability: 68,
			MatchScore:           80,
			RecommendReason:      "专业契合度较高",
			RiskLevel:            "medium",
			SuggestedMajors:      []FrontendMajor{{Name: "软件工程"}},
		},
		{
			University:           FrontendUniversity{Name: "浙江大学", Province: "浙江", City: "杭州", Type: "综合"},
			AdmissionProbability: 45,
			MatchScore:           73,
			RecommendReason:      "可作为冲刺选择",
			RiskLevel:            "high",
			SuggestedMajors:      []FrontendMajor{{Name: "人工智能"}},
		},
	}
}

func intPtr(v int) *int { return &v }

func TestGenerateCacheKeyIgnoresStudentIdentityAndMapOrder(t *testing.T) {
	handler := NewSimpleRecommendationHandler(nil, cache.NewMemoryCache(), nil)

	req1 := &cppbridge.RecommendationRequest{
		StudentID:          "student-a",
		Name:               "Alice",
		TotalScore:         620,
		Ranking:            12000,
		Province:           "河北",
		SubjectCombination: "物理",
		MaxRecommendations: 30,
		Algorithm:          "hybrid",
		Preferences: map[string]interface{}{
			"risk_tolerance":   "moderate",
			"major_categories": []string{"计算机", "电子信息"},
		},
	}
	req2 := &cppbridge.RecommendationRequest{
		StudentID:          "student-b",
		Name:               "Bob",
		TotalScore:         620,
		Ranking:            12000,
		Province:           "河北",
		SubjectCombination: "物理",
		MaxRecommendations: 30,
		Algorithm:          "hybrid",
		Preferences: map[string]interface{}{
			"major_categories": []string{"计算机", "电子信息"},
			"risk_tolerance":   "moderate",
		},
	}

	key1 := handler.generateCacheKey(req1)
	key2 := handler.generateCacheKey(req2)
	if key1 != key2 {
		t.Fatalf("expected same cache key, got %q vs %q", key1, key2)
	}
}

func TestWarmRecommendationCacheStoresHotRequests(t *testing.T) {
	bridge := &stubBridge{}
	handler := NewSimpleRecommendationHandler(bridge, cache.NewMemoryCache(), nil)
	requests := []*cppbridge.RecommendationRequest{
		{
			StudentID:          "warm-1",
			TotalScore:         610,
			Province:           "山东",
			SubjectCombination: "物理",
			MaxRecommendations: 30,
			Algorithm:          "hybrid",
			Preferences: map[string]interface{}{
				"risk_tolerance": "moderate",
			},
		},
		{
			StudentID:          "warm-2",
			TotalScore:         610,
			Province:           "山东",
			SubjectCombination: "物理",
			MaxRecommendations: 30,
			Algorithm:          "hybrid",
			Preferences: map[string]interface{}{
				"risk_tolerance": "moderate",
			},
		},
	}

	summary := handler.WarmRecommendationCache(context.Background(), requests)
	if summary.Attempted != 2 || summary.Warmed != 1 || summary.Skipped != 1 || summary.Failed != 0 {
		t.Fatalf("unexpected summary: %+v", summary)
	}

	bridge.mu.Lock()
	calls := bridge.calls
	bridge.mu.Unlock()
	if calls != 1 {
		t.Fatalf("expected 1 bridge call, got %d", calls)
	}
}

func TestWarmRecommendationCacheContinuesOnFailure(t *testing.T) {
	bridge := &stubBridge{
		failProvince: map[string]error{
			"河北": errors.New("upstream failed"),
		},
	}
	handler := NewSimpleRecommendationHandler(bridge, cache.NewMemoryCache(), nil)
	requests := []*cppbridge.RecommendationRequest{
		{
			StudentID:          "warm-fail",
			TotalScore:         580,
			Province:           "河北",
			SubjectCombination: "物理",
			MaxRecommendations: 30,
			Algorithm:          "hybrid",
		},
		{
			StudentID:          "warm-ok",
			TotalScore:         605,
			Province:           "河南",
			SubjectCombination: "物理",
			MaxRecommendations: 30,
			Algorithm:          "hybrid",
		},
	}

	summary := handler.WarmRecommendationCache(context.Background(), requests)
	if summary.Attempted != 2 || summary.Warmed != 1 || summary.Failed != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}
