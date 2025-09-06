package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// TestConfig 测试配置
type TestConfig struct {
	BaseURL         string        `json:"base_url"`
	ConcurrentUsers int           `json:"concurrent_users"`
	RequestsPerUser int           `json:"requests_per_user"`
	RequestTimeout  time.Duration `json:"request_timeout"`
}

// TestResult 测试结果
type TestResult struct {
	TotalRequests    int           `json:"total_requests"`
	SuccessRequests  int           `json:"success_requests"`
	FailedRequests   int           `json:"failed_requests"`
	TotalTime        time.Duration `json:"total_time"`
	AvgResponseTime  time.Duration `json:"avg_response_time"`
	MaxResponseTime  time.Duration `json:"max_response_time"`
	MinResponseTime  time.Duration `json:"min_response_time"`
	RequestsPerSec   float64       `json:"requests_per_sec"`
	ErrorRate        float64       `json:"error_rate"`
}

// StudentRequest 学生推荐请求
type StudentRequest struct {
	StudentID          string   `json:"student_id"`
	Name               string   `json:"name"`
	TotalScore         int      `json:"total_score"`
	Ranking            int      `json:"ranking"`
	Province           string   `json:"province"`
	SubjectCombination string   `json:"subject_combination"`
	PreferredCities    []string `json:"preferred_cities"`
	PreferredMajors    []string `json:"preferred_majors"`
	MaxVolunteers      int      `json:"max_volunteers"`
}

func main() {
	config := TestConfig{
		BaseURL:         "http://localhost:8083",
		ConcurrentUsers: 10,
		RequestsPerUser: 100,
		RequestTimeout:  30 * time.Second,
	}

	fmt.Printf("开始性能测试...\n")
	fmt.Printf("并发用户数: %d\n", config.ConcurrentUsers)
	fmt.Printf("每用户请求数: %d\n", config.RequestsPerUser)
	fmt.Printf("总请求数: %d\n", config.ConcurrentUsers*config.RequestsPerUser)

	// 先测试服务是否可用
	if !testServiceHealth(config.BaseURL) {
		log.Fatal("服务不可用，请确保推荐服务正在运行")
	}

	// 执行性能测试
	result := runPerformanceTest(config)

	// 输出结果
	printResults(result)

	// 测试不同的API端点
	testAPIEndpoints(config.BaseURL)
}

// testServiceHealth 测试服务健康状况
func testServiceHealth(baseURL string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("健康检查失败: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("健康检查返回状态码: %d\n", resp.StatusCode)
		return false
	}

	fmt.Println("✓ 服务健康检查通过")
	return true
}

// runPerformanceTest 运行性能测试
func runPerformanceTest(config TestConfig) TestResult {
	var wg sync.WaitGroup
	resultChan := make(chan time.Duration, config.ConcurrentUsers*config.RequestsPerUser)
	errorChan := make(chan error, config.ConcurrentUsers*config.RequestsPerUser)

	startTime := time.Now()

	// 创建并发用户
	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			runUserRequests(config, userID, resultChan, errorChan)
		}(i)
	}

	wg.Wait()
	close(resultChan)
	close(errorChan)

	endTime := time.Now()
	totalTime := endTime.Sub(startTime)

	// 统计结果
	var responseTimes []time.Duration
	successCount := 0

	for responseTime := range resultChan {
		responseTimes = append(responseTimes, responseTime)
		successCount++
	}

	errorCount := 0
	for range errorChan {
		errorCount++
	}

	// 计算统计信息
	result := TestResult{
		TotalRequests:   config.ConcurrentUsers * config.RequestsPerUser,
		SuccessRequests: successCount,
		FailedRequests:  errorCount,
		TotalTime:       totalTime,
	}

	if len(responseTimes) > 0 {
		var totalResponseTime time.Duration
		minTime := responseTimes[0]
		maxTime := responseTimes[0]

		for _, rt := range responseTimes {
			totalResponseTime += rt
			if rt < minTime {
				minTime = rt
			}
			if rt > maxTime {
				maxTime = rt
			}
		}

		result.AvgResponseTime = totalResponseTime / time.Duration(len(responseTimes))
		result.MinResponseTime = minTime
		result.MaxResponseTime = maxTime
	}

	result.RequestsPerSec = float64(result.TotalRequests) / totalTime.Seconds()
	result.ErrorRate = float64(errorCount) / float64(result.TotalRequests) * 100

	return result
}

