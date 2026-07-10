package configs

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	JWTIssuer   string
	JWTAudience string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "3000"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/dc_express"),
		JWTSecret:   getEnv("JWT_SECRET", "super-secret-key-change-in-production"),
		JWTIssuer:   getEnv("JWT_ISSUER", "dc-express"),
		JWTAudience: getEnv("JWT_AUDIENCE", "dc-express"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
