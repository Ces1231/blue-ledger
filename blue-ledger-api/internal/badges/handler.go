package badges

import (
	"errors"
	"net/http"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// Handler handles all badges HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler creates a new badges handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts badge routes on the given Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group, jwtMW echo.MiddlewareFunc) {
	g.Use(jwtMW)
	g.GET("", h.List)
	g.GET("/mine", h.Mine)
	g.GET("/:id", h.GetByID)
	g.POST("", h.Create, auth.RoleGate("admin"))
	g.PUT("/:id", h.Update, auth.RoleGate("admin"))
	g.DELETE("/:id", h.Delete, auth.RoleGate("admin"))
	g.POST("/:id/award", h.Award, auth.RoleGate("admin"))
}

// Mine handles GET /badges/mine — returns all badges earned by the calling member.
func (h *Handler) Mine(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	memberID := auth.GetMemberID(c)
	items, err := h.svc.Mine(c.Request().Context(), chapterID, memberID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list my badges")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": items})
}

// List handles GET /badges
func (h *Handler) List(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	items, err := h.svc.List(c.Request().Context(), chapterID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list badges")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": items})
}

// GetByID handles GET /badges/:id
func (h *Handler) GetByID(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	item, err := h.svc.GetByID(c.Request().Context(), chapterID, c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "badge not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get badge")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": item})
}

// Create handles POST /badges
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
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create badge")
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": item})
}

// Update handles PUT /badges/:id
func (h *Handler) Update(c echo.Context) error {
	var req UpdateInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	chapterID := auth.GetChapterID(c)
	item, err := h.svc.Update(c.Request().Context(), chapterID, c.Param("id"), req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "badge not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update badge")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": item})
}

// Delete handles DELETE /badges/:id
func (h *Handler) Delete(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	if err := h.svc.Delete(c.Request().Context(), chapterID, c.Param("id")); err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "badge not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete badge")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": map[string]string{"message": "badge deleted"}})
}

// Award handles POST /badges/:id/award
func (h *Handler) Award(c echo.Context) error {
	var req AwardInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}
	chapterID := auth.GetChapterID(c)
	memberID := auth.GetMemberID(c)
	mb, err := h.svc.Award(c.Request().Context(), chapterID, c.Param("id"), req.MemberID, memberID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "badge not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to award badge")
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": mb})
}
