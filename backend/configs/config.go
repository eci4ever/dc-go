package configs

import (
	"errors"
	"os"
	"strconv"
	"unicode"
)

type Config struct {
	Port         string
	DatabaseURL  string
	RedisURL     string
	JWTSecret    string
	JWTIssuer    string
	JWTAudience  string
	CookieSecure bool
	Environment  string
	S3Endpoint   string
	S3AccessKey  string
	S3SecretKey  string
	S3Bucket     string
	S3Region     string
	S3UseSSL     bool
	S3PathStyle  bool
}

func Load() (Config, error) {
	environment := getEnv("ENVIRONMENT", "development")
	s3UseSSL, err := getBoolEnv("S3_USE_SSL", false)
	if err != nil {
		return Config{}, err
	}
	s3PathStyle, err := getBoolEnv("S3_FORCE_PATH_STYLE", true)
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		Port:         getEnv("PORT", "3000"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		RedisURL:     os.Getenv("REDIS_URL"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		JWTIssuer:    getEnv("JWT_ISSUER", "dc-go"),
		JWTAudience:  getEnv("JWT_AUDIENCE", "dc-go"),
		Environment:  environment,
		CookieSecure: environment == "production",
		S3Endpoint:   os.Getenv("S3_ENDPOINT"),
		S3AccessKey:  os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey:  os.Getenv("S3_SECRET_KEY"),
		S3Bucket:     getEnv("S3_BUCKET", "dc-go"),
		S3Region:     getEnv("S3_REGION", "us-east-1"),
		S3UseSSL:     s3UseSSL,
		S3PathStyle:  s3PathStyle,
	}
	if cfg.DatabaseURL == "" || cfg.RedisURL == "" || !strongSecret(cfg.JWTSecret) || cfg.S3Endpoint == "" || cfg.S3AccessKey == "" || cfg.S3SecretKey == "" {
		return Config{}, errors.New("DATABASE_URL, REDIS_URL, a strong JWT_SECRET, and S3 connection settings are required")
	}
	return cfg, nil
}

func getBoolEnv(key string, fallback bool) (bool, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, errors.New(key + " must be true or false")
	}
	return parsed, nil
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
