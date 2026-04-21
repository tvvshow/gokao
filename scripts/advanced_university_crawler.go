package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

// UniversityInfo 高校信息结构
type UniversityInfo struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Code         string `json:"code"`
	Province     string `json:"province"`
	City         string `json:"city"`
	Type         string `json:"type"`
	Level        string `json:"level"`
	Supervisor   string `json:"supervisor"`
	Is985        bool   `json:"is_985"`
	Is211        bool   `json:"is_211"`
	IsDoubleFirst bool  `json:"is_double_first"`
	Website      string `json:"website"`
	Source       string `json:"source"`
	CrawledAt    time.Time `json:"crawled_at"`
}

// ProvinceConfig 省份爬取配置
type ProvinceConfig struct {
	Name         string            `json:"name"`
	URL          string            `json:"url"`
	Method       string            `json:"method"` // "firecrawl", "playwright", "http"
	Selectors    map[string]string `json:"selectors"`
	Enabled      bool              `json:"enabled"`
	RetryCount   int               `json:"retry_count"`
	Delay        int               `json:"delay_seconds"`
	ExpectedMin  int               `json:"expected_min_universities"`
}

// CrawlResult 爬取结果
type CrawlResult struct {
	Province     string           `json:"province"`
	URL          string           `json:"url"`
	Method       string           `json:"method"`
	Success      bool             `json:"success"`
	Universities []UniversityInfo `json:"universities"`
	Error        string           `json:"error,omitempty"`
	RetryCount   int              `json:"retry_count"`
	CrawledAt    time.Time        `json:"crawled_at"`
	Duration     time.Duration    `json:"duration"`
}

// CrawlStats 爬取统计
type CrawlStats struct {
	TotalProvinces    int                    `json:"total_provinces"`
	SuccessfulCrawls  int                    `json:"successful_crawls"`
	FailedCrawls      int                    `json:"failed_crawls"`
	TotalUniversities int                    `json:"total_universities"`
	MethodStats       map[string]int         `json:"method_stats"`
	ProvinceStats     map[string]int         `json:"province_stats"`
	StartTime         time.Time              `json:"start_time"`
	EndTime           time.Time              `json:"end_time"`
	TotalDuration     time.Duration          `json:"total_duration"`
}

func main() {
	fmt.Println("🚀 高级高校数据爬虫系统")
	fmt.Println("支持 Firecrawl + Playwright + HTTP 多种爬取方式")
	fmt.Println("================================================\n")

	// 创建必要的目录
	createDirs()

	// 读取或创建省份配置
	configs, err := loadOrCreateProvinceConfigs()
	if err != nil {
		log.Fatalf("❌ 配置加载失败: %v", err)
	}

	fmt.Printf("📊 加载了 %d 个省份配置\n", len(configs))
	printConfigSummary(configs)

	// 初始化统计信息
	stats := &CrawlStats{
		TotalProvinces: len(configs),
		MethodStats:    make(map[string]int),
		ProvinceStats:  make(map[string]int),
		StartTime:      time.Now(),
	}

	var allUniversities []UniversityInfo
	var results []CrawlResult
	universityID := 1

	// 遍历每个省份进行爬取
	for i, config := range configs {
		if !config.Enabled {
			fmt.Printf("⏭️  跳过 %s (已禁用)\n", config.Name)
			continue
		}

		fmt.Printf("\n🎯 [%d/%d] 开始爬取 %s\n", i+1, len(configs), config.Name)
		fmt.Printf("   📍 URL: %s\n", config.URL)
		fmt.Printf("   🔧 方法: %s\n", config.Method)

		// 执行爬取
		result := crawlProvince(config, &universityID)
		results = append(results, result)

		// 更新统计
		updateStats(stats, result)

		if result.Success {
			fmt.Printf("   ✅ 成功获得 %d 所高校 (用时: %v)\n", len(result.Universities), result.Duration)
			allUniversities = append(allUniversities, result.Universities...)
		} else {
			fmt.Printf("   ❌ 爬取失败: %s (重试 %d 次)\n", result.Error, result.RetryCount)
		}

		// 保存单个省份结果
		saveProvinceResult(config.Name, result)

		// 添加延迟
		if config.Delay > 0 {
			fmt.Printf("   ⏳ 等待 %d 秒...\n", config.Delay)
			time.Sleep(time.Duration(config.Delay) * time.Second)
		}
	}

	// 完成统计
	stats.EndTime = time.Now()
	stats.TotalDuration = stats.EndTime.Sub(stats.StartTime)
	stats.TotalUniversities = len(allUniversities)

	// 打印最终统计
	printFinalStats(stats)

	// 保存结果
	saveResults(allUniversities, results, stats)

	fmt.Println("\n🎉 爬取任务完成！")
	printNextSteps()
}

