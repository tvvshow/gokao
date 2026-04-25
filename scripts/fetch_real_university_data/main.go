package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/oktetopython/gaokao/pkg/shared"
)

// 实用工具箱API配置
const (
	APIURL    = "https://www.idcd.com/api/school"
	ClientID  = "949f0f5e-86fc-463f-b9c8-087060953e78"                             // 需要注册获取
	SecretKey = "0a165bea285e518625b142880c1b5a0deaaea85f075cf9382e6129c0a79ce9f6" // 需要注册获取
)

// 高校数据结构
type UniversityData struct {
	Name         string `json:"name"`
	Code         string `json:"code"`
	Department   string `json:"department"`
	Level        string `json:"level"`
	Property     string `json:"property"`
	Province     string `json:"province"`
	Is985        int    `json:"is_985"`
	Is211        int    `json:"is_211"`
	IsWorldA     int    `json:"is_world_a"`
	IsWorldB     int    `json:"is_world_b"`
	IsWorldClass int    `json:"is_world_class"`
}

// API响应结构
type APIResponse struct {
	Status    bool   `json:"status"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Data      struct {
		Total    int              `json:"total"`
		Page     int              `json:"page"`
		PageSize int              `json:"page_size"`
		Items    []UniversityData `json:"items"`
	} `json:"data"`
}

// 生成随机字符串
func generateNonce(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// 生成HMAC-SHA256签名
func generateSignature(clientID, nonce, timestamp, secretKey string) string {
	return shared.GenerateSignature(clientID, nonce, timestamp, secretKey)
}

// 获取高校数据
func fetchUniversityData(page, pageSize int) (*APIResponse, error) {
	// 检查API密钥
	if ClientID == "" || SecretKey == "" {
		return nil, fmt.Errorf("请先配置API密钥：ClientID和SecretKey")
	}

	// 生成请求参数
	nonce := generateNonce(32)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature := generateSignature(ClientID, nonce, timestamp, SecretKey)

	// 构建请求URL
	url := fmt.Sprintf("%s?page=%d&page_size=%d", APIURL, page, pageSize)

	// 创建HTTP请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("ClientID", ClientID)
	req.Header.Set("Nonce", nonce)
	req.Header.Set("Timestamp", timestamp)
	req.Header.Set("Signature", signature)
	req.Header.Set("SignatureMethod", "HmacSHA256")

	// 发送请求
	client := shared.HTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析JSON响应
	var apiResp APIResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		return nil, err
	}

	return &apiResp, nil
}

// 获取所有高校数据
func fetchAllUniversityData() ([]UniversityData, error) {
	var allUniversities []UniversityData
	pageSize := 100 // 每页获取100条数据
	page := 1

	for {
		fmt.Printf("正在获取第 %d 页数据...\n", page)

		resp, err := fetchUniversityData(page, pageSize)
		if err != nil {
			return nil, err
		}

		if !resp.Status {
			return nil, fmt.Errorf("API请求失败: %s", resp.Message)
		}

		// 添加当前页数据
		allUniversities = append(allUniversities, resp.Data.Items...)

		fmt.Printf("已获取 %d/%d 所高校数据\n", len(allUniversities), resp.Data.Total)

		// 检查是否还有更多数据
		if len(resp.Data.Items) < pageSize || len(allUniversities) >= resp.Data.Total {
			break
		}

		page++
		// 添加延迟避免请求过于频繁
		time.Sleep(1 * time.Second)
	}

	return allUniversities, nil
}

// 保存数据到JSON文件
func saveToJSON(universities []UniversityData, filename string) error {
	data, err := json.MarshalIndent(universities, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, data, 0644)
}

func main() {
	fmt.Println("开始获取全国高校真实数据...")
	fmt.Println("数据源：实用工具箱API (https://www.idcd.com)")
	fmt.Println("")

	// 检查API配置
	if ClientID == "" || SecretKey == "" {
		fmt.Println("⚠️  请先配置API密钥：")
		fmt.Println("1. 访问 https://www.idcd.com 注册账号")
		fmt.Println("2. 获取 ClientID 和 SecretKey")
		fmt.Println("3. 在代码中填入对应的密钥")
		fmt.Println("")
		fmt.Println("配置完成后重新运行此脚本")
		return
	}

	// 获取所有高校数据
	universities, err := fetchAllUniversityData()
	if err != nil {
		log.Fatalf("获取高校数据失败: %v", err)
	}

	fmt.Printf("\n✅ 成功获取 %d 所高校的真实数据\n", len(universities))

	// 统计数据
	var count985, count211, countWorldA, countWorldB, countWorldClass int
	for _, uni := range universities {
		if uni.Is985 == 1 {
			count985++
		}
		if uni.Is211 == 1 {
			count211++
		}
		if uni.IsWorldA == 1 {
			countWorldA++
		}
		if uni.IsWorldB == 1 {
			countWorldB++
		}
		if uni.IsWorldClass == 1 {
			countWorldClass++
		}
	}

	fmt.Println("\n📊 数据统计：")
	fmt.Printf("- 985工程院校: %d 所\n", count985)
	fmt.Printf("- 211工程院校: %d 所\n", count211)
	fmt.Printf("- 世界一流大学A类: %d 所\n", countWorldA)
	fmt.Printf("- 世界一流大学B类: %d 所\n", countWorldB)
	fmt.Printf("- 世界一流学科: %d 所\n", countWorldClass)

	// 保存到JSON文件
	filename := "real_universities_data.json"
	err = saveToJSON(universities, filename)
	if err != nil {
		log.Fatalf("保存数据失败: %v", err)
	}

	fmt.Printf("\n💾 数据已保存到: %s\n", filename)
	fmt.Println("\n🎉 真实高校数据获取完成！")
}
