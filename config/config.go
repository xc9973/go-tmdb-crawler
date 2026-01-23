package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	TMDB      TMDBConfig
	Telegraph TelegraphConfig
	Scheduler SchedulerConfig
	Paths     PathsConfig
	CORS      CORSConfig
	Timezone  TimezoneConfig
	Auth      AuthConfig
}

// AppConfig holds application configuration
type AppConfig struct {
	Env      string
	Port     int
	LogLevel string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type     string // sqlite, postgres
	Path     string // for sqlite
	Host     string // for postgres
	Port     int    // for postgres
	Name     string // for postgres
	User     string // for postgres
	Password string // for postgres
	SSLMode  string // for postgres
}

// TMDBConfig holds TMDB API configuration
type TMDBConfig struct {
	APIKey   string
	BaseURL  string
	Language string
}

// TelegraphConfig holds Telegraph configuration
type TelegraphConfig struct {
	Token      string
	ShortName  string
	AuthorName string
	AuthorURL  string
}

// SchedulerConfig holds scheduler configuration
type SchedulerConfig struct {
	Enabled bool
	Cron    string
	TZ      string
}

// PathsConfig holds paths configuration
type PathsConfig struct {
	Web  string
	Log  string
	Data string
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins string
	AllowedMethods string
	AllowedHeaders string
}

// TimezoneConfig holds timezone configuration
type TimezoneConfig struct {
	// Default timezone for date/time operations
	// Examples: "UTC", "Asia/Shanghai", "America/New_York"
	Default string
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	// SecretKey 用于JWT签名和API Key验证
	// 如果为空，则跳过认证（开发环境）
	SecretKey string

	// AllowRemote 是否允许远程访问管理接口
	AllowRemote bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if exists
	if err := godotenv.Load(); err != nil {
		// .env file is optional, continue without it
		fmt.Println("No .env file found, using environment variables")
	}

	cfg := &Config{
		App: AppConfig{
			Env:      getEnv("APP_ENV", "development"),
			Port:     getEnvAsInt("APP_PORT", 8080),
			LogLevel: getEnv("APP_LOG_LEVEL", "info"),
		},
		Database: DatabaseConfig{
			Type:     getEnv("DB_TYPE", "sqlite"),
			Path:     getEnv("DB_PATH", "./tmdb.db"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			Name:     getEnv("DB_NAME", "tmdb"),
			User:     getEnv("DB_USER", "tmdb"),
			Password: getEnv("DB_PASSWORD", ""),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		TMDB: TMDBConfig{
			APIKey:   getEnv("TMDB_API_KEY", ""),
			BaseURL:  getEnv("TMDB_BASE_URL", "https://api.themoviedb.org/3"),
			Language: getEnv("TMDB_LANGUAGE", "zh-CN"),
		},
		Telegraph: TelegraphConfig{
			Token:      getEnv("TELEGRAPH_TOKEN", ""),
			ShortName:  getEnv("TELEGRAPH_SHORT_NAME", "tmdb_crawler"),
			AuthorName: getEnv("TELEGRAPH_AUTHOR_NAME", "剧集更新助手"),
			AuthorURL:  getEnv("TELEGRAPH_AUTHOR_URL", ""),
		},
		Scheduler: SchedulerConfig{
			Enabled: getEnvAsBool("ENABLE_SCHEDULER", true),
			Cron:    getEnv("DAILY_CRON", "0 8 * * *"),
			TZ:      getEnv("SCHEDULER_TZ", "Asia/Shanghai"),
		},
		Paths: PathsConfig{
			Web:  getEnv("WEB_DIR", "./web"),
			Log:  getEnv("LOG_DIR", "./logs"),
			Data: getEnv("DATA_DIR", "./data"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", ""),
			AllowedMethods: getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS"),
			AllowedHeaders: getEnv("CORS_ALLOWED_HEADERS", "Content-Type,Authorization,X-Admin-API-Key"),
		},
		Timezone: TimezoneConfig{
			Default: getEnv("DEFAULT_TIMEZONE", "Asia/Shanghai"),
		},
		Auth: AuthConfig{
			SecretKey:   getEnv("ADMIN_API_KEY", ""),
			AllowRemote: getEnvAsBool("ALLOW_REMOTE_ADMIN", false),
		},
	}

	// Validate required fields
	if cfg.Database.Type == "postgres" && cfg.Database.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required for PostgreSQL")
	}
	if cfg.TMDB.APIKey == "" {
		return nil, fmt.Errorf("TMDB_API_KEY is required")
	}
	if cfg.App.Port < 1 || cfg.App.Port > 65535 {
		return nil, fmt.Errorf("APP_PORT must be between 1 and 65535")
	}
	if cfg.Database.Port < 1 || cfg.Database.Port > 65535 {
		return nil, fmt.Errorf("DB_PORT must be between 1 and 65535")
	}
	if cfg.Database.Type != "sqlite" && cfg.Database.Type != "postgres" {
		return nil, fmt.Errorf("DB_TYPE must be sqlite or postgres")
	}

	// Validate CORS configuration
	if cfg.CORS.AllowedOrigins == "" {
		// If not configured, use localhost for development
		cfg.CORS.AllowedOrigins = "http://localhost:8080,http://127.0.0.1:8080"
	}

	// Validate timezone configuration
	if _, err := time.LoadLocation(cfg.Timezone.Default); err != nil {
		return nil, fmt.Errorf("invalid DEFAULT_TIMEZONE: %s (error: %w)", cfg.Timezone.Default, err)
	}

	return cfg, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsBool gets an environment variable as boolean
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

// GetDSN returns the database connection string for PostgreSQL
func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.User,
		c.Database.Password, c.Database.Name, c.Database.SSLMode,
	)
}

// GetSQLitePath returns the SQLite database path
func (c *Config) GetSQLitePath() string {
	return c.Database.Path
}
