package events

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// GenerateQRToken creates a signed HMAC-SHA256 event QR token.
// Format: "eventID:chapterID:timestamp:signature"
// The token is stored in events.qr_code_token and displayed as a QR code for scanners.
func GenerateQRToken(secretKey, eventID, chapterID string) (string, error) {
	if secretKey == "" {
		return "", fmt.Errorf("QR secret key is not configured")
	}

	ts := strconv.FormatInt(time.Now().Unix(), 10)
	payload := fmt.Sprintf("%s:%s:%s", eventID, chapterID, ts)

	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("%s:%s", payload, sig), nil
}

// ValidateQRToken verifies an event QR scanner token.
// Returns the event ID and chapter ID if valid.
func ValidateQRToken(secretKey, token string) (eventID, chapterID string, err error) {
	parts := strings.SplitN(token, ":", 4)
	if len(parts) != 4 {
		return "", "", fmt.Errorf("malformed QR token")
	}

	eventID = parts[0]
	chapterID = parts[1]
	ts := parts[2]
	sig := parts[3]

	// Verify signature
	payload := fmt.Sprintf("%s:%s:%s", eventID, chapterID, ts)
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return "", "", fmt.Errorf("invalid QR token signature")
	}

	return eventID, chapterID, nil
}
