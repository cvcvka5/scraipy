package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv attempts to load a .env file from the root directory.
// It does not return an error because it's common to use environment
// variables in production and .env only for local development.
func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables.")
	}
}

// GetEnv is a helper to fetch a key or return a default value.
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
