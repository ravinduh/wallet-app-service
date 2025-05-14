package config

import (
	"os"
	"strconv"

	"github.com/ravindu/wallet-app-service/pkg/database"
)

// Config holds all the configuration for the application
type Config struct {
	Server   ServerConfig
	Postgres database.PostgresConfig
	Redis    database.RedisConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Server config
	port := getEnv("SERVER_PORT", "8080")

	// Postgres config
	pgHost := getEnv("POSTGRES_HOST", "localhost")
	pgPort, _ := strconv.Atoi(getEnv("POSTGRES_PORT", "5432"))
	pgUser := getEnv("POSTGRES_USER", "postgres")
	pgPassword := getEnv("POSTGRES_PASSWORD", "postgres")
	pgDBName := getEnv("POSTGRES_DBNAME", "wallet")
	pgSSLMode := getEnv("POSTGRES_SSLMODE", "disable")
	
	// Redis config
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort, _ := strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	return &Config{
		Server: ServerConfig{
			Port: port,
		},
		Postgres: database.PostgresConfig{
			Host:     pgHost,
			Port:     pgPort,
			User:     pgUser,
			Password: pgPassword,
			DBName:   pgDBName,
			SSLMode:  pgSSLMode,
		},
		Redis: database.RedisConfig{
			Host:     redisHost,
			Port:     redisPort,
			Password: redisPassword,
			DB:       redisDB,
		},
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}