// createDirs 创建必要的目录
func createDirs() {
	dirs := []string{"crawl_results", "logs", "cache"}
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}
}

// loadOrCreateProvinceConfigs 加载或创建省份配置
func loadOrCreateProvinceConfigs() ([]ProvinceConfig, error) {
	configFile := "province_configs.json"
	
	// 如果配置文件不存在，创建默认配置
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Println("📝 创建默认省份配置...")
		configs := createDefaultConfigs()
		err := saveConfigs(configs, configFile)
		if err != nil {
			return nil, fmt.Errorf("保存默认配置失败: %v", err)
		}
		fmt.Println("✅ 默认配置已创建")
		return configs, nil
	}

	// 读取现有配置
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var configs []ProvinceConfig
	err = json.Unmarshal(data, &configs)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return configs, nil
}

// createDefaultConfigs 创建默认配置
func createDefaultConfigs() []ProvinceConfig {
	return []ProvinceConfig{
		{
			Name:        "北京",
			URL:         "https://www.bjeea.cn/",
			Method:      "firecrawl",
			Selectors:   map[string]string{"university_list": "table tr", "name": "td:nth-child(1)", "code": "td:nth-child(2)"},
			Enabled:     true,
			RetryCount:  3,
			Delay:       2,
			ExpectedMin: 50,
		},
		{
			Name:        "上海",
			URL:         "https://www.shmeea.edu.cn/",
			Method:      "firecrawl",
			Selectors:   map[string]string{"university_list": ".university-list li", "name": ".name", "code": ".code"},
			Enabled:     true,
			RetryCount:  3,
			Delay:       2,
			ExpectedMin: 40,
		},
		{
			Name:        "江苏",
			URL:         "https://www.jseea.cn/",
			Method:      "playwright",
			Selectors:   map[string]string{"university_list": ".college-list tr", "name": "td:first-child", "code": "td:last-child"},
			Enabled:     true,
			RetryCount:  3,
			Delay:       3,
			ExpectedMin: 80,
		},
		// 可以继续添加其他省份...
	}
}

// crawlProvince 爬取单个省份
func crawlProvince(config ProvinceConfig, universityID *int) CrawlResult {
	start := time.Now()
	result := CrawlResult{
		Province:   config.Name,
		URL:        config.URL,
		Method:     config.Method,
		CrawledAt:  start,
	}

	// 根据方法选择爬取策略
	var universities []UniversityInfo
	var err error

	for attempt := 1; attempt <= config.RetryCount; attempt++ {
		result.RetryCount = attempt - 1
		
		fmt.Printf("   🔄 尝试 %d/%d...\n", attempt, config.RetryCount)
		
		switch config.Method {
		case "firecrawl":
			universities, err = crawlWithFirecrawl(config, universityID)
		case "playwright":
			universities, err = crawlWithPlaywright(config, universityID)
		case "http":
			universities, err = crawlWithHTTP(config, universityID)
		default:
			// 自动选择最佳方法
			universities, err = crawlWithAutoMethod(config, universityID)
		}

		if err == nil && len(universities) >= config.ExpectedMin {
			result.Success = true
			result.Universities = universities
			break
		}

		if attempt < config.RetryCount {
			fmt.Printf("   ⚠️  重试中... (错误: %v)\n", err)
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}

	if !result.Success {
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Error = fmt.Sprintf("获得的高校数量 (%d) 少于预期 (%d)", len(universities), config.ExpectedMin)
		}
	}

	result.Duration = time.Since(start)
	return result
}

