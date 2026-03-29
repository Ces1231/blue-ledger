package fundraising

import (
	"errors"
	"net/http"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// Handler handles fundraising HTTP endpoints.
type Handler struct{ svc Service }

// NewHandler creates a new fundraising handler.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// RegisterRoutes mounts fundraising routes.
func (h *Handler) RegisterRoutes(g *echo.Group, jwtMW echo.MiddlewareFunc) {
	g.Use(jwtMW)
	g.GET("", h.List)
	g.GET("/:id", h.GetByID)
	g.POST("", h.Create, auth.RoleGate("admin"))
	g.PUT("/:id", h.Update, auth.RoleGate("admin"))
	g.DELETE("/:id", h.Delete, auth.RoleGate("admin"))
	g.POST("/:id/contribute", h.Contribute)
}

func (h *Handler) List(c echo.Context) error {
	items, err := h.svc.List(c.Request().Context(), auth.GetChapterID(c))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list campaigns")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) GetByID(c echo.Context) error {
	item, err := h.svc.GetByID(c.Request().Context(), auth.GetChapterID(c), c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "campaign not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get campaign")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) Create(c echo.Context) error {
	var req CreateInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}
	item, err := h.svc.Create(c.Request().Context(), auth.GetChapterID(c), auth.GetMemberID(c), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create campaign")
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) Update(c echo.Context) error {
	var req UpdateInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	item, err := h.svc.Update(c.Request().Context(), auth.GetChapterID(c), c.Param("id"), req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "campaign not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update campaign")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) Delete(c echo.Context) error {
	if err := h.svc.Delete(c.Request().Context(), auth.GetChapterID(c), c.Param("id")); err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "campaign not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete campaign")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": map[string]string{"message": "campaign deleted"}})
}

func (h *Handler) Contribute(c echo.Context) error {
	var req ContributeInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}
	memberID := auth.GetMemberID(c)
	contrib, err := h.svc.Contribute(c.Request().Context(), c.Param("id"), memberID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to record contribution")
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": contrib})
}
