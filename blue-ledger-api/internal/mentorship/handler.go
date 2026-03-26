package mentorship

import (
	"errors"
	"net/http"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// Handler handles all mentorship HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler creates a new mentorship handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts mentorship routes on the given Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group, jwtMiddleware echo.MiddlewareFunc) {
	g.Use(jwtMiddleware)
	g.GET("", h.List)
	g.POST("/become-mentor", h.BecomeMentor)
	g.POST("/:id/request", h.RequestMatch)
	g.GET("/my-match", h.GetMyMatch)
}

// List handles GET /mentorship
func (h *Handler) List(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	mentors, err := h.svc.List(c.Request().Context(), chapterID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list mentors")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": mentors})
}

// BecomeMentor handles POST /mentorship/become-mentor
func (h *Handler) BecomeMentor(c echo.Context) error {
	var req BecomeMentorInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	chapterID := auth.GetChapterID(c)
	memberID := auth.GetMemberID(c)

	mentor, err := h.svc.BecomeMentor(c.Request().Context(), chapterID, memberID, req)
	if err != nil {
		if errors.Is(err, ErrAlreadyMentor) {
			return echo.NewHTTPError(http.StatusConflict, "already registered as a mentor")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to register as mentor")
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": mentor})
}

// RequestMatch handles POST /mentorship/:id/request
func (h *Handler) RequestMatch(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	menteeID := auth.GetMemberID(c)
	mentorID := c.Param("id")

	match, err := h.svc.RequestMatch(c.Request().Context(), chapterID, menteeID, RequestMatchInput{
		MentorID: mentorID,
	})
	if err != nil {
		if errors.Is(err, ErrAlreadyMatched) {
			return echo.NewHTTPError(http.StatusConflict, "you already have an active mentorship match")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to request mentorship match")
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": match})
}

// GetMyMatch handles GET /mentorship/my-match
func (h *Handler) GetMyMatch(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	memberID := auth.GetMemberID(c)

	match, err := h.svc.GetMyMatch(c.Request().Context(), chapterID, memberID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get mentorship match")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": match})
}
