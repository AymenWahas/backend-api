package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// SecurityHeaders adds a minimal secure headers set to all responses.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}

// RequestSizeLimit rejects oversized request bodies to reduce abuse.
func RequestSizeLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > maxBytes {
				http.Error(w, "request entity too large", http.StatusRequestEntityTooLarge)
				return
			}

			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimit blocks abusive clients using a simple in-memory token bucket / sliding-window approach.
func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	type requestRecord struct {
		times []time.Time
	}

	var (
		mu      sync.Mutex
		clients = map[string]*requestRecord{}
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientKey := r.Header.Get("X-Forwarded-For")
			if clientKey == "" {
				clientKey = r.RemoteAddr
			}
			clientKey = strings.Split(clientKey, ",")[0]

			mu.Lock()
			record, ok := clients[clientKey]
			if !ok {
				record = &requestRecord{times: make([]time.Time, 0, limit)}
				clients[clientKey] = record
			}

			now := time.Now()
			cutoff := now.Add(-window)
			filtered := record.times[:0]
			for _, t := range record.times {
				if t.After(cutoff) {
					filtered = append(filtered, t)
				}
			}
			record.times = filtered

			if len(record.times) >= limit {
				mu.Unlock()
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			record.times = append(record.times, now)
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

// CORS allows only trusted origins and keeps browser-facing APIs conservative.
func CORS(allowedOrigins string) func(http.Handler) http.Handler {
	origins := strings.Split(allowedOrigins, ",")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			for _, allowed := range origins {
				if strings.TrimSpace(allowed) == "*" || strings.TrimSpace(allowed) == origin {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
					w.Header().Set("Access-Control-Max-Age", "600")
					break
				}
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Validation helper for common request abuse checks.
func ValidateIdentifier(value string, fieldName string, maxLen int) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	if len(trimmed) > maxLen {
		return fmt.Errorf("%s is too long", fieldName)
	}
	return nil
}
