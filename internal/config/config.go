package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	AppEnv          string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	AIBaseURL       string
	AIModel         string
	MidtransServerKey string
	MidtransClientKey string
	JWTSecret       string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		Port:              getEnv("PORT", "8080"),
		AppEnv:            getEnv("APP_ENV", "development"),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBUser:            getEnv("DB_USER", "vivian"),
		DBPassword:        getEnv("DB_PASSWORD", "admin"),
		DBName:            getEnv("DB_NAME", "smartcv"),
		AIBaseURL:         getEnv("AI_BASE_URL", "http://localhost:20128/v1"),
		AIModel:           getEnv("AI_MODEL", "GLM"),
		MidtransServerKey: getEnv("MIDTRANS_SERVER_KEY", ""),
		MidtransClientKey: getEnv("MIDTRANS_CLIENT_KEY", ""),
		JWTSecret:         getEnv("JWT_SECRET", "smartcv-secret-key"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
