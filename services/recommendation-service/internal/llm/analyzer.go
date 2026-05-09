package llm

import (
	"context"
	"fmt"
	"strings"
)

// OpenAICompatibleAnalyzer 基于 OpenAI 兼容客户端生成分析。
type OpenAICompatibleAnalyzer struct {
	client       Client
	model        string
	temperature  float64
	maxTokens    int
	systemPrompt string
	fallback     Analyzer
}

// NewOpenAICompatibleAnalyzer 创建分析器。
func NewOpenAICompatibleAnalyzer(client Client, model string, temperature float64, maxTokens int, systemPrompt string, fallback Analyzer) *OpenAICompatibleAnalyzer {
	if systemPrompt == "" {
		systemPrompt = DefaultSystemPrompt
	}
	return &OpenAICompatibleAnalyzer{
		client:       client,
		model:        model,
		temperature:  temperature,
		maxTokens:    maxTokens,
		systemPrompt: systemPrompt,
		fallback:     fallback,
	}
}

// AnalyzeRecommendation 生成分析报告。
func (a *OpenAICompatibleAnalyzer) AnalyzeRecommendation(ctx context.Context, input RecommendationAnalysisInput) (string, error) {
	if a == nil || a.client == nil || a.model == "" {
		if a != nil && a.fallback != nil {
			return a.fallback.AnalyzeRecommendation(ctx, input)
		}
		return "", fmt.Errorf("llm analyzer is not configured")
	}

	prompt := buildRecommendationAnalysisPrompt(input)
	req := ChatCompletionRequest{
		Model:       a.model,
		Temperature: a.temperature,
		MaxTokens:   a.maxTokens,
		Messages: []Message{
			{Role: "system", Content: a.systemPrompt},
			{Role: "user", Content: prompt},
		},
	}

	resp, err := a.client.CreateChatCompletion(ctx, req)
	if err != nil {
		if a.fallback != nil {
			return a.fallback.AnalyzeRecommendation(ctx, input)
		}
		return "", err
	}
	if len(resp.Choices) == 0 {
		if a.fallback != nil {
			return a.fallback.AnalyzeRecommendation(ctx, input)
		}
		return "", fmt.Errorf("llm response has no choices")
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	if content == "" {
		if a.fallback != nil {
			return a.fallback.AnalyzeRecommendation(ctx, input)
		}
		return "", fmt.Errorf("llm response content is empty")
	}
	return content, nil
}

// Status 返回当前分析器运行状态，便于 system/status 暴露。
//
// 安全原则：base_url、api_key、system_prompt 等"上游接口指纹"不在此处导出，避免
// 通过公开 status 端点泄露分析后端供攻击者直接调用或进行配额消耗。运维端如需排查可
// 直接查 LLM_BASE_URL 环境变量或服务启动日志。
func (a *OpenAICompatibleAnalyzer) Status() map[string]interface{} {
	status := map[string]interface{}{
		"enabled":       a != nil && a.client != nil && strings.TrimSpace(a.model) != "",
		"provider":      "openai-compatible",
		"model":         "",
		"temperature":   0.0,
		"max_tokens":    0,
		"fallback_mode": fallbackMode(a),
		"status":        analyzerStatus(a),
	}
	if a == nil {
		return status
	}
	status["model"] = a.model
	status["temperature"] = a.temperature
	status["max_tokens"] = a.maxTokens
	return status
}

func buildRecommendationAnalysisPrompt(input RecommendationAnalysisInput) string {
	var b strings.Builder
	b.WriteString("请基于以下学生信息和志愿推荐结果生成一段分析报告，要求：中文、简洁、专业、可执行；明确给出整体判断、风险分布、填报建议、关注点。\n\n")
	b.WriteString(fmt.Sprintf("学生：%s\n", blankOrPlaceholder(input.StudentName, "未提供")))
	b.WriteString(fmt.Sprintf("分数：%d\n", input.Score))
	b.WriteString(fmt.Sprintf("省份：%s\n", blankOrPlaceholder(input.Province, "未提供")))
	b.WriteString(fmt.Sprintf("选科：%s\n", blankOrPlaceholder(input.SubjectCombination, "未提供")))
	if input.Rank != nil {
		b.WriteString(fmt.Sprintf("位次：%d\n", *input.Rank))
	}
	if input.RiskTolerance != "" {
		b.WriteString(fmt.Sprintf("风险偏好：%s\n", input.RiskTolerance))
	}
	if len(input.PreferredRegions) > 0 {
		b.WriteString(fmt.Sprintf("地区偏好：%s\n", strings.Join(input.PreferredRegions, ", ")))
	}
	if len(input.PreferredMajors) > 0 {
		b.WriteString(fmt.Sprintf("专业偏好：%s\n", strings.Join(input.PreferredMajors, ", ")))
	}
	if len(input.UniversityTypes) > 0 {
		b.WriteString(fmt.Sprintf("院校类型偏好：%s\n", strings.Join(input.UniversityTypes, ", ")))
	}
	if input.SpecialRequirements != "" {
		b.WriteString(fmt.Sprintf("特殊要求：%s\n", input.SpecialRequirements))
	}
	b.WriteString(fmt.Sprintf("推荐总数：%d\n\n", input.TotalCount))
	b.WriteString("推荐摘要：\n")
	for i, rec := range input.Recommendations {
		if i >= 8 {
			break
		}
		b.WriteString(fmt.Sprintf("%d. %s / %s，录取概率%d%%，匹配度%d，风险%s，院校类型%s，地区%s %s\n",
			i+1, rec.SchoolName, rec.MajorName, rec.AdmissionProbability, rec.MatchScore, blankOrPlaceholder(rec.RiskLevel, "未知"), blankOrPlaceholder(rec.Type, "未知"), blankOrPlaceholder(rec.Province, ""), blankOrPlaceholder(rec.City, "")))
		if rec.Reason != "" {
			b.WriteString(fmt.Sprintf("   理由：%s\n", rec.Reason))
		}
	}
	b.WriteString("\n请输出一段 150~300 字的分析。")
	return b.String()
}

func blankOrPlaceholder(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

// LocalFallbackAnalyzer 使用现有规则生成简短分析。
type LocalFallbackAnalyzer struct{}

func NewLocalFallbackAnalyzer() *LocalFallbackAnalyzer { return &LocalFallbackAnalyzer{} }

func (a *LocalFallbackAnalyzer) AnalyzeRecommendation(ctx context.Context, input RecommendationAnalysisInput) (string, error) {
	if len(input.Recommendations) == 0 {
		return "暂无推荐结果", nil
	}

	stableCount := 0
	moderateCount := 0
	reachCount := 0
	for _, rec := range input.Recommendations {
		switch {
		case rec.Probability >= 0.8:
			stableCount++
		case rec.Probability >= 0.6:
			moderateCount++
		default:
			reachCount++
		}
	}

	return fmt.Sprintf("根据您的分数和偏好，为您推荐了%d所院校。其中稳妥选择%d个，适中选择%d个，冲刺选择%d个。建议合理搭配，确保志愿填报的科学性和安全性。",
		len(input.Recommendations), stableCount, moderateCount, reachCount), nil
}

func (a *LocalFallbackAnalyzer) Status() map[string]interface{} {
	return map[string]interface{}{
		"enabled":       false,
		"provider":      "local-fallback",
		"status":        "degraded",
		"model":         "rule-based-summary",
		"max_tokens":    0,
		"fallback_mode": "local_rules",
	}
}

func analyzerStatus(a *OpenAICompatibleAnalyzer) string {
	if a == nil || a.client == nil || strings.TrimSpace(a.model) == "" {
		return "not_configured"
	}
	if a.fallback != nil {
		return "degraded"
	}
	return "healthy"
}

func fallbackMode(a *OpenAICompatibleAnalyzer) string {
	if a == nil || a.fallback == nil {
		return "none"
	}
	if statusReporter, ok := a.fallback.(StatusReporter); ok {
		if provider, ok := statusReporter.Status()["fallback_mode"].(string); ok && provider != "" {
			return provider
		}
	}
	return "static_fallback"
}
