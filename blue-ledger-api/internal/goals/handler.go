package goals

import (
	"errors"
	"net/http"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// Handler handles all goals HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler creates a new goals handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts goals routes on the given Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group, jwtMiddleware echo.MiddlewareFunc) {
	g.Use(jwtMiddleware)
	g.GET("", h.List)
	g.POST("", h.Create, auth.RoleGate("admin"))
	g.PUT("/:id", h.Update, auth.RoleGate("admin"))
	g.PUT("/:id/progress", h.UpdateProgress)
}

// List handles GET /goals
func (h *Handler) List(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	items, err := h.svc.List(c.Request().Context(), chapterID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list goals")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": items})
}

// Create handles POST /goals
func (h *Handler) Create(c echo.Context) error {
	var req CreateInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	chapterID := auth.GetChapterID(c)
	goal, err := h.svc.Create(c.Request().Context(), chapterID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create goal")
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": goal})
}

// Update handles PUT /goals/:id
func (h *Handler) Update(c echo.Context) error {
	var req UpdateInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	chapterID := auth.GetChapterID(c)
	id := c.Param("id")

	goal, err := h.svc.Update(c.Request().Context(), chapterID, id, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "goal not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update goal")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": goal})
}

// UpdateProgress handles PUT /goals/:id/progress
func (h *Handler) UpdateProgress(c echo.Context) error {
	var req ProgressInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	chapterID := auth.GetChapterID(c)
	id := c.Param("id")

	goal, err := h.svc.UpdateProgress(c.Request().Context(), chapterID, id, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "goal not found or inactive")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update goal progress")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": goal})
}
