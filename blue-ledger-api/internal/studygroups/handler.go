package studygroups

import (
	"errors"
	"net/http"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// Handler handles study-groups HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler creates a new study-groups handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts study-groups routes on the given Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group, jwtMiddleware echo.MiddlewareFunc) {
	g.Use(jwtMiddleware)
	g.GET("", h.List)
	g.POST("", h.Create)
	g.POST("/:id/join", h.Join)
	g.DELETE("/:id", h.Delete, auth.RoleGate("admin", "chair"))
}

// List handles GET /study-groups
func (h *Handler) List(c echo.Context) error {
	chapterID := auth.GetChapterID(c)

	items, err := h.svc.List(c.Request().Context(), chapterID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list study groups")
	}
	if items == nil {
		items = []*StudyGroup{}
	}
	return c.JSON(http.StatusOK, map[string]any{"data": items})
}

// Create handles POST /study-groups
func (h *Handler) Create(c echo.Context) error {
	var req CreateInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	chapterID := auth.GetChapterID(c)
	hostID := auth.GetMemberID(c)

	item, err := h.svc.Create(c.Request().Context(), chapterID, hostID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create study group")
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": item})
}

// Join handles POST /study-groups/:id/join
// Appends the caller's member_id to member_ids; operation is idempotent.
func (h *Handler) Join(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	memberID := auth.GetMemberID(c)
	id := c.Param("id")

	if err := h.svc.Join(c.Request().Context(), chapterID, id, memberID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "study group not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to join study group")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": map[string]string{"message": "joined study group"}})
}

// Delete handles DELETE /study-groups/:id (admin / chair only)
func (h *Handler) Delete(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	id := c.Param("id")

	if err := h.svc.Delete(c.Request().Context(), chapterID, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "study group not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete study group")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": map[string]string{"message": "study group deleted"}})
}
