package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	ServerPort string
	GinMode    string

	ClerkSecretKey string

	// MinIO Configuration
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	MinIOUseSSL    bool

	// Redis Configuration
	RedisURL string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	config := &Config{
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "filevault_user"),
		DBPassword:     getEnv("DB_PASSWORD", "filevault_password"),
		DBName:         getEnv("DB_NAME", "filevault"),
		DBSSLMode:      getEnv("DB_SSL_MODE", "disable"),
		ServerPort:     getEnv("PORT", getEnv("SERVER_PORT", "8080")), // Railway uses PORT
		GinMode:        getEnv("GIN_MODE", "debug"),
		ClerkSecretKey: getEnv("CLERK_SECRET_KEY", ""),

		MinIOEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin123"),
		MinIOBucket:    getEnv("MINIO_BUCKET", "filevault-files"),
		MinIOUseSSL:    getEnv("MINIO_USE_SSL", "false") == "true",

		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379"),
	}

	// Handle Railway DATABASE_URL
	if databaseURL := getEnv("DATABASE_URL", ""); databaseURL != "" {
		if err := config.parsePostgresURL(databaseURL); err != nil {
			return nil, fmt.Errorf("failed to parse DATABASE_URL: %w", err)
		}
	}

	return config, nil
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
}

// parsePostgresURL parses Railway's DATABASE_URL format
func (c *Config) parsePostgresURL(databaseURL string) error {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return err
	}

	c.DBHost = u.Hostname()
	c.DBPort = u.Port()
	if c.DBPort == "" {
		c.DBPort = "5432"
	}

	if u.User != nil {
		c.DBUser = u.User.Username()
		if password, ok := u.User.Password(); ok {
			c.DBPassword = password
		}
	}

	c.DBName = strings.TrimPrefix(u.Path, "/")

	// Railway typically requires SSL
	if strings.Contains(u.Query().Get("sslmode"), "require") {
		c.DBSSLMode = "require"
	} else {
		c.DBSSLMode = "disable"
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
