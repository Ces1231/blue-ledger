package health

import (
	"net/http"
	"strconv"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// Handler handles all chapter health analytics HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler creates a new health handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts health analytics routes on the given Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group, jwtMiddleware echo.MiddlewareFunc) {
	g.Use(jwtMiddleware, auth.RoleGate("admin", "chair", "pia"))
	g.GET("", h.GetOverview)
	g.GET("/attendance", h.GetAttendance)
	g.GET("/xp", h.GetXPDistribution)
	g.GET("/service", h.GetServiceStats)
	g.GET("/dues", h.GetDuesBreakdown)
}

// GetOverview handles GET /health
func (h *Handler) GetOverview(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	overview, err := h.svc.GetOverview(c.Request().Context(), chapterID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get chapter health overview")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": overview})
}

// GetAttendance handles GET /health/attendance
func (h *Handler) GetAttendance(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	stats, err := h.svc.GetAttendance(c.Request().Context(), chapterID, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get attendance stats")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": stats})
}

// GetXPDistribution handles GET /health/xp
func (h *Handler) GetXPDistribution(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	stats, err := h.svc.GetXPDistribution(c.Request().Context(), chapterID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get XP distribution")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": stats})
}

// GetServiceStats handles GET /health/service
func (h *Handler) GetServiceStats(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	months, _ := strconv.Atoi(c.QueryParam("months"))

	stats, err := h.svc.GetServiceStats(c.Request().Context(), chapterID, months)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get service stats")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": stats})
}

// GetDuesBreakdown handles GET /health/dues
func (h *Handler) GetDuesBreakdown(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	stats, err := h.svc.GetDuesBreakdown(c.Request().Context(), chapterID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get dues breakdown")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": stats})
}
