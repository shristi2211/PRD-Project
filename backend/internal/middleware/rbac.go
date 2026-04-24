package middleware

import (
	"net/http"

	"golf-score-lottery/backend/internal/utils"
)

// RequireRole returns middleware that enforces role-based access control.
// Only users with one of the specified roles are allowed through.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	roleSet := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		roleSet[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetUserClaims(r.Context())
			if !ok {
				utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			if _, allowed := roleSet[claims.Role]; !allowed {
				utils.ErrorJSON(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireSubscription returns middleware that checks if the user has an active subscription.
func RequireSubscription() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetUserClaims(r.Context())
			if !ok {
				utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			if !claims.SubscriptionActive {
				utils.ErrorJSON(w, http.StatusForbidden, "Active subscription required")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
