package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSecurityHeaders(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	res := httptest.NewRecorder()

	SecurityHeaders(next).ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}

	for key, want := range map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "no-referrer",
		"X-XSS-Protection":       "1; mode=block",
	} {
		if got := res.Header().Get(key); got != want {
			t.Fatalf("header %s = %q, want %q", key, got, want)
		}
	}
}

func TestRequestSizeLimit(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := RequestSizeLimit(10)(next)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", strings.NewReader("12345678901"))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", res.Code)
	}
}
func TestRateLimitBlocksRequestsAfterLimit(t *testing.T) {
	const (
		limit  = 3
		window = time.Minute
	)

	calls := 0

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	})

	handler := RateLimit(limit, window)(next)

	for i := 0; i < limit; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Forwarded-For", "10.0.0.1")

		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"request %d: expected status 200, got %d",
				i+1,
				rec.Code,
			)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf(
			"expected status 429 after exceeding rate limit, got %d",
			rec.Code,
		)
	}

	if calls != limit {
		t.Fatalf(
			"expected next handler to be called %d times, got %d",
			limit,
			calls,
		)
	}
}
