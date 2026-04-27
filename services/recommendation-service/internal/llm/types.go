package llm

import "context"

// Message 表示对话消息，兼容 OpenAI Chat Completions 风格。
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest 表示兼容 OpenAI 的对话补全请求。
type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// ChatCompletionResponse 表示兼容 OpenAI 的对话补全响应。
type ChatCompletionResponse struct {
	ID      string                 `json:"id,omitempty"`
	Object  string                 `json:"object,omitempty"`
	Created int64                  `json:"created,omitempty"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   *ChatCompletionUsage   `json:"usage,omitempty"`
}

// ChatCompletionChoice 表示返回结果中的单条候选。
type ChatCompletionChoice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason,omitempty"`
}

// ChatCompletionUsage 表示 token 统计。
type ChatCompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Client 是兼容 OpenAI 的最小客户端接口。
type Client interface {
	CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error)
}

// Analyzer 生成分析报告。
type Analyzer interface {
	AnalyzeRecommendation(ctx context.Context, input RecommendationAnalysisInput) (string, error)
}

// RecommendationAnalysisInput 是推荐分析输入。
type RecommendationAnalysisInput struct {
	StudentName         string
	Score               int
	Province            string
	SubjectCombination  string
	Rank                *int
	RiskTolerance       string
	PreferredRegions    []string
	PreferredMajors     []string
	UniversityTypes     []string
	SpecialRequirements string
	Recommendations     []RecommendationCandidate
	TotalCount          int
}

// RecommendationCandidate 是模型分析时使用的推荐候选摘要。
type RecommendationCandidate struct {
	SchoolName           string
	MajorName            string
	Probability          float64
	AdmissionProbability int
	MatchScore           int
	RiskLevel            string
	Type                 string
	Province             string
	City                 string
	Reason               string
}

// DefaultSystemPrompt 默认系统提示词。
const DefaultSystemPrompt = "你是一名高考志愿分析助手。请基于学生分数、地区偏好、风险偏好和推荐结果，输出简洁、专业、可执行的中文分析。"

// StatusReporter 暴露分析器当前运行状态。
type StatusReporter interface {
	Status() map[string]interface{}
}
