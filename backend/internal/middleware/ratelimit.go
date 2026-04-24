package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"

	"golf-score-lottery/backend/internal/utils"
)

// RateLimiter provides Redis-backed rate limiting per IP address.
// Uses a sliding window counter implemented with Redis INCR + EXPIRE.
type RateLimiter struct {
	redisClient *redis.Client
	maxRequests int
	window      time.Duration
}

// NewRateLimiter creates a new Redis-backed rate limiter.
func NewRateLimiter(redisClient *redis.Client, maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
		maxRequests: maxRequests,
		window:      window,
	}
}

// Middleware returns an HTTP middleware that enforces rate limits.
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractIP(r)
			key := fmt.Sprintf("ratelimit:%s", ip)

			allowed, err := rl.allow(r.Context(), key)
			if err != nil {
				// If Redis is down, allow the request (fail-open for availability)
				next.ServeHTTP(w, r)
				return
			}

			if !allowed {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(rl.window.Seconds())))
				utils.ErrorJSON(w, http.StatusTooManyRequests, "Too many requests. Please try again later.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// allow checks if the request is within the rate limit using Redis INCR + EXPIRE.
func (rl *RateLimiter) allow(ctx context.Context, key string) (bool, error) {
	// Use a Lua script for atomic increment + expire
	script := redis.NewScript(`
		local current = redis.call("INCR", KEYS[1])
		if current == 1 then
			redis.call("EXPIRE", KEYS[1], ARGV[1])
		end
		return current
	`)

	result, err := script.Run(ctx, rl.redisClient, []string{key}, int(rl.window.Seconds())).Int()
	if err != nil {
		return false, err
	}

	return result <= rl.maxRequests, nil
}

// extractIP gets the client IP address, respecting X-Forwarded-For and X-Real-IP headers.
func extractIP(r *http.Request) string {
	// Check X-Forwarded-For first (may contain multiple IPs)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP (original client)
		if idx := len(xff); idx > 0 {
			for i := 0; i < len(xff); i++ {
				if xff[i] == ',' {
					return xff[:i]
				}
			}
			return xff
		}
	}

	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fallback to remote address
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
