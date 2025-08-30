package services

import (
	"context"
	"data-service/internal/database"
	"data-service/internal/models"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// AdmissionService 录取数据服务
type AdmissionService struct {
	db     *database.DB
	logger *logrus.Logger
}

// NewAdmissionService 创建录取数据服务实例
func NewAdmissionService(db *database.DB, logger *logrus.Logger) *AdmissionService {
	return &AdmissionService{
		db:     db,
		logger: logger,
	}
}

// AdmissionQueryParams 录取数据查询参数
type AdmissionQueryParams struct {
	// 基本查询
	UniversityID string `form:"university_id"`
	MajorID      string `form:"major_id"`
	
	// 年份筛选
	Year     int   `form:"year"`
	MinYear  int   `form:"min_year"`
	MaxYear  int   `form:"max_year"`
	
	// 地区筛选
	Province string `form:"province"`
	
	// 批次筛选
	Batch    string `form:"batch"`    // early_admission, first_batch, second_batch, third_batch, specialized
	Category string `form:"category"` // science, liberal_arts, comprehensive
	
	// 分数筛选
	MinScore float64 `form:"min_score"`
	MaxScore float64 `form:"max_score"`
	
	// 排名筛选
	MinRank int `form:"min_rank"`
	MaxRank int `form:"max_rank"`
	
	// 难度筛选
	Difficulty string `form:"difficulty"` // very_easy, easy, medium, hard, very_hard
	
	// 排序选项
	SortBy    string `form:"sort_by"`    // year, score, rank
	SortOrder string `form:"sort_order"` // asc, desc
	
	// 分页参数
	Page     int `form:"page,default=1"`
	PageSize int `form:"page_size,default=20"`
	
	// 关联数据
	IncludeUniversity bool `form:"include_university"`
	IncludeMajor      bool `form:"include_major"`
}

// AdmissionListResponse 录取数据列表响应
type AdmissionListResponse struct {
	AdmissionData []models.AdmissionData `json:"admission_data"`
	Total         int64                  `json:"total"`
	Page          int                    `json:"page"`
	PageSize      int                    `json:"page_size"`
	TotalPages    int                    `json:"total_pages"`
}

// AdmissionAnalysis 录取数据分析结果
type AdmissionAnalysis struct {
	University     *models.University `json:"university,omitempty"`
	Major          *models.Major      `json:"major,omitempty"`
	Years          []int              `json:"years"`
	ScoreTrend     []ScorePoint       `json:"score_trend"`
	RankTrend      []RankPoint        `json:"rank_trend"`
	AvgScore       float64            `json:"avg_score"`
	MinScore       float64            `json:"min_score"`
	MaxScore       float64            `json:"max_score"`
	AvgRank        int                `json:"avg_rank"`
	MinRank        int                `json:"min_rank"`
	MaxRank        int                `json:"max_rank"`
	Difficulty     string             `json:"difficulty"`
	Competition    float64            `json:"competition"`
	Recommendation string             `json:"recommendation"`
}

// ScorePoint 分数趋势点
type ScorePoint struct {
	Year  int     `json:"year"`
	Score float64 `json:"score"`
}

// RankPoint 排名趋势点
type RankPoint struct {
	Year int `json:"year"`
	Rank int `json:"rank"`
}

// PredictionRequest 录取预测请求
type PredictionRequest struct {
	UniversityID string  `json:"university_id" validate:"required"`
	MajorID      string  `json:"major_id,omitempty"`
	Province     string  `json:"province" validate:"required"`
	Category     string  `json:"category" validate:"required"`
	Score        float64 `json:"score" validate:"required,min=0,max=1000"`
	Rank         int     `json:"rank,omitempty"`
}

// PredictionResponse 录取预测响应
type PredictionResponse struct {
	Probability    float64 `json:"probability"`    // 录取概率 0-1
	Recommendation string  `json:"recommendation"` // 推荐级别: safe, moderate, risky, very_risky
	MinScore       float64 `json:"min_score"`      // 历史最低分
	AvgScore       float64 `json:"avg_score"`      // 历史平均分
	MaxScore       float64 `json:"max_score"`      // 历史最高分
	ScoreGap       float64 `json:"score_gap"`      // 与平均分的差距
	Analysis       string  `json:"analysis"`       // 分析说明
}

// ListAdmissionData 获取录取数据列表
func (s *AdmissionService) ListAdmissionData(ctx context.Context, params AdmissionQueryParams) (*AdmissionListResponse, error) {
	// 验证分页参数
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > s.db.Config.MaxPageSize {
		params.PageSize = s.db.Config.DefaultPageSize
	}

	// 生成缓存键
	cacheKey := s.generateAdmissionCacheKey("admission:list", params)
	
	// 尝试从缓存获取
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		cached, err := s.db.Redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var response AdmissionListResponse
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				s.logger.Debugf("从缓存获取录取数据列表")
				return &response, nil
			}
		} else if err != redis.Nil {
			s.logger.Warnf("获取缓存失败: %v", err)
		}
	}

	// 构建查询
	query := s.db.PostgreSQL.Model(&models.AdmissionData{})
	
	// 应用筛选条件
	s.applyAdmissionFilters(query, params)
	
	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("计算总数失败: %w", err)
	}

	// 应用排序
	s.applyAdmissionSort(query, params)
	
	// 应用分页
	offset := (params.Page - 1) * params.PageSize
	query = query.Offset(offset).Limit(params.PageSize)
	
	// 预加载关联数据
	if params.IncludeUniversity {
		query = query.Preload("University")
	}
	if params.IncludeMajor {
		query = query.Preload("Major")
	}

	// 执行查询
	var admissionData []models.AdmissionData
	if err := query.Find(&admissionData).Error; err != nil {
		return nil, fmt.Errorf("查询录取数据失败: %w", err)
	}

	// 构建响应
	totalPages := int((total + int64(params.PageSize) - 1) / int64(params.PageSize))
	response := &AdmissionListResponse{
		AdmissionData: admissionData,
		Total:         total,
		Page:          params.Page,
		PageSize:      params.PageSize,
		TotalPages:    totalPages,
	}

	// 缓存结果
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		data, _ := json.Marshal(response)
		s.db.Redis.Set(ctx, cacheKey, data, s.db.Config.CacheDefaultTTL).Err()
	}

	return response, nil
}

