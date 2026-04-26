package feature_engineering

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// AdvancedFeatureEngineering 高级特征工程框架
type AdvancedFeatureEngineering struct {
	// 自动特征选择器
	autoFeatureSelector *AutoMLFeatureSelector
	// 特征交互发现器
	interactionDetector *FeatureInteractionEngine
	// 在线特征更新器
	onlineFeatureUpdater *StreamingFeatureUpdater
	// 特征重要性分析器
	featureImportanceAnalyzer *FeatureImportanceAnalyzer

	logger *logrus.Logger
	mu     sync.RWMutex
}

// AutoMLFeatureSelector 自动机器学习特征选择器
type AutoMLFeatureSelector struct {
	selectedFeatures map[string]bool
	featureScores    map[string]float64
	selectionMethod  string
}

// FeatureInteractionEngine 特征交互发现引擎
type FeatureInteractionEngine struct {
	interactionGraph     map[string]map[string]float64
	interactionThreshold float64
}

// StreamingFeatureUpdater 流式特征更新器
type StreamingFeatureUpdater struct {
	featureWindowSize int
	updateInterval    time.Duration
	featureStatistics map[string]*FeatureStatistics
}

// FeatureImportanceAnalyzer 特征重要性分析器
type FeatureImportanceAnalyzer struct {
	importanceScores map[string]float64
	analysisMethod   string
}

// FeatureStatistics 特征统计信息
type FeatureStatistics struct {
	Count     int
	Sum       float64
	Mean      float64
	StdDev    float64
	Min       float64
	Max       float64
	Histogram map[float64]int
}

// NewAdvancedFeatureEngineering 创建新的高级特征工程框架
func NewAdvancedFeatureEngineering(logger *logrus.Logger) (*AdvancedFeatureEngineering, error) {
	engine := &AdvancedFeatureEngineering{
		logger: logger,
	}

	// 初始化各个组件
	if err := engine.initComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize feature engineering components: %w", err)
	}

	return engine, nil
}

// ExtractFeatures 提取特征
func (e *AdvancedFeatureEngineering) ExtractFeatures(ctx context.Context,
	rawData map[string]interface{}) (map[string]float64, error) {

	e.mu.RLock()
	defer e.mu.RUnlock()

	// 基础特征提取
	baseFeatures, err := e.extractBaseFeatures(rawData)
	if err != nil {
		return nil, fmt.Errorf("failed to extract base features: %w", err)
	}

	// 自动特征选择
	selectedFeatures := e.autoFeatureSelector.SelectFeatures(baseFeatures)

	// 特征交互发现
	interactionFeatures := e.interactionDetector.DiscoverInteractions(selectedFeatures)

	// 合并特征
	finalFeatures := make(map[string]float64)
	for feature, value := range selectedFeatures {
		finalFeatures[feature] = value
	}
	for feature, value := range interactionFeatures {
		finalFeatures[feature] = value
	}

	// 特征缩放和标准化
	finalFeatures = e.normalizeFeatures(finalFeatures)

	return finalFeatures, nil
}

// extractBaseFeatures 提取基础特征
func (e *AdvancedFeatureEngineering) extractBaseFeatures(rawData map[string]interface{}) (map[string]float64, error) {
	features := make(map[string]float64)

	// 数值型特征直接提取
	for key, value := range rawData {
		switch v := value.(type) {
		case int:
			features[key] = float64(v)
		case float64:
			features[key] = v
		case float32:
			features[key] = float64(v)
		case bool:
			if v {
				features[key] = 1.0
			} else {
				features[key] = 0.0
			}
		}
	}

	// 计算派生特征
	features = e.calculateDerivedFeatures(features)

	return features, nil
}

// calculateDerivedFeatures 计算派生特征
func (e *AdvancedFeatureEngineering) calculateDerivedFeatures(features map[string]float64) map[string]float64 {
	// 成绩相关派生特征
	if totalScore, ok := features["total_score"]; ok {
		features["score_log"] = math.Log(totalScore + 1)
		features["score_sqrt"] = math.Sqrt(totalScore)
		features["score_normalized"] = (totalScore - 400) / 300 // 假设分数范围400-700
	}

	// 排名相关派生特征
	if ranking, ok := features["ranking"]; ok && ranking > 0 {
		features["ranking_reciprocal"] = 1.0 / ranking
		features["ranking_log"] = math.Log(ranking + 1)
	}

	// 时间相关派生特征
	if year, ok := features["year"]; ok {
		currentYear := float64(time.Now().Year())
		features["years_since"] = currentYear - year
	}

	// 交互特征（基础版）
	if score, ok := features["total_score"]; ok {
		if minScore, ok := features["min_admission_score"]; ok {
			features["score_gap"] = score - minScore
			features["score_ratio"] = score / minScore
			features["score_gap_squared"] = features["score_gap"] * features["score_gap"]
		}
	}

	return features
}

