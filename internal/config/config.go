package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBDriver       string
	JWTSecret      string
	GoogleClientID string
	StockfishPath  string
	EnginePoolSize int

	AdminFrontendURL string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		Port:           getEnv("PORT", "8080"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "chess_user"),
		DBPassword:     getEnv("DB_PASSWORD", "chess_password"),
		DBName:         getEnv("DB_NAME", "chess_puzzle_db"),
		DBDriver:       getEnv("DB_DRIVER", "postgres"),
		JWTSecret:      getEnv("JWT_SECRET", "supersecretkey"),
		GoogleClientID: getEnv("GOOGLE_CLIENT_ID", ""),
		StockfishPath:  getEnv("STOCKFISH_PATH", "/opt/homebrew/bin/stockfish"),
		EnginePoolSize: getEnvInt("ENGINE_POOL_SIZE", 2),
		AdminFrontendURL: getEnv("ADMIN_FRONTEND_URL", "http://localhost:3000"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if v, err := strconv.Atoi(value); err == nil && v > 0 {
			return v
		}
	}
	return fallback
}
