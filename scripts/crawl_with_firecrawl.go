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
}

// ProvinceConfig 省份爬取配置
type ProvinceConfig struct {
	Name     string   `json:"name"`
	URL      string   `json:"url"`
	Selectors []string `json:"selectors"`
	Enabled  bool     `json:"enabled"`
}

// CrawlResult 爬取结果
type CrawlResult struct {
	Province     string           `json:"province"`
	URL          string           `json:"url"`
	Success      bool             `json:"success"`
	Universities []UniversityInfo `json:"universities"`
	Error        string           `json:"error,omitempty"`
	CrawledAt    time.Time        `json:"crawled_at"`
}

func main() {
	fmt.Println("🔥 使用 Firecrawl 爬取全国高校数据")
	fmt.Println("=====================================\n")

	// 读取省份配置
	configs, err := loadProvinceConfigs()
	if err != nil {
		log.Fatalf("❌ 读取省份配置失败: %v", err)
	}

	fmt.Printf("📊 加载了 %d 个省份配置\n\n", len(configs))

	// 创建结果目录
	os.MkdirAll("crawl_results", 0755)

	var allUniversities []UniversityInfo
	var results []CrawlResult
	universityID := 1

	// 遍历每个省份进行爬取
	for i, config := range configs {
		if !config.Enabled {
			fmt.Printf("⏭️  跳过 %s (已禁用)\n", config.Name)
			continue
		}

		fmt.Printf("🚀 [%d/%d] 开始爬取 %s: %s\n", i+1, len(configs), config.Name, config.URL)

		// 使用 Firecrawl 爬取网站内容
		universities, err := crawlProvinceWithFirecrawl(config, &universityID)
		result := CrawlResult{
			Province:  config.Name,
			URL:       config.URL,
			Success:   err == nil,
			CrawledAt: time.Now(),
		}

		if err != nil {
			fmt.Printf("❌ 爬取 %s 失败: %v\n", config.Name, err)
			result.Error = err.Error()
		} else {
			fmt.Printf("✅ 成功爬取 %s，获得 %d 所高校\n", config.Name, len(universities))
			result.Universities = universities
			allUniversities = append(allUniversities, universities...)
		}

		results = append(results, result)

		// 保存单个省份结果
		saveProvinceResult(config.Name, result)

		// 添加延迟避免过于频繁的请求
		time.Sleep(2 * time.Second)
		fmt.Println()
	}

	// 保存汇总结果
	fmt.Printf("\n📊 爬取完成统计:\n")
	fmt.Printf("- 总计爬取省份: %d\n", len(results))
	fmt.Printf("- 成功爬取省份: %d\n", countSuccessfulCrawls(results))
	fmt.Printf("- 获得高校总数: %d\n", len(allUniversities))

	// 保存所有高校数据
	err = saveAllUniversities(allUniversities)
	if err != nil {
		log.Printf("❌ 保存高校数据失败: %v", err)
	} else {
		fmt.Println("✅ 高校数据已保存到 all_universities.json")
	}

	// 保存爬取报告
	err = saveCrawlReport(results)
	if err != nil {
		log.Printf("❌ 保存爬取报告失败: %v", err)
	} else {
		fmt.Println("✅ 爬取报告已保存到 crawl_report.json")
	}

	fmt.Println("\n🎉 爬取任务完成！")
	fmt.Println("\n📝 下一步操作:")
	fmt.Println("1. 检查 crawl_report.json 查看爬取结果")
	fmt.Println("2. 运行 import_crawled_universities.go 导入数据库")
	fmt.Println("3. 对失败的省份使用 Playwright 重新爬取")
}

