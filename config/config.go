package config

import (
	"log"
	"os"
)

type Config struct {
	WhoisAPIKey string
}

func New() *Config {
	return &Config{
		WhoisAPIKey: getRequiredEnvString("WHOIS_API_KEY"),
	}
}

func getRequiredEnvString(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	log.Fatalf("Environment variable %s must be set", key)
	return ""
}
