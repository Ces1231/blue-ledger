package announcements

import (
	"errors"
	"net/http"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// Handler handles all announcements HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler creates a new announcements handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts announcement routes on the given Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group, jwtMiddleware echo.MiddlewareFunc) {
	g.Use(jwtMiddleware)
	g.GET("", h.List)
	g.POST("", h.Create, auth.RoleGate("admin", "pia", "chair"))
	g.PUT("/:id", h.Update, auth.RoleGate("admin", "pia", "chair"))
	g.DELETE("/:id", h.Delete, auth.RoleGate("admin"))
	g.PUT("/:id/pin", h.Pin, auth.RoleGate("admin", "pia"))
}

// List handles GET /announcements
func (h *Handler) List(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	items, err := h.svc.List(c.Request().Context(), chapterID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list announcements")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": items})
}

// Create handles POST /announcements
func (h *Handler) Create(c echo.Context) error {
	var req CreateInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	chapterID := auth.GetChapterID(c)
	memberID := auth.GetMemberID(c)

	item, err := h.svc.Create(c.Request().Context(), chapterID, memberID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create announcement")
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": item})
}

// Update handles PUT /announcements/:id
func (h *Handler) Update(c echo.Context) error {
	var req UpdateInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	chapterID := auth.GetChapterID(c)
	id := c.Param("id")

	item, err := h.svc.Update(c.Request().Context(), chapterID, id, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "announcement not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update announcement")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": item})
}

// Delete handles DELETE /announcements/:id
func (h *Handler) Delete(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	id := c.Param("id")

	if err := h.svc.Delete(c.Request().Context(), chapterID, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "announcement not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete announcement")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": map[string]string{"message": "announcement deleted"}})
}

// Pin handles PUT /announcements/:id/pin
func (h *Handler) Pin(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	id := c.Param("id")

	var body struct {
		Pinned bool `json:"pinned"`
	}
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	item, err := h.svc.Pin(c.Request().Context(), chapterID, id, body.Pinned)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "announcement not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to pin announcement")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": item})
}
