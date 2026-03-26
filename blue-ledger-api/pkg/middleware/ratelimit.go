package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

const (
	defaultLimit      = 100
	defaultWindowSecs = 60
)

// RateLimiter returns an Echo middleware that enforces a fixed-window rate limit
// using Redis. If client is nil (Redis unavailable), the middleware is a no-op.
//
// Limit:  max requests per window (default 100)
// Window: rolling window duration in seconds (default 60s)
//
// Key format: rate:<ip>:<window_bucket>
// Response headers: X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset
func RateLimiter(client *redis.Client) echo.MiddlewareFunc {
	return RateLimiterWithConfig(client, defaultLimit, defaultWindowSecs)
}

// RateLimiterWithConfig is the configurable version of RateLimiter.
func RateLimiterWithConfig(client *redis.Client, limit int, windowSecs int) echo.MiddlewareFunc {
	if client == nil {
		// Redis unavailable — pass through without rate limiting.
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}

	window := time.Duration(windowSecs) * time.Second

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			now := time.Now()
			bucket := now.Unix() / int64(windowSecs)
			key := fmt.Sprintf("rate:%s:%d", ip, bucket)

			ctx := context.Background()

			// Increment counter atomically; set TTL on first write.
			count, err := client.Incr(ctx, key).Result()
			if err != nil {
				// Redis error — fail open (don't block the request).
				return next(c)
			}

			if count == 1 {
				client.Expire(ctx, key, window+time.Second) //nolint:errcheck
			}

			remaining := int64(limit) - count
			if remaining < 0 {
				remaining = 0
			}
			reset := (bucket+1)*int64(windowSecs)

			c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			c.Response().Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			c.Response().Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", reset))

			if count > int64(limit) {
				return echo.NewHTTPError(http.StatusTooManyRequests, "rate limit exceeded")
			}

			return next(c)
		}
	}
}
