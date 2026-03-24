package members

import (
	"math"
	"net/http"
	"strconv"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// Handler handles all members HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler creates a new members handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts members routes on the given Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group, jwtMW echo.MiddlewareFunc) {
	g.GET("", h.List, jwtMW)
	g.GET("/:id", h.Get, jwtMW)
	g.PUT("/:id", h.Update, jwtMW)
	g.DELETE("/:id", h.Delete, jwtMW, auth.RoleGate("admin"))
	g.GET("/:id/xp-history", h.GetXPHistory, jwtMW)
}

// --- Pagination helper ---

type pageMeta struct {
	Total   int `json:"total"`
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Pages   int `json:"pages"`
}

func buildMeta(total, page, perPage int) pageMeta {
	pages := int(math.Ceil(float64(total) / float64(perPage)))
	if pages < 1 {
		pages = 1
	}
	return pageMeta{Total: total, Page: page, PerPage: perPage, Pages: pages}
}

func queryInt(c echo.Context, key string, defaultVal int) int {
	v := c.QueryParam(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return defaultVal
	}
	return n
}

// List handles GET /members
func (h *Handler) List(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	page := queryInt(c, "page", 1)
	perPage := queryInt(c, "per_page", 25)

	members, total, err := h.svc.List(c.Request().Context(), chapterID, page, perPage)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list members")
	}
	if members == nil {
		members = []*Member{}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": members,
		"meta": buildMeta(total, page, perPage),
	})
}

// Get handles GET /members/:id
func (h *Handler) Get(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	memberID := c.Param("id")

	m, err := h.svc.Get(c.Request().Context(), chapterID, memberID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get member")
	}
	if m == nil {
		return echo.NewHTTPError(http.StatusNotFound, "member not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"data": m})
}

// updateRequest for PUT /members/:id
type updateRequest struct {
	InductedYear *int    `json:"inducted_year"`
	Employer     *string `json:"employer"`
	JobTitle     *string `json:"job_title"`
	City         *string `json:"city"`
	LinkedinURL  *string `json:"linkedin_url"`
	AvatarBg     *string `json:"avatar_bg"`
	AvatarFg     *string `json:"avatar_fg"`
	// Admin-only
	Role       *string `json:"role"`
	Status     *string `json:"status"`
	DuesStatus *string `json:"dues_status"`
}

// Update handles PUT /members/:id
func (h *Handler) Update(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	memberID := c.Param("id")
	currentMemberID := auth.GetMemberID(c)
	role := auth.GetRole(c)

	// Members can only edit their own profile (unless admin/sysadmin)
	if role != "admin" && role != "sysadmin" && memberID != currentMemberID {
		return echo.NewHTTPError(http.StatusForbidden, "cannot edit another member's profile")
	}

	var req updateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	input := UpdateInput{
		InductedYear: req.InductedYear,
		Employer:     req.Employer,
		JobTitle:     req.JobTitle,
		City:         req.City,
		LinkedinURL:  req.LinkedinURL,
		AvatarBg:     req.AvatarBg,
		AvatarFg:     req.AvatarFg,
		Role:         req.Role,
		Status:       req.Status,
		DuesStatus:   req.DuesStatus,
	}

	m, err := h.svc.Update(c.Request().Context(), chapterID, memberID, input, role)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update member")
	}
	if m == nil {
		return echo.NewHTTPError(http.StatusNotFound, "member not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"data": m})
}

// Delete handles DELETE /members/:id
func (h *Handler) Delete(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	memberID := c.Param("id")

	if err := h.svc.Delete(c.Request().Context(), chapterID, memberID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete member")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": map[string]string{"message": "member removed"},
	})
}

// GetXPHistory handles GET /members/:id/xp-history
func (h *Handler) GetXPHistory(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	memberID := c.Param("id")
	currentMemberID := auth.GetMemberID(c)
	role := auth.GetRole(c)

	// Members can only view their own history
	if role != "admin" && role != "sysadmin" && memberID != currentMemberID {
		return echo.NewHTTPError(http.StatusForbidden, "cannot view another member's XP history")
	}

	page := queryInt(c, "page", 1)
	perPage := queryInt(c, "per_page", 25)

	entries, total, err := h.svc.GetXPHistory(c.Request().Context(), chapterID, memberID, page, perPage)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get XP history")
	}
	if entries == nil {
		entries = []*XPHistoryEntry{}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": entries,
		"meta": buildMeta(total, page, perPage),
	})
}
