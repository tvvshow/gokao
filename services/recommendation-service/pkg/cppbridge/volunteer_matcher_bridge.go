//go:build cgo && linux && amd64
// +build cgo,linux,amd64

package cppbridge

/*
#cgo CFLAGS: -I../../../../cpp-modules/volunteer-matcher/include
#cgo LDFLAGS: -L../../../../cpp-modules/volunteer-matcher/build -l:libvolunteer_matcher.a -ljsoncpp -lstdc++

#include <stdlib.h>
#include "c_interface.h"
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
	"unsafe"
)

type cppVolunteerPlan struct {
	StudentID       string `json:"student_id"`
	Recommendations []struct {
		UniversityID         string   `json:"university_id"`
		UniversityName       string   `json:"university_name"`
		MajorID              string   `json:"major_id"`
		MajorName            string   `json:"major_name"`
		AdmissionProbability float64  `json:"admission_probability"`
		RiskLevel            string   `json:"risk_level"`
		ScoreGap             int      `json:"score_gap"`
		RankingGap           int      `json:"ranking_gap"`
		MatchScore           float64  `json:"match_score"`
		RecommendationReason string   `json:"recommendation_reason"`
		RiskFactors          []string `json:"risk_factors"`
	} `json:"recommendations"`
	TotalVolunteers         int      `json:"total_volunteers"`
	RushCount               int      `json:"rush_count"`
	StableCount             int      `json:"stable_count"`
	SafeCount               int      `json:"safe_count"`
	OverallRiskScore        float64  `json:"overall_risk_score"`
	PlanQuality             string   `json:"plan_quality"`
	OptimizationSuggestions []string `json:"optimization_suggestions"`
	GeneratedTime           int64    `json:"generated_time"`
}

// CppVolunteerMatcherBridge 使用 C++ 志愿匹配引擎提供推荐能力。
//
// 并发模型：C++ 引擎自身通过 std::shared_mutex 保证线程安全（见
// cpp-modules/volunteer-matcher/include/volunteer_matcher.h:300、
// src/volunteer_matcher.cpp:52）。此处 RWMutex 仅用于隔离 handle 生命周期：
//   - 读路径（推理、查询）持 RLock，允许 N 个 goroutine 并发进入 C++；
//   - 写路径（Close / ClearCache 重建 handle）持 Lock，独占重建。
type CppVolunteerMatcherBridge struct {
	mu        sync.RWMutex
	handle    *C.VolunteerMatcherHandle
	cfg       BridgeConfig
	dataFiles bridgeDataFiles
	startedAt time.Time
}

// NewHybridRecommendationBridge 创建基于 C++ volunteer matcher 的桥接器。
func NewHybridRecommendationBridge(cfg BridgeConfig) (HybridRecommendationBridge, error) {
	dataFiles, err := resolveBridgeDataFiles(cfg)
	if err != nil {
		return nil, err
	}

	bridge := &CppVolunteerMatcherBridge{
		cfg:       cfg,
		dataFiles: dataFiles,
		startedAt: time.Now(),
	}

	if err := bridge.initialize(); err != nil {
		bridge.Close()
		return nil, err
	}

	return bridge, nil
}

func (b *CppVolunteerMatcherBridge) initialize() error {
	b.handle = C.CreateVolunteerMatcher()
	if b.handle == nil {
		return fmt.Errorf("create volunteer matcher failed")
	}

	if err := callCResult(func() *C.CResult {
		configPath := C.CString(b.cfg.ConfigPath)
		defer C.free(unsafe.Pointer(configPath))
		return C.InitializeVolunteerMatcher(b.handle, configPath)
	}); err != nil {
		return fmt.Errorf("initialize volunteer matcher: %w", err)
	}

	if err := b.loadDataFiles(); err != nil {
		return err
	}

	return nil
}

func (b *CppVolunteerMatcherBridge) loadDataFiles() error {
	if err := callCResult(func() *C.CResult {
		path := C.CString(b.dataFiles.universities)
		defer C.free(unsafe.Pointer(path))
		return C.LoadUniversities(b.handle, path)
	}); err != nil {
		return fmt.Errorf("load universities: %w", err)
	}

	if err := callCResult(func() *C.CResult {
		path := C.CString(b.dataFiles.majors)
		defer C.free(unsafe.Pointer(path))
		return C.LoadMajors(b.handle, path)
	}); err != nil {
		return fmt.Errorf("load majors: %w", err)
	}

	if err := callCResult(func() *C.CResult {
		path := C.CString(b.dataFiles.historical)
		defer C.free(unsafe.Pointer(path))
		return C.LoadHistoricalData(b.handle, path)
	}); err != nil {
		return fmt.Errorf("load historical data: %w", err)
	}

	return nil
}

func (b *CppVolunteerMatcherBridge) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.handle != nil {
		C.DestroyVolunteerMatcher(b.handle)
		b.handle = nil
	}
	return nil
}

func (b *CppVolunteerMatcherBridge) GenerateRecommendations(request *RecommendationRequest) (*RecommendationResponse, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.handle == nil {
		return nil, fmt.Errorf("cpp bridge is not initialized")
	}

	payload, err := json.Marshal(buildCPPStudentPayload(request))
	if err != nil {
		return nil, fmt.Errorf("marshal cpp request: %w", err)
	}

	studentJSON := C.CString(string(payload))
	defer C.free(unsafe.Pointer(studentJSON))

	result := C.GenerateVolunteerPlan(b.handle, studentJSON, C.int(normalizeMaxRecommendations(request.MaxRecommendations)))
	defer C.FreeCResult(result)
	if result == nil {
		return nil, fmt.Errorf("generate volunteer plan returned nil result")
	}
	if result.error_code != 0 {
		return nil, fmt.Errorf("cpp volunteer matcher failed: %s", C.GoString(result.message))
	}

	var plan cppVolunteerPlan
	if err := json.Unmarshal([]byte(C.GoString(result.data)), &plan); err != nil {
		return nil, fmt.Errorf("decode cpp response: %w", err)
	}

	response := &RecommendationResponse{
		StudentID:       request.StudentID,
		Algorithm:       "cpp_engine",
		Success:         true,
		GeneratedAt:     plan.GeneratedTime,
		TotalCount:      len(plan.Recommendations),
		Recommendations: make([]Recommendation, 0, len(plan.Recommendations)),
	}

	for i, rec := range plan.Recommendations {
		reasons := rec.RiskFactors
		if rec.RecommendationReason != "" {
			reasons = append([]string{rec.RecommendationReason}, reasons...)
		}

		response.Recommendations = append(response.Recommendations, Recommendation{
			SchoolID:       rec.UniversityID,
			SchoolName:     rec.UniversityName,
			MajorID:        rec.MajorID,
			MajorName:      rec.MajorName,
			AdmissionScore: request.TotalScore - rec.ScoreGap,
			Probability:    rec.AdmissionProbability,
			RiskLevel:      rec.RiskLevel,
			Ranking:        i + 1,
			Algorithm:      "cpp_engine",
			Reasons:        reasons,
			Score:          rec.MatchScore,
		})
	}

	return response, nil
}

func (b *CppVolunteerMatcherBridge) GetHybridConfig() (map[string]interface{}, error) {
	return map[string]interface{}{
		"engine_type":        "cpp_engine",
		"config_path":        b.cfg.ConfigPath,
		"universities_path":  b.dataFiles.universities,
		"majors_path":        b.dataFiles.majors,
		"historical_path":    b.dataFiles.historical,
		"cpp_bridge_enabled": true,
	}, nil
}

func (b *CppVolunteerMatcherBridge) UpdateFusionWeights(weights map[string]float64) error {
	return fmt.Errorf("cpp volunteer matcher does not expose runtime fusion weight updates")
}

func (b *CppVolunteerMatcherBridge) CompareRecommendations(request *RecommendationRequest) (map[string]interface{}, error) {
	response, err := b.GenerateRecommendations(request)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status":          "supported",
		"engine":          "cpp_engine",
		"recommendations": response.Recommendations,
		"total_count":     response.TotalCount,
	}, nil
}

func (b *CppVolunteerMatcherBridge) GetPerformanceMetrics() (map[string]interface{}, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.handle == nil {
		return nil, fmt.Errorf("cpp bridge is not initialized")
	}

	var stats C.CPerformanceStats
	if err := callCResult(func() *C.CResult {
		return C.GetPerformanceStats(b.handle, &stats)
	}); err != nil {
		return nil, fmt.Errorf("get cpp performance stats: %w", err)
	}

	return map[string]interface{}{
		"engine":              "cpp_engine",
		"total_requests":      uint64(stats.total_requests),
		"successful_requests": uint64(stats.successful_requests),
		"avg_response_time":   float64(stats.avg_response_time),
		"max_response_time":   float64(stats.max_response_time),
		"memory_usage":        uint64(stats.memory_usage),
		"uptime_seconds":      int64(time.Since(b.startedAt).Seconds()),
		"cpp_bridge_enabled":  true,
	}, nil
}

func (b *CppVolunteerMatcherBridge) GenerateHybridPlan(request *RecommendationRequest) (map[string]interface{}, error) {
	response, err := b.GenerateRecommendations(request)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status":          "supported",
		"engine":          "cpp_engine",
		"student_id":      response.StudentID,
		"recommendations": response.Recommendations,
		"total_count":     response.TotalCount,
	}, nil
}

func (b *CppVolunteerMatcherBridge) ClearCache() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.handle != nil {
		C.DestroyVolunteerMatcher(b.handle)
		b.handle = nil
	}

	b.startedAt = time.Now()
	return b.initialize()
}

func (b *CppVolunteerMatcherBridge) UpdateModel(modelPath string) error {
	return fmt.Errorf("cpp volunteer matcher does not support runtime model updates")
}

func (b *CppVolunteerMatcherBridge) GetSystemStatus() (map[string]interface{}, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.handle == nil {
		return nil, fmt.Errorf("cpp bridge is not initialized")
	}

	status := map[string]interface{}{
		"status":             "healthy",
		"engine":             "cpp_engine",
		"cpp_bridge_enabled": true,
		"uptime":             time.Since(b.startedAt).String(),
		"config_path":        b.cfg.ConfigPath,
	}

	rawStatus, err := cResultJSON(func() *C.CResult {
		return C.GetEngineStatus(b.handle)
	})
	if err == nil {
		for k, v := range rawStatus {
			status[k] = v
		}
	} else {
		status["engine_status_error"] = err.Error()
	}

	status["data_files"] = map[string]string{
		"universities": b.dataFiles.universities,
		"majors":       b.dataFiles.majors,
		"historical":   b.dataFiles.historical,
	}

	return status, nil
}

func buildCPPStudentPayload(request *RecommendationRequest) map[string]interface{} {
	preferredCities := extractStringSlice(request.Preferences, "regions")
	preferredMajors := extractStringSlice(request.Preferences, "major_categories")

	return map[string]interface{}{
		"student_id":            request.StudentID,
		"name":                  request.Name,
		"total_score":           request.TotalScore,
		"ranking":               request.Ranking,
		"province":              request.Province,
		"subject_combination":   request.SubjectCombination,
		"chinese_score":         request.ChineseScore,
		"math_score":            request.MathScore,
		"english_score":         request.EnglishScore,
		"physics_score":         request.Physics,
		"chemistry_score":       request.Chemistry,
		"biology_score":         request.Biology,
		"history_score":         request.History,
		"geography_score":       request.Geography,
		"politics_score":        request.Politics,
		"preferred_cities":      preferredCities,
		"preferred_majors":      preferredMajors,
		"avoided_majors":        []string{},
		"city_weight":           0.3,
		"major_weight":          0.4,
		"school_ranking_weight": 0.3,
		"is_minority":           false,
		"has_sports_specialty":  false,
		"has_art_specialty":     false,
	}
}

func extractStringSlice(source map[string]interface{}, key string) []string {
	if len(source) == 0 {
		return nil
	}

	raw, ok := source[key]
	if !ok {
		return nil
	}

	switch typed := raw.(type) {
	case []string:
		return typed
	case []interface{}:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			if str, ok := item.(string); ok && strings.TrimSpace(str) != "" {
				result = append(result, str)
			}
		}
		return result
	default:
		return nil
	}
}

func normalizeMaxRecommendations(max int) int {
	if max <= 0 {
		return 30
	}
	return max
}

func callCResult(fn func() *C.CResult) error {
	result := fn()
	if result == nil {
		return fmt.Errorf("nil C result")
	}
	defer C.FreeCResult(result)

	if result.error_code != 0 {
		return fmt.Errorf("%s", C.GoString(result.message))
	}
	return nil
}

func cResultJSON(fn func() *C.CResult) (map[string]interface{}, error) {
	result := fn()
	if result == nil {
		return nil, fmt.Errorf("nil C result")
	}
	defer C.FreeCResult(result)

	if result.error_code != 0 {
		return nil, fmt.Errorf("%s", C.GoString(result.message))
	}

	payload := C.GoString(result.data)
	if strings.TrimSpace(payload) == "" {
		return nil, fmt.Errorf("empty C JSON payload")
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		return nil, err
	}

	return decoded, nil
}

var _ HybridRecommendationBridge = (*CppVolunteerMatcherBridge)(nil)
