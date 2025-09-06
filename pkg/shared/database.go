package shared

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// 数据库配置
const (
	DBHost     = "localhost"
	DBPort     = 5432
	DBUser     = "postgres"
	DBPassword = "password"
	DBName     = "gaokao_data"
)

// ConnectDB 连接到PostgreSQL数据库
func ConnectDB() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
		DBHost, DBPort, DBUser, DBPassword, DBName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}