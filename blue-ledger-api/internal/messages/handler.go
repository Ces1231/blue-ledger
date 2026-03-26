package messages

import (
	"errors"
	"net/http"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// Handler handles all messages HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler creates a new messages handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts message routes on the given Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group, jwtMiddleware echo.MiddlewareFunc) {
	g.Use(jwtMiddleware)
	g.GET("/threads", h.ListThreads)
	g.POST("/threads", h.CreateThread)
	g.GET("/threads/:id", h.GetThread)
	g.POST("/threads/:id/messages", h.SendMessage)
}

// ListThreads handles GET /messages/threads
func (h *Handler) ListThreads(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	memberID := auth.GetMemberID(c)

	threads, err := h.svc.ListThreads(c.Request().Context(), chapterID, memberID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list threads")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": threads})
}

// CreateThread handles POST /messages/threads
func (h *Handler) CreateThread(c echo.Context) error {
	var req CreateThreadInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	chapterID := auth.GetChapterID(c)
	memberID := auth.GetMemberID(c)

	thread, err := h.svc.CreateThread(c.Request().Context(), chapterID, memberID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create thread")
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": thread})
}

// GetThread handles GET /messages/threads/:id
func (h *Handler) GetThread(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	memberID := auth.GetMemberID(c)
	id := c.Param("id")

	msgs, err := h.svc.GetThread(c.Request().Context(), chapterID, id, memberID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "thread not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get thread")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": msgs})
}

// SendMessage handles POST /messages/threads/:id/messages
func (h *Handler) SendMessage(c echo.Context) error {
	var req SendMessageInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	chapterID := auth.GetChapterID(c)
	senderID := auth.GetMemberID(c)
	id := c.Param("id")

	msg, err := h.svc.SendMessage(c.Request().Context(), chapterID, id, senderID, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "thread not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to send message")
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": msg})
}
