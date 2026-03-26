package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// newTestTokenManager creates a TokenManager with an ephemeral in-memory RSA key.
func newTestTokenManager(t *testing.T, accessExpiry, refreshExpiry time.Duration) *TokenManager {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	return &TokenManager{
		privateKey:    key,
		publicKey:     &key.PublicKey,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

func TestGenerateAccessToken(t *testing.T) {
	tm := newTestTokenManager(t, 15*time.Minute, 720*time.Hour)

	token, err := tm.GenerateAccessToken("user-1", "chapter-1", "member-1", "admin", "test@chapter.org", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestValidateAccessToken_RoundTrip(t *testing.T) {
	tm := newTestTokenManager(t, 15*time.Minute, 720*time.Hour)

	userID := "user-abc"
	chapterID := "chapter-xyz"
	memberID := "member-123"
	role := "chair"
	email := "chair@chapter.org"

	tokenStr, err := tm.GenerateAccessToken(userID, chapterID, memberID, role, email, false)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	claims, err := tm.ValidateAccessToken(tokenStr)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID: got %q want %q", claims.UserID, userID)
	}
	if claims.ChapterID != chapterID {
		t.Errorf("ChapterID: got %q want %q", claims.ChapterID, chapterID)
	}
	if claims.MemberID != memberID {
		t.Errorf("MemberID: got %q want %q", claims.MemberID, memberID)
	}
	if claims.Role != role {
		t.Errorf("Role: got %q want %q", claims.Role, role)
	}
	if claims.Email != email {
		t.Errorf("Email: got %q want %q", claims.Email, email)
	}
}

func TestValidateAccessToken_Expired(t *testing.T) {
	// Issue a token that expires immediately.
	tm := newTestTokenManager(t, -1*time.Second, 720*time.Hour)

	tokenStr, err := tm.GenerateAccessToken("u", "c", "m", "member", "e@e.com", false)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	_, err = tm.ValidateAccessToken(tokenStr)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestValidateAccessToken_Tampered(t *testing.T) {
	tm := newTestTokenManager(t, 15*time.Minute, 720*time.Hour)

	tokenStr, err := tm.GenerateAccessToken("u", "c", "m", "member", "e@e.com", false)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	// Append garbage to invalidate the signature.
	_, err = tm.ValidateAccessToken(tokenStr + "tampered")
	if err == nil {
		t.Fatal("expected error for tampered token, got nil")
	}
}

func TestValidateAccessToken_WrongKey(t *testing.T) {
	tm1 := newTestTokenManager(t, 15*time.Minute, 720*time.Hour)
	tm2 := newTestTokenManager(t, 15*time.Minute, 720*time.Hour)

	tokenStr, err := tm1.GenerateAccessToken("u", "c", "m", "member", "e@e.com", false)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	// Validate with a different key — should fail.
	_, err = tm2.ValidateAccessToken(tokenStr)
	if err == nil {
		t.Fatal("expected error when validating with wrong key, got nil")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	tm := newTestTokenManager(t, 15*time.Minute, 720*time.Hour)

	t1, err := tm.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(t1) == 0 {
		t.Fatal("expected non-empty refresh token")
	}

	t2, err := tm.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Two tokens should never be equal (extremely improbable with 32 random bytes).
	if t1 == t2 {
		t.Fatal("two refresh tokens should not be equal")
	}
}

func TestTokenManager_Expiry(t *testing.T) {
	access := 10 * time.Minute
	refresh := 24 * time.Hour
	tm := newTestTokenManager(t, access, refresh)

	if tm.AccessExpiry() != access {
		t.Errorf("AccessExpiry: got %v want %v", tm.AccessExpiry(), access)
	}
	if tm.RefreshExpiry() != refresh {
		t.Errorf("RefreshExpiry: got %v want %v", tm.RefreshExpiry(), refresh)
	}
}

func TestGenerateAccessToken_SysadminFlag(t *testing.T) {
	tm := newTestTokenManager(t, 15*time.Minute, 720*time.Hour)

	tokenStr, err := tm.GenerateAccessToken("u", "c", "m", "sysadmin", "sys@sys.io", true)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	claims, err := tm.ValidateAccessToken(tokenStr)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}

	if !claims.IsSysadmin {
		t.Error("expected IsSysadmin=true")
	}
}

// Ensure JWT signing method is RS256 (not HS256 — which would be a security bug).
func TestToken_SigningMethod_IsRS256(t *testing.T) {
	tm := newTestTokenManager(t, 15*time.Minute, 720*time.Hour)
	tokenStr, err := tm.GenerateAccessToken("u", "c", "m", "member", "e@e.com", false)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	// Parse without validation to inspect the header.
	token, _, err := jwt.NewParser().ParseUnverified(tokenStr, &Claims{})
	if err != nil {
		t.Fatalf("parse unverified: %v", err)
	}
	if token.Method.Alg() != "RS256" {
		t.Errorf("signing method: got %q want RS256", token.Method.Alg())
	}
}
