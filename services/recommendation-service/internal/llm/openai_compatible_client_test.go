package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestOpenAICompatibleClientCreateChatCompletion(t *testing.T) {
	var capturedReq ChatCompletionRequest
	client := NewOpenAICompatibleClientWithHTTPClient("http://example.com/v1", "test-key", &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/v1/chat/completions" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			if r.Header.Get("Authorization") != "Bearer test-key" {
				t.Fatalf("unexpected auth header: %s", r.Header.Get("Authorization"))
			}
			if err := json.NewDecoder(r.Body).Decode(&capturedReq); err != nil {
				t.Fatalf("decode request failed: %v", err)
			}
			body, _ := json.Marshal(ChatCompletionResponse{
				Choices: []ChatCompletionChoice{{Index: 0, Message: Message{Role: "assistant", Content: "分析完成"}}},
			})
			return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(body)), Header: make(http.Header)}, nil
		}),
	})

	resp, err := client.CreateChatCompletion(context.Background(), ChatCompletionRequest{
		Model:    "gpt-test",
		Messages: []Message{{Role: "user", Content: "hello"}},
	})
	if err != nil {
		t.Fatalf("CreateChatCompletion failed: %v", err)
	}
	if capturedReq.Model != "gpt-test" {
		t.Fatalf("unexpected model: %s", capturedReq.Model)
	}
	if len(resp.Choices) != 1 || resp.Choices[0].Message.Content != "分析完成" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestOpenAICompatibleAnalyzerFallback(t *testing.T) {
	analyzer := NewOpenAICompatibleAnalyzer(nil, "", 0.3, 200, "", NewLocalFallbackAnalyzer())
	report, err := analyzer.AnalyzeRecommendation(context.Background(), RecommendationAnalysisInput{
		Score:           620,
		Province:        "广东",
		TotalCount:      2,
		Recommendations: []RecommendationCandidate{{SchoolName: "A", Probability: 0.82}, {SchoolName: "B", Probability: 0.55}},
	})
	if err != nil {
		t.Fatalf("AnalyzeRecommendation failed: %v", err)
	}
	if report == "" {
		t.Fatal("expected fallback report")
	}
}
