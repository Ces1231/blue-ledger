package store

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// Handler handles all store HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler creates a new store handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts store routes on the given Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group, jwtMiddleware echo.MiddlewareFunc) {
	g.Use(jwtMiddleware)
	g.GET("", h.List)
	g.POST("", h.Create, auth.RoleGate("admin"))
	g.POST("/:id/purchase", h.Purchase)
	g.GET("/orders", h.ListOrders, auth.RoleGate("admin"))
}

// List handles GET /store
func (h *Handler) List(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	items, err := h.svc.List(c.Request().Context(), chapterID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list store items")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": items})
}

// Create handles POST /store
func (h *Handler) Create(c echo.Context) error {
	var req CreateItemInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	chapterID := auth.GetChapterID(c)
	item, err := h.svc.Create(c.Request().Context(), chapterID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create store item")
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": item})
}

// Purchase handles POST /store/:id/purchase
func (h *Handler) Purchase(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	memberID := auth.GetMemberID(c)
	itemID := c.Param("id")

	order, err := h.svc.Purchase(c.Request().Context(), chapterID, memberID, itemID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "store item not found")
		}
		if errors.Is(err, ErrInsufficientXP) {
			return echo.NewHTTPError(http.StatusPaymentRequired, "insufficient XP balance for this purchase")
		}
		if errors.Is(err, ErrItemUnavailable) {
			return echo.NewHTTPError(http.StatusConflict, "item is not available or out of stock")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "purchase failed")
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": order})
}

// ListOrders handles GET /store/orders
func (h *Handler) ListOrders(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	orders, total, err := h.svc.ListOrders(c.Request().Context(), chapterID, page, perPage)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list orders")
	}
	return c.JSON(http.StatusOK, map[string]any{
		"data": orders,
		"meta": map[string]any{"total": total},
	})
}
