package events

import (
	"math"
	"net/http"
	"strconv"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/ces1231/blue-ledger-api/internal/streaks"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler handles all events HTTP endpoints.
type Handler struct {
	svc            EventsService
	qrSecret       string
	streaksService *streaks.Service
}

// NewHandler creates a new events handler.
func NewHandler(svc EventsService, qrSecret string, streaksService *streaks.Service) *Handler {
	return &Handler{
		svc:            svc,
		qrSecret:       qrSecret,
		streaksService: streaksService,
	}
}

// RegisterRoutes mounts events routes on the given Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group, jwtMW echo.MiddlewareFunc) {
	g.GET("", h.List, jwtMW)
	g.POST("", h.Create, jwtMW, auth.RoleGate("admin", "chair"))
	g.GET("/:id", h.Get, jwtMW)
	g.PUT("/:id", h.Update, jwtMW, auth.RoleGate("admin", "chair"))
	g.DELETE("/:id", h.Delete, jwtMW, auth.RoleGate("admin"))
	g.POST("/:id/rsvp", h.RSVP, jwtMW)
	g.GET("/:id/qr-token", h.GetQRToken, jwtMW, auth.RoleGate("admin", "chair"))
	g.POST("/:id/checkin/qr", h.CheckInQR, jwtMW, auth.RoleGate("admin", "chair"))
}

// List handles GET /events
func (h *Handler) List(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	upcoming := c.QueryParam("upcoming") != "false"

	page := 1
	if v := c.QueryParam("page"); v != "" {
		if n, _ := strconv.Atoi(v); n > 0 {
			page = n
		}
	}
	perPage := 25
	if v := c.QueryParam("per_page"); v != "" {
		if n, _ := strconv.Atoi(v); n > 0 && n <= 100 {
			perPage = n
		}
	}

	events, total, err := h.svc.List(c.Request().Context(), chapterID, upcoming, page, perPage)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list events")
	}
	if events == nil {
		events = []*Event{}
	}

	pages := int(math.Ceil(float64(total) / float64(perPage)))
	if pages < 1 {
		pages = 1
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": events,
		"meta": map[string]interface{}{
			"total": total, "page": page, "per_page": perPage, "pages": pages,
		},
	})
}

// createEventRequest for POST /events
type createEventRequest struct {
	Name         string  `json:"name" validate:"required,min=3"`
	Description  *string `json:"description"`
	EventDate    string  `json:"event_date" validate:"required"`
	EventTime    *string `json:"event_time"`
	Location     *string `json:"location"`
	EventType    string  `json:"event_type" validate:"required"`
	XPAttend     int     `json:"xp_attend"`
	XPRsvp       int     `json:"xp_rsvp"`
	RSVPDeadline *string `json:"rsvp_deadline"`
	Capacity     *int    `json:"capacity"`
}

// Create handles POST /events
func (h *Handler) Create(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	createdBy := auth.GetMemberID(c)

	var req createEventRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	event, err := h.svc.Create(c.Request().Context(), chapterID, createdBy, CreateEventInput{
		Name:         req.Name,
		Description:  req.Description,
		EventDate:    req.EventDate,
		EventTime:    req.EventTime,
		Location:     req.Location,
		EventType:    req.EventType,
		XPAttend:     req.XPAttend,
		XPRsvp:       req.XPRsvp,
		RSVPDeadline: req.RSVPDeadline,
		Capacity:     req.Capacity,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create event")
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{"data": event})
}

// Get handles GET /events/:id
func (h *Handler) Get(c echo.Context) error {
	event, err := h.svc.Get(c.Request().Context(), c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get event")
	}
	if event == nil {
		return echo.NewHTTPError(http.StatusNotFound, "event not found")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": event})
}

// Update handles PUT /events/:id
func (h *Handler) Update(c echo.Context) error {
	var req createEventRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	event, err := h.svc.Update(c.Request().Context(), c.Param("id"), CreateEventInput{
		Name:         req.Name,
		Description:  req.Description,
		EventDate:    req.EventDate,
		EventTime:    req.EventTime,
		Location:     req.Location,
		EventType:    req.EventType,
		XPAttend:     req.XPAttend,
		XPRsvp:       req.XPRsvp,
		RSVPDeadline: req.RSVPDeadline,
		Capacity:     req.Capacity,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update event")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": event})
}

// Delete handles DELETE /events/:id
func (h *Handler) Delete(c echo.Context) error {
	if err := h.svc.Delete(c.Request().Context(), c.Param("id")); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete event")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": map[string]string{"message": "event cancelled"},
	})
}

// rsvpRequest for POST /events/:id/rsvp
type rsvpRequest struct {
	Status string `json:"status" validate:"required,oneof=yes no maybe"`
}

// RSVP handles POST /events/:id/rsvp
func (h *Handler) RSVP(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	memberID := auth.GetMemberID(c)

	var req rsvpRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.svc.RSVP(c.Request().Context(), chapterID, c.Param("id"), memberID, req.Status); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to RSVP")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": map[string]string{"message": "RSVP recorded", "status": req.Status},
	})
}

// GetQRToken handles GET /events/:id/qr-token
func (h *Handler) GetQRToken(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	eventID := c.Param("id")

	token, err := h.svc.GetQRToken(c.Request().Context(), chapterID, eventID, h.qrSecret)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get QR token")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": map[string]string{"token": token},
	})
}

// checkInQRRequest for POST /events/:id/checkin/qr
type checkInQRRequest struct {
	MemberDisplayID string `json:"member_display_id" validate:"required"`
	ScannerToken    string `json:"scanner_token" validate:"required"`
}

// CheckInQR handles POST /events/:id/checkin/qr
func (h *Handler) CheckInQR(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	scannerMemberID := auth.GetMemberID(c)

	var req checkInQRRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	result, err := h.svc.CheckInQR(
		c.Request().Context(),
		chapterID, c.Param("id"),
		req.MemberDisplayID, req.ScannerToken,
		scannerMemberID, h.qrSecret,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Update streak for attendance
	if result != nil && h.streaksService != nil {
		// Convert string MemberID to UUID for streaks service
		if memberID, err := uuid.Parse(result.MemberID); err == nil {
			// Don't block response if streak update fails
			_, _ = h.streaksService.UpdateStreakForAttendance(c.Request().Context(), memberID)
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"data": result})
}
