package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL string
	Port        string
}

func LoadFromEnv() Config {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		user := getenv("POSTGRES_USER", "pruser")
		pass := getenv("POSTGRES_PASSWORD", "prpass")
		host := getenv("DB_HOST", "db")  // Docker-сервис
		port := getenv("DB_PORT", "5432")
		db := getenv("POSTGRES_DB", "pr_review")
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, db)
	}
	return Config{DatabaseURL: dsn, Port: getenv("PORT", "8080")}
}

func getenv(k, d string) string {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	return v
}
