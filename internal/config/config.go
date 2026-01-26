package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	ServerPort string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

// Load reads configuration from environment variables
// .env file is OPTIONAL (for Docker/deployed environments where env vars are injected)
// Will fail only if required environment variables are missing
func Load() (*Config, error) {
	// Try to load .env file (optional - for local development)
	// In Docker, env vars are already injected by docker-compose
	_ = godotenv.Load() // Ignore error - .env file is optional

	// Load configuration from environment variables (NO DEFAULTS for security)
	config := &Config{
		ServerPort: getRequiredEnv("SERVER_PORT"),
		DBHost:     getRequiredEnv("DB_HOST"),
		DBPort:     getRequiredEnv("DB_PORT"),
		DBUser:     getRequiredEnv("DB_USER"),
		DBPassword: getRequiredEnv("DB_PASSWORD"),
		DBName:     getRequiredEnv("DB_NAME"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	// Log only non-sensitive information
	log.Printf("Configuration loaded successfully: ServerPort=%s, DBHost=%s, DBName=%s",
		config.ServerPort, config.DBHost, config.DBName)

	return config, nil
}

// getRequiredEnv gets environment variable WITHOUT fallback
// Returns error if variable is empty or not set
func getRequiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		return ""
	}
	return value
}

// getEnv gets environment variable with fallback (only for non-sensitive config)
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// validateConfig validates the configuration values
func validateConfig(cfg *Config) error {
	// Validate ServerPort
	if cfg.ServerPort == "" {
		return errors.New("SERVER_PORT environment variable is required. Set it in .env file or docker-compose.yml")
	}
	port, err := strconv.Atoi(cfg.ServerPort)
	if err != nil {
		return fmt.Errorf("invalid SERVER_PORT: must be a number (got: %s)", cfg.ServerPort)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid SERVER_PORT: must be between 1 and 65535 (got: %d)", port)
	}

	// Validate DBPort
	if cfg.DBPort == "" {
		return errors.New("DB_PORT is required")
	}
	dbPort, err := strconv.Atoi(cfg.DBPort)
	if err != nil {
		return fmt.Errorf("invalid DB_PORT: must be a number (got: %s)", cfg.DBPort)
	}
	if dbPort < 1 || dbPort > 65535 {
		return fmt.Errorf("invalid DB_PORT: must be between 1 and 65535 (got: %d)", dbPort)
	}

	// Validate DBUser
	if cfg.DBUser == "" {
		return errors.New("DB_USER is required")
	}

	// Validate DBPassword
	if cfg.DBPassword == "" {
		return errors.New("DB_PASSWORD is required")
	}

	// Validate DBName
	if cfg.DBName == "" {
		return errors.New("DB_NAME is required")
	}

	// Validate DBHost
	if cfg.DBHost == "" {
		return errors.New("DB_HOST is required")
	}

	return nil
}

// GetDSN returns PostgreSQL connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost,
		c.DBPort,
		c.DBUser,
		c.DBPassword,
		c.DBName,
		c.DBSSLMode,
	)
}

// GetServerAddr returns server address
func (c *Config) GetServerAddr() string {
	return ":" + c.ServerPort
}
