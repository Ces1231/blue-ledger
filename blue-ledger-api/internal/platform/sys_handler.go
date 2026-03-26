package platform

import (
	"net/http"
	"strconv"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// SysHandler handles sysadmin platform HTTP endpoints.
type SysHandler struct {
	svc SysService
}

// NewSysHandler creates a new sysadmin handler.
func NewSysHandler(svc SysService) *SysHandler {
	return &SysHandler{svc: svc}
}

// RegisterSysRoutes mounts sysadmin routes.
func (h *SysHandler) RegisterSysRoutes(g *echo.Group, jwtMiddleware echo.MiddlewareFunc) {
	sys := g.Group("/sys", jwtMiddleware, sysadminOnly)
	sys.GET("/chapters", h.ListChapters)
	sys.POST("/chapters", h.CreateChapter)
	sys.GET("/chapters/:id", h.GetChapter)
	sys.PUT("/chapters/:id/subscription", h.UpdateSubscription)
	sys.DELETE("/chapters/:id", h.DeleteChapter)
	sys.GET("/users", h.ListUsers)
	sys.GET("/audit-log", h.GetAuditLog)
	sys.POST("/xp/recalculate-all", h.RecalculateAllXP)
}

// sysadminOnly middleware restricts to sysadmin role.
func sysadminOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !auth.IsSysadmin(c) {
			return echo.NewHTTPError(http.StatusForbidden, "sysadmin access required")
		}
		return next(c)
	}
}

// ListChapters handles GET /sys/chapters
func (h *SysHandler) ListChapters(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	chapters, total, err := h.svc.ListChapters(c.Request().Context(), page, perPage)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list chapters")
	}
	return c.JSON(http.StatusOK, map[string]any{
		"data": chapters,
		"meta": map[string]any{"total": total},
	})
}

// CreateChapter handles POST /sys/chapters
func (h *SysHandler) CreateChapter(c echo.Context) error {
	var req CreateChapterInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	chapter, err := h.svc.CreateChapter(c.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create chapter")
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": chapter})
}

// GetChapter handles GET /sys/chapters/:id
func (h *SysHandler) GetChapter(c echo.Context) error {
	id := c.Param("id")
	chapter, err := h.svc.GetChapter(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get chapter")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": chapter})
}

// UpdateSubscription handles PUT /sys/chapters/:id/subscription
func (h *SysHandler) UpdateSubscription(c echo.Context) error {
	var req UpdateSubscriptionInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	id := c.Param("id")
	chapter, err := h.svc.UpdateSubscription(c.Request().Context(), id, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update subscription")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": chapter})
}

// DeleteChapter handles DELETE /sys/chapters/:id
func (h *SysHandler) DeleteChapter(c echo.Context) error {
	id := c.Param("id")
	if err := h.svc.DeleteChapter(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete chapter")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": map[string]string{"message": "chapter deleted"}})
}

// ListUsers handles GET /sys/users
func (h *SysHandler) ListUsers(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	users, total, err := h.svc.ListUsers(c.Request().Context(), page, perPage)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list users")
	}
	return c.JSON(http.StatusOK, map[string]any{
		"data": users,
		"meta": map[string]any{"total": total},
	})
}

// GetAuditLog handles GET /sys/audit-log
func (h *SysHandler) GetAuditLog(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	entries, total, err := h.svc.GetAuditLog(c.Request().Context(), page, perPage)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get audit log")
	}
	return c.JSON(http.StatusOK, map[string]any{
		"data": entries,
		"meta": map[string]any{"total": total},
	})
}

// RecalculateAllXP handles POST /sys/xp/recalculate-all
func (h *SysHandler) RecalculateAllXP(c echo.Context) error {
	count, err := h.svc.RecalculateAllXP(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "XP recalculation failed")
	}
	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{"members_updated": count},
	})
}
