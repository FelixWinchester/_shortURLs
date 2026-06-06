package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort    string
	DatabaseURL   string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	BaseURL       string
}

func Load() *Config {
	return &Config{
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://shorturls:shorturls@localhost:5432/shorturls?sslmode=disable"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),
		BaseURL:       getEnv("BASE_URL", "http://localhost:8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
