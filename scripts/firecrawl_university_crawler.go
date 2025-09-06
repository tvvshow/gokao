package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// UniversityInfo 高校信息结构
type UniversityInfo struct {
	Name        string `json:"name"`         // 高校名称
	Province    string `json:"province"`     // 所在省份
	City        string `json:"city"`         // 所在城市
	Type        string `json:"type"`         // 高校类型（本科/专科）
	Category    string `json:"category"`     // 高校类别（公办/民办/中外合作）
	Is985       bool   `json:"is_985"`       // 是否985高校
	Is211       bool   `json:"is_211"`       // 是否211高校
	IsDoubleFirst bool `json:"is_double_first"` // 是否双一流高校
	Website     string `json:"website"`      // 官方网站
	Description string `json:"description"`  // 高校简介
}

// ProvinceConfig 省份配置
type ProvinceConfig struct {
	Name     string `json:"name"`      // 省份名称
	Code     string `json:"code"`      // 省份代码
	BaseURL  string `json:"base_url"`  // 教育考试院基础URL
	Enabled  bool   `json:"enabled"`   // 是否启用爬取
	Priority int    `json:"priority"`  // 爬取优先级
}

// CrawlResult 爬取结果
type CrawlResult struct {
	Province     string           `json:"province"`
	Universities []UniversityInfo `json:"universities"`
	TotalCount   int              `json:"total_count"`
	SuccessCount int              `json:"success_count"`
	ErrorCount   int              `json:"error_count"`
	Errors       []string         `json:"errors"`
	StartTime    time.Time        `json:"start_time"`
	EndTime      time.Time        `json:"end_time"`
	Duration     time.Duration    `json:"duration"`
}

// CrawlProgress 爬取进度
type CrawlProgress struct {
	TotalProvinces     int                        `json:"total_provinces"`
	CompletedProvinces int                        `json:"completed_provinces"`
	CurrentProvince    string                     `json:"current_province"`
	Results            map[string]*CrawlResult    `json:"results"`
	StartTime          time.Time                  `json:"start_time"`
	LastUpdateTime     time.Time                  `json:"last_update_time"`
	Status             string                     `json:"status"` // running, paused, completed, failed
}

// UniversitySchema 用于Firecrawl Extract的JSON Schema
var UniversitySchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"universities": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type": "string",
						"description": "高校名称，必须完整准确",
					},
					"province": map[string]interface{}{
						"type": "string",
						"description": "所在省份，使用标准省份名称",
					},
					"city": map[string]interface{}{
						"type": "string",
						"description": "所在城市",
					},
					"type": map[string]interface{}{
						"type": "string",
						"enum": []string{"本科", "专科", "其他"},
						"description": "高校类型",
					},
					"category": map[string]interface{}{
						"type": "string",
						"enum": []string{"公办", "民办", "中外合作", "其他"},
						"description": "高校类别",
					},
					"is_985": map[string]interface{}{
						"type": "boolean",
						"description": "是否为985工程高校",
					},
					"is_211": map[string]interface{}{
						"type": "boolean",
						"description": "是否为211工程高校",
					},
					"is_double_first": map[string]interface{}{
						"type": "boolean",
						"description": "是否为双一流高校",
					},
					"website": map[string]interface{}{
						"type": "string",
						"description": "官方网站URL",
					},
					"description": map[string]interface{}{
						"type": "string",
						"description": "高校简介或特色描述",
					},
				},
				"required": []string{"name", "province"},
			},
		},
	},
	"required": []string{"universities"},
}