// normalizeFeatures 特征标准化
func (e *AdvancedFeatureEngineering) normalizeFeatures(features map[string]float64) map[string]float64 {
	normalized := make(map[string]float64)

	for feature, value := range features {
		// 获取特征统计信息
		stats := e.onlineFeatureUpdater.GetFeatureStatistics(feature)

		if stats != nil && stats.StdDev > 0 {
			// Z-score标准化
			normalized[feature] = (value - stats.Mean) / stats.StdDev
		} else {
			// Min-Max标准化
			minVal, maxVal := e.getFeatureRange(feature)
			if maxVal > minVal {
				normalized[feature] = (value - minVal) / (maxVal - minVal)
			} else {
				normalized[feature] = value
			}
		}
	}

	return normalized
}

// getFeatureRange 获取特征值范围
func (e *AdvancedFeatureEngineering) getFeatureRange(feature string) (float64, float64) {
	// 这里应该从特征统计信息中获取
	// 简化实现
	switch feature {
	case "total_score":
		return 400.0, 700.0
	case "ranking":
		return 1.0, 10000.0
	case "min_admission_score":
		return 400.0, 700.0
	default:
		return 0.0, 1.0
	}
}

// UpdateFeatureStatistics 更新特征统计信息
func (e *AdvancedFeatureEngineering) UpdateFeatureStatistics(ctx context.Context,
	features map[string]float64) error {

	e.mu.Lock()
	defer e.mu.Unlock()

	// 更新在线特征统计
	if err := e.onlineFeatureUpdater.Update(features); err != nil {
		return fmt.Errorf("failed to update feature statistics: %w", err)
	}

	// 更新特征重要性
	if err := e.featureImportanceAnalyzer.AnalyzeImportance(features); err != nil {
		return fmt.Errorf("failed to analyze feature importance: %w", err)
	}

	// 更新自动特征选择
	if err := e.autoFeatureSelector.UpdateSelection(features); err != nil {
		return fmt.Errorf("failed to update feature selection: %w", err)
	}

	return nil
}

// GetFeatureImportance 获取特征重要性
func (e *AdvancedFeatureEngineering) GetFeatureImportance() map[string]float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.featureImportanceAnalyzer.GetImportanceScores()
}

// GetFeatureInteractions 获取特征交互
func (e *AdvancedFeatureEngineering) GetFeatureInteractions() map[string]map[string]float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.interactionDetector.GetInteractionGraph()
}

// GetSelectedFeatures 获取选择的特征
func (e *AdvancedFeatureEngineering) GetSelectedFeatures() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.autoFeatureSelector.GetSelectedFeatures()
}

// ExportFeaturePipeline 导出特征处理管道
func (e *AdvancedFeatureEngineering) ExportFeaturePipeline() ([]byte, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	pipeline := map[string]interface{}{
		"selected_features":    e.autoFeatureSelector.Export(),
		"feature_importance":   e.featureImportanceAnalyzer.Export(),
		"feature_interactions": e.interactionDetector.Export(),
		"feature_statistics":   e.onlineFeatureUpdater.Export(),
		"timestamp":            time.Now().Unix(),
	}

	return json.Marshal(pipeline)
}

