package minutes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// Handler handles all meeting minutes HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler creates a new minutes handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts minutes routes on the given Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group, jwtMiddleware echo.MiddlewareFunc) {
	g.Use(jwtMiddleware)
	g.GET("", h.List)
	g.POST("", h.Create, auth.RoleGate("admin", "chair"))
	g.GET("/:id", h.Get)
	g.PUT("/:id", h.Update, auth.RoleGate("admin", "chair"))
	g.PUT("/:id/finalize", h.Finalize, auth.RoleGate("admin", "chair"))
}

// List handles GET /minutes
func (h *Handler) List(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	items, total, err := h.svc.List(c.Request().Context(), chapterID, page, perPage)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list minutes")
	}
	return c.JSON(http.StatusOK, map[string]any{
		"data": items,
		"meta": map[string]any{"total": total},
	})
}

// Create handles POST /minutes
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

	mn, err := h.svc.Create(c.Request().Context(), chapterID, memberID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create minutes")
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": mn})
}

// Get handles GET /minutes/:id
func (h *Handler) Get(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	id := c.Param("id")

	mn, err := h.svc.Get(c.Request().Context(), chapterID, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "minutes not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get minutes")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": mn})
}

// Update handles PUT /minutes/:id
func (h *Handler) Update(c echo.Context) error {
	var req UpdateInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	chapterID := auth.GetChapterID(c)
	id := c.Param("id")

	mn, err := h.svc.Update(c.Request().Context(), chapterID, id, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "minutes not found or already finalized")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update minutes")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": mn})
}

// Finalize handles PUT /minutes/:id/finalize
func (h *Handler) Finalize(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	memberID := auth.GetMemberID(c)
	id := c.Param("id")

	mn, err := h.svc.Finalize(c.Request().Context(), chapterID, id, memberID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "minutes not found or already finalized")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to finalize minutes")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": mn})
}
