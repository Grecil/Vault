package config

import (
	"fmt"
	"os"

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
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		GinMode:        getEnv("GIN_MODE", "debug"),
		ClerkSecretKey: getEnv("CLERK_SECRET_KEY", ""),
	}

	return config, nil
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
