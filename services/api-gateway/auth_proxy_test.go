package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func makeBearerToken(secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  "u-1",
		"username": "tester",
		"role":     "user",
		"exp":      time.Now().Add(10 * time.Minute).Unix(),
	})
	signed, _ := token.SignedString([]byte(secret))
	return "Bearer " + signed
}

func TestProxyAuth_UserRoutesRequireJWT(t *testing.T) {
	secret := "test-secret"
	_ = os.Setenv("JWT_SECRET", secret)
	t.Cleanup(func() { _ = os.Unsetenv("JWT_SECRET") })

	r := setupRouterWithLimiter(1000, 1000)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/profile", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for protected user route, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "MISSING_TOKEN") {
		t.Fatalf("expected missing token response, got %s", w.Body.String())
	}
}

func TestProxyAuth_UserAuthRoutesStayPublic(t *testing.T) {
	secret := "test-secret"
	_ = os.Setenv("JWT_SECRET", secret)
	t.Cleanup(func() { _ = os.Unsetenv("JWT_SECRET") })

	r := setupRouterWithLimiter(1000, 1000)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/auth/login", nil)
	r.ServeHTTP(w, req)

	if w.Code == http.StatusUnauthorized {
		t.Fatalf("expected public auth route to bypass JWT, got 401 body=%s", w.Body.String())
	}
}

func TestProxyAuth_PaymentRoutesRequireJWT(t *testing.T) {
	secret := "test-secret"
	_ = os.Setenv("JWT_SECRET", secret)
	t.Cleanup(func() { _ = os.Unsetenv("JWT_SECRET") })

	r := setupRouterWithLimiter(1000, 1000)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/orders", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for payment route, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestProxyAuth_DataRoutesAllowAnonymous(t *testing.T) {
	r := setupRouterWithLimiter(1000, 1000)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/data/universities", nil)
	r.ServeHTTP(w, req)

	if w.Code == http.StatusUnauthorized {
		t.Fatalf("expected anonymous data route, got 401 body=%s", w.Body.String())
	}
}

func TestProxyAuth_UserRoutesAcceptValidJWT(t *testing.T) {
	secret := "test-secret"
	_ = os.Setenv("JWT_SECRET", secret)
	t.Cleanup(func() { _ = os.Unsetenv("JWT_SECRET") })

	r := setupRouterWithLimiter(1000, 1000)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/profile", nil)
	req.Header.Set("Authorization", makeBearerToken(secret))
	r.ServeHTTP(w, req)

	if w.Code == http.StatusUnauthorized {
		t.Fatalf("expected valid JWT to pass auth, got 401 body=%s", w.Body.String())
	}
}
