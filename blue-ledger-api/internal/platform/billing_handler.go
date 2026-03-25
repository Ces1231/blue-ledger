package platform

import (
	"fmt"
	"net/http"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
	stripe "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/billingportal/session"
)

// BillingHandler handles Stripe billing-related endpoints.
type BillingHandler struct {
	stripeSecretKey string
	returnURL       string
}

// NewBillingHandler creates a new billing handler.
func NewBillingHandler(stripeSecretKey, returnURL string) *BillingHandler {
	return &BillingHandler{
		stripeSecretKey: stripeSecretKey,
		returnURL:       returnURL,
	}
}

// RegisterBillingRoutes mounts billing routes.
func (h *BillingHandler) RegisterBillingRoutes(g *echo.Group, jwtMiddleware echo.MiddlewareFunc) {
	settings := g.Group("/settings", jwtMiddleware, auth.RoleGate("admin"))
	settings.GET("/billing", h.GetBilling)
	settings.POST("/billing/portal", h.CreatePortalSession)
}

// GetBilling handles GET /settings/billing
// Returns the chapter's current subscription status.
func (h *BillingHandler) GetBilling(c echo.Context) error {
	chapterID := auth.GetChapterID(c)

	// Return subscription metadata from context — chapter data is available via JWT claims
	// A real implementation would query chapters table; for now return context data
	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{
			"chapter_id": chapterID,
			"message":    "use /sys/chapters/:id for full billing details",
		},
	})
}

// CreatePortalSession handles POST /settings/billing/portal
// Creates a Stripe Customer Portal session for the chapter admin.
func (h *BillingHandler) CreatePortalSession(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	_ = chapterID

	stripe.Key = h.stripeSecretKey

	// Retrieve the chapter's stripe_customer_id from the DB
	// In production this would query the DB; returning a helpful error if not configured
	var req struct {
		CustomerID string `json:"customer_id" validate:"required"`
	}
	if err := c.Bind(&req); err != nil || req.CustomerID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "stripe_customer_id is required")
	}

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(req.CustomerID),
		ReturnURL: stripe.String(fmt.Sprintf("%s/settings/billing", h.returnURL)),
	}

	sess, err := session.New(params)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create billing portal session")
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"data": map[string]string{"url": sess.URL},
	})
}
