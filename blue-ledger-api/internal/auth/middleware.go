package auth

import (
	"net/http"
	"strings"

	"github.com/ces1231/blue-ledger-api/pkg/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// ContextKey type avoids collisions in context values.
type ContextKey string

const (
	ContextKeyUserID    ContextKey = "user_id"
	ContextKeyChapterID ContextKey = "chapter_id"
	ContextKeyMemberID  ContextKey = "member_id"
	ContextKeyRole      ContextKey = "role"
	ContextKeyIsSysadmin ContextKey = "is_sysadmin"
)

// UserFromContext extracts the authenticated user claims from an Echo context.
func UserFromContext(c echo.Context) *Claims {
	if claims, ok := c.Get(string(ContextKeyUserID)).(*Claims); ok {
		return claims
	}
	return nil
}

// JWTMiddleware validates the Bearer token in the Authorization header.
// On success, the parsed Claims are stored in the Echo context.
// Routes that set `optional: true` in the skip func will pass through unauthenticated.
func JWTMiddleware(tm *TokenManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "authorization header required")
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}

			claims, err := tm.ValidateAccessToken(parts[1])
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			// Store individual fields for easy access
			c.Set(string(ContextKeyUserID), claims.UserID)
			c.Set(string(ContextKeyChapterID), claims.ChapterID)
			c.Set(string(ContextKeyMemberID), claims.MemberID)
			c.Set(string(ContextKeyRole), claims.Role)
			c.Set(string(ContextKeyIsSysadmin), claims.IsSysadmin)
			// Also store the full claims object
			c.Set("claims", claims)

			return next(c)
		}
	}
}

// TenantSetter sets the PostgreSQL RLS session variable for the current request.
// Must run after JWTMiddleware so chapter_id is available in the context.
func TenantSetter(pool *pgxpool.Pool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			chapterID, ok := c.Get(string(ContextKeyChapterID)).(string)
			if !ok || chapterID == "" {
				// Sysadmin requests or unauthenticated paths skip tenant setting
				return next(c)
			}

			if err := db.SetChapterContext(c.Request().Context(), pool, chapterID); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to set tenant context")
			}

			return next(c)
		}
	}
}

// RoleGate returns middleware that restricts access to the listed roles.
// Sysadmin always bypasses role checks.
func RoleGate(allowedRoles ...string) echo.MiddlewareFunc {
	allowed := make(map[string]bool, len(allowedRoles))
	for _, r := range allowedRoles {
		allowed[r] = true
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, _ := c.Get(string(ContextKeyRole)).(string)
			isSysadmin, _ := c.Get(string(ContextKeyIsSysadmin)).(bool)

			if isSysadmin || allowed[role] {
				return next(c)
			}

			return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
		}
	}
}

// GetChapterID is a convenience helper that pulls chapter_id from the Echo context.
func GetChapterID(c echo.Context) string {
	v, _ := c.Get(string(ContextKeyChapterID)).(string)
	return v
}

// GetUserID is a convenience helper that pulls user_id from the Echo context.
func GetUserID(c echo.Context) string {
	v, _ := c.Get(string(ContextKeyUserID)).(string)
	return v
}

// GetMemberID is a convenience helper that pulls member_id from the Echo context.
func GetMemberID(c echo.Context) string {
	v, _ := c.Get(string(ContextKeyMemberID)).(string)
	return v
}

// GetRole returns the authenticated user's role from the Echo context.
func GetRole(c echo.Context) string {
	v, _ := c.Get(string(ContextKeyRole)).(string)
	return v
}

// IsSysadmin returns true if the current user is a platform sysadmin.
func IsSysadmin(c echo.Context) bool {
	v, _ := c.Get(string(ContextKeyIsSysadmin)).(bool)
	return v
}
