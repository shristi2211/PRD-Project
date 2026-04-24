package middleware

import (
	"net/http"
	"strings"

	"github.com/go-chi/cors"
)

// CORSMiddleware returns a configured CORS handler.
// origins is a comma-separated list of allowed origins (e.g., "http://localhost:5173,https://app.example.com").
func CORSMiddleware(origins string) func(http.Handler) http.Handler {
	allowedOrigins := parseOrigins(origins)

	return cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		ExposedHeaders:   []string{"Link", "X-Request-Id"},
		AllowCredentials: true,
		MaxAge:           300, // 5 minutes preflight cache
	})
}

func parseOrigins(origins string) []string {
	parts := strings.Split(origins, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return []string{"http://localhost:5173"}
	}
	return result
}
