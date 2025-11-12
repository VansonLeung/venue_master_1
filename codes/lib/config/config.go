package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config aggregates runtime configuration for a service.
type Config struct {
	AppEnv      string
	ServiceName string

	httpHost string
	httpPort int

	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	AWS      AWSConfig
	Stripe   StripeConfig
	SendGrid SendGridConfig
	FCM      FCMConfig
}

// DatabaseConfig captures connection attributes for PostgreSQL.
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

// RedisConfig stores caching/session settings.
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// JWTConfig groups token behavior.
type JWTConfig struct {
	Secret        string
	Issuer        string
	Audience      string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

// AWSConfig configures S3-compatible storage.
type AWSConfig struct {
	Region   string
	Bucket   string
	Endpoint string
}

// StripeConfig stores payment credentials.
type StripeConfig struct {
	APIKey string
}

// SendGridConfig stores email credentials.
type SendGridConfig struct {
	APIKey string
}

// FCMConfig stores push notification credentials.
type FCMConfig struct {
	ServiceAccountPath string
}

// Load reads environment variables (and optional .env) into Config.
func Load(serviceName string) (*Config, error) {
	_ = godotenv.Load(".env")

	cfg := &Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		ServiceName: serviceName,
		httpHost:    getEnv("SERVICE_HOST", "0.0.0.0"),
		httpPort:    getEnvAsInt("SERVICE_PORT", 8080),
	}

	cfg.Database = DatabaseConfig{
		Host:     getEnv("POSTGRES_HOST", "postgres"),
		Port:     getEnvAsInt("POSTGRES_PORT", 5432),
		User:     getEnv("POSTGRES_USER", "postgres"),
		Password: getEnv("POSTGRES_PASSWORD", "postgres"),
		Name:     getEnv("POSTGRES_DB", defaultDBName(serviceName)),
		SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
	}

	cfg.Redis = RedisConfig{
		Addr:     getEnv("REDIS_ADDR", "redis:6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       getEnvAsInt("REDIS_DB", 0),
	}

	cfg.JWT = JWTConfig{
		Secret:        getEnv("JWT_SECRET", "super-secret"),
		Issuer:        getEnv("JWT_ISSUER", "venue-master"),
		Audience:      getEnv("JWT_AUDIENCE", "venue-master-clients"),
		AccessExpiry:  time.Duration(getEnvAsInt("JWT_ACCESS_EXP_MINUTES", 15)) * time.Minute,
		RefreshExpiry: time.Duration(getEnvAsInt("JWT_REFRESH_EXP_MINUTES", 7*24*60)) * time.Minute,
	}

	cfg.AWS = AWSConfig{
		Region:   getEnv("AWS_REGION", "us-east-1"),
		Bucket:   getEnv("AWS_S3_BUCKET", "venue-master"),
		Endpoint: getEnv("AWS_ENDPOINT", ""),
	}

	cfg.Stripe = StripeConfig{APIKey: getEnv("STRIPE_API_KEY", "")}
	cfg.SendGrid = SendGridConfig{APIKey: getEnv("SENDGRID_API_KEY", "")}
	cfg.FCM = FCMConfig{ServiceAccountPath: getEnv("FCM_SERVICE_ACCOUNT", "")}

	return cfg, nil
}

// HTTPAddr returns the HTTP bind address.
func (c *Config) HTTPAddr() string {
	return fmt.Sprintf("%s:%d", c.httpHost, c.httpPort)
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if valStr, ok := os.LookupEnv(key); ok && valStr != "" {
		if val, err := strconv.Atoi(valStr); err == nil {
			return val
		}
	}
	return fallback
}

func defaultDBName(serviceName string) string {
	name := strings.ReplaceAll(serviceName, "-", "_")
	name = strings.ReplaceAll(name, " ", "_")
	if name == "" {
		name = "service"
	}
	return fmt.Sprintf("%s_db", name)
}
