package config

import (
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://loguser:logpassword@localhost:5435/logdb?sslmode=disable"
	}

	return &Config{
		Port:        port,
		DatabaseURL: dbURL,
	}
}
