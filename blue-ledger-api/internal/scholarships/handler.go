package scholarships

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// Handler handles all scholarship HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler creates a new scholarships handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts scholarship routes on the given Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group, jwtMiddleware echo.MiddlewareFunc) {
	g.Use(jwtMiddleware, auth.RoleGate("pia", "admin"))
	g.GET("", h.List)
	g.POST("", h.Create)
	g.GET("/:id", h.Get)
	g.PUT("/:id", h.Update)
}

// List handles GET /scholarships
func (h *Handler) List(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	items, total, err := h.svc.List(c.Request().Context(), chapterID, page, perPage)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list scholarship applications")
	}
	return c.JSON(http.StatusOK, map[string]any{
		"data": items,
		"meta": map[string]any{"total": total},
	})
}

// Create handles POST /scholarships
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

	app, err := h.svc.Create(c.Request().Context(), chapterID, memberID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create scholarship application")
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": app})
}

// Get handles GET /scholarships/:id
func (h *Handler) Get(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	id := c.Param("id")

	app, err := h.svc.Get(c.Request().Context(), chapterID, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "scholarship application not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get scholarship application")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": app})
}

// Update handles PUT /scholarships/:id
func (h *Handler) Update(c echo.Context) error {
	var req UpdateInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	chapterID := auth.GetChapterID(c)
	id := c.Param("id")

	app, err := h.svc.Update(c.Request().Context(), chapterID, id, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "scholarship application not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update scholarship application")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": app})
}