// AnalyzeAdmissionData 分析录取数据趋势
func (s *AdmissionService) AnalyzeAdmissionData(ctx context.Context, universityID, majorID, province, category string) (*AdmissionAnalysis, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("admission:analysis:%s:%s:%s:%s", universityID, majorID, province, category)
	
	// 尝试从缓存获取
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		cached, err := s.db.Redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var analysis AdmissionAnalysis
			if err := json.Unmarshal([]byte(cached), &analysis); err == nil {
				s.logger.Debugf("从缓存获取录取分析")
				return &analysis, nil
			}
		} else if err != redis.Nil {
			s.logger.Warnf("获取缓存失败: %v", err)
		}
	}

	// 构建基础查询
	query := s.db.PostgreSQL.Model(&models.AdmissionData{}).
		Where("university_id = ?", universityID).
		Where("province = ?", province).
		Where("category = ?", category)
	
	if majorID != "" {
		query = query.Where("major_id = ?", majorID)
	}

	// 获取录取数据
	var admissionData []models.AdmissionData
	if err := query.Order("year ASC").Find(&admissionData).Error; err != nil {
		return nil, fmt.Errorf("查询录取数据失败: %w", err)
	}

	if len(admissionData) == 0 {
		return nil, fmt.Errorf("未找到相关录取数据")
	}

	// 构建分析结果
	analysis := &AdmissionAnalysis{}
	
	// 获取院校和专业信息
	if universityID != "" {
		var university models.University
		if err := s.db.PostgreSQL.Where("id = ?", universityID).First(&university).Error; err == nil {
			analysis.University = &university
		}
	}
	
	if majorID != "" {
		var major models.Major
		if err := s.db.PostgreSQL.Where("id = ?", majorID).First(&major).Error; err == nil {
			analysis.Major = &major
		}
	}

	// 计算趋势数据
	years := make(map[int]bool)
	var scores []float64
	var ranks []int
	
	for _, data := range admissionData {
		years[data.Year] = true
		
		// 分数趋势
		if data.AvgScore > 0 {
			analysis.ScoreTrend = append(analysis.ScoreTrend, ScorePoint{
				Year:  data.Year,
				Score: data.AvgScore,
			})
			scores = append(scores, data.AvgScore)
		}
		
		// 排名趋势
		if data.AvgRank > 0 {
			analysis.RankTrend = append(analysis.RankTrend, RankPoint{
				Year: data.Year,
				Rank: data.AvgRank,
			})
			ranks = append(ranks, data.AvgRank)
		}
	}

	// 年份列表
	for year := range years {
		analysis.Years = append(analysis.Years, year)
	}

	// 分数统计
	if len(scores) > 0 {
		analysis.AvgScore = s.calculateAverage(scores)
		analysis.MinScore = s.findMin(scores)
		analysis.MaxScore = s.findMax(scores)
	}

	// 排名统计
	if len(ranks) > 0 {
		analysis.AvgRank = int(s.calculateAverageInt(ranks))
		analysis.MinRank = s.findMinInt(ranks)
		analysis.MaxRank = s.findMaxInt(ranks)
	}

	// 计算难度和竞争程度
	analysis.Difficulty = s.calculateDifficulty(admissionData)
	analysis.Competition = s.calculateCompetition(admissionData)
	analysis.Recommendation = s.generateRecommendation(analysis)

	// 缓存结果
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		data, _ := json.Marshal(analysis)
		s.db.Redis.Set(ctx, cacheKey, data, s.db.Config.CacheDefaultTTL).Err()
	}

	return analysis, nil
}

