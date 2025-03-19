package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           int
	DatabaseURL    string
	JWTSecret      string
	AllowedOrigins []string
	Environment    string
}

// Load loads configuration from environment variables
func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it. Using environment variables.")
	}

	port, _ := strconv.Atoi(getEnv("PORT", "8080"))

	return &Config{
		Port:           port,
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://localhost:5432/feedback_collector?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "super_secret_key_change_this_in_production"),
		AllowedOrigins: []string{getEnv("ALLOWED_ORIGIN", "http://localhost:3000")},
		Environment:    getEnv("ENVIRONMENT", "development"),
	}
}

// Helper function to get environment variable with a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetPortString returns the port as a formatted string for HTTP server
func (c *Config) GetPortString() string {
	return fmt.Sprintf(":%d", c.Port)
}
