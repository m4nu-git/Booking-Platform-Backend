package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)


func Load() {
	err := godotenv.Load()

	if err != nil {
		// Log the error if the .env file is not found ot cannot br loaded
		fmt.Println("Error loading .env file")
	}
}

func GetString(key string, fallback string) string {
	Load()

	value, ok := os.LookupEnv(key)

	if !ok {
		return fallback
	}

	return value
}