// PredictAdmission 预测录取概率
func (s *AdmissionService) PredictAdmission(ctx context.Context, req PredictionRequest) (*PredictionResponse, error) {
	// 获取历史录取数据
	var admissionData []models.AdmissionData
	query := s.db.PostgreSQL.Model(&models.AdmissionData{}).
		Where("university_id = ?", req.UniversityID).
		Where("province = ?", req.Province).
		Where("category = ?", req.Category).
		Where("year >= ?", time.Now().Year()-5) // 近5年数据

	if req.MajorID != "" {
		query = query.Where("major_id = ?", req.MajorID)
	}

	if err := query.Find(&admissionData).Error; err != nil {
		return nil, fmt.Errorf("查询历史录取数据失败: %w", err)
	}

	if len(admissionData) == 0 {
		return nil, fmt.Errorf("未找到相关历史录取数据")
	}

	// 计算预测结果
	response := &PredictionResponse{}
	
	// 统计分数信息
	var scores []float64
	var ranks []int
	
	for _, data := range admissionData {
		if data.AvgScore > 0 {
			scores = append(scores, data.AvgScore)
		}
		if data.AvgRank > 0 {
			ranks = append(ranks, data.AvgRank)
		}
	}

	if len(scores) > 0 {
		response.MinScore = s.findMin(scores)
		response.AvgScore = s.calculateAverage(scores)
		response.MaxScore = s.findMax(scores)
		response.ScoreGap = req.Score - response.AvgScore
		
		// 计算录取概率
		response.Probability = s.calculateProbability(req.Score, scores)
		
		// 生成推荐
		response.Recommendation = s.generatePredictionRecommendation(response.Probability)
		
		// 生成分析说明
		response.Analysis = s.generateAnalysis(req, response)
	}

	return response, nil
}

