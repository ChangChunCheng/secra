// internal/config/config.go
package config

import (
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	GRPCPort    string
	HTTPPort    string
	PostgresDSN string
	NvdURLv1    string
}

func Load() *AppConfig {
	_ = godotenv.Load()
	return &AppConfig{
		GRPCPort:    getenv("GRPC_PORT", ":50051"),
		HTTPPort:    getenv("HTTP_PORT", ":8080"),
		PostgresDSN: getenv("POSTGRES_DSN", "postgres://postgres:password@localhost:5432/secra?sslmode=disable"),
		NvdURLv1:    getenv("NVD_URL_V1", "https://nvd.nist.gov/feeds/json/cve/1.1/"),
	}
}

func getenv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
