package milestones

import (
	"errors"
	"net/http"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// Handler handles milestones HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler creates a new milestones handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts milestones routes on the given Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group, jwtMiddleware echo.MiddlewareFunc) {
	g.Use(jwtMiddleware)
	g.GET("", h.List)
	g.POST("", h.Create)
	g.DELETE("/:id", h.Delete, auth.RoleGate("admin", "chair"))
}

// List handles GET /milestones
// Optional query param: ?member_id=<uuid> to filter by member.
func (h *Handler) List(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	memberID := c.QueryParam("member_id")

	items, err := h.svc.List(c.Request().Context(), chapterID, memberID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list milestones")
	}
	if items == nil {
		items = []*Milestone{}
	}
	return c.JSON(http.StatusOK, map[string]any{"data": items})
}

// Create handles POST /milestones
func (h *Handler) Create(c echo.Context) error {
	var req CreateInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	chapterID := auth.GetChapterID(c)

	item, err := h.svc.Create(c.Request().Context(), chapterID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create milestone")
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": item})
}

// Delete handles DELETE /milestones/:id
func (h *Handler) Delete(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	id := c.Param("id")

	if err := h.svc.Delete(c.Request().Context(), chapterID, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "milestone not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete milestone")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": map[string]string{"message": "milestone deleted"}})
}