// GetAdmissionStatistics 获取录取数据统计
func (s *AdmissionService) GetAdmissionStatistics(ctx context.Context, year int) (map[string]interface{}, error) {
	cacheKey := fmt.Sprintf("admission:statistics:%d", year)
	
	// 尝试从缓存获取
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		cached, err := s.db.Redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var stats map[string]interface{}
			if err := json.Unmarshal([]byte(cached), &stats); err == nil {
				return stats, nil
			}
		}
	}

	stats := make(map[string]interface{})
	
	// 基础查询
	baseQuery := s.db.PostgreSQL.Model(&models.AdmissionData{})
	if year > 0 {
		baseQuery = baseQuery.Where("year = ?", year)
	}

	// 总数统计
	var total int64
	baseQuery.Count(&total)
	stats["total"] = total
	
	// 按省份统计
	var provinceStats []struct {
		Province string `json:"province"`
		Count    int64  `json:"count"`
	}
	baseQuery.Select("province, count(*) as count").
		Group("province").
		Order("count DESC").
		Limit(20).
		Scan(&provinceStats)
	stats["by_province"] = provinceStats
	
	// 按批次统计
	var batchStats []struct {
		Batch string `json:"batch"`
		Count int64  `json:"count"`
	}
	baseQuery.Select("batch, count(*) as count").
		Group("batch").
		Scan(&batchStats)
	stats["by_batch"] = batchStats
	
	// 分数分布统计
	var scoreStats struct {
		AvgScore float64 `json:"avg_score"`
		MinScore float64 `json:"min_score"`
		MaxScore float64 `json:"max_score"`
	}
	baseQuery.Select("AVG(avg_score) as avg_score, MIN(min_score) as min_score, MAX(max_score) as max_score").
		Where("avg_score > 0").
		Scan(&scoreStats)
	stats["score_distribution"] = scoreStats

	// 缓存结果
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		data, _ := json.Marshal(stats)
		s.db.Redis.Set(ctx, cacheKey, data, 30*time.Minute).Err()
	}

	return stats, nil
}

// 辅助方法

// applyAdmissionFilters 应用录取数据筛选条件
func (s *AdmissionService) applyAdmissionFilters(query *gorm.DB, params AdmissionQueryParams) {
	if params.UniversityID != "" {
		query.Where("university_id = ?", params.UniversityID)
	}
	if params.MajorID != "" {
		query.Where("major_id = ?", params.MajorID)
	}
	if params.Year > 0 {
		query.Where("year = ?", params.Year)
	}
	if params.MinYear > 0 {
		query.Where("year >= ?", params.MinYear)
	}
	if params.MaxYear > 0 {
		query.Where("year <= ?", params.MaxYear)
	}
	if params.Province != "" {
		query.Where("province = ?", params.Province)
	}
	if params.Batch != "" {
		query.Where("batch = ?", params.Batch)
	}
	if params.Category != "" {
		query.Where("category = ?", params.Category)
	}
	if params.MinScore > 0 {
		query.Where("avg_score >= ?", params.MinScore)
	}
	if params.MaxScore > 0 {
		query.Where("avg_score <= ?", params.MaxScore)
	}
	if params.MinRank > 0 {
		query.Where("avg_rank >= ?", params.MinRank)
	}
	if params.MaxRank > 0 {
		query.Where("avg_rank <= ?", params.MaxRank)
	}
	if params.Difficulty != "" {
		query.Where("difficulty = ?", params.Difficulty)
	}
}

// applyAdmissionSort 应用排序
func (s *AdmissionService) applyAdmissionSort(query *gorm.DB, params AdmissionQueryParams) {
	sortBy := params.SortBy
	if sortBy == "" {
		sortBy = "year"
	}
	
	sortOrder := strings.ToUpper(params.SortOrder)
	if sortOrder != "ASC" {
		sortOrder = "DESC"
	}
	
	switch sortBy {
	case "year":
		query.Order("year " + sortOrder)
	case "score":
		query.Order("avg_score " + sortOrder + ", year DESC")
	case "rank":
		query.Order("avg_rank " + (map[string]string{"DESC": "ASC", "ASC": "DESC"}[sortOrder]) + ", year DESC")
	default:
		query.Order("year DESC")
	}
}

// calculateAverage 计算平均值
func (s *AdmissionService) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateAverageInt 计算整数平均值
func (s *AdmissionService) calculateAverageInt(values []int) float64 {
	if len(values) == 0 {
		return 0
	}
	
	sum := 0
	for _, v := range values {
		sum += v
	}
	return float64(sum) / float64(len(values))
}