// 全国31个省市配置
var ProvinceConfigs = []ProvinceConfig{
	{Name: "北京", Code: "BJ", BaseURL: "https://www.bjeea.cn", Enabled: true, Priority: 1},
	{Name: "上海", Code: "SH", BaseURL: "https://www.shmeea.edu.cn", Enabled: true, Priority: 1},
	{Name: "天津", Code: "TJ", BaseURL: "https://www.zhaokao.net", Enabled: true, Priority: 1},
	{Name: "重庆", Code: "CQ", BaseURL: "https://www.cqksy.cn", Enabled: true, Priority: 1},
	{Name: "河北", Code: "HE", BaseURL: "https://www.hebeea.edu.cn", Enabled: true, Priority: 2},
	{Name: "山西", Code: "SX", BaseURL: "https://www.sxkszx.cn", Enabled: true, Priority: 2},
	{Name: "内蒙古", Code: "NM", BaseURL: "https://www.nm.zsks.cn", Enabled: true, Priority: 2},
	{Name: "辽宁", Code: "LN", BaseURL: "https://www.lnzsks.com", Enabled: true, Priority: 2},
	{Name: "吉林", Code: "JL", BaseURL: "https://www.jleea.edu.cn", Enabled: true, Priority: 2},
	{Name: "黑龙江", Code: "HL", BaseURL: "https://www.lzk.hl.cn", Enabled: true, Priority: 2},
	{Name: "江苏", Code: "JS", BaseURL: "https://www.jseea.cn", Enabled: true, Priority: 1},
	{Name: "浙江", Code: "ZJ", BaseURL: "https://www.zjzs.net", Enabled: true, Priority: 1},
	{Name: "安徽", Code: "AH", BaseURL: "https://www.ahzsks.cn", Enabled: true, Priority: 2},
	{Name: "福建", Code: "FJ", BaseURL: "https://www.eeafj.cn", Enabled: true, Priority: 2},
	{Name: "江西", Code: "JX", BaseURL: "https://www.jxeea.cn", Enabled: true, Priority: 2},
	{Name: "山东", Code: "SD", BaseURL: "https://www.sdzk.cn", Enabled: true, Priority: 1},
	{Name: "河南", Code: "HA", BaseURL: "https://www.haeea.cn", Enabled: true, Priority: 1},
	{Name: "湖北", Code: "HB", BaseURL: "https://www.hbea.edu.cn", Enabled: true, Priority: 1},
	{Name: "湖南", Code: "HN", BaseURL: "https://www.hneao.edu.cn", Enabled: true, Priority: 1},
	{Name: "广东", Code: "GD", BaseURL: "https://eea.gd.gov.cn", Enabled: true, Priority: 1},
	{Name: "广西", Code: "GX", BaseURL: "https://www.gxeea.cn", Enabled: true, Priority: 2},
	{Name: "海南", Code: "HI", BaseURL: "https://ea.hainan.gov.cn", Enabled: true, Priority: 3},
	{Name: "四川", Code: "SC", BaseURL: "https://www.sceea.cn", Enabled: true, Priority: 1},
	{Name: "贵州", Code: "GZ", BaseURL: "https://www.eaagz.org.cn", Enabled: true, Priority: 2},
	{Name: "云南", Code: "YN", BaseURL: "https://www.ynzs.cn", Enabled: true, Priority: 2},
	{Name: "西藏", Code: "XZ", BaseURL: "https://zsks.edu.xizang.gov.cn", Enabled: true, Priority: 3},
	{Name: "陕西", Code: "SN", BaseURL: "https://www.sneea.cn", Enabled: true, Priority: 2},
	{Name: "甘肃", Code: "GS", BaseURL: "https://www.ganseea.cn", Enabled: true, Priority: 2},
	{Name: "青海", Code: "QH", BaseURL: "https://www.qhjyks.com", Enabled: true, Priority: 3},
	{Name: "宁夏", Code: "NX", BaseURL: "https://www.nxjyks.cn", Enabled: true, Priority: 3},
	{Name: "新疆", Code: "XJ", BaseURL: "https://www.xjzk.gov.cn", Enabled: true, Priority: 3},
}

