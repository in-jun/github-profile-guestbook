package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	GitHubClientID     string
	GitHubClientSecret string
	OriginURL          string
	Port               string
	JWTSecret          string
	AccessTokenTTL     int
	RefreshTokenTTL    int
}

func Load() *Config {
	cfg := &Config{
		DBHost:             mustEnv("DB_HOST"),
		DBPort:             mustEnv("DB_PORT"),
		DBUser:             mustEnv("DB_USER"),
		DBPassword:         mustEnv("DB_PASSWORD"),
		DBName:             mustEnv("DB_DB"),
		GitHubClientID:     mustEnv("GITHUB_CLIENT_ID"),
		GitHubClientSecret: mustEnv("GITHUB_CLIENT_SECRET"),
		OriginURL:          mustEnv("ORIGIN_URL"),
		Port:               envWithDefault("PORT", "8080"),
		JWTSecret:          mustEnv("JWT_SECRET"),
		AccessTokenTTL:     envInt("ACCESS_TOKEN_TTL", 900),
		RefreshTokenTTL:    envInt("REFRESH_TOKEN_TTL", 604800),
	}
	return cfg
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("environment variable %s is required", key))
	}
	return v
}

func envInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}

func envWithDefault(key, defaultVal string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	return v
}