// findMin 找最小值
func (s *AdmissionService) findMin(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	min := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

// findMax 找最大值
func (s *AdmissionService) findMax(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

// findMinInt 找整数最小值
func (s *AdmissionService) findMinInt(values []int) int {
	if len(values) == 0 {
		return 0
	}
	
	min := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

// findMaxInt 找整数最大值
func (s *AdmissionService) findMaxInt(values []int) int {
	if len(values) == 0 {
		return 0
	}
	
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

// calculateDifficulty 计算录取难度
func (s *AdmissionService) calculateDifficulty(data []models.AdmissionData) string {
	if len(data) == 0 {
		return "unknown"
	}
	
	// 简单的难度计算逻辑，实际应该更复杂
	var avgAdmissionRate float64
	count := 0
	
	for _, d := range data {
		if d.AdmissionRate > 0 {
			avgAdmissionRate += d.AdmissionRate
			count++
		}
	}
	
	if count > 0 {
		avgAdmissionRate /= float64(count)
		
		if avgAdmissionRate >= 0.8 {
			return "very_easy"
		} else if avgAdmissionRate >= 0.6 {
			return "easy"
		} else if avgAdmissionRate >= 0.4 {
			return "medium"
		} else if avgAdmissionRate >= 0.2 {
			return "hard"
		} else {
			return "very_hard"
		}
	}
	
	return "medium"
}

// calculateCompetition 计算竞争激烈程度
func (s *AdmissionService) calculateCompetition(data []models.AdmissionData) float64 {
	if len(data) == 0 {
		return 0
	}
	
	var totalCompetition float64
	count := 0
	
	for _, d := range data {
		if d.Competition > 0 {
			totalCompetition += d.Competition
			count++
		}
	}
	
	if count > 0 {
		return totalCompetition / float64(count)
	}
	
	return 0
}

// calculateProbability 计算录取概率
func (s *AdmissionService) calculateProbability(score float64, historicalScores []float64) float64 {
	if len(historicalScores) == 0 {
		return 0
	}
	
	// 简单的概率计算，实际应该使用更复杂的统计模型
	lowerCount := 0
	for _, s := range historicalScores {
		if score >= s {
			lowerCount++
		}
	}
	
	return float64(lowerCount) / float64(len(historicalScores))
}

// generateRecommendation 生成推荐
func (s *AdmissionService) generateRecommendation(analysis *AdmissionAnalysis) string {
	if analysis.Competition > 0.8 {
		return "竞争激烈，建议谨慎考虑"
	} else if analysis.Competition > 0.6 {
		return "竞争较激烈，建议充分准备"
	} else if analysis.Competition > 0.4 {
		return "竞争适中，可以考虑"
	} else {
		return "竞争相对较小，推荐报考"
	}
}

// generatePredictionRecommendation 生成预测推荐
func (s *AdmissionService) generatePredictionRecommendation(probability float64) string {
	if probability >= 0.8 {
		return "safe"
	} else if probability >= 0.6 {
		return "moderate"
	} else if probability >= 0.3 {
		return "risky"
	} else {
		return "very_risky"
	}
}

// generateAnalysis 生成分析说明
func (s *AdmissionService) generateAnalysis(req PredictionRequest, resp *PredictionResponse) string {
	var analysis strings.Builder
	
	if resp.ScoreGap > 0 {
		analysis.WriteString(fmt.Sprintf("您的分数比历年平均分高%.1f分，", resp.ScoreGap))
	} else {
		analysis.WriteString(fmt.Sprintf("您的分数比历年平均分低%.1f分，", -resp.ScoreGap))
	}
	
	switch resp.Recommendation {
	case "safe":
		analysis.WriteString("录取把握较大，建议作为保底志愿。")
	case "moderate":
		analysis.WriteString("录取有一定把握，可以作为稳妥志愿。")
	case "risky":
		analysis.WriteString("录取存在风险，建议作为冲刺志愿。")
	case "very_risky":
		analysis.WriteString("录取风险较大，不建议填报。")
	}
	
	return analysis.String()
}

// generateAdmissionCacheKey 生成缓存键
func (s *AdmissionService) generateAdmissionCacheKey(prefix string, params AdmissionQueryParams) string {
	data, _ := json.Marshal(params)
	return fmt.Sprintf("%s:%x", prefix, data)
}