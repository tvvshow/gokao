package scripts

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClientConfig HTTP客户端配置
type HTTPClientConfig struct {
	Timeout     time.Duration
	MaxRetries  int
	RetryDelay  time.Duration
	UserAgent   string
	ContentType string
}

// DefaultHTTPClientConfig 默认HTTP客户端配置
func DefaultHTTPClientConfig() *HTTPClientConfig {
	return &HTTPClientConfig{
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		RetryDelay:  1 * time.Second,
		UserAgent:   "Gaokao-Script-Client/1.0",
		ContentType: "application/json",
	}
}

// HTTPClient HTTP客户端
type HTTPClient struct {
	client  *http.Client
	config  *HTTPClientConfig
}

// NewHTTPClient 创建新的HTTP客户端
func NewHTTPClient(config *HTTPClientConfig) *HTTPClient {
	if config == nil {
		config = DefaultHTTPClientConfig()
	}
	
	client := &http.Client{
		Timeout: config.Timeout,
	}
	
	return &HTTPClient{
		client: client,
		config: config,
	}
}

// Get 发送GET请求
func (c *HTTPClient) Get(url string, headers map[string]string) ([]byte, error) {
	return c.doRequest("GET", url, nil, headers)
}

// Post 发送POST请求
func (c *HTTPClient) Post(url string, body io.Reader, headers map[string]string) ([]byte, error) {
	return c.doRequest("POST", url, body, headers)
}

// doRequest 执行HTTP请求
func (c *HTTPClient) doRequest(method, url string, body io.Reader, headers map[string]string) ([]byte, error) {
	var lastErr error
	
	for i := 0; i <= c.config.MaxRetries; i++ {
		// 创建请求
		req, err := http.NewRequest(method, url, body)
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %w", err)
		}
		
		// 设置默认头部
		req.Header.Set("User-Agent", c.config.UserAgent)
		req.Header.Set("Content-Type", c.config.ContentType)
		
		// 设置自定义头部
		for key, value := range headers {
			req.Header.Set(key, value)
		}
		
		// 发送请求
		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = err
			if i < c.config.MaxRetries {
				time.Sleep(c.config.RetryDelay * time.Duration(i+1))
				continue
			}
			return nil, fmt.Errorf("请求失败: %w", err)
		}
		
		defer resp.Body.Close()
		
		// 读取响应
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			if i < c.config.MaxRetries {
				time.Sleep(c.config.RetryDelay * time.Duration(i+1))
				continue
			}
			return nil, fmt.Errorf("读取响应失败: %w", err)
		}
		
		// 检查状态码
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
			if i < c.config.MaxRetries {
				time.Sleep(c.config.RetryDelay * time.Duration(i+1))
				continue
			}
			return nil, fmt.Errorf("请求失败: HTTP %d: %s, 响应内容: %s", resp.StatusCode, resp.Status, string(data))
		}
		
		return data, nil
	}
	
	return nil, fmt.Errorf("达到最大重试次数，最后错误: %w", lastErr)
}

// WithTimeout 设置请求超时
func (c *HTTPClient) WithTimeout(timeout time.Duration) *HTTPClient {
	c.client.Timeout = timeout
	return c
}

// WithContext 设置请求上下文
func (c *HTTPClient) WithContext(ctx context.Context) *HTTPClient {
	// 创建一个新的客户端，因为http.Client不支持直接设置上下文
	// 上下文应该在NewRequest时设置
	return c
}