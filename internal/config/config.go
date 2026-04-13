package config

import (
	"log"
	"os"

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

	// Microservices URLs (for Gateway routing)
	IdentityURL string
	PuzzlesURL  string
	EngineURL   string
	AdminURL    string
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
		IdentityURL:    getEnv("IDENTITY_URL", "http://localhost:8081"),
		PuzzlesURL:     getEnv("PUZZLES_URL", "http://localhost:8082"),
		EngineURL:      getEnv("ENGINE_URL", "http://localhost:8083"),
		AdminURL:       getEnv("ADMIN_URL", "http://localhost:8084"),
		AdminFrontendURL: getEnv("ADMIN_FRONTEND_URL", "http://localhost:3000"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
