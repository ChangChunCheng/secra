// internal/config/config.go
package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	GRPCPort    string
	HTTPPort    string
	PostgresDSN string
	NvdURLv1    string
	NvdURLv2    string
	JWTConfig   JWTConfig
}

// JWTConfig holds JWT settings.
type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

func Load() *AppConfig {
	_ = godotenv.Load()
	return &AppConfig{
		GRPCPort:    getenv("GRPC_PORT", ":50051"),
		HTTPPort:    getenv("HTTP_PORT", ":8080"),
		PostgresDSN: getenv("POSTGRES_DSN", "postgres://postgres:password@localhost:5432/secra?sslmode=disable"),
		NvdURLv1:    getenv("NVD_URL_V1", "https://nvd.nist.gov/feeds/json/cve/1.1/"),
		NvdURLv2:    getenv("NVD_URL_V2", "https://services.nvd.nist.gov/rest/json/cves/2.0"),
		JWTConfig: JWTConfig{
			Secret: getenv("JWT_SECRET", "default_secret"),
			Expiry: 24 * time.Hour,
		},
	}
}

func getenv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
