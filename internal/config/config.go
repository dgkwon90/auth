package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/gofiber/fiber/v2/log"
	"github.com/joho/godotenv"
)

// Config 환경 변수 구조체
type Config struct {
	Port         string
	JwtSecret    string
	SmtpServer   string
	SmtpPort     string
	SmtpId       string
	SmtpPassword string
	DatabaseUrl  string
}

var (
	config Config
	once   sync.Once
)

// LoadConfig 환경 변수 로드
func LoadConfig(filenames ...string) Config {
	once.Do(func() {
		err := godotenv.Load(filenames...)
		if err != nil {
			log.Warn("Error loading .env file", "error", err)
		}

		dbHost := getEnv("DB_HOST", "localhost")
		dbPort := getEnv("DB_PORT", "5432")
		dbUser := getEnv("DB_USER", "postgres")
		dbPassword := getEnv("DB_PASSWORD", "")
		dbName := getEnv("DB_NAME", "postgres")

		// PostgreSQL 연결 문자열 생성
		databaseUrl := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			dbHost, dbPort, dbUser, dbPassword, dbName,
		)

		config = Config{
			Port:         getEnv("PORT", "3000"),
			JwtSecret:    getEnv("JWT_SECRET", ""),
			SmtpServer:   getEnv("SMTP_SERVER", ""),
			SmtpPort:     getEnv("SMTP_PORT", ""),
			SmtpId:       getEnv("SMTP_ID", ""),
			SmtpPassword: getEnv("SMTP_PASSWORD", ""),
			DatabaseUrl:  databaseUrl,
		}
	})
	return config
}

// getEnv 환경 변수 가져오기
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
