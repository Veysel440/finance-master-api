package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr       string
	DSN        string
	JWTSecret  string
	AccessTTL  time.Duration
	RefreshTTL time.Duration

	AllowedOrigins []string
	AuthRateRPM    int
	OtherRateRPM   int

	RatesURL       string
	RatesTTL       time.Duration
	RatesStaleTTL  time.Duration
	RatesWarmEvery time.Duration
	RatesWarmBases []string

	CaptchaProvider  string
	RecaptchaSecret  string
	TurnstileSecret  string
	CaptchaThreshold int

	RetainAuditDays   int
	RetainSessionDays int
	CleanupEvery      time.Duration
}

func Load() Config {
	return Config{
		Addr:       getenv("ADDR", ":8080"),
		DSN:        getenv("DSN", "root:root@tcp(127.0.0.1:3306)/finance_master?parseTime=true&multiStatements=true"),
		JWTSecret:  getenv("JWT_SECRET", "dev-secret"),
		AccessTTL:  getdur("ACCESS_TTL", time.Hour),
		RefreshTTL: getdur("REFRESH_TTL", 24*time.Hour*15),

		AllowedOrigins: splitCSV(getenv("CORS_ALLOWED_ORIGINS", "")),
		AuthRateRPM:    getint("AUTH_RATE_RPM", 60),
		OtherRateRPM:   getint("OTHER_RATE_RPM", 200),

		RatesURL:       getenv("RATES_URL", "https://open.er-api.com/v6/latest"),
		RatesTTL:       getdur("RATES_TTL", 30*time.Minute),
		RatesStaleTTL:  getdur("RATES_STALE_TTL", 24*time.Hour),
		RatesWarmEvery: getdur("RATES_WARM_EVERY", time.Duration(0)),
		RatesWarmBases: splitCSV(getenv("RATES_WARM_BASES", "TRY,USD,EUR")),

		CaptchaProvider:   strings.ToLower(getenv("CAPTCHA_PROVIDER", "")),
		RecaptchaSecret:   getenv("RECAPTCHA_SECRET", ""),
		TurnstileSecret:   getenv("TURNSTILE_SECRET", ""),
		CaptchaThreshold:  getint("CAPTCHA_THRESHOLD", 5),
		RetainAuditDays:   getint("RETAIN_AUDIT_DAYS", 180),
		RetainSessionDays: getint("RETAIN_SESSION_DAYS", 30),
		CleanupEvery:      getdur("CLEANUP_EVERY", time.Hour*6),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
func getint(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
func getdur(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	ss := strings.Split(s, ",")
	out := make([]string, 0, len(ss))
	for _, p := range ss {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
