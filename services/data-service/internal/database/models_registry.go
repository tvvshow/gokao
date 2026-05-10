package database

import "github.com/tvvshow/gokao/services/data-service/internal/models"

func migratableModels() []interface{} {
	return []interface{}{
		&models.University{},
		&models.Major{},
		&models.AdmissionData{},
		&models.SearchIndex{},
		&models.AnalysisResult{},
		&models.HotSearch{},
		&models.DataStatistics{},
		&models.UniversityStatistics{},
		&models.MajorStatistics{},
	}
}

