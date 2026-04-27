//go:build cgo && cppengine
// +build cgo,cppengine

package cppbridge

import "testing"

func TestCppVolunteerMatcherBridgeGenerateRecommendations(t *testing.T) {
	bridge, err := NewHybridRecommendationBridge(BridgeConfig{
		ConfigPath: "../../config/hybrid_config.json",
	})
	if err != nil {
		t.Fatalf("NewHybridRecommendationBridge returned error: %v", err)
	}
	defer func() {
		_ = bridge.Close()
	}()

	response, err := bridge.GenerateRecommendations(&RecommendationRequest{
		StudentID:          "student-1",
		Name:               "tester",
		TotalScore:         620,
		Ranking:            12000,
		Province:           "北京",
		SubjectCombination: "physics-chemistry-biology",
		ChineseScore:       120,
		MathScore:          130,
		EnglishScore:       125,
		MaxRecommendations: 12,
		Preferences: map[string]interface{}{
			"regions":          []interface{}{"北京", "上海"},
			"major_categories": []interface{}{"工科", "理科"},
		},
	})
	if err != nil {
		t.Fatalf("GenerateRecommendations returned error: %v", err)
	}

	if !response.Success {
		t.Fatalf("expected success response, got %#v", response)
	}
	if response.Algorithm != "cpp_engine" {
		t.Fatalf("expected algorithm cpp_engine, got %s", response.Algorithm)
	}
	if response.TotalCount == 0 {
		t.Fatalf("expected non-empty recommendations")
	}

	status, err := bridge.GetSystemStatus()
	if err != nil {
		t.Fatalf("GetSystemStatus returned error: %v", err)
	}
	if status["engine"] != "cpp_engine" {
		t.Fatalf("expected engine cpp_engine, got %#v", status["engine"])
	}
}
