package challenges

import (
	"errors"
	"net/http"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// Handler handles HTTP endpoints for the challenges domain.
type Handler struct {
	svc Service
}

// NewHandler returns a new challenges Handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts challenges routes on the given Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group, jwtMW echo.MiddlewareFunc) {
	g.Use(jwtMW)
	g.GET("", h.List)
	g.POST("", h.Send)
	g.GET("/:id", h.GetByID)
	g.POST("/:id/accept", h.Accept)
	g.POST("/:id/decline", h.Decline)
	g.POST("/:id/submit", h.Submit)
}

// List handles GET /v1/challenges — lists all challenges for the authenticated member.
func (h *Handler) List(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	memberID := auth.GetMemberID(c)
	items, err := h.svc.List(c.Request().Context(), chapterID, memberID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list challenges")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": items})
}

// GetByID handles GET /v1/challenges/:id
func (h *Handler) GetByID(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	ch, err := h.svc.GetByID(c.Request().Context(), chapterID, c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "challenge not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get challenge")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": ch})
}

// Send handles POST /v1/challenges — sends a new challenge.
func (h *Handler) Send(c echo.Context) error {
	var req SendInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}
	chapterID := auth.GetChapterID(c)
	challengerID := auth.GetMemberID(c)
	ch, err := h.svc.Send(c.Request().Context(), chapterID, challengerID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrSelfChallenge):
			return echo.NewHTTPError(http.StatusBadRequest, "cannot challenge yourself")
		case errors.Is(err, ErrAlreadyActive):
			return echo.NewHTTPError(http.StatusConflict, "an active challenge already exists with this member")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to send challenge")
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": ch})
}

// Accept handles POST /v1/challenges/:id/accept
func (h *Handler) Accept(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	memberID := auth.GetMemberID(c)
	ch, err := h.svc.Accept(c.Request().Context(), chapterID, memberID, c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return echo.NewHTTPError(http.StatusNotFound, "challenge not found")
		case errors.Is(err, ErrUnauthorized):
			return echo.NewHTTPError(http.StatusForbidden, "not your challenge to accept")
		case errors.Is(err, ErrBadStatus):
			return echo.NewHTTPError(http.StatusConflict, "challenge is not pending")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to accept challenge")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": ch})
}

// Decline handles POST /v1/challenges/:id/decline
func (h *Handler) Decline(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	memberID := auth.GetMemberID(c)
	ch, err := h.svc.Decline(c.Request().Context(), chapterID, memberID, c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return echo.NewHTTPError(http.StatusNotFound, "challenge not found")
		case errors.Is(err, ErrUnauthorized):
			return echo.NewHTTPError(http.StatusForbidden, "not your challenge to decline")
		case errors.Is(err, ErrBadStatus):
			return echo.NewHTTPError(http.StatusConflict, "challenge is not pending")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to decline challenge")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": ch})
}

// Submit handles POST /v1/challenges/:id/submit — submits trivia answers.
func (h *Handler) Submit(c echo.Context) error {
	var req SubmitInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}
	chapterID := auth.GetChapterID(c)
	memberID := auth.GetMemberID(c)
	ch, err := h.svc.Submit(c.Request().Context(), chapterID, memberID, c.Param("id"), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return echo.NewHTTPError(http.StatusNotFound, "challenge not found")
		case errors.Is(err, ErrUnauthorized):
			return echo.NewHTTPError(http.StatusForbidden, "not a participant in this challenge")
		case errors.Is(err, ErrBadStatus):
			return echo.NewHTTPError(http.StatusConflict, "challenge is not active")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to submit challenge")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": ch})
}