// ImportFeaturePipeline 导入特征处理管道
func (e *AdvancedFeatureEngineering) ImportFeaturePipeline(data []byte) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	var pipeline map[string]interface{}
	if err := json.Unmarshal(data, &pipeline); err != nil {
		return fmt.Errorf("failed to unmarshal feature pipeline: %w", err)
	}

	// 导入各个组件
	if err := e.autoFeatureSelector.Import(pipeline["selected_features"]); err != nil {
		return fmt.Errorf("failed to import feature selection: %w", err)
	}

	if err := e.featureImportanceAnalyzer.Import(pipeline["feature_importance"]); err != nil {
		return fmt.Errorf("failed to import feature importance: %w", err)
	}

	if err := e.interactionDetector.Import(pipeline["feature_interactions"]); err != nil {
		return fmt.Errorf("failed to import feature interactions: %w", err)
	}

	if err := e.onlineFeatureUpdater.Import(pipeline["feature_statistics"]); err != nil {
		return fmt.Errorf("failed to import feature statistics: %w", err)
	}

	return nil
}

// SelectFeatures 自动特征选择
func (s *AutoMLFeatureSelector) SelectFeatures(features map[string]float64) map[string]float64 {
	if len(s.selectedFeatures) == 0 {
		for k := range features {
			s.selectedFeatures[k] = true
		}
	}
	selected := make(map[string]float64)
	for k, v := range features {
		if s.selectedFeatures[k] {
			selected[k] = v
		}
	}
	return selected
}

// UpdateSelection 更新特征选择
func (s *AutoMLFeatureSelector) UpdateSelection(features map[string]float64) error {
	for k, v := range features {
		score := math.Abs(v)
		s.featureScores[k] = score
		s.selectedFeatures[k] = score >= 0.01
	}
	return nil
}

