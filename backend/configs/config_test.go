package configs

import "testing"

const validSecret = "A-Strong-Secret-With-Upper-Lower-123!"

func setValidS3Config(t *testing.T) {
	t.Helper()
	t.Setenv("S3_ENDPOINT", "http://storage:8333")
	t.Setenv("S3_ACCESS_KEY", "test-access-key")
	t.Setenv("S3_SECRET_KEY", "test-secret-key")
}

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
	setValidS3Config(t)

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
	setValidS3Config(t)

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

func TestLoadRejectsInvalidS3Boolean(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("JWT_SECRET", validSecret)
	setValidS3Config(t)
	t.Setenv("S3_FORCE_PATH_STYLE", "sometimes")
	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject an invalid S3_FORCE_PATH_STYLE value")
	}
}
