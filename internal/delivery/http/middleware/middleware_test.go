package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAcceptJSON(t *testing.T) {
	tests := []struct {
		name           string
		accept         string
		expectedStatus int
		nextCalled     bool
	}{
		{
			name:           "application json",
			accept:         "application/json",
			expectedStatus: http.StatusOK,
			nextCalled:     true,
		},
		{
			name:           "wildcard",
			accept:         "*/*",
			expectedStatus: http.StatusOK,
			nextCalled:     true,
		},
		{
			name:           "empty accept",
			accept:         "",
			expectedStatus: http.StatusOK,
			nextCalled:     true,
		},
		{
			name:           "unsupported accept",
			accept:         "text/html",
			expectedStatus: http.StatusNotAcceptable,
			nextCalled:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false

			next := http.HandlerFunc(func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			handler := AcceptJSON(next)

			req := httptest.NewRequest(
				http.MethodGet,
				"/employees",
				nil,
			)

			req.Header.Set("Accept", tt.accept)

			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf(
					"expected status %d, got %d",
					tt.expectedStatus,
					rec.Code,
				)
			}

			if nextCalled != tt.nextCalled {
				t.Fatalf(
					"expected nextCalled=%v, got %v",
					tt.nextCalled,
					nextCalled,
				)
			}
		})
	}
}

func TestJSONContentType(t *testing.T) {
	tests := []struct {
		name           string
		contentType    string
		expectedStatus int
		nextCalled     bool
	}{
		{
			name:           "application json",
			contentType:    "application/json",
			expectedStatus: http.StatusOK,
			nextCalled:     true,
		},
		{
			name:           "application json with charset",
			contentType:    "application/json; charset=utf-8",
			expectedStatus: http.StatusOK,
			nextCalled:     true,
		},
		{
			name:           "empty content type",
			contentType:    "",
			expectedStatus: http.StatusUnsupportedMediaType,
			nextCalled:     false,
		},
		{
			name:           "text html",
			contentType:    "text/html",
			expectedStatus: http.StatusUnsupportedMediaType,
			nextCalled:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false

			next := http.HandlerFunc(func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			handler := JSONContentType(next)

			req := httptest.NewRequest(
				http.MethodPost,
				"/employees",
				strings.NewReader(`{}`),
			)

			req.Header.Set(
				"Content-Type",
				tt.contentType,
			)

			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf(
					"expected status %d, got %d",
					tt.expectedStatus,
					rec.Code,
				)
			}

			if nextCalled != tt.nextCalled {
				t.Fatalf(
					"expected nextCalled=%v, got %v",
					tt.nextCalled,
					nextCalled,
				)
			}
		})
	}
}

func TestRequestID(t *testing.T) {
	t.Run("existing request id", func(t *testing.T) {
		next := http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			requestID := GetRequestID(r.Context())

			if requestID != "test-request-id" {
				t.Fatalf(
					"expected request id %q, got %q",
					"test-request-id",
					requestID,
				)
			}

			w.WriteHeader(http.StatusOK)
		})

		handler := RequestID(next)

		req := httptest.NewRequest(
			http.MethodGet,
			"/health",
			nil,
		)

		req.Header.Set(
			"X-Request-ID",
			"test-request-id",
		)

		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Header().Get("X-Request-ID") != "test-request-id" {
			t.Fatalf(
				"expected response request id %q, got %q",
				"test-request-id",
				rec.Header().Get("X-Request-ID"),
			)
		}
	})

	t.Run("generates request id", func(t *testing.T) {
		next := http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			requestID := GetRequestID(r.Context())

			if requestID == "" {
				t.Fatal("expected generated request id")
			}

			w.WriteHeader(http.StatusOK)
		})

		handler := RequestID(next)

		req := httptest.NewRequest(
			http.MethodGet,
			"/health",
			nil,
		)

		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		responseRequestID := rec.Header().Get("X-Request-ID")

		if responseRequestID == "" {
			t.Fatal("expected X-Request-ID response header")
		}
	})
}

func TestRecovery(t *testing.T) {
	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		panic("test panic")
	})

	handler := Recovery(next)

	req := httptest.NewRequest(
		http.MethodGet,
		"/panic",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status 500, got %d",
			rec.Code,
		)
	}

	if !strings.Contains(
		rec.Body.String(),
		"INTERNAL_ERROR",
	) {
		t.Fatalf(
			"expected INTERNAL_ERROR in response, got %s",
			rec.Body.String(),
		)
	}
}

func TestTimeout(t *testing.T) {
	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		select {
		case <-r.Context().Done():
			w.WriteHeader(http.StatusRequestTimeout)

		case <-time.After(100 * time.Millisecond):
			w.WriteHeader(http.StatusOK)
		}
	})

	handler := Timeout(10 * time.Millisecond)(next)

	req := httptest.NewRequest(
		http.MethodGet,
		"/slow",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestTimeout {
		t.Fatalf(
			"expected status 408, got %d",
			rec.Code,
		)
	}
}

func TestRequestLogger(t *testing.T) {
	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		w.WriteHeader(http.StatusCreated)
	})

	handler := RequestLogger(next)

	req := httptest.NewRequest(
		http.MethodPost,
		"/employees",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected status 201, got %d",
			rec.Code,
		)
	}
}