// runUserRequests 运行单个用户的请求
func runUserRequests(config TestConfig, userID int, resultChan chan<- time.Duration, errorChan chan<- error) {
	client := &http.Client{Timeout: config.RequestTimeout}

	for i := 0; i < config.RequestsPerUser; i++ {
		// 构建测试请求
		request := StudentRequest{
			StudentID:          fmt.Sprintf("test_user_%d_%d", userID, i),
			Name:               fmt.Sprintf("测试学生%d", userID),
			TotalScore:         580 + (userID+i)%100,
			Ranking:            1000 + (userID+i)*10,
			Province:           "北京市",
			SubjectCombination: "物理+化学+生物",
			PreferredCities:    []string{"北京", "上海", "深圳"},
			PreferredMajors:    []string{"计算机科学与技术", "软件工程"},
			MaxVolunteers:      24,
		}

		requestBody, _ := json.Marshal(request)

		startTime := time.Now()
		resp, err := client.Post(
			config.BaseURL+"/api/v1/hybrid/plan",
			"application/json",
			bytes.NewBuffer(requestBody),
		)
		responseTime := time.Since(startTime)

		if err != nil {
			errorChan <- err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			errorChan <- fmt.Errorf("HTTP %d", resp.StatusCode)
			resp.Body.Close()
			continue
		}

		// 读取响应体（确保完整处理）
		_, err = io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			errorChan <- err
			continue
		}

		resultChan <- responseTime
	}
}

// printResults 打印测试结果
func printResults(result TestResult) {
	fmt.Printf("\n=== 性能测试结果 ===\n")
	fmt.Printf("总请求数: %d\n", result.TotalRequests)
	fmt.Printf("成功请求数: %d\n", result.SuccessRequests)
	fmt.Printf("失败请求数: %d\n", result.FailedRequests)
	fmt.Printf("总测试时间: %v\n", result.TotalTime)
	fmt.Printf("平均响应时间: %v\n", result.AvgResponseTime)
	fmt.Printf("最小响应时间: %v\n", result.MinResponseTime)
	fmt.Printf("最大响应时间: %v\n", result.MaxResponseTime)
	fmt.Printf("请求/秒: %.2f\n", result.RequestsPerSec)
	fmt.Printf("错误率: %.2f%%\n", result.ErrorRate)

	// 性能评估
	fmt.Printf("\n=== 性能评估 ===\n")
	if result.AvgResponseTime < 100*time.Millisecond {
		fmt.Println("✓ 响应时间: 优秀 (< 100ms)")
	} else if result.AvgResponseTime < 500*time.Millisecond {
		fmt.Println("✓ 响应时间: 良好 (< 500ms)")
	} else {
		fmt.Println("⚠ 响应时间: 需要优化 (> 500ms)")
	}

	if result.RequestsPerSec > 100 {
		fmt.Println("✓ 吞吐量: 优秀 (> 100 req/s)")
	} else if result.RequestsPerSec > 50 {
		fmt.Println("✓ 吞吐量: 良好 (> 50 req/s)")
	} else {
		fmt.Println("⚠ 吞吐量: 需要优化 (< 50 req/s)")
	}

	if result.ErrorRate < 1 {
		fmt.Println("✓ 错误率: 优秀 (< 1%)")
	} else if result.ErrorRate < 5 {
		fmt.Println("✓ 错误率: 可接受 (< 5%)")
	} else {
		fmt.Println("⚠ 错误率: 需要优化 (> 5%)")
	}
}

// testAPIEndpoints 测试不同的API端点
func testAPIEndpoints(baseURL string) {
	fmt.Printf("\n=== API端点测试 ===\n")

	endpoints := []struct {
		name   string
		method string
		path   string
		body   interface{}
	}{
		{"健康检查", "GET", "/health", nil},
		{"性能指标", "GET", "/api/v1/analytics/performance", nil},
		{"融合统计", "GET", "/api/v1/analytics/fusion-stats", nil},
		{"推荐趋势", "GET", "/api/v1/analytics/trends?time_range=1h", nil},
		{"质量报告", "POST", "/api/v1/analytics/quality-report", map[string]interface{}{
			"start_time": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			"end_time":   time.Now().Format(time.RFC3339),
		}},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	for _, endpoint := range endpoints {
		fmt.Printf("测试 %s (%s %s)... ", endpoint.name, endpoint.method, endpoint.path)

		var req *http.Request
		var err error

		if endpoint.body != nil {
			bodyBytes, _ := json.Marshal(endpoint.body)
			req, err = http.NewRequest(endpoint.method, baseURL+endpoint.path, bytes.NewBuffer(bodyBytes))
			if err == nil {
				req.Header.Set("Content-Type", "application/json")
			}
		} else {
			req, err = http.NewRequest(endpoint.method, baseURL+endpoint.path, nil)
		}

		if err != nil {
			fmt.Printf("❌ 创建请求失败: %v\n", err)
			continue
		}

		startTime := time.Now()
		resp, err := client.Do(req)
		responseTime := time.Since(startTime)

		if err != nil {
			fmt.Printf("❌ 请求失败: %v\n", err)
			continue
		}

		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			fmt.Printf("✓ 成功 (%d) - %v\n", resp.StatusCode, responseTime)
		} else {
			fmt.Printf("⚠ 状态码 %d - %v\n", resp.StatusCode, responseTime)
		}
	}
}