package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App           AppConfig
	Database      DatabaseConfig
	JWT           JWTConfig
	Storage       StorageConfig
	CORS          CORSConfig
	Log           LogConfig
	Razorpay      RazorpayConfig
	Notifications NotificationConfig
}

type NotificationConfig struct {
	Provider      string // "log", "gupshup"
	GupshupAPIKey string
	GupshupApp    string
	WebhookSecret string
}

type RazorpayConfig struct {
	KeyID         string
	KeySecret     string
	WebhookSecret string
}

type AppConfig struct {
	Name string
	Env  string
	Port string
}

type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	Name         string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode,
	)
}

type JWTConfig struct {
	AccessSecret      string
	RefreshSecret     string
	AccessExpiration  time.Duration
	RefreshExpiration time.Duration
	Issuer            string
}

type StorageConfig struct {
	Provider        string
	Endpoint        string
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	UsePathStyle    bool
}

type CORSConfig struct {
	AllowedOrigins []string
}

type LogConfig struct {
	Level  string
	Format string
}

func Load() (*Config, error) {
	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	maxOpenConns, _ := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "25"))
	maxIdleConns, _ := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "10"))
	maxLifetime, _ := time.ParseDuration(getEnv("DB_MAX_LIFETIME", "5m"))

	accessExp, _ := time.ParseDuration(getEnv("JWT_ACCESS_EXPIRATION", "15m"))
	refreshExp, _ := time.ParseDuration(getEnv("JWT_REFRESH_EXPIRATION", "168h"))

	cfg := &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "stayflow-track"),
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         dbPort,
			User:         getEnv("DB_USER", "stayflow"),
			Password:     getEnv("DB_PASSWORD", ""),
			Name:         getEnv("DB_NAME", "stayflow_track"),
			SSLMode:      getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns: maxOpenConns,
			MaxIdleConns: maxIdleConns,
			MaxLifetime:  maxLifetime,
		},
		JWT: JWTConfig{
			AccessSecret:      getEnv("JWT_ACCESS_SECRET", ""),
			RefreshSecret:     getEnv("JWT_REFRESH_SECRET", ""),
			AccessExpiration:  accessExp,
			RefreshExpiration: refreshExp,
			Issuer:            getEnv("JWT_ISSUER", "stayflow-track"),
		},
		Storage: StorageConfig{
			Provider:        getEnv("STORAGE_PROVIDER", "s3"),
			Endpoint:        getEnv("STORAGE_ENDPOINT", ""),
			Region:          getEnv("STORAGE_REGION", "us-east-1"),
			Bucket:          getEnv("STORAGE_BUCKET", "stayflow-documents"),
			AccessKeyID:     getEnv("STORAGE_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("STORAGE_SECRET_ACCESS_KEY", ""),
			UsePathStyle:    getEnv("STORAGE_USE_PATH_STYLE", "false") == "true",
		},
		CORS: CORSConfig{
			AllowedOrigins: strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"), ","),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Razorpay: RazorpayConfig{
			KeyID:         getEnv("RAZORPAY_KEY_ID", ""),
			KeySecret:     getEnv("RAZORPAY_KEY_SECRET", ""),
			WebhookSecret: getEnv("RAZORPAY_WEBHOOK_SECRET", ""),
		},
		Notifications: NotificationConfig{
			Provider:      getEnv("NOTIFICATION_PROVIDER", "log"),
			GupshupAPIKey: getEnv("GUPSHUP_API_KEY", ""),
			GupshupApp:    getEnv("GUPSHUP_APP_NAME", ""),
			WebhookSecret: getEnv("NOTIFICATION_WEBHOOK_SECRET", ""),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}
	if c.JWT.AccessSecret == "" {
		return fmt.Errorf("JWT_ACCESS_SECRET is required")
	}
	if c.JWT.RefreshSecret == "" {
		return fmt.Errorf("JWT_REFRESH_SECRET is required")
	}
	// Reject weak secrets in production
	if c.App.Env == "production" {
		if len(c.JWT.AccessSecret) < 32 {
			return fmt.Errorf("JWT_ACCESS_SECRET must be at least 32 characters in production")
		}
		if len(c.JWT.RefreshSecret) < 32 {
			return fmt.Errorf("JWT_REFRESH_SECRET must be at least 32 characters in production")
		}
		if c.JWT.AccessSecret == c.JWT.RefreshSecret {
			return fmt.Errorf("JWT_ACCESS_SECRET and JWT_REFRESH_SECRET must be different")
		}
		if strings.Contains(strings.ToLower(c.JWT.AccessSecret), "dev") ||
			strings.Contains(strings.ToLower(c.JWT.AccessSecret), "change") {
			return fmt.Errorf("JWT_ACCESS_SECRET contains unsafe default value")
		}
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
