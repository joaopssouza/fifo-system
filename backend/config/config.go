// backend/config/config.go
package config

import (
	"log"
	"os"
	"time" // <-- ADICIONAR IMPORT

	"github.com/joho/godotenv"
)

// Config struct holds all configuration for the application
type Config struct {
	DatabaseURL       string
	JWTSecret         string
	JWTExpirationTime time.Duration
}

var AppConfig *Config

// LoadConfig loads config from .env file
func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Carrega o tempo de expiração e converte-o para time.Duration
	duration, err := time.ParseDuration(os.Getenv("JWT_EXPIRATION_TIME"))
	if err != nil {
		log.Printf("Invalid JWT_EXPIRATION_TIME format, using default 24h. Error: %v", err)
		duration = time.Hour * 24 // Define um padrão seguro em caso de erro
	}

	AppConfig = &Config{
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		JWTExpirationTime: duration, // <-- USAR O VALOR CARREGADO
	}
}