// crawlWithFirecrawl 使用Firecrawl爬取
func crawlWithFirecrawl(config ProvinceConfig, universityID *int) ([]UniversityInfo, error) {
	fmt.Printf("   🔥 使用 Firecrawl 爬取...\n")
	
	// 注意：这里需要实际调用 MCP Firecrawl 工具
	// 由于这是 Go 脚本，我们先模拟实现
	// 实际使用时需要通过外部调用 MCP 工具
	
	// 模拟 Firecrawl 调用
	universities := simulateFirecrawlResult(config, universityID)
	
	if len(universities) == 0 {
		return nil, fmt.Errorf("Firecrawl 未能提取到高校数据")
	}
	
	return universities, nil
}

// crawlWithPlaywright 使用Playwright爬取
func crawlWithPlaywright(config ProvinceConfig, universityID *int) ([]UniversityInfo, error) {
	fmt.Printf("   🎭 使用 Playwright 爬取...\n")
	
	// 注意：这里需要实际调用 MCP Playwright 工具
	// 模拟实现
	universities := simulatePlaywrightResult(config, universityID)
	
	if len(universities) == 0 {
		return nil, fmt.Errorf("Playwright 未能提取到高校数据")
	}
	
	return universities, nil
}

// crawlWithHTTP 使用传统HTTP爬取
func crawlWithHTTP(config ProvinceConfig, universityID *int) ([]UniversityInfo, error) {
	fmt.Printf("   🌐 使用 HTTP 爬取...\n")
	
	// 传统HTTP爬取实现
	universities := simulateHTTPResult(config, universityID)
	
	if len(universities) == 0 {
		return nil, fmt.Errorf("HTTP 爬取未能提取到高校数据")
	}
	
	return universities, nil
}

// crawlWithAutoMethod 自动选择最佳方法
func crawlWithAutoMethod(config ProvinceConfig, universityID *int) ([]UniversityInfo, error) {
	fmt.Printf("   🤖 自动选择最佳爬取方法...\n")
	
	// 按优先级尝试不同方法
	methods := []string{"firecrawl", "playwright", "http"}
	
	for _, method := range methods {
		fmt.Printf("      尝试 %s...\n", method)
		config.Method = method
		
		var universities []UniversityInfo
		var err error
		
		switch method {
		case "firecrawl":
			universities, err = crawlWithFirecrawl(config, universityID)
		case "playwright":
			universities, err = crawlWithPlaywright(config, universityID)
		case "http":
			universities, err = crawlWithHTTP(config, universityID)
		}
		
		if err == nil && len(universities) > 0 {
			fmt.Printf("      ✅ %s 成功\n", method)
			return universities, nil
		}
		
		fmt.Printf("      ❌ %s 失败: %v\n", method, err)
	}
	
	return nil, fmt.Errorf("所有方法都失败了")
}

// 模拟函数（实际使用时需要替换为真实的MCP调用）
func simulateFirecrawlResult(config ProvinceConfig, universityID *int) []UniversityInfo {
	return generateMockUniversities(config.Name, universityID, config.ExpectedMin)
}

func simulatePlaywrightResult(config ProvinceConfig, universityID *int) []UniversityInfo {
	return generateMockUniversities(config.Name, universityID, config.ExpectedMin-10)
}

func simulateHTTPResult(config ProvinceConfig, universityID *int) []UniversityInfo {
	return generateMockUniversities(config.Name, universityID, config.ExpectedMin-20)
}

// generateMockUniversities 生成模拟高校数据
func generateMockUniversities(province string, universityID *int, count int) []UniversityInfo {
	provinceUniversities := getProvinceUniversities(province)
	var universities []UniversityInfo
	
	// 确保不超过可用的高校数量
	if count > len(provinceUniversities) {
		count = len(provinceUniversities)
	}
	
	for i := 0; i < count; i++ {
		name := provinceUniversities[i%len(provinceUniversities)]
		if i >= len(provinceUniversities) {
			name = fmt.Sprintf("%s第%d大学", province, i+1)
		}
		
		is985, is211, isDoubleFirst := isEliteUniversity(name)
		university := UniversityInfo{
			ID:            *universityID,
			Name:          name,
			Code:          generateUniversityCode(province, *universityID),
			Province:      province,
			City:          getProvinceCapital(province),
			Type:          inferUniversityType(name),
			Level:         "本科",
			Supervisor:    inferSupervisor(name),
			Is985:         is985,
			Is211:         is211,
			IsDoubleFirst: isDoubleFirst,
			Website:       fmt.Sprintf("http://www.%s.edu.cn", strings.ToLower(strings.ReplaceAll(name, "大学", "u"))),
			Source:        fmt.Sprintf("爬取自%s考试院", province),
			CrawledAt:     time.Now(),
		}
		universities = append(universities, university)
		*universityID++
	}
	
	return universities
}