// GetSelectedFeatures 返回当前选择的特征列表
func (s *AutoMLFeatureSelector) GetSelectedFeatures() []string {
	keys := make([]string, 0, len(s.selectedFeatures))
	for k, selected := range s.selectedFeatures {
		if selected {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

// Export 导出特征选择器状态
func (s *AutoMLFeatureSelector) Export() map[string]interface{} {
	return map[string]interface{}{
		"selected_features": s.selectedFeatures,
		"feature_scores":    s.featureScores,
		"selection_method":  s.selectionMethod,
	}
}

// Import 导入特征选择器状态
func (s *AutoMLFeatureSelector) Import(data interface{}) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	var payload struct {
		SelectedFeatures map[string]bool    `json:"selected_features"`
		FeatureScores    map[string]float64 `json:"feature_scores"`
		SelectionMethod  string             `json:"selection_method"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}
	if payload.SelectedFeatures != nil {
		s.selectedFeatures = payload.SelectedFeatures
	}
	if payload.FeatureScores != nil {
		s.featureScores = payload.FeatureScores
	}
	if payload.SelectionMethod != "" {
		s.selectionMethod = payload.SelectionMethod
	}
	return nil
}

// DiscoverInteractions 发现特征交互项
func (f *FeatureInteractionEngine) DiscoverInteractions(features map[string]float64) map[string]float64 {
	result := make(map[string]float64)
	keys := make([]string, 0, len(features))
	for k := range features {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			a, b := keys[i], keys[j]
			score := math.Abs(features[a] * features[b])
			if score >= f.interactionThreshold {
				name := a + "_x_" + b
				result[name] = features[a] * features[b]
				if _, ok := f.interactionGraph[a]; !ok {
					f.interactionGraph[a] = make(map[string]float64)
				}
				f.interactionGraph[a][b] = score
			}
		}
	}
	return result
}

// GetInteractionGraph 返回交互图
func (f *FeatureInteractionEngine) GetInteractionGraph() map[string]map[string]float64 {
	return f.interactionGraph
}

// Export 导出交互引擎状态
func (f *FeatureInteractionEngine) Export() map[string]interface{} {
	return map[string]interface{}{
		"interaction_graph":     f.interactionGraph,
		"interaction_threshold": f.interactionThreshold,
	}
}

// Import 导入交互引擎状态
func (f *FeatureInteractionEngine) Import(data interface{}) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	var payload struct {
		InteractionGraph     map[string]map[string]float64 `json:"interaction_graph"`
		InteractionThreshold float64                       `json:"interaction_threshold"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}
	if payload.InteractionGraph != nil {
		f.interactionGraph = payload.InteractionGraph
	}
	if payload.InteractionThreshold > 0 {
		f.interactionThreshold = payload.InteractionThreshold
	}
	return nil
}

// GetFeatureStatistics 返回某个特征统计
func (s *StreamingFeatureUpdater) GetFeatureStatistics(feature string) *FeatureStatistics {
	return s.featureStatistics[feature]
}

// Update 增量更新特征统计
func (s *StreamingFeatureUpdater) Update(features map[string]float64) error {
	for k, v := range features {
		stats, ok := s.featureStatistics[k]
		if !ok {
			s.featureStatistics[k] = &FeatureStatistics{
				Count:     1,
				Sum:       v,
				Mean:      v,
				StdDev:    0,
				Min:       v,
				Max:       v,
				Histogram: map[float64]int{v: 1},
			}
			continue
		}
		stats.Count++
		stats.Sum += v
		prevMean := stats.Mean
		stats.Mean = stats.Sum / float64(stats.Count)
		stats.StdDev = math.Sqrt(((stats.StdDev * stats.StdDev * float64(stats.Count-1)) + ((v - prevMean) * (v - stats.Mean))) / float64(stats.Count))
		if v < stats.Min {
			stats.Min = v
		}
		if v > stats.Max {
			stats.Max = v
		}
		if stats.Histogram == nil {
			stats.Histogram = make(map[float64]int)
		}
		stats.Histogram[v]++
	}
	return nil
}

// Export 导出流式更新器状态
func (s *StreamingFeatureUpdater) Export() map[string]interface{} {
	return map[string]interface{}{
		"feature_window_size": s.featureWindowSize,
		"update_interval_ns":  int64(s.updateInterval),
		"feature_statistics":  s.featureStatistics,
	}
}

// Import 导入流式更新器状态
func (s *StreamingFeatureUpdater) Import(data interface{}) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	var payload struct {
		FeatureWindowSize int                           `json:"feature_window_size"`
		UpdateIntervalNS  int64                         `json:"update_interval_ns"`
		FeatureStatistics map[string]*FeatureStatistics `json:"feature_statistics"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}
	if payload.FeatureWindowSize > 0 {
		s.featureWindowSize = payload.FeatureWindowSize
	}
	if payload.UpdateIntervalNS > 0 {
		s.updateInterval = time.Duration(payload.UpdateIntervalNS)
	}
	if payload.FeatureStatistics != nil {
		s.featureStatistics = payload.FeatureStatistics
	}
	return nil
}

// AnalyzeImportance 分析特征重要性
func (f *FeatureImportanceAnalyzer) AnalyzeImportance(features map[string]float64) error {
	for k, v := range features {
		f.importanceScores[k] = math.Abs(v)
	}
	return nil
}

// GetImportanceScores 返回特征重要性分数
func (f *FeatureImportanceAnalyzer) GetImportanceScores() map[string]float64 {
	return f.importanceScores
}

// Export 导出重要性分析器状态
func (f *FeatureImportanceAnalyzer) Export() map[string]interface{} {
	return map[string]interface{}{
		"importance_scores": f.importanceScores,
		"analysis_method":   f.analysisMethod,
	}
}

// Import 导入重要性分析器状态
func (f *FeatureImportanceAnalyzer) Import(data interface{}) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	var payload struct {
		ImportanceScores map[string]float64 `json:"importance_scores"`
		AnalysisMethod   string             `json:"analysis_method"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}
	if payload.ImportanceScores != nil {
		f.importanceScores = payload.ImportanceScores
	}
	if payload.AnalysisMethod != "" {
		f.analysisMethod = payload.AnalysisMethod
	}
	return nil
}

// 初始化各个组件
func (e *AdvancedFeatureEngineering) initComponents() error {
	// 初始化自动特征选择器
	e.autoFeatureSelector = &AutoMLFeatureSelector{
		selectedFeatures: make(map[string]bool),
		featureScores:    make(map[string]float64),
		selectionMethod:  "random_forest",
	}

	// 初始化特征交互发现器
	e.interactionDetector = &FeatureInteractionEngine{
		interactionGraph:     make(map[string]map[string]float64),
		interactionThreshold: 0.3,
	}

	// 初始化在线特征更新器
	e.onlineFeatureUpdater = &StreamingFeatureUpdater{
		featureWindowSize: 1000,
		updateInterval:    time.Minute,
		featureStatistics: make(map[string]*FeatureStatistics),
	}

	// 初始化特征重要性分析器
	e.featureImportanceAnalyzer = &FeatureImportanceAnalyzer{
		importanceScores: make(map[string]float64),
		analysisMethod:   "permutation_importance",
	}

	return nil
}
