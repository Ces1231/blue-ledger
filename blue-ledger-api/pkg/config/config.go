package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Server
	AppEnv     string
	Port       string
	APIBaseURL string

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// JWT
	JWTPrivateKeyPath string
	JWTPublicKeyPath  string
	JWTAccessExpiry   time.Duration
	JWTRefreshExpiry  time.Duration

	// Magic Links
	MagicLinkExpiry     time.Duration
	MagicLinkHMACSecret string

	// Email (Resend)
	ResendAPIKey    string
	EmailFrom       string
	EmailFromName   string

	// Storage (Cloudflare R2)
	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2Bucket          string
	R2PublicURL       string

	// Stripe
	StripeSecretKey       string
	StripeWebhookSecret   string
	StripeStarterPriceID  string
	StripeGrowthPriceID   string
	StripeProPriceID      string

	// AI Assistant
	AnthropicAPIKey string
	OpenAIAPIKey    string

	// Sentry
	SentryDSN string

	// CORS
	AllowedOrigins string

	// Rate limiting
	RateLimitPerMinuteIP   int
	RateLimitPerMinuteUser int
}

// Load reads configuration from environment variables, applying defaults where appropriate.
func Load() *Config {
	return &Config{
		AppEnv:     getEnv("APP_ENV", "development"),
		Port:       getEnv("PORT", "8080"),
		APIBaseURL: getEnv("API_BASE_URL", "http://localhost:8080"),

		DatabaseURL: getEnv("DATABASE_URL", "postgres://blue_ledger:localpassword@localhost:5432/blue_ledger?sslmode=disable"),

		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379"),

		JWTPrivateKeyPath: getEnv("JWT_PRIVATE_KEY_PATH", "./keys/private.pem"),
		JWTPublicKeyPath:  getEnv("JWT_PUBLIC_KEY_PATH", "./keys/public.pem"),
		JWTAccessExpiry:   getDuration("JWT_ACCESS_EXPIRY", 15*time.Minute),
		JWTRefreshExpiry:  getDuration("JWT_REFRESH_EXPIRY", 720*time.Hour),

		MagicLinkExpiry:     getDuration("MAGIC_LINK_EXPIRY", 15*time.Minute),
		MagicLinkHMACSecret: getEnv("MAGIC_LINK_HMAC_SECRET", ""),

		ResendAPIKey:  getEnv("RESEND_API_KEY", ""),
		EmailFrom:     getEnv("EMAIL_FROM", "noreply@blueledger.io"),
		EmailFromName: getEnv("EMAIL_FROM_NAME", "The Blue Ledger"),

		R2AccountID:       getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2Bucket:          getEnv("R2_BUCKET", "blue-ledger-assets"),
		R2PublicURL:       getEnv("R2_PUBLIC_URL", "https://assets.blueledger.io"),

		StripeSecretKey:      getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret:  getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripeStarterPriceID: getEnv("STRIPE_STARTER_PRICE_ID", ""),
		StripeGrowthPriceID:  getEnv("STRIPE_GROWTH_PRICE_ID", ""),
		StripeProPriceID:     getEnv("STRIPE_PRO_PRICE_ID", ""),

		AnthropicAPIKey: getEnv("ANTHROPIC_API_KEY", ""),
		OpenAIAPIKey:    getEnv("OPENAI_API_KEY", ""),

		SentryDSN: getEnv("SENTRY_DSN", ""),

		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:3001"),

		RateLimitPerMinuteIP:   getInt("RATE_LIMIT_PER_MINUTE_IP", 100),
		RateLimitPerMinuteUser: getInt("RATE_LIMIT_PER_MINUTE_USER", 1000),
	}
}

// IsDevelopment returns true when running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}

// IsProduction returns true when running in production mode.
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getDuration(key string, defaultVal time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return defaultVal
	}
	return d
}

func getInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return n
}
