package main

import (
	"fmt"
	"log"

	"github.com/oktetopython/gaokao/pkg/scripts"
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

// 省级考试院配置（基于用户提供的31个网站）
var provinceConfigs = []ProvinceConfig{
	{Name: "北京", URL: "https://www.bjeea.cn", Selectors: []string{"university-list", "school-info"}, Enabled: true},
	{Name: "天津", URL: "http://www.zhaokao.net", Selectors: []string{"college-list"}, Enabled: true},
	{Name: "河北", URL: "http://www.hebeea.edu.cn", Selectors: []string{"university-data"}, Enabled: true},
	{Name: "山西", URL: "http://www.sxkszx.cn", Selectors: []string{"school-list"}, Enabled: true},
	{Name: "内蒙古", URL: "https://www.nm.zsks.cn", Selectors: []string{"college-info"}, Enabled: true},
	{Name: "辽宁", URL: "https://www.lnzsks.com", Selectors: []string{"university-list"}, Enabled: true},
	{Name: "吉林", URL: "http://www.jleea.edu.cn", Selectors: []string{"school-data"}, Enabled: true},
	{Name: "黑龙江", URL: "https://www.lzk.hl.cn", Selectors: []string{"college-list"}, Enabled: true},
	{Name: "上海", URL: "https://www.shmeea.edu.cn", Selectors: []string{"university-info"}, Enabled: true},
	{Name: "江苏", URL: "https://www.jseea.cn", Selectors: []string{"school-list"}, Enabled: true},
	{Name: "浙江", URL: "https://www.zjzs.net", Selectors: []string{"college-data"}, Enabled: true},
	{Name: "安徽", URL: "https://www.ahzsks.cn", Selectors: []string{"university-list"}, Enabled: true},
	{Name: "福建", URL: "https://www.eeafj.cn", Selectors: []string{"school-info"}, Enabled: true},
	{Name: "江西", URL: "http://www.jxeea.cn", Selectors: []string{"college-list"}, Enabled: true},
	{Name: "山东", URL: "http://www.sdzk.cn", Selectors: []string{"university-data"}, Enabled: true},
	{Name: "河南", URL: "http://www.haeea.cn", Selectors: []string{"school-list"}, Enabled: true},
	{Name: "湖北", URL: "http://www.hbea.edu.cn", Selectors: []string{"college-info"}, Enabled: true},
	{Name: "湖南", URL: "http://jyt.hunan.gov.cn/jyt/sjyt/hnsjyksy", Selectors: []string{"university-list"}, Enabled: true},
	{Name: "广东", URL: "https://eea.gd.gov.cn", Selectors: []string{"school-data"}, Enabled: true},
	{Name: "广西", URL: "https://www.gxeea.cn", Selectors: []string{"college-list"}, Enabled: true},
	{Name: "海南", URL: "http://ea.hainan.gov.cn", Selectors: []string{"university-info"}, Enabled: true},
	{Name: "重庆", URL: "https://www.cqksy.cn", Selectors: []string{"school-list"}, Enabled: true},
	{Name: "四川", URL: "https://www.sceea.cn", Selectors: []string{"college-data"}, Enabled: true},
	{Name: "贵州", URL: "http://zsksy.guizhou.gov.cn", Selectors: []string{"university-list"}, Enabled: true},
	{Name: "云南", URL: "https://www.ynzs.cn", Selectors: []string{"school-info"}, Enabled: true},
	{Name: "西藏", URL: "http://zsks.edu.xizang.gov.cn", Selectors: []string{"college-list"}, Enabled: true},
	{Name: "陕西", URL: "http://www.sneac.com", Selectors: []string{"university-data"}, Enabled: true},
	{Name: "甘肃", URL: "https://www.ganseea.cn", Selectors: []string{"school-list"}, Enabled: true},
	{Name: "青海", URL: "http://www.qhjyks.com", Selectors: []string{"college-info"}, Enabled: true},
	{Name: "宁夏", URL: "https://www.nxjyks.cn", Selectors: []string{"university-list"}, Enabled: true},
	{Name: "新疆", URL: "http://www.xjzk.gov.cn", Selectors: []string{"school-data"}, Enabled: true},
}

func main() {
	fmt.Println("🚀 开始爬取全国高校数据...")
	fmt.Printf("📊 计划爬取 %d 个省份的考试院网站\n", len(provinceConfigs))

	// 创建结果目录
	fileWriter := scripts.NewFileWriter()
	err := fileWriter.EnsureDir("crawl_results")
	if err != nil {
		log.Printf("❌ 创建目录失败: %v", err)
	}

	// 保存省份配置
	err = saveProvinceConfigs()
	if err != nil {
		log.Printf("❌ 保存省份配置失败: %v", err)
	}

	fmt.Println("\n📝 爬虫配置已生成:")
	fmt.Println("1. province_configs.json - 省份爬取配置")
	fmt.Println("2. 请先使用 Firecrawl 工具测试各省网站的可访问性")
	fmt.Println("3. 然后使用 Playwright 工具处理需要 JavaScript 的网站")
	fmt.Println("4. 最后运行完整的爬取流程")

	fmt.Println("\n🔧 下一步操作:")
	fmt.Println("- 运行 test_province_websites.go 测试网站可访问性")
	fmt.Println("- 运行 crawl_with_firecrawl.go 使用 Firecrawl 爬取数据")
	fmt.Println("- 运行 crawl_with_playwright.go 处理复杂网站")
}

// saveProvinceConfigs 保存省份配置到JSON文件
func saveProvinceConfigs() error {
	fileWriter := scripts.NewFileWriter()
	return fileWriter.SaveJSON(provinceConfigs, "province_configs.json")
}

// 985高校名单
var universities985 = []string{
	"北京大学", "清华大学", "复旦大学", "上海交通大学", "浙江大学", "中国科学技术大学",
	"南京大学", "华中科技大学", "西安交通大学", "哈尔滨工业大学", "中山大学", "四川大学",
	"北京理工大学", "华南理工大学", "大连理工大学", "北京航空航天大学", "东南大学",
	"天津大学", "华东师范大学", "北京师范大学", "同济大学", "厦门大学", "中南大学",
	"东北大学", "重庆大学", "湖南大学", "西北工业大学", "兰州大学", "电子科技大学",
	"华东理工大学", "中国农业大学", "东北师范大学", "西北农林科技大学", "中央民族大学",
	"国防科技大学", "中国海洋大学", "西北大学", "中南财经政法大学", "华中师范大学",
}

// 211高校名单（部分重点）
var universities211 = []string{
	"北京交通大学", "北京工业大学", "北京科技大学", "北京化工大学", "北京邮电大学",
	"北京林业大学", "北京中医药大学", "北京外国语大学", "中国传媒大学", "中央财经大学",
	"对外经济贸易大学", "北京体育大学", "中央音乐学院", "中国政法大学", "华北电力大学",
	"中国矿业大学", "中国石油大学", "中国地质大学", "南开大学", "天津医科大学",
	"河北工业大学", "太原理工大学", "内蒙古大学", "辽宁大学", "东北大学",
	"大连海事大学", "延边大学", "东北林业大学", "东北农业大学", "上海大学",
}

// isEliteUniversity 判断是否为重点大学
func isEliteUniversity(name string) (bool, bool, bool) {
	dataProcessor := scripts.NewDataProcessor()
	is985 := dataProcessor.ContainsFuzzy(universities985, name)
	is211 := dataProcessor.ContainsFuzzy(universities211, name) || is985
	isDoubleFirst := is985 // 简化处理，985通常也是双一流
	return is985, is211, isDoubleFirst
}

// inferUniversityType 根据名称推断大学类型
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

	dataProcessor := scripts.NewDataProcessor()
	for schoolType, patterns := range typePatterns {
		for _, pattern := range patterns {
			if dataProcessor.IsInString(name, pattern) {
				return schoolType
			}
		}
	}

	return "综合类"
}