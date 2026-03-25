package intake

import (
	"errors"
	"net/http"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// Handler handles all intake HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler creates a new intake handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts intake routes on the given Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group, jwtMiddleware echo.MiddlewareFunc) {
	g.Use(jwtMiddleware, auth.RoleGate("chair", "admin"))
	g.GET("", h.List)
	g.POST("", h.Create)
	g.PUT("/:id", h.Update)
	g.PUT("/:id/stage", h.UpdateStage)
	g.DELETE("/:id", h.Archive)
}

// List handles GET /intake
func (h *Handler) List(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	items, err := h.svc.List(c.Request().Context(), chapterID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list prospects")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": items})
}

// Create handles POST /intake
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

	prospect, err := h.svc.Create(c.Request().Context(), chapterID, memberID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create prospect")
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": prospect})
}

// Update handles PUT /intake/:id
func (h *Handler) Update(c echo.Context) error {
	var req UpdateInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	chapterID := auth.GetChapterID(c)
	id := c.Param("id")

	prospect, err := h.svc.Update(c.Request().Context(), chapterID, id, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "prospect not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update prospect")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": prospect})
}

// UpdateStage handles PUT /intake/:id/stage
func (h *Handler) UpdateStage(c echo.Context) error {
	var body struct {
		Stage string `json:"stage" validate:"required"`
	}
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&body); err != nil {
		return err
	}

	chapterID := auth.GetChapterID(c)
	id := c.Param("id")

	prospect, err := h.svc.UpdateStage(c.Request().Context(), chapterID, id, body.Stage)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "prospect not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update prospect stage")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": prospect})
}

// Archive handles DELETE /intake/:id
func (h *Handler) Archive(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	id := c.Param("id")

	if err := h.svc.Archive(c.Request().Context(), chapterID, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "prospect not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to archive prospect")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": map[string]string{"message": "prospect archived"}})
}
