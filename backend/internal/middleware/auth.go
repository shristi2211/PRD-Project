package middleware

import (
	"context"
	"net/http"
	"strings"

	"golf-score-lottery/backend/internal/service"
	"golf-score-lottery/backend/internal/utils"
)

// contextKey is an unexported type for context keys to prevent collisions.
type contextKey string

const claimsKey contextKey = "user_claims"

// AuthMiddleware creates a JWT verification middleware.
// It extracts the Bearer token, validates it with RS256 public key,
// checks the Redis blacklist, and injects claims into the request context.
func AuthMiddleware(keyManager *utils.KeyManager, authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.ErrorJSON(w, http.StatusUnauthorized, "Authorization header is required")
				return
			}

			// Must be "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				utils.ErrorJSON(w, http.StatusUnauthorized, "Authorization header must be: Bearer <token>")
				return
			}

			tokenStr := parts[1]
			if tokenStr == "" {
				utils.ErrorJSON(w, http.StatusUnauthorized, "Token is required")
				return
			}

			// Validate JWT (RS256 signature, expiry, issuer)
			claims, err := utils.ValidateAccessToken(tokenStr, keyManager.PublicKey())
			if err != nil {
				utils.ErrorJSON(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			// Check Redis blacklist for revoked tokens
			if claims.ID != "" {
				blacklisted, err := authService.IsAccessTokenBlacklisted(r.Context(), claims.ID)
				if err == nil && blacklisted {
					utils.ErrorJSON(w, http.StatusUnauthorized, "Token has been revoked")
					return
				}
				// If Redis is down, allow the request (fail-open for availability)
			}

			// Inject claims into context
			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserClaims extracts the JWT claims from the request context.
func GetUserClaims(ctx context.Context) (*utils.CustomClaims, bool) {
	claims, ok := ctx.Value(claimsKey).(*utils.CustomClaims)
	return claims, ok
}
