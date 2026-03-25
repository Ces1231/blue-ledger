package config

import (
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	// Unset any env overrides by running in a clean environment.
	// t.Setenv restores the original value after the test.
	t.Setenv("APP_ENV", "")
	t.Setenv("PORT", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("REDIS_URL", "")
	t.Setenv("JWT_ACCESS_EXPIRY", "")
	t.Setenv("JWT_REFRESH_EXPIRY", "")
	t.Setenv("RATE_LIMIT_PER_MINUTE_IP", "")
	t.Setenv("RATE_LIMIT_PER_MINUTE_USER", "")

	cfg := Load()

	cases := []struct {
		name string
		got  string
		want string
	}{
		{"AppEnv", cfg.AppEnv, "development"},
		{"Port", cfg.Port, "8080"},
		{"EmailFrom", cfg.EmailFrom, "noreply@blueledger.io"},
		{"R2Bucket", cfg.R2Bucket, "blue-ledger-assets"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("got %q want %q", tc.got, tc.want)
			}
		})
	}

	if cfg.JWTAccessExpiry != 15*time.Minute {
		t.Errorf("JWTAccessExpiry: got %v want 15m", cfg.JWTAccessExpiry)
	}
	if cfg.JWTRefreshExpiry != 720*time.Hour {
		t.Errorf("JWTRefreshExpiry: got %v want 720h", cfg.JWTRefreshExpiry)
	}
	if cfg.RateLimitPerMinuteIP != 100 {
		t.Errorf("RateLimitPerMinuteIP: got %d want 100", cfg.RateLimitPerMinuteIP)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("PORT", "9090")
	t.Setenv("RATE_LIMIT_PER_MINUTE_IP", "50")
	t.Setenv("JWT_ACCESS_EXPIRY", "30m")

	cfg := Load()

	if cfg.AppEnv != "production" {
		t.Errorf("AppEnv: got %q want production", cfg.AppEnv)
	}
	if cfg.Port != "9090" {
		t.Errorf("Port: got %q want 9090", cfg.Port)
	}
	if cfg.RateLimitPerMinuteIP != 50 {
		t.Errorf("RateLimitPerMinuteIP: got %d want 50", cfg.RateLimitPerMinuteIP)
	}
	if cfg.JWTAccessExpiry != 30*time.Minute {
		t.Errorf("JWTAccessExpiry: got %v want 30m", cfg.JWTAccessExpiry)
	}
}

func TestIsDevelopment(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	cfg := Load()
	if !cfg.IsDevelopment() {
		t.Error("expected IsDevelopment()=true for APP_ENV=development")
	}
	if cfg.IsProduction() {
		t.Error("expected IsProduction()=false for APP_ENV=development")
	}
}

func TestIsProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	cfg := Load()
	if !cfg.IsProduction() {
		t.Error("expected IsProduction()=true for APP_ENV=production")
	}
	if cfg.IsDevelopment() {
		t.Error("expected IsDevelopment()=false for APP_ENV=production")
	}
}

func TestLoad_InvalidDuration_FallsBackToDefault(t *testing.T) {
	t.Setenv("JWT_ACCESS_EXPIRY", "not-a-duration")
	cfg := Load()
	// Should fall back to default (15m) rather than panicking.
	if cfg.JWTAccessExpiry != 15*time.Minute {
		t.Errorf("expected fallback to 15m, got %v", cfg.JWTAccessExpiry)
	}
}

func TestLoad_InvalidInt_FallsBackToDefault(t *testing.T) {
	t.Setenv("RATE_LIMIT_PER_MINUTE_IP", "not-a-number")
	cfg := Load()
	if cfg.RateLimitPerMinuteIP != 100 {
		t.Errorf("expected fallback to 100, got %d", cfg.RateLimitPerMinuteIP)
	}
}
