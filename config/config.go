package config

import (
	"fmt"
	"os"
	"strconv"

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
	AuthorName string
	AuthorURL  string
}

// SchedulerConfig holds scheduler configuration
type SchedulerConfig struct {
	Enabled bool
	Cron    string
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
			AuthorName: getEnv("TELEGRAPH_AUTHOR_NAME", "剧集更新助手"),
			AuthorURL:  getEnv("TELEGRAPH_AUTHOR_URL", ""),
		},
		Scheduler: SchedulerConfig{
			Enabled: getEnvAsBool("ENABLE_SCHEDULER", true),
			Cron:    getEnv("DAILY_CRON", "0 8 * * *"),
		},
		Paths: PathsConfig{
			Web:  getEnv("WEB_DIR", "./web"),
			Log:  getEnv("LOG_DIR", "./logs"),
			Data: getEnv("DATA_DIR", "./data"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "*"),
			AllowedMethods: getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS"),
			AllowedHeaders: getEnv("CORS_ALLOWED_HEADERS", "*"),
		},
	}

	// Validate required fields
	if cfg.Database.Type == "postgres" && cfg.Database.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required for PostgreSQL")
	}
	if cfg.TMDB.APIKey == "" {
		return nil, fmt.Errorf("TMDB_API_KEY is required")
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