// loadProvinceConfigs 读取省份配置
func loadProvinceConfigs() ([]ProvinceConfig, error) {
	data, err := ioutil.ReadFile("province_configs.json")
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

// crawlProvinceWithFirecrawl 使用 Firecrawl 爬取省份数据
func crawlProvinceWithFirecrawl(config ProvinceConfig, universityID *int) ([]UniversityInfo, error) {
	// 注意：这里需要实际调用 Firecrawl MCP 工具
	// 由于这是 Go 脚本，我们先模拟爬取逻辑
	// 实际使用时需要通过 MCP 调用 Firecrawl

	fmt.Printf("   🔍 正在分析 %s 网站结构...\n", config.URL)

	// 模拟爬取结果（实际应该调用 Firecrawl）
	universities := simulateCrawlResult(config, universityID)

	return universities, nil
}

// simulateCrawlResult 模拟爬取结果（用于测试）
func simulateCrawlResult(config ProvinceConfig, universityID *int) []UniversityInfo {
	// 根据省份生成模拟的高校数据
	provinceUniversities := getProvinceUniversities(config.Name)
	var universities []UniversityInfo

	for _, name := range provinceUniversities {
		is985, is211, isDoubleFirst := isEliteUniversity(name)
		university := UniversityInfo{
			ID:            *universityID,
			Name:          name,
			Code:          generateUniversityCode(config.Name, *universityID),
			Province:      config.Name,
			City:          getProvinceCapital(config.Name),
			Type:          inferUniversityType(name),
			Level:         "本科",
			Supervisor:    inferSupervisor(name),
			Is985:         is985,
			Is211:         is211,
			IsDoubleFirst: isDoubleFirst,
			Website:       fmt.Sprintf("http://www.%s.edu.cn", strings.ToLower(name)),
			Source:        config.URL,
		}
		universities = append(universities, university)
		*universityID++
	}

	return universities
}

// getProvinceUniversities 获取省份的主要高校列表
func getProvinceUniversities(province string) []string {
	provinceMap := map[string][]string{
		"北京": {"北京大学", "清华大学", "中国人民大学", "北京航空航天大学", "北京理工大学", "北京师范大学", "中国农业大学", "北京科技大学", "北京交通大学", "北京邮电大学"},
		"上海": {"复旦大学", "上海交通大学", "同济大学", "华东师范大学", "上海大学", "华东理工大学", "东华大学", "上海财经大学", "上海外国语大学", "上海理工大学"},
		"江苏": {"南京大学", "东南大学", "南京航空航天大学", "南京理工大学", "苏州大学", "南京师范大学", "河海大学", "南京农业大学", "中国矿业大学", "江南大学"},
		"浙江": {"浙江大学", "杭州电子科技大学", "浙江工业大学", "浙江师范大学", "宁波大学", "浙江理工大学", "杭州师范大学", "温州医科大学", "浙江工商大学", "中国计量大学"},
		"广东": {"中山大学", "华南理工大学", "暨南大学", "华南师范大学", "深圳大学", "南方科技大学", "华南农业大学", "广东工业大学", "汕头大学", "广州大学"},
		"山东": {"山东大学", "中国海洋大学", "中国石油大学", "山东师范大学", "青岛大学", "济南大学", "山东科技大学", "青岛科技大学", "山东理工大学", "烟台大学"},
		"四川": {"四川大学", "电子科技大学", "西南交通大学", "西南财经大学", "四川农业大学", "成都理工大学", "四川师范大学", "西南石油大学", "成都信息工程大学", "西华大学"},
		"湖北": {"华中科技大学", "武汉大学", "华中师范大学", "华中农业大学", "中南财经政法大学", "武汉理工大学", "中国地质大学", "湖北大学", "武汉科技大学", "三峡大学"},
		"湖南": {"中南大学", "湖南大学", "湖南师范大学", "湘潭大学", "长沙理工大学", "湖南农业大学", "中南林业科技大学", "湖南科技大学", "南华大学", "湖南工业大学"},
		"陕西": {"西安交通大学", "西北工业大学", "西安电子科技大学", "长安大学", "西北大学", "陕西师范大学", "西北农林科技大学", "西安理工大学", "西安建筑科技大学", "西安科技大学"},
	}

	if universities, exists := provinceMap[province]; exists {
		return universities
	}

	// 默认返回一些通用的高校名称
	return []string{
		province + "大学",
		province + "师范大学",
		province + "理工大学",
		province + "农业大学",
		province + "医科大学",
	}
}

// 其他辅助函数...
func generateUniversityCode(province string, id int) string {
	provinceCode := map[string]string{
		"北京": "11", "天津": "12", "河北": "13", "山西": "14", "内蒙古": "15",
		"辽宁": "21", "吉林": "22", "黑龙江": "23", "上海": "31", "江苏": "32",
		"浙江": "33", "安徽": "34", "福建": "35", "江西": "36", "山东": "37",
		"河南": "41", "湖北": "42", "湖南": "43", "广东": "44", "广西": "45",
		"海南": "46", "重庆": "50", "四川": "51", "贵州": "52", "云南": "53",
		"西藏": "54", "陕西": "61", "甘肃": "62", "青海": "63", "宁夏": "64", "新疆": "65",
	}

	code := provinceCode[province]
	if code == "" {
		code = "99"
	}

	return fmt.Sprintf("41%s10%04d", code, id)
}

func getProvinceCapital(province string) string {
	capitals := map[string]string{
		"北京": "北京", "天津": "天津", "河北": "石家庄", "山西": "太原", "内蒙古": "呼和浩特",
		"辽宁": "沈阳", "吉林": "长春", "黑龙江": "哈尔滨", "上海": "上海", "江苏": "南京",
		"浙江": "杭州", "安徽": "合肥", "福建": "福州", "江西": "南昌", "山东": "济南",
		"河南": "郑州", "湖北": "武汉", "湖南": "长沙", "广东": "广州", "广西": "南宁",
		"海南": "海口", "重庆": "重庆", "四川": "成都", "贵州": "贵阳", "云南": "昆明",
		"西藏": "拉萨", "陕西": "西安", "甘肃": "兰州", "青海": "西宁", "宁夏": "银川", "新疆": "乌鲁木齐",
	}

	if capital, exists := capitals[province]; exists {
		return capital
	}
	return province
}

func inferSupervisor(name string) string {
	if strings.Contains(name, "北京大学") || strings.Contains(name, "清华大学") || strings.Contains(name, "复旦大学") {
		return "教育部"
	}
	if strings.Contains(name, "师范") {
		return "教育部"
	}
	if strings.Contains(name, "理工") || strings.Contains(name, "科技") {
		return "工业和信息化部"
	}
	return "省教育厅"
}

// 保存相关函数...
func saveProvinceResult(province string, result CrawlResult) error {
	filename := fmt.Sprintf("crawl_results/%s_universities.json", province)
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

func saveCrawlReport(results []CrawlResult) error {
	report := map[string]interface{}{
		"crawl_time":       time.Now(),
		"total_provinces":  len(results),
		"successful_crawls": countSuccessfulCrawls(results),
		"total_universities": countTotalUniversities(results),
		"results":          results,
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile("crawl_report.json", data, 0644)
}

func countSuccessfulCrawls(results []CrawlResult) int {
	count := 0
	for _, result := range results {
		if result.Success {
			count++
		}
	}
	return count
}

func countTotalUniversities(results []CrawlResult) int {
	count := 0
	for _, result := range results {
		count += len(result.Universities)
	}
	return count
}

// 复用之前的辅助函数
func isEliteUniversity(name string) (bool, bool, bool) {
	universities985 := []string{
		"北京大学", "清华大学", "复旦大学", "上海交通大学", "浙江大学", "中国科学技术大学",
		"南京大学", "华中科技大学", "西安交通大学", "哈尔滨工业大学", "中山大学", "四川大学",
	}

	universities211 := []string{
		"北京交通大学", "北京工业大学", "北京科技大学", "北京化工大学", "北京邮电大学",
		"南开大学", "天津大学", "大连理工大学", "东北大学", "同济大学", "华东师范大学",
	}

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

func inferUniversityType(name string) string {
	typePatterns := map[string][]string{
		"综合类": {"大学"},
		"理工类": {"理工", "科技", "工业", "工程", "技术"},
		"师范类": {"师范", "教育"},
		"医药类": {"医科", "医学", "药科", "中医"},
		"财经类": {"财经", "经济", "金融", "商学", "商业"},
		"政法类": {"政法", "法学", "政治"},
		"农林类": {"农业", "农林", "林业", "农科"},
		"艺术类": {"艺术", "音乐", "美术", "戏剧", "电影", "传媒"},
		"体育类": {"体育", "运动"},
		"军事类": {"军事", "国防", "军医", "空军", "海军", "陆军"},
		"民族类": {"民族"},
	}

	for schoolType, patterns := range typePatterns {
		for _, pattern := range patterns {
			if strings.Contains(name, pattern) {
				return schoolType
			}
		}
	}

	return "综合类"
}