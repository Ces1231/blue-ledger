package servicelog

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// Handler handles all service log HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler creates a new service log handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts service log routes on the given Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group, jwtMiddleware echo.MiddlewareFunc) {
	g.Use(jwtMiddleware)
	g.GET("", h.List)
	g.POST("", h.Create)
	g.PUT("/:id/verify", h.Verify, auth.RoleGate("chair", "admin"))
	g.GET("/member/:id", h.GetByMember)
}

// List handles GET /service
func (h *Handler) List(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	items, total, err := h.svc.List(c.Request().Context(), chapterID, page, perPage)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list service log")
	}
	return c.JSON(http.StatusOK, map[string]any{
		"data": items,
		"meta": map[string]any{"total": total},
	})
}

// Create handles POST /service
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

	entry, err := h.svc.Create(c.Request().Context(), chapterID, memberID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to log service hours")
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": entry})
}

// Verify handles PUT /service/:id/verify
func (h *Handler) Verify(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	verifierID := auth.GetMemberID(c)
	id := c.Param("id")

	entry, err := h.svc.Verify(c.Request().Context(), chapterID, id, verifierID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "service log entry not found or already verified")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to verify service hours")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": entry})
}

// GetByMember handles GET /service/member/:id
func (h *Handler) GetByMember(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	memberID := c.Param("id")
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	items, total, err := h.svc.GetByMember(c.Request().Context(), chapterID, memberID, page, perPage)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get member service log")
	}
	return c.JSON(http.StatusOK, map[string]any{
		"data": items,
		"meta": map[string]any{"total": total},
	})
}
