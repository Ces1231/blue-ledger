package xp

import (
	"math"
	"net/http"
	"strconv"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// Handler handles XP and leaderboard HTTP endpoints.
type Handler struct {
	svc XPService
}

// NewHandler creates a new XP handler.
func NewHandler(svc XPService) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts XP routes on the given Echo groups.
func (h *Handler) RegisterRoutes(rootGroup *echo.Group, jwtMW echo.MiddlewareFunc) {
	rootGroup.GET("/leaderboard", h.GetLeaderboard, jwtMW)
	rootGroup.POST("/xp/award", h.AwardXP, jwtMW, auth.RoleGate("admin"))
	rootGroup.POST("/xp/recalculate", h.RecalculateAll, jwtMW, auth.RoleGate("admin", "sysadmin"))
	rootGroup.GET("/engagement-log", h.GetEngagementLog, jwtMW, auth.RoleGate("admin", "chair"))
}

// GetLeaderboard handles GET /leaderboard
func (h *Handler) GetLeaderboard(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	mode := c.QueryParam("mode") // "alltime" or "semester"
	if mode == "" {
		mode = "alltime"
	}

	entries, err := h.svc.GetLeaderboard(c.Request().Context(), chapterID, mode)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get leaderboard")
	}
	if entries == nil {
		entries = []*LeaderboardEntry{}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": entries,
		"meta": map[string]string{"mode": mode},
	})
}

// awardXPRequest for POST /xp/award
type awardXPRequest struct {
	MemberID string  `json:"member_id" validate:"required"`
	XPAmount int     `json:"xp_amount" validate:"required,min=1"`
	Activity string  `json:"activity" validate:"required"`
	Note     *string `json:"note"`
	Semester *string `json:"semester"`
}

// AwardXP handles POST /xp/award
func (h *Handler) AwardXP(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	awardedBy := auth.GetMemberID(c)

	var req awardXPRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	input := AwardXPInput{
		MemberID: req.MemberID,
		XPAmount: req.XPAmount,
		Activity: req.Activity,
		Note:     req.Note,
		Semester: req.Semester,
	}

	if err := h.svc.AwardXP(c.Request().Context(), chapterID, awardedBy, input); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to award XP")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"message":   "XP awarded",
			"xp_amount": req.XPAmount,
			"member_id": req.MemberID,
		},
	})
}

// RecalculateAll handles POST /xp/recalculate
func (h *Handler) RecalculateAll(c echo.Context) error {
	chapterID := auth.GetChapterID(c)

	count, err := h.svc.RecalculateAll(c.Request().Context(), chapterID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to recalculate XP")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"message":          "XP recalculated",
			"members_updated":  count,
		},
	})
}

// GetEngagementLog handles GET /engagement-log
func (h *Handler) GetEngagementLog(c echo.Context) error {
	chapterID := auth.GetChapterID(c)

	page := 1
	if v := c.QueryParam("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	perPage := 25
	if v := c.QueryParam("per_page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			perPage = n
		}
	}

	entries, total, err := h.svc.GetEngagementLog(c.Request().Context(), chapterID, page, perPage)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get engagement log")
	}
	if entries == nil {
		entries = []*EngagementLogEntry{}
	}

	pages := int(math.Ceil(float64(total) / float64(perPage)))
	if pages < 1 {
		pages = 1
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": entries,
		"meta": map[string]interface{}{
			"total": total, "page": page, "per_page": perPage, "pages": pages,
		},
	})
}
