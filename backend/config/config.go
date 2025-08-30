// backend/config/config.go
package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config struct holds all configuration for the application
type Config struct {
	DatabaseURL string
	JWTSecret   string
}

var AppConfig *Config

// LoadConfig loads config from .env file
func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	AppConfig = &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
	}
}
