package configs

import "testing"

const validSecret = "A-Strong-Secret-With-Upper-Lower-123!"

func TestStrongSecret(t *testing.T) {
	tests := []struct {
		name   string
		secret string
		want   bool
	}{
		{name: "valid", secret: validSecret, want: true},
		{name: "too short", secret: "Short-1", want: false},
		{name: "one character class", secret: "abcdefghijklmnopqrstuvwxyzabcdef", want: false},
		{name: "two character classes", secret: "abcdefghijklmnopqrstuvwxyz123456", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := strongSecret(tt.secret); got != tt.want {
				t.Fatalf("strongSecret() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("JWT_SECRET", validSecret)
	t.Setenv("ENVIRONMENT", "")
	t.Setenv("PORT", "")
	t.Setenv("JWT_ISSUER", "")
	t.Setenv("JWT_AUDIENCE", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "3000" || cfg.JWTIssuer != "dc-go" || cfg.JWTAudience != "dc-go" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if cfg.CookieSecure {
		t.Fatal("development cookies must not be secure")
	}
}

func TestLoadProductionEnablesSecureCookies(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("JWT_SECRET", validSecret)
	t.Setenv("ENVIRONMENT", "production")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.CookieSecure {
		t.Fatal("production cookies must be secure")
	}
}

func TestLoadRejectsInvalidConfiguration(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("JWT_SECRET", "weak")
	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject missing database URL and weak secret")
	}
}
