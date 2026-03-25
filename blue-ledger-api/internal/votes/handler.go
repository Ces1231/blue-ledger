package votes

import (
	"errors"
	"net/http"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// Handler handles all votes HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler creates a new votes handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts vote routes on the given Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group, jwtMiddleware echo.MiddlewareFunc) {
	g.Use(jwtMiddleware)
	g.GET("", h.List)
	g.POST("", h.Create, auth.RoleGate("admin", "chair"))
	g.POST("/:id/respond", h.Respond)
	g.GET("/:id/results", h.GetResults)
}

// List handles GET /votes
func (h *Handler) List(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	memberID := auth.GetMemberID(c)

	items, err := h.svc.List(c.Request().Context(), chapterID, memberID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list votes")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": items})
}

// Create handles POST /votes
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

	vote, err := h.svc.Create(c.Request().Context(), chapterID, memberID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create vote")
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": vote})
}

// Respond handles POST /votes/:id/respond
func (h *Handler) Respond(c echo.Context) error {
	var req RespondInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	chapterID := auth.GetChapterID(c)
	memberID := auth.GetMemberID(c)
	id := c.Param("id")

	if err := h.svc.Vote(c.Request().Context(), chapterID, id, memberID, req); err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "vote not found")
		}
		if errors.Is(err, ErrAlreadyVoted) {
			return echo.NewHTTPError(http.StatusConflict, "you have already voted on this item")
		}
		if errors.Is(err, ErrVoteClosed) {
			return echo.NewHTTPError(http.StatusForbidden, "voting is closed")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to record vote")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": map[string]string{"message": "vote recorded"}})
}

// GetResults handles GET /votes/:id/results
func (h *Handler) GetResults(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	id := c.Param("id")

	results, err := h.svc.GetResults(c.Request().Context(), chapterID, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "vote not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get vote results")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": results})
}
