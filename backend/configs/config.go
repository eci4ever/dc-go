package configs

import (
	"errors"
	"os"
	"strings"
	"unicode"
)

type Config struct {
	Port           string
	DatabaseURL    string
	JWTSecret      string
	JWTIssuer      string
	JWTAudience    string
	AllowedOrigins []string
	CookieSecure   bool
	Environment    string
}

func Load() (Config, error) {
	cfg := Config{
		Port:         getEnv("PORT", "3000"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		JWTIssuer:    getEnv("JWT_ISSUER", "dc-express"),
		JWTAudience:  getEnv("JWT_AUDIENCE", "dc-express"),
		Environment:  getEnv("ENVIRONMENT", "development"),
		CookieSecure: os.Getenv("COOKIE_SECURE") == "true",
	}
	if origins := os.Getenv("ALLOWED_ORIGINS"); origins != "" {
		for _, origin := range strings.Split(origins, ",") {
			if origin = strings.TrimSpace(origin); origin != "" {
				cfg.AllowedOrigins = append(cfg.AllowedOrigins, origin)
			}
		}
	}
	if cfg.DatabaseURL == "" || !strongSecret(cfg.JWTSecret) {
		return Config{}, errors.New("DATABASE_URL and a JWT_SECRET of at least 32 characters are required")
	}
	if cfg.Environment == "production" && (len(cfg.AllowedOrigins) == 0 || !cfg.CookieSecure) {
		return Config{}, errors.New("production requires ALLOWED_ORIGINS and COOKIE_SECURE=true")
	}
	return cfg, nil
}

func strongSecret(s string) bool {
	if len(s) < 32 {
		return false
	}
	var upper, lower, digit, symbol bool
	for _, r := range s {
		upper = upper || unicode.IsUpper(r)
		lower = lower || unicode.IsLower(r)
		digit = digit || unicode.IsDigit(r)
		symbol = symbol || unicode.IsPunct(r) || unicode.IsSymbol(r)
	}
	classes := 0
	for _, v := range []bool{upper, lower, digit, symbol} {
		if v {
			classes++
		}
	}
	return classes >= 3
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
