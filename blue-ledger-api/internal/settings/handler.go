package settings

import (
	"errors"
	"net/http"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// Handler handles all chapter settings HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler creates a new settings handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts settings routes on the given Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group, jwtMiddleware echo.MiddlewareFunc) {
	g.Use(jwtMiddleware)
	g.GET("", h.GetSettings)
	g.PUT("", h.UpdateSettings, auth.RoleGate("admin"))
	g.GET("/point-economy", h.GetPointEconomy)
	g.PUT("/point-economy", h.UpdatePointEconomy, auth.RoleGate("admin"))
	g.POST("/zeffy-config", h.SaveZeffyConfig, auth.RoleGate("admin"))
}

// GetSettings handles GET /settings
func (h *Handler) GetSettings(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	cfg, err := h.svc.GetSettings(c.Request().Context(), chapterID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "chapter not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get settings")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": cfg})
}

// UpdateSettings handles PUT /settings
func (h *Handler) UpdateSettings(c echo.Context) error {
	var req UpdateSettingsInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	chapterID := auth.GetChapterID(c)
	cfg, err := h.svc.UpdateSettings(c.Request().Context(), chapterID, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "chapter not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update settings")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": cfg})
}

// GetPointEconomy handles GET /settings/point-economy
func (h *Handler) GetPointEconomy(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	pe, err := h.svc.GetPointEconomy(c.Request().Context(), chapterID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get point economy")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": pe})
}

// UpdatePointEconomy handles PUT /settings/point-economy
func (h *Handler) UpdatePointEconomy(c echo.Context) error {
	var req PointEconomy
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	chapterID := auth.GetChapterID(c)
	pe, err := h.svc.UpdatePointEconomy(c.Request().Context(), chapterID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update point economy")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": pe})
}

// SaveZeffyConfig handles POST /settings/zeffy-config
func (h *Handler) SaveZeffyConfig(c echo.Context) error {
	var req ZeffyConfigInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	chapterID := auth.GetChapterID(c)
	cfg, err := h.svc.SaveZeffyConfig(c.Request().Context(), chapterID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save Zeffy config")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": cfg})
}