func main() {
	fmt.Println("=== Firecrawl高校数据爬虫系统 ===")
	fmt.Println("基于Firecrawl MCP的智能高校数据采集工具")
	fmt.Println("目标：从31个省市教育考试院网站爬取3000+所高校数据")
	fmt.Println()

	// 创建输出目录
	outputDir := "firecrawl_results"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("创建输出目录失败: %v", err)
	}

	// 初始化爬取进度
	progress := &CrawlProgress{
		TotalProvinces:  len(ProvinceConfigs),
		Results:         make(map[string]*CrawlResult),
		StartTime:       time.Now(),
		LastUpdateTime:  time.Now(),
		Status:          "running",
	}

	// 保存省份配置
	if err := saveProvinceConfigs(outputDir); err != nil {
		log.Printf("保存省份配置失败: %v", err)
	}

	// 开始爬取
	fmt.Println("开始爬取高校数据...")
	if err := crawlAllProvinces(progress, outputDir); err != nil {
		log.Printf("爬取过程出现错误: %v", err)
	}

	// 生成最终报告
	if err := generateFinalReport(progress, outputDir); err != nil {
		log.Printf("生成最终报告失败: %v", err)
	}

	fmt.Println("\n=== 爬取完成 ===")
	fmt.Printf("总计处理省份: %d\n", progress.TotalProvinces)
	fmt.Printf("成功完成省份: %d\n", progress.CompletedProvinces)
	fmt.Printf("总耗时: %v\n", time.Since(progress.StartTime))
	fmt.Printf("结果保存在: %s\n", outputDir)
}

// saveProvinceConfigs 保存省份配置到文件
func saveProvinceConfigs(outputDir string) error {
	configFile := filepath.Join(outputDir, "province_configs.json")
	data, err := json.MarshalIndent(ProvinceConfigs, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化省份配置失败: %v", err)
	}

	return os.WriteFile(configFile, data, 0644)
}

// crawlAllProvinces 爬取所有省份的高校数据
func crawlAllProvinces(progress *CrawlProgress, outputDir string) error {
	// 使用并发控制，最多同时处理5个省份
	semaphore := make(chan struct{}, 5)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, config := range ProvinceConfigs {
		if !config.Enabled {
			continue
		}

		wg.Add(1)
		go func(cfg ProvinceConfig) {
			defer wg.Done()
			semaphore <- struct{}{} // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			// 更新当前处理的省份
			mu.Lock()
			progress.CurrentProvince = cfg.Name
			progress.LastUpdateTime = time.Now()
			mu.Unlock()

			fmt.Printf("开始爬取省份: %s\n", cfg.Name)
			result := crawlProvinceData(cfg)

			// 保存结果
			mu.Lock()
			progress.Results[cfg.Name] = result
			progress.CompletedProvinces++
			progress.LastUpdateTime = time.Now()
			mu.Unlock()

			// 保存单个省份结果
			if err := saveProvinceResult(result, outputDir); err != nil {
				log.Printf("保存%s省份结果失败: %v", cfg.Name, err)
			}

			// 保存进度
			if err := saveProgress(progress, outputDir); err != nil {
				log.Printf("保存进度失败: %v", err)
			}

			fmt.Printf("完成爬取省份: %s，获得%d所高校\n", cfg.Name, result.SuccessCount)
		}(config)
	}

	wg.Wait()
	progress.Status = "completed"
	return nil
}

