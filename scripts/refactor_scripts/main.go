package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// 检查scripts目录结构
	fmt.Println("检查scripts目录结构...")
	
	// 获取当前目录
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取当前目录失败: %v\n", err)
		return
	}
	
	fmt.Printf("当前目录: %s\n", currentDir)
	
	// 检查scripts目录是否存在
	scriptsDir := filepath.Join(currentDir, "..")
	if _, err := os.Stat(scriptsDir); os.IsNotExist(err) {
		fmt.Printf("scripts目录不存在: %s\n", scriptsDir)
		return
	}
	
	fmt.Printf("scripts目录存在: %s\n", scriptsDir)
	
	// 检查几个重要的脚本子目录
	importantScripts := []string{
		"convert_json_to_go",
		"check_admission_table",
		"crawl_with_firecrawl",
		"advanced_university_crawler",
	}
	
	for _, script := range importantScripts {
		scriptDir := filepath.Join(scriptsDir, script)
		if _, err := os.Stat(scriptDir); os.IsNotExist(err) {
			fmt.Printf("❌ 脚本目录不存在: %s\n", scriptDir)
		} else {
			fmt.Printf("✅ 脚本目录存在: %s\n", scriptDir)
			
			// 检查main.go文件
			mainFile := filepath.Join(scriptDir, "main.go")
			if _, err := os.Stat(mainFile); os.IsNotExist(err) {
				fmt.Printf("❌ main.go文件不存在: %s\n", mainFile)
			} else {
				fmt.Printf("✅ main.go文件存在: %s\n", mainFile)
			}
			
			// 检查go.mod文件
			modFile := filepath.Join(scriptDir, "go.mod")
			if _, err := os.Stat(modFile); os.IsNotExist(err) {
				fmt.Printf("❌ go.mod文件不存在: %s\n", modFile)
			} else {
				fmt.Printf("✅ go.mod文件存在: %s\n", modFile)
			}
		}
		fmt.Println()
	}
	
	fmt.Println("目录结构检查完成!")
}