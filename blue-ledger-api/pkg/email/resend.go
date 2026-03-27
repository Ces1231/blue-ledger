package email

import (
	"fmt"
	"strings"

	"github.com/resendlabs/resend-go"
)

// Client wraps the Resend email API.
type Client struct {
	client    *resend.Client
	fromEmail string
	fromName  string
	appURL    string
	// devMode is true when no real Resend API key is configured.
	// In dev mode all emails are printed to stdout instead of sent.
	devMode bool
}

// New creates a new email client backed by Resend.
// When apiKey is empty or a placeholder, the client runs in dev mode:
// emails are printed to stdout so magic links can be copy-pasted from logs.
func New(apiKey, fromEmail, fromName, appURL string) *Client {
	devMode := apiKey == "" || strings.HasPrefix(apiKey, "re_placeholder")
	return &Client{
		client:    resend.NewClient(apiKey),
		fromEmail: fromEmail,
		fromName:  fromName,
		appURL:    appURL,
		devMode:   devMode,
	}
}

// SendMagicLink sends a sign-in magic link email to the given address.
// In dev mode (no real Resend key) the link is printed to stdout instead.
func (c *Client) SendMagicLink(toEmail, token string) error {
	magicURL := fmt.Sprintf("%s/magic?token=%s", c.appURL, token)

	if c.devMode {
		fmt.Printf("\n╔══════════════════════════════════════════════════════════════╗\n")
		fmt.Printf("║  [DEV] MAGIC LINK — copy into browser to sign in             ║\n")
		fmt.Printf("║  To: %-56s ║\n", toEmail)
		fmt.Printf("╠══════════════════════════════════════════════════════════════╣\n")
		fmt.Printf("║  %s\n", magicURL)
		fmt.Printf("╚══════════════════════════════════════════════════════════════╝\n\n")
		return nil
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family:'Helvetica Neue',Arial,sans-serif;background:#FAF8F3;padding:40px 20px;">
  <div style="max-width:500px;margin:0 auto;background:#fff;border-radius:14px;border:1px solid #E2DDD4;overflow:hidden;">
    <div style="background:#001A4D;padding:32px;text-align:center;">
      <div style="font-family:Georgia,serif;font-size:24px;color:#C9A84C;margin-bottom:4px;">The Blue Ledger</div>
      <div style="font-size:11px;color:rgba(255,255,255,.4);font-family:'Courier New',monospace;letter-spacing:2px;text-transform:uppercase;">Chapter Engagement Platform</div>
    </div>
    <div style="padding:32px;">
      <h2 style="font-family:Georgia,serif;color:#001A4D;margin-bottom:8px;">Sign In</h2>
      <p style="color:#6B6657;font-size:14px;margin-bottom:24px;">Click the button below to sign in to The Blue Ledger. This link expires in 15 minutes and can only be used once.</p>
      <a href="%s" style="display:inline-block;background:#001A4D;color:#C9A84C;text-decoration:none;padding:14px 28px;border-radius:10px;font-weight:700;font-size:15px;">Sign In to Blue Ledger</a>
      <p style="color:#A8A099;font-size:12px;margin-top:24px;">If you didn't request this, you can safely ignore this email.</p>
      <p style="color:#A8A099;font-size:11px;margin-top:8px;">Or copy this link:<br><span style="color:#001A4D;word-break:break-all;">%s</span></p>
    </div>
  </div>
</body>
</html>`, magicURL, magicURL)

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", c.fromName, c.fromEmail),
		To:      []string{toEmail},
		Subject: "Sign in to The Blue Ledger",
		Html:    html,
	}

	_, err := c.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("send magic link email: %w", err)
	}
	return nil
}

// SendDuesReminder sends a dues reminder email to the given member.
func (c *Client) SendDuesReminder(toEmail, memberName, semester string, amountCents int, dueDate string) error {
	amount := float64(amountCents) / 100.0

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family:'Helvetica Neue',Arial,sans-serif;background:#FAF8F3;padding:40px 20px;">
  <div style="max-width:500px;margin:0 auto;background:#fff;border-radius:14px;border:1px solid #E2DDD4;overflow:hidden;">
    <div style="background:#001A4D;padding:32px;text-align:center;">
      <div style="font-family:Georgia,serif;font-size:24px;color:#C9A84C;">The Blue Ledger</div>
    </div>
    <div style="padding:32px;">
      <h2 style="font-family:Georgia,serif;color:#001A4D;">Dues Reminder</h2>
      <p style="color:#6B6657;">Brother %s,</p>
      <p style="color:#6B6657;margin-top:12px;">Your chapter dues of <strong>$%.2f</strong> for <strong>%s</strong> are due by <strong>%s</strong>.</p>
      <p style="color:#6B6657;margin-top:12px;">Log in to Blue Ledger to pay via your chapter's Zeffy form.</p>
      <a href="%s/dues" style="display:inline-block;background:#C9A84C;color:#001A4D;text-decoration:none;padding:12px 24px;border-radius:10px;font-weight:700;margin-top:20px;">Pay Dues Now</a>
    </div>
  </div>
</body>
</html>`, memberName, amount, semester, dueDate, c.appURL)

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", c.fromName, c.fromEmail),
		To:      []string{toEmail},
		Subject: fmt.Sprintf("Dues Reminder: %s — %.2f due", semester, amount),
		Html:    html,
	}

	_, err := c.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("send dues reminder: %w", err)
	}
	return nil
}

// SendBadgeEarned sends a congratulatory email when a member earns a badge.
func (c *Client) SendBadgeEarned(toEmail, memberName, badgeName, badgeDescription, badgeIcon string) error {
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family:'Helvetica Neue',Arial,sans-serif;background:#FAF8F3;padding:40px 20px;">
  <div style="max-width:500px;margin:0 auto;background:#fff;border-radius:14px;border:1px solid #E2DDD4;overflow:hidden;">
    <div style="background:#001A4D;padding:32px;text-align:center;">
      <div style="font-family:Georgia,serif;font-size:24px;color:#C9A84C;">The Blue Ledger</div>
    </div>
    <div style="padding:32px;text-align:center;">
      <div style="font-size:48px;margin-bottom:16px;">%s</div>
      <h2 style="font-family:Georgia,serif;color:#001A4D;">Badge Earned!</h2>
      <p style="color:#6B6657;margin-top:8px;">Brother %s, you've earned the <strong>%s</strong> badge!</p>
      <p style="color:#A8A099;font-size:13px;margin-top:8px;">%s</p>
      <a href="%s/quests" style="display:inline-block;background:#001A4D;color:#C9A84C;text-decoration:none;padding:12px 24px;border-radius:10px;font-weight:700;margin-top:20px;">View Your Badges</a>
    </div>
  </div>
</body>
</html>`, badgeIcon, memberName, badgeName, badgeDescription, c.appURL)

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", c.fromName, c.fromEmail),
		To:      []string{toEmail},
		Subject: fmt.Sprintf("You earned the %s badge!", badgeName),
		Html:    html,
	}

	_, err := c.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("send badge earned: %w", err)
	}
	return nil
}