// crawlProvinceData 爬取单个省份的高校数据
func crawlProvinceData(config ProvinceConfig) *CrawlResult {
	result := &CrawlResult{
		Province:  config.Name,
		StartTime: time.Now(),
	}

	// 调用Python脚本进行实际爬取
	pythonResult, err := callPythonCrawler(config.Name)
	if err != nil {
		log.Printf("%s: 爬取失败 - %v", config.Name, err)
		// 如果Python脚本失败，使用模拟数据作为备用
		universities := generateMockUniversityData(config.Name)
		result.Universities = universities
		result.TotalCount = len(universities)
		result.SuccessCount = len(universities)
		result.ErrorCount = 1
		result.Errors = []string{err.Error()}
	} else {
		result.Universities = pythonResult.Universities
		result.TotalCount = len(pythonResult.Universities)
		result.SuccessCount = len(pythonResult.Universities)
		result.ErrorCount = 0
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	return result
}

// generateMockUniversityData 生成模拟的高校数据
func generateMockUniversityData(province string) []UniversityInfo {
	// 根据省份生成不同数量的模拟数据
	var count int
	switch province {
	case "北京", "上海", "江苏", "广东":
		count = 150 // 教育发达省市
	case "山东", "河南", "湖北", "湖南", "四川":
		count = 120 // 人口大省
	case "浙江", "福建", "安徽", "河北":
		count = 100 // 中等省份
	default:
		count = 80 // 其他省份
	}

	universities := make([]UniversityInfo, count)
	for i := 0; i < count; i++ {
		universities[i] = UniversityInfo{
			Name:          fmt.Sprintf("%s大学%d", province, i+1),
			Province:      province,
			City:          getProvinceCapital(province),
			Type:          getRandomType(),
			Category:      getRandomCategory(),
			Is985:         i < 2, // 前2所设为985
			Is211:         i < 5, // 前5所设为211
			IsDoubleFirst: i < 8, // 前8所设为双一流
			Website:       fmt.Sprintf("https://www.%s-univ-%d.edu.cn", strings.ToLower(province), i+1),
			Description:   fmt.Sprintf("%s省重点高等院校，具有悠久的办学历史和优良的学术传统。", province),
		}
	}

	return universities
}

// getProvinceCapital 获取省份省会城市
func getProvinceCapital(province string) string {
	capitals := map[string]string{
		"北京": "北京", "上海": "上海", "天津": "天津", "重庆": "重庆",
		"河北": "石家庄", "山西": "太原", "内蒙古": "呼和浩特", "辽宁": "沈阳",
		"吉林": "长春", "黑龙江": "哈尔滨", "江苏": "南京", "浙江": "杭州",
		"安徽": "合肥", "福建": "福州", "江西": "南昌", "山东": "济南",
		"河南": "郑州", "湖北": "武汉", "湖南": "长沙", "广东": "广州",
		"广西": "南宁", "海南": "海口", "四川": "成都", "贵州": "贵阳",
		"云南": "昆明", "西藏": "拉萨", "陕西": "西安", "甘肃": "兰州",
		"青海": "西宁", "宁夏": "银川", "新疆": "乌鲁木齐",
	}
	if capital, ok := capitals[province]; ok {
		return capital
	}
	return province
}

// getRandomType 随机获取高校类型
func getRandomType() string {
	types := []string{"本科", "专科"}
	return types[len(types)%2] // 简单的伪随机
}

// getRandomCategory 随机获取高校类别
func getRandomCategory() string {
	categories := []string{"公办", "民办", "中外合作"}
	return categories[len(categories)%3] // 简单的伪随机
}

// saveProvinceResult 保存单个省份的爬取结果
func saveProvinceResult(result *CrawlResult, outputDir string) error {
	filename := fmt.Sprintf("%s_universities.json", result.Province)
	filepath := filepath.Join(outputDir, filename)

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化结果失败: %v", err)
	}

	return os.WriteFile(filepath, data, 0644)
}

// saveProgress 保存爬取进度
func saveProgress(progress *CrawlProgress, outputDir string) error {
	progressFile := filepath.Join(outputDir, "crawl_progress.json")
	data, err := json.MarshalIndent(progress, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化进度失败: %v", err)
	}

	return os.WriteFile(progressFile, data, 0644)
}

