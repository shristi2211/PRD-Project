package utils

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"golf-score-lottery/backend/internal/models"
)

const (
	issuer = "golf-score-lottery"
)

// CustomClaims extends JWT standard claims with app-specific fields.
type CustomClaims struct {
	UserID             uuid.UUID `json:"user_id"`
	Email              string    `json:"email"`
	Role               string    `json:"role"`
	SubscriptionActive bool      `json:"subscription_active"`
	jwt.RegisteredClaims
}

// RefreshClaims holds minimal data for refresh tokens.
type RefreshClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a signed RS256 JWT access token.
func GenerateAccessToken(user *models.User, privateKey *rsa.PrivateKey, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := CustomClaims{
		UserID:             user.ID,
		Email:              user.Email,
		Role:               user.Role,
		SubscriptionActive: user.SubscriptionActive,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			ID:        uuid.New().String(), // unique JTI for blacklisting
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return signedToken, nil
}

// GenerateRefreshToken creates a signed RS256 JWT refresh token with a unique JTI.
func GenerateRefreshToken(userID uuid.UUID, privateKey *rsa.PrivateKey, expiry time.Duration) (string, string, error) {
	now := time.Now()
	jti := uuid.New().String()

	claims := RefreshClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return signedToken, jti, nil
}

// ValidateAccessToken verifies an RS256 JWT access token and returns the claims.
func ValidateAccessToken(tokenString string, publicKey *rsa.PublicKey) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Enforce RS256 algorithm to prevent algorithm confusion attacks
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	},
		jwt.WithIssuer(issuer),
		jwt.WithValidMethods([]string{"RS256"}),
	)

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// ValidateRefreshToken verifies an RS256 JWT refresh token and returns the claims.
func ValidateRefreshToken(tokenString string, publicKey *rsa.PublicKey) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	},
		jwt.WithIssuer(issuer),
		jwt.WithValidMethods([]string{"RS256"}),
	)

	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid refresh token claims")
	}

	return claims, nil
}
