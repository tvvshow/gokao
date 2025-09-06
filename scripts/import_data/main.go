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

// InsertUniversities inserts universities into the database
func InsertUniversities(db *sql.DB, universities []University) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO universities(id, name, province, city, level, ranking) VALUES($1, $2, $3, $4, $5, $6)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, uni := range universities {
		_, err := stmt.Exec(uni.ID, uni.Name, uni.Province, uni.City, uni.Level, uni.Ranking)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// InsertMajors inserts majors into the database
func InsertMajors(db *sql.DB, majors []Major) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO majors(id, name, category, university_id) VALUES($1, $2, $3, $4)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, major := range majors {
		_, err := stmt.Exec(major.ID, major.Name, major.Category, major.UniversityID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func main() {
	if len(os.Args) < 4 {
		log.Fatal("Usage: go run import_data.go <config.json> <universities.json> <majors.json>")
	}

	configPath := os.Args[1]
	universitiesPath := os.Args[2]
	majorsPath := os.Args[3]

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

	// Read universities data
	universityData, err := ioutil.ReadFile(universitiesPath)
	if err != nil {
		log.Fatalf("Failed to read universities file: %v", err)
	}

	var universities []University
	err = json.Unmarshal(universityData, &universities)
	if err != nil {
		log.Fatalf("Failed to parse universities data: %v", err)
	}

	// Read majors data
	majorData, err := ioutil.ReadFile(majorsPath)
	if err != nil {
		log.Fatalf("Failed to read majors file: %v", err)
	}

	var majors []Major
	err = json.Unmarshal(majorData, &majors)
	if err != nil {
		log.Fatalf("Failed to parse majors data: %v", err)
	}

	// Insert data
	err = InsertUniversities(db, universities)
	if err != nil {
		log.Fatalf("Failed to insert universities: %v", err)
	}

	err = InsertMajors(db, majors)
	if err != nil {
		log.Fatalf("Failed to insert majors: %v", err)
	}

	fmt.Println("Data imported successfully!")
}