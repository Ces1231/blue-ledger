package committees

import (
	"errors"
	"net/http"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// Handler handles committee HTTP endpoints.
type Handler struct{ svc Service }

// NewHandler creates a new committees handler.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// RegisterRoutes mounts committee routes.
func (h *Handler) RegisterRoutes(g *echo.Group, jwtMW echo.MiddlewareFunc) {
	g.Use(jwtMW)
	g.GET("", h.List)
	g.GET("/:id", h.GetByID)
	g.POST("", h.Create, auth.RoleGate("admin"))
	g.PUT("/:id", h.Update, auth.RoleGate("admin"))
	g.DELETE("/:id", h.Delete, auth.RoleGate("admin"))
	g.POST("/:id/members", h.AddMember, auth.RoleGate("admin", "chair"))
	g.DELETE("/:id/members/:member_id", h.RemoveMember, auth.RoleGate("admin", "chair"))
}

func (h *Handler) List(c echo.Context) error {
	items, err := h.svc.List(c.Request().Context(), auth.GetChapterID(c))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list committees")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) GetByID(c echo.Context) error {
	item, err := h.svc.GetByID(c.Request().Context(), auth.GetChapterID(c), c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "committee not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get committee")
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
	item, err := h.svc.Create(c.Request().Context(), auth.GetChapterID(c), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create committee")
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
			return echo.NewHTTPError(http.StatusNotFound, "committee not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update committee")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) Delete(c echo.Context) error {
	if err := h.svc.Delete(c.Request().Context(), auth.GetChapterID(c), c.Param("id")); err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "committee not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete committee")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": map[string]string{"message": "committee deleted"}})
}

func (h *Handler) AddMember(c echo.Context) error {
	var req AddMemberInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}
	if err := h.svc.AddMember(c.Request().Context(), c.Param("id"), req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add member")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": map[string]string{"message": "member added"}})
}

func (h *Handler) RemoveMember(c echo.Context) error {
	if err := h.svc.RemoveMember(c.Request().Context(), c.Param("id"), c.Param("member_id")); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to remove member")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": map[string]string{"message": "member removed"}})
}
