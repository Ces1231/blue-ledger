package streaks

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler handles HTTP requests for streak endpoints
type Handler struct {
	service *Service
}

// NewHandler creates a new streak handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetStreakInfo handles GET /v1/members/:id/streak
func (h *Handler) GetStreakInfo(c echo.Context) error {
	memberIDStr := c.Param("id")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid member id"})
	}

	streak, err := h.service.GetStreakInfo(c.Request().Context(), memberID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "streak not found"})
	}

	return c.JSON(http.StatusOK, streak)
}

// GetStreakLeaderboard handles GET /v1/chapters/:chapterId/streaks/leaderboard
func (h *Handler) GetStreakLeaderboard(c echo.Context) error {
	chapterIDStr := c.Param("chapterId")
	chapterID, err := uuid.Parse(chapterIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid chapter id"})
	}

	limit := 100
	if l := c.QueryParam("limit"); l != "" {
		// Parse limit if provided
	}

	leaderboard, err := h.service.GetStreakLeaderboard(c.Request().Context(), chapterID, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get leaderboard"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  leaderboard,
		"total": len(leaderboard),
	})
}

// RegisterRoutes registers streak routes
func (h *Handler) RegisterRoutes(group *echo.Group, jwtMW echo.MiddlewareFunc) {
	// Member streak info - GET /v1/streaks/members/:id
	group.GET("/members/:id", h.GetStreakInfo, jwtMW)

	// Streak leaderboard - GET /v1/streaks/chapters/:chapterId/leaderboard
	group.GET("/chapters/:chapterId/leaderboard", h.GetStreakLeaderboard, jwtMW)
}
