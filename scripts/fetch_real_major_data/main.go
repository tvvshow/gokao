package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/oktetopython/gaokao/pkg/scripts"
)

// 咕咕数据API配置
const (
	MajorAPIURL = "https://api.gugudata.com/metadata/ceemajor"
	APPKEY      = "" // 需要注册获取
)

// 专业数据结构
type MajorData struct {
	EducationLevel          string `json:"EducationLevel"`          // 学历层次
	DisciplinaryCategory    string `json:"DisciplinaryCategory"`    // 学科门类
	DisciplinarySubCategory string `json:"DisciplinarySubCategory"` // 学科专业类
	MajorCode               string `json:"MajorCode"`               // 专业代码
	MajorName               string `json:"MajorName"`               // 专业名称
	MajorIntroduction       string `json:"MajorIntroduction"`       // 专业介绍
	Courses                 []struct {
		CourseName       string `json:"CourseName"`       // 课程名称
		CourseDifficulty string `json:"CourseDifficulty"` // 课程难度
	} `json:"Courses"`
	GraduateScale    string   `json:"GraduateScale"`    // 毕业生规模
	MaleFemaleRatio  string   `json:"MaleFemaleRatio"`  // 男女比例
	RecommendSchools []string `json:"RecommendSchools"` // 推荐院校
}

// API响应结构
type MajorAPIResponse struct {
	DataStatus struct {
		StatusCode        int    `json:"StatusCode"`
		StatusDescription string `json:"StatusDescription"`
		ResponseDateTime  string `json:"ResponseDateTime"`
		DataTotalCount    int    `json:"DataTotalCount"`
	} `json:"DataStatus"`
	Data []MajorData `json:"Data"`
}

func main() {
	fmt.Println("开始获取全国高校专业真实数据...")
	fmt.Println("数据源：咕咕数据API (https://www.gugudata.com)")
	fmt.Println("")

	// 检查API配置
	if APPKEY == "" {
		fmt.Println("⚠️  请先配置API密钥：")
		fmt.Println("1. 访问 https://www.gugudata.com 注册账号")
		fmt.Println("2. 获取专业数据API的 APPKEY")
		fmt.Println("3. 在代码中填入对应的密钥")
		fmt.Println("")
		fmt.Println("配置完成后重新运行此脚本")
		return
	}

	// 获取所有专业数据
	majors, err := fetchAllMajorData()
	if err != nil {
		log.Fatalf("获取专业数据失败: %v", err)
	}

	fmt.Printf("\n✅ 成功获取 %d 个专业的真实数据\n", len(majors))

	// 统计分析数据
	analyzeMajorData(majors)

	// 保存到JSON文件
	fileWriter := scripts.NewFileWriter()
	filename := "real_majors_data.json"
	err = fileWriter.SaveJSON(majors, filename)
	if err != nil {
		log.Fatalf("保存数据失败: %v", err)
	}

	fmt.Printf("\n💾 数据已保存到: %s\n", filename)
	fmt.Println("\n🎉 真实专业数据获取完成！")
}

// 获取专业数据
func fetchMajorData(keywords string, pageSize, pageIndex int) (*MajorAPIResponse, error) {
	// 检查API密钥
	if APPKEY == "" {
		return nil, fmt.Errorf("请先配置API密钥：APPKEY")
	}

	// 构建请求URL
	baseURL, err := url.Parse(MajorAPIURL)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("appkey", APPKEY)
	params.Add("keywords", keywords)
	params.Add("pagesize", fmt.Sprintf("%d", pageSize))
	params.Add("pageindex", fmt.Sprintf("%d", pageIndex))
	baseURL.RawQuery = params.Encode()

	// 使用共享的HTTP客户端发送请求
	httpClient := scripts.NewHTTPClient(nil)
	headers := map[string]string{
		"User-Agent": "Gaokao-Major-Data-Fetcher/1.0",
	}

	data, err := httpClient.Get(baseURL.String(), headers)
	if err != nil {
		return nil, err
	}

	// 解析JSON响应
	var apiResp MajorAPIResponse
	err = json.Unmarshal(data, &apiResp)
	if err != nil {
		return nil, err
	}

	return &apiResp, nil
}

// 获取所有专业数据
func fetchAllMajorData() ([]MajorData, error) {
	var allMajors []MajorData
	pageSize := 20 // API限制每页最多20条
	pageIndex := 1

	// 首先获取总数据量
	fmt.Println("正在获取专业数据总量...")
	resp, err := fetchMajorData("", pageSize, pageIndex)
	if err != nil {
		return nil, err
	}

	if resp.DataStatus.StatusCode != 200 {
		return nil, fmt.Errorf("API请求失败: %s", resp.DataStatus.StatusDescription)
	}

	totalCount := resp.DataStatus.DataTotalCount
	totalPages := (totalCount + pageSize - 1) / pageSize

	fmt.Printf("共有 %d 个专业，需要获取 %d 页数据\n\n", totalCount, totalPages)

	// 获取所有页面的数据
	for pageIndex = 1; pageIndex <= totalPages; pageIndex++ {
		fmt.Printf("正在获取第 %d/%d 页数据...\n", pageIndex, totalPages)

		resp, err := fetchMajorData("", pageSize, pageIndex)
		if err != nil {
			return nil, err
		}

		if resp.DataStatus.StatusCode != 200 {
			return nil, fmt.Errorf("API请求失败: %s", resp.DataStatus.StatusDescription)
		}

		// 添加当前页数据
		allMajors = append(allMajors, resp.Data...)

		fmt.Printf("已获取 %d/%d 个专业数据\n", len(allMajors), totalCount)

		// 添加延迟避免请求过于频繁
		if pageIndex < totalPages {
			time.Sleep(1 * time.Second)
		}
	}

	return allMajors, nil
}

// 统计专业数据
func analyzeMajorData(majors []MajorData) {
	// 按学科门类统计
	categoryCount := make(map[string]int)
	// 按学历层次统计
	levelCount := make(map[string]int)
	// 按专业类统计
	subCategoryCount := make(map[string]int)

	for _, major := range majors {
		categoryCount[major.DisciplinaryCategory]++
		levelCount[major.EducationLevel]++
		subCategoryCount[major.DisciplinarySubCategory]++
	}

	fmt.Println("\n📊 专业数据统计：")
	fmt.Printf("- 总专业数: %d 个\n", len(majors))

	fmt.Println("\n按学历层次分布：")
	for level, count := range levelCount {
		fmt.Printf("  - %s: %d 个\n", level, count)
	}

	fmt.Println("\n按学科门类分布：")
	for category, count := range categoryCount {
		if category != "" {
			fmt.Printf("  - %s: %d 个\n", category, count)
		}
	}

	fmt.Println("\n主要专业类别（前10）：")
	type categoryItem struct {
		name  string
		count int
	}
	var sortedCategories []categoryItem
	for name, count := range subCategoryCount {
		if name != "" {
			sortedCategories = append(sortedCategories, categoryItem{name, count})
		}
	}
	// 简单排序（冒泡排序）
	for i := 0; i < len(sortedCategories)-1; i++ {
		for j := 0; j < len(sortedCategories)-1-i; j++ {
			if sortedCategories[j].count < sortedCategories[j+1].count {
				sortedCategories[j], sortedCategories[j+1] = sortedCategories[j+1], sortedCategories[j]
			}
		}
	}

	maxShow := 10
	if len(sortedCategories) < maxShow {
		maxShow = len(sortedCategories)
	}
	for i := 0; i < maxShow; i++ {
		fmt.Printf("  - %s: %d 个\n", sortedCategories[i].name, sortedCategories[i].count)
	}
}