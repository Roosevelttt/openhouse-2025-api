package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort           string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPass             string
	DBName             string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	JWTSecret          string
	CORSOrigins        string
	SMTPHost           string
	SMTPPort           string
	SMTPUser           string
	SMTPPass           string
	SMTPFrom           string // e.g., "OpenHouse 2025 <no-reply@yourdomain.com>"
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Load() *Config {
	// Load .env if present
	_ = godotenv.Load()

	cfg := &Config{
		HTTPPort:           getenv("HTTP_PORT", "8080"),
		DBHost:             getenv("DB_HOST", "127.0.0.1"),
		DBPort:             getenv("DB_PORT", "3306"),
		DBUser:             getenv("DB_USERNAME", "root"),
		DBPass:             getenv("DB_PASSWORD", ""),
		DBName:             getenv("DB_DATABASE", "openhouse-2025"),
		GoogleClientID:     getenv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getenv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getenv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/auth/google/callback"),
		JWTSecret:          getenv("JWT_SECRET", "change-me"),
		CORSOrigins:        getenv("CORS_ORIGINS", "*"),
		SMTPHost:           getenv("SMTP_HOST", ""),
		SMTPPort:           getenv("SMTP_PORT", "587"),
		SMTPUser:           getenv("SMTP_USER", ""),
		SMTPPass:           getenv("SMTP_PASS", ""),
		SMTPFrom:           getenv("SMTP_FROM", ""),
	}
	log.Printf("config loaded: HTTP_PORT=%s DB_HOST=%s DB_NAME=%s", cfg.HTTPPort, cfg.DBHost, cfg.DBName)
	return cfg
}
