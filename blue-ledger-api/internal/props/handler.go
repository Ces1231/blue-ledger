package props

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// Handler handles all props HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler creates a new props handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts props routes on the given Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group, jwtMiddleware echo.MiddlewareFunc) {
	g.Use(jwtMiddleware)
	g.GET("", h.List)
	g.POST("", h.Give)
	g.GET("/received", h.GetReceived)
}

// List handles GET /props
func (h *Handler) List(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	items, total, err := h.svc.List(c.Request().Context(), chapterID, page, perPage)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list props")
	}
	return c.JSON(http.StatusOK, map[string]any{
		"data": items,
		"meta": map[string]any{"total": total},
	})
}

// Give handles POST /props
func (h *Handler) Give(c echo.Context) error {
	var req GiveInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	chapterID := auth.GetChapterID(c)
	memberID := auth.GetMemberID(c)

	prop, err := h.svc.Give(c.Request().Context(), chapterID, memberID, req)
	if err != nil {
		if errors.Is(err, ErrSelfProps) {
			return echo.NewHTTPError(http.StatusBadRequest, "cannot give props to yourself")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to give props")
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": prop})
}

// GetReceived handles GET /props/received
func (h *Handler) GetReceived(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	memberID := auth.GetMemberID(c)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	items, total, err := h.svc.GetReceived(c.Request().Context(), chapterID, memberID, page, perPage)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get received props")
	}
	return c.JSON(http.StatusOK, map[string]any{
		"data": items,
		"meta": map[string]any{"total": total},
	})
}
