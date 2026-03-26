package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the payload of a Blue Ledger access token.
type Claims struct {
	jwt.RegisteredClaims
	UserID    string `json:"user_id"`
	ChapterID string `json:"chapter_id"`
	MemberID  string `json:"member_id"`
	Role      string `json:"role"`
	Email     string `json:"email"`
	IsSysadmin bool  `json:"is_sysadmin"`
}

// TokenManager handles JWT generation and validation using RS256.
type TokenManager struct {
	privateKey     *rsa.PrivateKey
	publicKey      *rsa.PublicKey
	accessExpiry   time.Duration
	refreshExpiry  time.Duration
}

// NewTokenManager loads RSA keys from disk and creates a token manager.
func NewTokenManager(privateKeyPath, publicKeyPath string, accessExpiry, refreshExpiry time.Duration) (*TokenManager, error) {
	privBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privBytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	pubBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(pubBytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	return &TokenManager{
		privateKey:    privateKey,
		publicKey:     publicKey,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}, nil
}

// GenerateAccessToken creates a signed RS256 JWT access token.
func (tm *TokenManager) GenerateAccessToken(userID, chapterID, memberID, role, email string, isSysadmin bool) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tm.accessExpiry)),
			Issuer:    "blue-ledger",
		},
		UserID:     userID,
		ChapterID:  chapterID,
		MemberID:   memberID,
		Role:       role,
		Email:      email,
		IsSysadmin: isSysadmin,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(tm.privateKey)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

// GenerateRefreshToken creates a cryptographically random opaque refresh token.
// The raw token is returned — the caller must hash it before storing.
func (tm *TokenManager) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// ValidateAccessToken parses and validates a JWT access token string.
// Returns the claims if valid, otherwise an error.
func (tm *TokenManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return tm.publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// AccessExpiry returns the access token lifetime.
func (tm *TokenManager) AccessExpiry() time.Duration {
	return tm.accessExpiry
}

// RefreshExpiry returns the refresh token lifetime.
func (tm *TokenManager) RefreshExpiry() time.Duration {
	return tm.refreshExpiry
}
