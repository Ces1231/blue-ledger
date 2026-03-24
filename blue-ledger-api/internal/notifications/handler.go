package notifications

import (
	"math"
	"net/http"
	"strconv"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// Handler handles notifications HTTP endpoints.
type Handler struct {
	svc NotificationsService
}

// NewHandler creates a new notifications handler.
func NewHandler(svc NotificationsService) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts notification routes on the given Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group, jwtMW echo.MiddlewareFunc) {
	g.GET("", h.List, jwtMW)
	g.PUT("/:id/read", h.MarkRead, jwtMW)
	g.PUT("/read-all", h.MarkAllRead, jwtMW)
	g.DELETE("/:id", h.Delete, jwtMW)
}

// List handles GET /notifications
func (h *Handler) List(c echo.Context) error {
	userID := auth.GetUserID(c)

	page := 1
	if v := c.QueryParam("page"); v != "" {
		if n, _ := strconv.Atoi(v); n > 0 {
			page = n
		}
	}
	perPage := 25

	notifications, total, err := h.svc.ListForUser(c.Request().Context(), userID, page, perPage)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get notifications")
	}
	if notifications == nil {
		notifications = []*Notification{}
	}

	pages := int(math.Ceil(float64(total) / float64(perPage)))
	if pages < 1 {
		pages = 1
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": notifications,
		"meta": map[string]interface{}{
			"total": total, "page": page, "per_page": perPage, "pages": pages,
		},
	})
}

// MarkRead handles PUT /notifications/:id/read
func (h *Handler) MarkRead(c echo.Context) error {
	userID := auth.GetUserID(c)
	notifID := c.Param("id")

	if err := h.svc.MarkRead(c.Request().Context(), notifID, userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to mark notification read")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": map[string]string{"message": "notification marked read"},
	})
}

// MarkAllRead handles PUT /notifications/read-all
func (h *Handler) MarkAllRead(c echo.Context) error {
	userID := auth.GetUserID(c)

	if err := h.svc.MarkAllRead(c.Request().Context(), userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to mark all notifications read")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": map[string]string{"message": "all notifications marked read"},
	})
}

// Delete handles DELETE /notifications/:id
func (h *Handler) Delete(c echo.Context) error {
	userID := auth.GetUserID(c)
	notifID := c.Param("id")

	if err := h.svc.Delete(c.Request().Context(), notifID, userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete notification")
	}

	return c.NoContent(http.StatusNoContent)
}
