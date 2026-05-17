package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	AppEnv             string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	AIBaseURL          string
	AIModel            string
	AIAPIKey           string
	MidtransServerKey  string
	MidtransClientKey  string
	MidtransProduction bool
	CORSAllowedOrigins string
	JWTSecret          string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		Port:               getEnv("PORT"),
		AppEnv:             getEnv("APP_ENV"),
		DBHost:             getEnv("DB_HOST"),
		DBPort:             getEnv("DB_PORT"),
		DBUser:             getEnv("DB_USER"),
		DBPassword:         getEnv("DB_PASSWORD"),
		DBName:             getEnv("DB_NAME"),
		AIBaseURL:          getEnv("AI_BASE_URL"),
		AIModel:            getEnv("AI_MODEL"),
		AIAPIKey:           getEnv("AI_API_KEY"),
		MidtransServerKey:  getEnv("MIDTRANS_SERVER_KEY"),
		MidtransClientKey:  getEnv("MIDTRANS_CLIENT_KEY"),
		MidtransProduction: getEnv("MIDTRANS_IS_PRODUCTION") == "true",
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS"),
		JWTSecret:          getEnv("JWT_SECRET"),
	}
}

func getEnv(key string) string {
	return os.Getenv(key)
}
