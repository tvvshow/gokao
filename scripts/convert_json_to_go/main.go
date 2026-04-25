package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/oktetopython/gaokao/pkg/scripts"
)

// ExtendedUniversity 扩展大学结构
type ExtendedUniversity struct {
	ID          int    `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Province    string `json:"province"`
	City        string `json:"city"`
	Type        string `json:"type"`
	Level       string `json:"level"`
	Ranking     int    `json:"ranking"`
	IsActive    bool   `json:"is_active"`
	Website     string `json:"website"`
	Description string `json:"description"`
	FoundedYear int    `json:"founded_year"`
	Students    int    `json:"students"`
}

func main() {
	// 读取JSON文件
	fileUtil := scripts.NewFileWriter()
	data, err := fileUtil.ReadFile("universities_data.json")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	// 解析JSON
	var universities []ExtendedUniversity
	err = json.Unmarshal(data, &universities)
	if err != nil {
		fmt.Printf("Error unmarshaling JSON: %v\n", err)
		return
	}

	// 生成Go代码
	var goCode strings.Builder
	goCode.WriteString("// 生成的2700+所大学数据\n")
	goCode.WriteString("func generateExtendedUniversities() []ExtendedUniversity {\n")
	goCode.WriteString("\tuniversities := []ExtendedUniversity{\n")

	// 转换每个大学数据
	for i, uni := range universities {
		// 转义描述中的引号
		description := strings.ReplaceAll(uni.Description, "\"", "\\\"")
		
		goCode.WriteString(fmt.Sprintf("\t\t{%d, \"%s\", \"%s\", \"%s\", \"%s\", \"%s\", \"%s\", %d, %t, \"%s\", \"%s\", %d, %d},\n",
			uni.ID, uni.Code, uni.Name, uni.Province, uni.City, uni.Type, uni.Level, uni.Ranking, uni.IsActive, uni.Website, description, uni.FoundedYear, uni.Students))
		
		// 每100个大学添加一个注释
		if (i+1)%100 == 0 {
			goCode.WriteString(fmt.Sprintf("\t\t// 已生成 %d 所大学\n", i+1))
		}
	}

	goCode.WriteString("\t}\n")
	goCode.WriteString("\treturn universities\n")
	goCode.WriteString("}\n")

	// 写入Go文件
	fileUtil = scripts.NewFileWriter()
	err = fileUtil.SaveToFile([]byte(goCode.String()), "universities_go_code.txt")
	if err != nil {
		fmt.Printf("Error writing Go code file: %v\n", err)
		return
	}

	fmt.Printf("Successfully converted %d universities to Go code\n", len(universities))
	fmt.Println("Generated file: universities_go_code.txt")
}