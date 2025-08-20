package config

import (
	"os"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	ServerPort  string
	ServerHost  string
	UploadDir   string
	MaxFileSize int64
	StunServer  string
	Environment string
	Debug       bool
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://telemed:telemed123@localhost:5432/telemed?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		ServerHost:  getEnv("SERVER_HOST", "localhost"),
		UploadDir:   getEnv("UPLOAD_DIR", "./uploads"),
		MaxFileSize: 10485760, // 10MB
		StunServer:  getEnv("STUN_SERVER", "stun:stun.l.google.com:19302"),
		Environment: getEnv("ENV", "development"),
		Debug:       getEnv("DEBUG", "true") == "true",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
