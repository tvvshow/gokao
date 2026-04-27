package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OpenAICompatibleClient 是可配置 BaseURL 的 OpenAI 兼容客户端。
type OpenAICompatibleClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewOpenAICompatibleClient 创建 OpenAI 兼容客户端。
func NewOpenAICompatibleClient(baseURL, apiKey string, timeout time.Duration) *OpenAICompatibleClient {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return NewOpenAICompatibleClientWithHTTPClient(baseURL, apiKey, &http.Client{Timeout: timeout})
}

// NewOpenAICompatibleClientWithHTTPClient 创建可注入 HTTP Client 的 OpenAI 兼容客户端。
func NewOpenAICompatibleClientWithHTTPClient(baseURL, apiKey string, httpClient *http.Client) *OpenAICompatibleClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &OpenAICompatibleClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

// CreateChatCompletion 调用 /v1/chat/completions。
func (c *OpenAICompatibleClient) CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("llm base url is empty")
	}
	if req.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	endpoint, err := c.chatCompletionsURL()
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var body bytes.Buffer
		_, _ = body.ReadFrom(resp.Body)
		return nil, fmt.Errorf("llm request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(body.String()))
	}

	var result ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *OpenAICompatibleClient) chatCompletionsURL() (string, error) {
	parsed, err := url.Parse(c.baseURL)
	if err != nil {
		return "", err
	}
	path := strings.TrimRight(parsed.Path, "/")
	if strings.HasSuffix(path, "/chat/completions") {
		return parsed.String(), nil
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/chat/completions"
	return parsed.String(), nil
}

// BaseURL 返回当前配置的基础地址。
func (c *OpenAICompatibleClient) BaseURL() string {
	if c == nil {
		return ""
	}
	return c.baseURL
}
