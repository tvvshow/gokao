package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// University represents a university entity
type University struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Province string `json:"province"`
	City     string `json:"city"`
	Level    string `json:"level"`
	Ranking  int    `json:"ranking"`
}

// Major represents a major entity
type Major struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	UniversityID string `json:"university_id"`
}

// CrawlerConfig holds crawler configuration
type CrawlerConfig struct {
	BaseURL    string `json:"base_url"`
	APIKey     string `json:"api_key"`
	OutputDir  string `json:"output_dir"`
	MaxRetries int    `json:"max_retries"`
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(configPath string) (*CrawlerConfig, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config CrawlerConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// FetchData fetches data from a URL with retry logic
func FetchData(url, apiKey string, maxRetries int) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	for i := 0; i < maxRetries; i++ {
		resp, err := client.Do(req)
		if err != nil {
			if i == maxRetries-1 {
				return nil, err
			}
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			if i == maxRetries-1 {
				return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
			}
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		return ioutil.ReadAll(resp.Body)
	}

	return nil, fmt.Errorf("max retries exceeded")
}

// SaveData saves data to a file
func SaveData(data []byte, filename, outputDir string) error {
	path := filepath.Join(outputDir, filename)
	return ioutil.WriteFile(path, data, 0644)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run crawl_data.go <config.json>")
	}

	configPath := os.Args[1]

	// Load configuration
	config, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create output directory if it doesn't exist
	err = os.MkdirAll(config.OutputDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Fetch universities data
	universityURL := config.BaseURL + "/universities"
	universityData, err := FetchData(universityURL, config.APIKey, config.MaxRetries)
	if err != nil {
		log.Fatalf("Failed to fetch universities data: %v", err)
	}

	err = SaveData(universityData, "universities.json", config.OutputDir)
	if err != nil {
		log.Fatalf("Failed to save universities data: %v", err)
	}

	// Fetch majors data
	majorURL := config.BaseURL + "/majors"
	majorData, err := FetchData(majorURL, config.APIKey, config.MaxRetries)
	if err != nil {
		log.Fatalf("Failed to fetch majors data: %v", err)
	}

	err = SaveData(majorData, "majors.json", config.OutputDir)
	if err != nil {
		log.Fatalf("Failed to save majors data: %v", err)
	}

	fmt.Println("Data crawled and saved successfully!")
}