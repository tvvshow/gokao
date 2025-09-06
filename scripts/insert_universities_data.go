package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	// 读取完整的大学数据
	universitiesData, err := ioutil.ReadFile("universities_go_code.txt")
	if err != nil {
		fmt.Printf("Error reading universities data: %v\n", err)
		return
	}

	// 读取extended-data-service.go文件
	serviceFile, err := ioutil.ReadFile("extended-data-service.go")
	if err != nil {
		fmt.Printf("Error reading service file: %v\n", err)
		return
	}

	// 将文件内容转换为字符串
	serviceContent := string(serviceFile)
	universitiesContent := string(universitiesData)

	// 查找并替换generateExtendedUniversities函数
	startMarker := "// 生成的2705所大学数据\nfunc generateExtendedUniversities() []ExtendedUniversity {"
	endMarker := "\t}\n\treturn universities\n\n// 生成生产级专业数据（500个专业）"

	// 查找函数开始位置
	startIndex := strings.Index(serviceContent, startMarker)
	if startIndex == -1 {
		fmt.Println("Could not find function start marker")
		return
	}

	// 查找函数结束位置（从开始位置之后查找）
	endIndex := strings.Index(serviceContent[startIndex:], endMarker)
	if endIndex == -1 {
		fmt.Println("Could not find function end marker")
		return
	}
	endIndex += startIndex + len(endMarker)

	// 构建新的文件内容
	newContent := serviceContent[:startIndex] + universitiesContent + serviceContent[endIndex:]

	// 写入更新后的文件
	err = ioutil.WriteFile("extended-data-service.go", []byte(newContent), 0644)
	if err != nil {
		fmt.Printf("Error writing updated file: %v\n", err)
		return
	}

	fmt.Println("Successfully updated extended-data-service.go with complete university data!")
}