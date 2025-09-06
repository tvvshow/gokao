package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
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

// Config holds database configuration
type Config struct {
	DBHost     string `json:"db_host"`
	DBPort     int    `json:"db_port"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
	DBName     string `json:"db_name"`
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(configPath string) (*Config, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// ConnectDB connects to the PostgreSQL database
func ConnectDB(config *Config) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)
	return sql.Open("postgres", connStr)
}

// CheckUniversityTable checks the universities table
func CheckUniversityTable(db *sql.DB) error {
	rows, err := db.Query("SELECT id, name, province, city, level, ranking FROM universities LIMIT 5")
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Println("Sample universities data:")
	for rows.Next() {
		var id, name, province, city, level string
		var ranking int
		err := rows.Scan(&id, &name, &province, &city, &level, &ranking)
		if err != nil {
			return err
		}
		fmt.Printf("ID: %s, Name: %s, Province: %s, City: %s, Level: %s, Ranking: %d\n", id, name, province, city, level, ranking)
	}

	return rows.Err()
}

// CheckMajorTable checks the majors table
func CheckMajorTable(db *sql.DB) error {
	rows, err := db.Query("SELECT id, name, category, university_id FROM majors LIMIT 5")
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Println("Sample majors data:")
	for rows.Next() {
		var id, name, category, universityID string
		err := rows.Scan(&id, &name, &category, &universityID)
		if err != nil {
			return err
		}
		fmt.Printf("ID: %s, Name: %s, Category: %s, University ID: %s\n", id, name, category, universityID)
	}

	return rows.Err()
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run check_data.go <config.json>")
	}

	configPath := os.Args[1]

	// Load configuration
	config, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := ConnectDB(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Check universities table
	err = CheckUniversityTable(db)
	if err != nil {
		log.Fatalf("Failed to check universities table: %v", err)
	}

	// Check majors table
	err = CheckMajorTable(db)
	if err != nil {
		log.Fatalf("Failed to check majors table: %v", err)
	}

	fmt.Println("Data check completed successfully!")
}