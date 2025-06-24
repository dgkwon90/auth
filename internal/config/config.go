// Package config provides configuration loading from environment variables and .env files.
package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/gofiber/fiber/v2/log"
	"github.com/joho/godotenv"
)

// Config 환경 변수 구조체
// Config holds all environment variables for the application.
type Config struct {
	Port         string
	JwtSecret    string
	SMTPServer   string
	SMTPPort     string
	SMTPID       string
	SMTPPassword string
	DatabaseURL  string
	DBType       string // "postgres" or "sqlite"
	SqlitePath   string // sqlite 파일 경로
}

var (
	config Config
	once   sync.Once
)

// LoadConfig 환경 변수 로드
// LoadConfig loads environment variables from .env files and returns a Config struct.
func LoadConfig(filenames ...string) Config {
	once.Do(func() {
		err := godotenv.Load(filenames...)
		if err != nil {
			log.Warn("Error loading .env file", "error", err)
		}

		dbType := getEnv("DB_TYPE", "postgres")
		sqlitePath := getEnv("SQLITE_PATH", "./local.db")
		var databaseURL string
		if dbType == "postgres" {
			dbHost := getEnv("DB_HOST", "localhost")
			dbPort := getEnv("DB_PORT", "5432")
			dbUser := getEnv("DB_USER", "postgres")
			dbPassword := getEnv("DB_PASSWORD", "")
			dbName := getEnv("DB_NAME", "postgres")
			// PostgreSQL 연결 문자열 생성
			databaseURL = fmt.Sprintf(
				"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
				dbHost, dbPort, dbUser, dbPassword, dbName,
			)
		}

		config = Config{
			Port:         getEnv("PORT", "3000"),
			JwtSecret:    getEnv("JWT_SECRET", ""),
			SMTPServer:   getEnv("SMTP_SERVER", ""),
			SMTPPort:     getEnv("SMTP_PORT", ""),
			SMTPID:       getEnv("SMTP_ID", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			DatabaseURL:  databaseURL,
			DBType:       dbType,
			SqlitePath:   sqlitePath,
		}
		log.Info("Configuration loaded successfully", config)
	})
	return config
}

// getEnv 환경 변수 가져오기
// getEnv returns the value of the environment variable or a default value if not set.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
