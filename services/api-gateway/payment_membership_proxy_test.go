package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestProxyPaymentMembershipRoutesRewriteAndForwardUserHeaders(t *testing.T) {
	secret := "test-secret"
	_ = os.Setenv("JWT_SECRET", secret)
	t.Cleanup(func() { _ = os.Unsetenv("JWT_SECRET") })

	requests := make(chan *http.Request, 1)
	oldTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		var body []byte
		if r.Body != nil {
			body, _ = io.ReadAll(r.Body)
			_ = r.Body.Close()
		}

		reqCopy := r.Clone(r.Context())
		reqCopy.Body = io.NopCloser(bytes.NewReader(body))
		reqCopy.ContentLength = int64(len(body))
		requests <- reqCopy

		return &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Access-Control-Allow-Origin": []string{"*"},
				"Content-Type":                []string{"application/json"},
			},
			Body:          io.NopCloser(strings.NewReader(`{"success":true}`)),
			ContentLength: int64(len(`{"success":true}`)),
			Request:       r,
		}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = oldTransport })

	oldPaymentURL := os.Getenv("PAYMENT_SERVICE_URL")
	_ = os.Setenv("PAYMENT_SERVICE_URL", "http://payment-service:8085")
	t.Cleanup(func() { _ = os.Setenv("PAYMENT_SERVICE_URL", oldPaymentURL) })

	r := setupRouterWithLimiter(1000, 1000)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/membership/plans", nil)
	req.Header.Set("Authorization", makeBearerToken(secret))
	req.Header.Set("X-Request-ID", "req-membership-1")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected gateway to strip backend CORS headers, got %q", got)
	}

	select {
	case proxied := <-requests:
		if proxied.URL.Path != "/api/v1/payments/membership/plans" {
			t.Fatalf("unexpected proxied path: %s", proxied.URL.Path)
		}
		if got := proxied.Header.Get("X-User-ID"); got != "u-1" {
			t.Fatalf("expected forwarded user id, got %q", got)
		}
		if got := proxied.Header.Get("X-Username"); got != "tester" {
			t.Fatalf("expected forwarded username, got %q", got)
		}
		if got := proxied.Header.Get("X-User-Role"); got != "user" {
			t.Fatalf("expected forwarded role, got %q", got)
		}
		if got := proxied.Header.Get("X-Request-ID"); got != "req-membership-1" {
			t.Fatalf("expected request id propagation, got %q", got)
		}
		if got := proxied.Header.Get("X-Forwarded-Service"); got != "payment-service" {
			t.Fatalf("expected forwarded service header, got %q", got)
		}
	default:
		t.Fatal("backend did not receive proxied request")
	}
}
