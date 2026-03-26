package dues

import (
	"fmt"
	"io"
	"net/http"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/customer"
	"github.com/stripe/stripe-go/v78/subscription"
	"github.com/stripe/stripe-go/v78/webhook"
)

// StripeWebhookEvent is a parsed Stripe event relevant to Blue Ledger.
type StripeWebhookEvent struct {
	Type           string
	SubscriptionID string
	CustomerID     string
	Status         string // subscription status
}

// VerifyStripeWebhook reads the raw request body and verifies the Stripe signature.
// Returns the parsed event on success.
func VerifyStripeWebhook(r *http.Request, webhookSecret string) (*stripe.Event, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("read stripe webhook body: %w", err)
	}

	event, err := webhook.ConstructEvent(body, r.Header.Get("Stripe-Signature"), webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("verify stripe webhook: %w", err)
	}

	return &event, nil
}

// CreateSubscription creates a Stripe subscription for a new chapter.
func CreateSubscription(secretKey, customerID, priceID string) (*stripe.Subscription, error) {
	stripe.Key = secretKey

	params := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{Price: stripe.String(priceID)},
		},
		TrialPeriodDays: stripe.Int64(30),
	}

	sub, err := subscription.New(params)
	if err != nil {
		return nil, fmt.Errorf("create stripe subscription: %w", err)
	}

	return sub, nil
}

// CreateCustomer creates a Stripe customer for a new chapter.
func CreateCustomer(secretKey, email, chapterName string) (*stripe.Customer, error) {
	stripe.Key = secretKey

	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(chapterName),
	}

	c, err := customer.New(params)
	if err != nil {
		return nil, fmt.Errorf("create stripe customer: %w", err)
	}

	return c, nil
}
