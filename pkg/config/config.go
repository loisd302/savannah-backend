package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for our application
type Config struct {
	Environment string
	Port        string
	Database    DatabaseConfig
	OIDC        OIDCConfig
	SMS         SMSConfig
	Redis       RedisConfig
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// OIDCConfig holds OpenID Connect configuration
type OIDCConfig struct {
	IssuerURL    string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// SMSConfig holds SMS service configuration
type SMSConfig struct {
	Username   string
	APIKey     string
	Shortcode  string
	BaseURL    string
	IsSandbox  bool
	RetryLimit int
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Port:        getEnv("PORT", "8080"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "devuser"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "backend_dev"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		OIDC: OIDCConfig{
			IssuerURL:    getEnv("OIDC_ISSUER_URL", ""),
			ClientID:     getEnv("OIDC_CLIENT_ID", ""),
			ClientSecret: getEnv("OIDC_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("OIDC_REDIRECT_URL", ""),
			Scopes:       getEnvSlice("OIDC_SCOPES", []string{"profile", "email"}),
		},
		SMS: SMSConfig{
			Username:   getEnv("SMS_USERNAME", ""),
			APIKey:     getEnv("SMS_API_KEY", ""),
			Shortcode:  getEnv("SMS_SHORTCODE", ""),
			BaseURL:    getEnv("SMS_BASE_URL", "https://api.sandbox.africastalking.com/version1"),
			IsSandbox:  getEnvBool("SMS_IS_SANDBOX", true),
			RetryLimit: getEnvInt("SMS_RETRY_LIMIT", 3),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an environment variable as integer with a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getEnvBool gets an environment variable as boolean with a default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

// getEnvSlice gets an environment variable as string slice with a default value
func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Split by comma and trim spaces
		parts := strings.Split(value, ",")
		result := make([]string, len(parts))
		for i, part := range parts {
			result[i] = strings.TrimSpace(part)
		}
		return result
	}
	return defaultValue
}
