package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/oktetopython/gaokao/services/recommendation-service/internal/cache"
	"github.com/oktetopython/gaokao/services/recommendation-service/internal/llm"
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
