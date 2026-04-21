package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

func main() {
	// Read the generated majors Go code
	majorsData, err := ioutil.ReadFile("majors_go_code.txt")
	if err != nil {
		fmt.Printf("Error reading majors_go_code.txt: %v\n", err)
		return
	}

	// Read the extended-data-service.go file
	serviceFile, err := ioutil.ReadFile("extended-data-service.go")
	if err != nil {
		fmt.Printf("Error reading extended-data-service.go: %v\n", err)
		return
	}

	// Convert to string for processing
	serviceContent := string(serviceFile)
	majorsContent := string(majorsData)

	// Find the generateExtendedMajors function and replace its content
	funcStart := regexp.MustCompile(`func generateExtendedMajors\(\) \[\]ExtendedMajor \{`)
	funcEnd := regexp.MustCompile(`\n\treturn majors\n\}`)

	// Find function start
	startMatch := funcStart.FindStringIndex(serviceContent)
	if startMatch == nil {
		fmt.Println("Could not find generateExtendedMajors function start")
		return
	}

	// Find function end
	endMatch := funcEnd.FindStringIndex(serviceContent)
	if endMatch == nil {
		fmt.Println("Could not find function end marker")
		return
	}

	// Replace the function content
	newFunctionContent := fmt.Sprintf(`func generateExtendedMajors() []ExtendedMajor {
	%s
	return majors
}`, majorsContent)

	// Build the new file content
	newContent := serviceContent[:startMatch[0]] + newFunctionContent + serviceContent[endMatch[1]:]

	// Write back to the file
	err = ioutil.WriteFile("extended-data-service.go", []byte(newContent), 0644)
	if err != nil {
		fmt.Printf("Error writing to extended-data-service.go: %v\n", err)
		return
	}

	fmt.Println("Successfully updated extended-data-service.go with complete majors data!")

	// Count the number of majors
	majorCount := strings.Count(majorsContent, `"id":`)
	fmt.Printf("Inserted %d majors into the service.\n", majorCount)
}