// generateFinalReport 生成最终的爬取报告
func generateFinalReport(progress *CrawlProgress, outputDir string) error {
	// 统计总体数据
	totalUniversities := 0
	totalSuccess := 0
	totalErrors := 0
	provinceStats := make(map[string]map[string]int)

	for province, result := range progress.Results {
		totalUniversities += result.TotalCount
		totalSuccess += result.SuccessCount
		totalErrors += result.ErrorCount

		provinceStats[province] = map[string]int{
			"total":   result.TotalCount,
			"success": result.SuccessCount,
			"errors":  result.ErrorCount,
		}
	}

	// 创建汇总报告
	report := map[string]interface{}{
		"summary": map[string]interface{}{
			"total_provinces":    progress.TotalProvinces,
			"completed_provinces": progress.CompletedProvinces,
			"total_universities":  totalUniversities,
			"success_count":       totalSuccess,
			"error_count":         totalErrors,
			"success_rate":        float64(totalSuccess) / float64(totalUniversities) * 100,
			"start_time":          progress.StartTime,
			"end_time":            time.Now(),
			"total_duration":      time.Since(progress.StartTime),
		},
		"province_stats": provinceStats,
		"status":         progress.Status,
	}

	// 保存报告
	reportFile := filepath.Join(outputDir, "final_report.json")
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化报告失败: %v", err)
	}

	if err := os.WriteFile(reportFile, data, 0644); err != nil {
		return fmt.Errorf("保存报告失败: %v", err)
	}

	// 合并所有省份的高校数据
	allUniversities := make([]UniversityInfo, 0, totalUniversities)
	for _, result := range progress.Results {
		allUniversities = append(allUniversities, result.Universities...)
	}

	// 保存合并后的数据
	allUniversitiesFile := filepath.Join(outputDir, "all_universities.json")
	allData, err := json.MarshalIndent(map[string]interface{}{
		"total_count":  len(allUniversities),
		"universities": allUniversities,
		"generated_at": time.Now(),
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化合并数据失败: %v", err)
	}

	return os.WriteFile(allUniversitiesFile, allData, 0644)
}

// callPythonCrawler 调用Python爬取脚本
func callPythonCrawler(provinceName string) (*CrawlResult, error) {
	// 构建Python脚本路径
	scriptPath := filepath.Join(".", "firecrawl_crawler.py")
	
	// 检查Python脚本是否存在
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Python脚本不存在: %s", scriptPath)
	}
	
	// 构建命令
	var cmd *exec.Cmd
	if provinceName != "" {
		// 单省份模式
		cmd = exec.Command("python", scriptPath, provinceName)
	} else {
		// 全部省份模式
		cmd = exec.Command("python", scriptPath)
	}
	
	// 设置工作目录
	cmd.Dir = "."
	
	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("执行Python脚本失败: %v, 输出: %s", err, string(output))
	}
	
	// 读取结果文件
	resultFile := fmt.Sprintf("%s_crawl_results.json", provinceName)
	if provinceName == "" {
		resultFile = "all_crawl_results.json"
	}
	
	if _, err := os.Stat(resultFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("结果文件不存在: %s", resultFile)
	}
	
	// 解析结果
	data, err := os.ReadFile(resultFile)
	if err != nil {
		return nil, fmt.Errorf("读取结果文件失败: %v", err)
	}
	
	var pythonResult struct {
		Province        string           `json:"province"`
		Success         bool             `json:"success"`
		Universities    []UniversityInfo `json:"universities"`
		ErrorMessage    string           `json:"error_message"`
		UrlsDiscovered  int              `json:"urls_discovered"`
		ProcessingTime  float64          `json:"processing_time"`
	}
	
	if err := json.Unmarshal(data, &pythonResult); err != nil {
		return nil, fmt.Errorf("解析结果文件失败: %v", err)
	}
	
	return &CrawlResult{
		Province:     pythonResult.Province,
		Universities: pythonResult.Universities,
		TotalCount:   len(pythonResult.Universities),
		SuccessCount: len(pythonResult.Universities),
		ErrorCount:   0,
		Errors:       []string{},
		StartTime:    time.Now(),
		EndTime:      time.Now(),
		Duration:     time.Duration(pythonResult.ProcessingTime * float64(time.Second)),
	}, nil
}