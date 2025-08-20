package config

import (
	"os"
	"time"
)

type Config struct {
	DSN        string
	Addr       string
	JWTSecret  string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func Env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func Load() Config {
	return Config{
		DSN:        Env("DB_DSN", "root:Veysel.12@tcp(127.0.0.1:3306)/finance_master?parseTime=true&loc=Local"),
		Addr:       Env("ADDR", ":8080"),
		JWTSecret:  Env("JWT_SECRET", "dev-secret-change"),
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 30 * 24 * time.Hour,
	}
}
