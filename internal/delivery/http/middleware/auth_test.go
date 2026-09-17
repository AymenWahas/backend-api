package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend-api/internal/auth"
)

func TestAuthMiddleware_NoAuthorization(t *testing.T) {
	handler := Auth("test-secret")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("next handler should not be called")
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	handler := Auth("test-secret")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("next handler should not be called")
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	secret := "test-secret"

	token, err := auth.GenerateAccessToken(
		123,
		456,
		secret,
		15*time.Minute,
	)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	handler := Auth(secret)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(UserIDContextKey).(int)
			if !ok {
				t.Fatal("user_id not found in context")
			}

			employeeID, ok := r.Context().Value(EmployeeIDContextKey).(int)
			if !ok {
				t.Fatal("employee_id not found in context")
			}

			if userID != 123 {
				t.Fatalf("expected user_id 123, got %d", userID)
			}

			if employeeID != 456 {
				t.Fatalf("expected employee_id 456, got %d", employeeID)
			}

			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}
func TestAuthMiddleware_DoesNotTrustUserRoleHeader(t *testing.T) {
	secret := "test-secret"

	token, err := auth.GenerateAccessToken(
		123,
		456,
		secret,
		15*time.Minute,
	)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	handler := Auth(secret)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// The middleware must not create authorization
			// information from a client-controlled role header.
			if role := r.Context().Value("role"); role != nil {
				t.Fatalf("role must not come from client header")
			}

			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)

	req.Header.Set("Authorization", "Bearer "+token)

	// Simulate an attacker trying to become admin.
	req.Header.Set("X-User-Role", "admin")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}
func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	secret := "test-secret"

	token, err := auth.GenerateAccessToken(
		123,
		456,
		secret,
		-1*time.Minute,
	)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	handler := Auth(secret)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("next handler should not be called")
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}