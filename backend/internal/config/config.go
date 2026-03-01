package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Build information (Injected via -ldflags)
var (
	Version   = "dev"
	BuildDate = "unknown"
	Commit    = "none"
	BuiltBy   = "unknown"
	OS        = "unknown"
	Arch      = "unknown"
)

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

type SMTPConfig struct {
	Host       string
	Port       int
	User       string
	Password   string
	From       string
	Encryption string // SSL, STARTTLS, or NONE
}

type AppConfig struct {
	GRPCPort    string
	HTTPPort    string
	PostgresDSN string
	NvdURLv1    string
	NvdURLv2    string
	NVDURL      string // Unified NVD URL

	JWTConfig  JWTConfig
	SMTPConfig SMTPConfig

	NvdMaxRetries int
	NvdRetryDelay time.Duration
	NvdAPIKey     string

	// Default Admin User
	DefaultAdminUsername string
	DefaultAdminPassword string

	// Auto Migration (default: true for convenience)
	AutoMigrate bool

	// Import Scheduler (cron format, default: every hour at 0 minutes)
	ImportSchedule string
	ImportEnabled  bool
}

func Load() *AppConfig {
	_ = godotenv.Load()

	maxRetries, _ := strconv.Atoi(getEnv("NVD_MAX_RETRIES", "5"))
	retryDelaySec, _ := strconv.Atoi(getEnv("NVD_RETRY_DELAY_SECONDS", "10"))
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	autoMigrate := getBoolEnv("AUTO_MIGRATE", true)
	importEnabled := getBoolEnv("IMPORT_ENABLED", true)

	// Build PostgresDSN from components or use explicit POSTGRES_DSN
	postgresDSN := getEnv("POSTGRES_DSN", "")
	if postgresDSN == "" {
		postgresUser := getEnv("POSTGRES_USER", "postgres")
		postgresPass := getEnv("POSTGRES_PASSWORD", "postgres")
		postgresDB := getEnv("POSTGRES_DB", "secra")
		postgresHost := getEnv("POSTGRES_HOST", "localhost")
		postgresPort := getEnv("POSTGRES_PORT", "5432")
		postgresDSN = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			postgresUser, postgresPass, postgresHost, postgresPort, postgresDB)
	}

	return &AppConfig{
		GRPCPort:    getEnv("GRPC_PORT", ":50051"),
		HTTPPort:    getEnv("HTTP_PORT", ":8081"),
		PostgresDSN: postgresDSN,
		NvdURLv1:    getEnv("NVD_URL_V1", "https://nvd.nist.gov/feeds/json/cve/1.1/"),
		NvdURLv2:    getEnv("NVD_URL_V2", "https://services.nvd.nist.gov/rest/json/cves/2.0/"),
		NVDURL:      getEnv("NVD_URL_V2", "https://services.nvd.nist.gov/rest/json/cves/2.0/"),
		JWTConfig: JWTConfig{
			Secret: getEnv("JWT_SECRET", "super-secret-key"),
			Expiry: 24 * time.Hour,
		},
		SMTPConfig: SMTPConfig{
			Host:       getEnv("SMTP_HOST", ""),
			Port:       smtpPort,
			User:       getEnv("SMTP_USER", ""),
			Password:   getEnv("SMTP_PASS", ""),
			From:       getEnv("SMTP_FROM", ""),
			Encryption: getEnv("SMTP_ENCRYPTION", "STARTTLS"),
		},
		NvdMaxRetries: maxRetries,
		NvdRetryDelay: time.Duration(retryDelaySec) * time.Second,
		NvdAPIKey:     getEnv("NVD_API_KEY", ""),

		DefaultAdminUsername: getEnv("SECRA_ADMIN_USER", "admin"),
		DefaultAdminPassword: getEnv("SECRA_ADMIN_PWD", "admin"),

		AutoMigrate:    autoMigrate,
		ImportSchedule: getEnv("IMPORT_SCHEDULE", "0 0 * * * *"), // Every hour at 0 minutes
		ImportEnabled:  importEnabled,
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getBoolEnv(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		b, err := strconv.ParseBool(value)
		if err == nil {
			return b
		}
	}
	return fallback
}
