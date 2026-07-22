package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

const defaultJWTSecret = "super-secret-key"

type Config struct {
	ServerPort         string
	DatabaseURL        string
	JWTSecret          string
	JWTExpiration      int
	CORSAllowedOrigins []string
}

func Load() *Config {
	exp, err := strconv.Atoi(getEnv("JWT_EXPIRATION", "3600"))
	if err != nil || exp <= 0 {
		exp = 3600
	}

	jwtSecret := getEnv("JWT_SECRET", defaultJWTSecret)
	if jwtSecret == defaultJWTSecret {
		log.Println("WARNING: using default JWT secret, set JWT_SECRET env variable for production")
	}

	return &Config{
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/booking?sslmode=disable"),
		JWTSecret:          jwtSecret,
		JWTExpiration:      exp,
		CORSAllowedOrigins: splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "*")),
	}
}

func splitCSV(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
