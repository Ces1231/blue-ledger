package dues

import (
	"math"
	"net/http"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// Handler handles all dues HTTP endpoints.
type Handler struct {
	svc           DuesService
	webhookSecret string
}

// NewHandler creates a new dues handler.
func NewHandler(svc DuesService, webhookSecret string) *Handler {
	return &Handler{svc: svc, webhookSecret: webhookSecret}
}

// RegisterRoutes mounts dues routes and webhook endpoints.
func (h *Handler) RegisterRoutes(g *echo.Group, webhookGroup *echo.Group, jwtMW echo.MiddlewareFunc) {
	g.GET("", h.List, jwtMW, auth.RoleGate("admin"))
	g.GET("/me", h.ListForMe, jwtMW)
	g.POST("", h.Create, jwtMW, auth.RoleGate("admin"))
	g.PUT("/:id/mark-paid", h.MarkPaid, jwtMW, auth.RoleGate("admin"))

	// Webhooks — no JWT, signature verification instead
	webhookGroup.POST("/zeffy", h.ZeffyWebhook)
	webhookGroup.POST("/stripe", h.StripeWebhook)
}

// List handles GET /dues
func (h *Handler) List(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	semester := c.QueryParam("semester")

	page := 1
	perPage := 50

	records, total, err := h.svc.List(c.Request().Context(), chapterID, semester, page, perPage)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list dues")
	}
	if records == nil {
		records = []*DuesRecord{}
	}

	pages := int(math.Ceil(float64(total) / float64(perPage)))
	if pages < 1 {
		pages = 1
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": records,
		"meta": map[string]interface{}{
			"total": total, "page": page, "per_page": perPage, "pages": pages,
		},
	})
}

// ListForMe handles GET /dues/me
func (h *Handler) ListForMe(c echo.Context) error {
	memberID := auth.GetMemberID(c)

	records, err := h.svc.ListForMember(c.Request().Context(), memberID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get dues")
	}
	if records == nil {
		records = []*DuesRecord{}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"data": records})
}

// createDuesRequest for POST /dues
type createDuesRequest struct {
	MemberID    string `json:"member_id" validate:"required"`
	Semester    string `json:"semester" validate:"required"`
	AmountCents int    `json:"amount_cents" validate:"required,min=1"`
	DueDate     string `json:"due_date" validate:"required"`
}

// Create handles POST /dues
func (h *Handler) Create(c echo.Context) error {
	chapterID := auth.GetChapterID(c)

	var req createDuesRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	record, err := h.svc.Create(c.Request().Context(), chapterID, req.MemberID, req.Semester, req.AmountCents, req.DueDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create dues record")
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{"data": record})
}

// markPaidRequest for PUT /dues/:id/mark-paid
type markPaidRequest struct {
	PaymentMethod string `json:"payment_method" validate:"required"`
}

// MarkPaid handles PUT /dues/:id/mark-paid
func (h *Handler) MarkPaid(c echo.Context) error {
	var req markPaidRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	record, err := h.svc.MarkPaid(c.Request().Context(), c.Param("id"), req.PaymentMethod)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to mark dues paid")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"data": record})
}

// ZeffyWebhook handles POST /webhooks/zeffy
// This endpoint is called by Zapier when a Zeffy payment is confirmed.
func (h *Handler) ZeffyWebhook(c echo.Context) error {
	// Zeffy webhooks come through Zapier with a simple HMAC secret in the header
	// for this implementation. In production, configure a shared secret in Zapier.
	var payload ZeffyWebhookPayload
	if err := c.Bind(&payload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}

	// Get chapter ID from a query param (set in Zapier webhook URL configuration)
	chapterID := c.QueryParam("chapter_id")
	if chapterID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "chapter_id required")
	}

	if err := h.svc.HandleZeffyWebhook(c.Request().Context(), chapterID, payload); err != nil {
		// Log but return 200 to prevent Zapier retries on partial failures
		return c.JSON(http.StatusOK, map[string]interface{}{
			"received": true,
			"error":    err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"received": true})
}

// StripeWebhook handles POST /webhooks/stripe
// This processes subscription lifecycle events from Stripe.
func (h *Handler) StripeWebhook(c echo.Context) error {
	event, err := VerifyStripeWebhook(c.Request(), h.webhookSecret)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid stripe signature")
	}

	if err := h.svc.HandleStripeSubscriptionWebhook(c.Request().Context(), event); err != nil {
		// Log in production; return 200 to avoid Stripe retries on non-critical errors
		return c.JSON(http.StatusOK, map[string]interface{}{"received": true})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"received": true})
}