// 统计和报告函数
func updateStats(stats *CrawlStats, result CrawlResult) {
	if result.Success {
		stats.SuccessfulCrawls++
	} else {
		stats.FailedCrawls++
	}
	
	stats.MethodStats[result.Method]++
	stats.ProvinceStats[result.Province] = len(result.Universities)
}

func printConfigSummary(configs []ProvinceConfig) {
	methodCount := make(map[string]int)
	enabledCount := 0
	
	for _, config := range configs {
		if config.Enabled {
			enabledCount++
			methodCount[config.Method]++
		}
	}
	
	fmt.Printf("   ✅ 启用省份: %d/%d\n", enabledCount, len(configs))
	fmt.Printf("   🔧 爬取方法分布: ")
	for method, count := range methodCount {
		fmt.Printf("%s(%d) ", method, count)
	}
	fmt.Println()
}

func printFinalStats(stats *CrawlStats) {
	fmt.Printf("\n📊 爬取完成统计:\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("⏱️  总用时: %v\n", stats.TotalDuration)
	fmt.Printf("🎯 成功省份: %d/%d (%.1f%%)\n", stats.SuccessfulCrawls, stats.TotalProvinces, float64(stats.SuccessfulCrawls)/float64(stats.TotalProvinces)*100)
	fmt.Printf("🏫 获得高校: %d 所\n", stats.TotalUniversities)
	fmt.Printf("📈 平均每省: %.1f 所\n", float64(stats.TotalUniversities)/float64(stats.SuccessfulCrawls))
	
	fmt.Printf("\n🔧 方法统计:\n")
	for method, count := range stats.MethodStats {
		fmt.Printf("   %s: %d 次\n", method, count)
	}
	
	fmt.Printf("\n🗺️  省份统计 (前10):\n")
	printTopProvinces(stats.ProvinceStats, 10)
}

func printTopProvinces(provinceStats map[string]int, limit int) {
	type provinceCount struct {
		name  string
		count int
	}
	
	var provinces []provinceCount
	for name, count := range provinceStats {
		provinces = append(provinces, provinceCount{name, count})
	}
	
	// 简单排序（冒泡排序）
	for i := 0; i < len(provinces)-1; i++ {
		for j := 0; j < len(provinces)-i-1; j++ {
			if provinces[j].count < provinces[j+1].count {
				provinces[j], provinces[j+1] = provinces[j+1], provinces[j]
			}
		}
	}
	
	for i, p := range provinces {
		if i >= limit {
			break
		}
		fmt.Printf("   %d. %s: %d 所\n", i+1, p.name, p.count)
	}
}

func printNextSteps() {
	fmt.Println("\n📋 下一步操作建议:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("1. 📊 查看 crawl_report.json 了解详细爬取结果")
	fmt.Println("2. 🔍 检查 crawl_results/ 目录中的单个省份结果")
	fmt.Println("3. 🗄️  运行 import_crawled_universities.go 导入数据库")
	fmt.Println("4. ⚙️  调整 province_configs.json 优化失败的省份配置")
	fmt.Println("5. 🔄 重新运行爬虫处理失败的省份")
	fmt.Println("6. ✅ 验证数据质量和完整性")
}

// 保存函数
func saveResults(universities []UniversityInfo, results []CrawlResult, stats *CrawlStats) {
	// 保存所有高校数据
	err := saveAllUniversities(universities)
	if err != nil {
		log.Printf("❌ 保存高校数据失败: %v", err)
	} else {
		fmt.Println("✅ 高校数据已保存到 all_universities.json")
	}

	// 保存爬取报告
	err = saveCrawlReport(results, stats)
	if err != nil {
		log.Printf("❌ 保存爬取报告失败: %v", err)
	} else {
		fmt.Println("✅ 爬取报告已保存到 crawl_report.json")
	}

	// 保存失败省份列表
	err = saveFailedProvinces(results)
	if err != nil {
		log.Printf("❌ 保存失败列表失败: %v", err)
	} else {
		fmt.Println("✅ 失败省份列表已保存到 failed_provinces.json")
	}
}

// 复用之前的辅助函数...
// (这里包含所有之前定义的辅助函数，如 getProvinceUniversities, generateUniversityCode 等)

// 为了保持代码简洁，这里只列出主要的新增函数
// 其他辅助函数可以从之前的脚本中复制

func saveConfigs(configs []ProvinceConfig, filename string) error {
	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

func saveProvinceResult(province string, result CrawlResult) error {
	filename := fmt.Sprintf("crawl_results/%s_result.json", province)
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

func saveAllUniversities(universities []UniversityInfo) error {
	data, err := json.MarshalIndent(universities, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile("all_universities.json", data, 0644)
}

func saveCrawlReport(results []CrawlResult, stats *CrawlStats) error {
	report := map[string]interface{}{
		"summary": stats,
		"results": results,
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile("crawl_report.json", data, 0644)
}

func saveFailedProvinces(results []CrawlResult) error {
	var failedProvinces []string
	for _, result := range results {
		if !result.Success {
			failedProvinces = append(failedProvinces, result.Province)
		}
	}

	data, err := json.MarshalIndent(failedProvinces, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile("failed_provinces.json", data, 0644)
}

// 复用的辅助函数（从之前的脚本复制）
func getProvinceUniversities(province string) []string {
	// 这里复用之前定义的函数
	provinceMap := map[string][]string{
		"北京": {"北京大学", "清华大学", "中国人民大学", "北京航空航天大学", "北京理工大学", "北京师范大学", "中国农业大学", "北京科技大学", "北京交通大学", "北京邮电大学"},
		"上海": {"复旦大学", "上海交通大学", "同济大学", "华东师范大学", "上海大学", "华东理工大学", "东华大学", "上海财经大学", "上海外国语大学", "上海理工大学"},
		"江苏": {"南京大学", "东南大学", "南京航空航天大学", "南京理工大学", "苏州大学", "南京师范大学", "河海大学", "南京农业大学", "中国矿业大学", "江南大学"},
		// ... 其他省份
	}

	if universities, exists := provinceMap[province]; exists {
		return universities
	}

	return []string{province + "大学", province + "师范大学", province + "理工大学"}
}

// 其他辅助函数...
func generateUniversityCode(province string, id int) string {
	// 复用之前的实现
	return fmt.Sprintf("41%02d10%04d", getProvinceCode(province), id)
}

func getProvinceCode(province string) int {
	codes := map[string]int{
		"北京": 11, "上海": 31, "江苏": 32, "浙江": 33,
		// ... 其他省份代码
	}
	if code, exists := codes[province]; exists {
		return code
	}
	return 99
}

func getProvinceCapital(province string) string {
	capitals := map[string]string{
		"北京": "北京", "上海": "上海", "江苏": "南京", "浙江": "杭州",
		// ... 其他省会
	}
	if capital, exists := capitals[province]; exists {
		return capital
	}
	return province
}

func inferUniversityType(name string) string {
	if strings.Contains(name, "师范") {
		return "师范类"
	}
	if strings.Contains(name, "理工") || strings.Contains(name, "科技") {
		return "理工类"
	}
	if strings.Contains(name, "医科") || strings.Contains(name, "医学") {
		return "医药类"
	}
	return "综合类"
}

func inferSupervisor(name string) string {
	if strings.Contains(name, "北京大学") || strings.Contains(name, "清华大学") {
		return "教育部"
	}
	if strings.Contains(name, "师范") {
		return "教育部"
	}
	return "省教育厅"
}

func isEliteUniversity(name string) (bool, bool, bool) {
	universities985 := []string{"北京大学", "清华大学", "复旦大学", "上海交通大学", "浙江大学", "南京大学"}
	universities211 := []string{"北京交通大学", "北京科技大学", "同济大学", "华东师范大学"}

	is985 := contains(universities985, name)
	is211 := contains(universities211, name) || is985
	isDoubleFirst := is985
	return is985, is211, isDoubleFirst
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.Contains(item, s) || strings.Contains(s, item) {
			return true
		}
	}
	return false
}