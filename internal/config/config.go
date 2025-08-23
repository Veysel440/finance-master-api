package config

import (
	"os"
	"time"
)

func Env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

type Config struct {
	DSN, Addr, JWTSecret  string
	AccessTTL, RefreshTTL time.Duration
	RatesURL              string
	RatesTTL              time.Duration
}

func Load() Config {
	return Config{
		DSN:        Env("DB_DSN", "fin:StrongPW!234@tcp(127.0.0.1:3306)/finance_master?parseTime=true&loc=Local"),
		Addr:       Env("ADDR", ":8080"),
		JWTSecret:  Env("JWT_SECRET", "dev-secret-change"),
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 30 * 24 * time.Hour,
		RatesURL:   Env("RATES_URL", "https://api.exchangerate.host/latest"),
		RatesTTL:   durEnv("RATES_TTL", "12h"),
	}
}
func durEnv(k, d string) time.Duration { v := Env(k, d); t, _ := time.ParseDuration(v); return t }
