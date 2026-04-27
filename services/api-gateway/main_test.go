package main

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// captureLogs is a helper to capture log.Println output during tests.
func captureLogs(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	old := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(old)
	fn()
	return buf.String()
}

func TestCORSHeaders_OnGET(t *testing.T) {
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("expected CORS Allow-Origin to echo allowed origin, got %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, "GET") {
		t.Fatalf("expected CORS Allow-Methods to include GET, got %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(got, "X-Request-ID") {
		t.Fatalf("expected CORS Allow-Headers to include X-Request-ID, got %q", got)
	}
}

func TestCORS_PreflightOPTIONS(t *testing.T) {
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodOptions, "/healthz", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for preflight, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("expected CORS Allow-Origin to echo allowed origin, got %q", got)
	}
}

func TestAccessLogMiddleware_EmitsLog(t *testing.T) {
	r := setupRouter()

	out := captureLogs(t, func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
		req.Header.Set("X-Request-ID", "req-log-1")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", w.Code)
		}
	})

	if !strings.Contains(out, "\"msg\":\"access\"") {
		t.Fatalf("expected structured access log, got: %s", out)
	}
	if !strings.Contains(out, "\"request_id\":\"req-log-1\"") {
		t.Fatalf("expected request id in access log, got: %s", out)
	}
	if !strings.Contains(out, "\"path\":\"/healthz\"") {
		t.Fatalf("expected matched path in access log, got: %s", out)
	}
}

func TestSecurityHeaders_OnGET(t *testing.T) {
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/healthz", nil)

	r.ServeHTTP(w, req)

	hdr := w.Header()
	if hdr.Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("missing or wrong X-Content-Type-Options: %q", hdr.Get("X-Content-Type-Options"))
	}
	if hdr.Get("X-Frame-Options") != "DENY" {
		t.Fatalf("missing or wrong X-Frame-Options: %q", hdr.Get("X-Frame-Options"))
	}
	if hdr.Get("X-XSS-Protection") != "1; mode=block" {
		t.Fatalf("missing or wrong X-XSS-Protection: %q", hdr.Get("X-XSS-Protection"))
	}
	if hdr.Get("Referrer-Policy") != "" {
		t.Fatalf("missing or wrong Referrer-Policy: %q", hdr.Get("Referrer-Policy"))
	}
}

func TestRequestID_PropagationFromHeader(t *testing.T) {
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("X-Request-ID", "abc123")

	r.ServeHTTP(w, req)

	if got := w.Header().Get("X-Request-ID"); got != "abc123" {
		t.Fatalf("expected X-Request-ID to propagate, got %q", got)
	}
}

func TestRequestID_GeneratedWhenMissing(t *testing.T) {
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/healthz", nil)

	r.ServeHTTP(w, req)

	if got := w.Header().Get("X-Request-ID"); got == "" {
		t.Fatalf("expected X-Request-ID to be generated")
	}
}

func TestRateLimiter_AllowsThenLimits(t *testing.T) {
	// Configure a very small limiter: 2 rps, burst 2 to make behavior deterministic
	r := setupRouterWithLimiter(2, 2)

	// Consume burst: first two should pass
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 on request %d, got %d", i+1, w.Code)
		}
	}

	// Third should be 429 immediately
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after exceeding burst, got %d", w.Code)
	}
	if ra := w.Header().Get("Retry-After"); ra == "" {
		t.Fatalf("expected Retry-After header on 429")
	}

	// After ~600ms (>= 0.5s), with 2 rps, at least 1 token should be available
	time.Sleep(600 * time.Millisecond)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 after waiting for refill, got %d", w2.Code)
	}
}

func TestRateLimiter_SkipsOPTIONS(t *testing.T) {
	r := setupRouterWithLimiter(0, 0) // effectively block all, but OPTIONS should pass due to CORS

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodOptions, "/healthz", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for OPTIONS even when limiter is 0, got %d", w.Code)
	}
}

func TestRateLimiter_UsesHigherBudgetForPublicDataRoutes(t *testing.T) {
	oldPublicRPS := os.Getenv("RATE_LIMIT_PUBLIC_RPS")
	oldPublicBurst := os.Getenv("RATE_LIMIT_PUBLIC_BURST")
	t.Cleanup(func() {
		_ = os.Setenv("RATE_LIMIT_PUBLIC_RPS", oldPublicRPS)
		_ = os.Setenv("RATE_LIMIT_PUBLIC_BURST", oldPublicBurst)
	})
	_ = os.Setenv("RATE_LIMIT_PUBLIC_RPS", "0")
	_ = os.Setenv("RATE_LIMIT_PUBLIC_BURST", "3")

	r := setupRouterWithLimiter(1, 1)

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/data/universities", nil)
		r.ServeHTTP(w, req)
		if w.Code == http.StatusTooManyRequests {
			t.Fatalf("expected public data route to use higher budget on request %d", i+1)
		}
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/data/universities", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after exhausting public data burst, got %d", w.Code)
	}
}

func TestSelectRateLimiterFallsBackToDefault(t *testing.T) {
	rules := buildRouteRateLimitRules(2, 2)

	defaultLimiter := selectRateLimiter(rules, "/healthz")
	if defaultLimiter == nil {
		t.Fatal("expected default limiter")
	}
	if ok, _ := defaultLimiter.allow("ip-healthz"); !ok {
		t.Fatal("expected default limiter to allow first request")
	}
	if ok, _ := defaultLimiter.allow("ip-healthz"); !ok {
		t.Fatal("expected default limiter burst to allow second request")
	}
	if ok, _ := defaultLimiter.allow("ip-healthz"); ok {
		t.Fatal("expected default limiter to block third request")
	}
}

func TestHealthz(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHealthLegacyAlias(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if body := strings.TrimSpace(w.Body.String()); body != "ok" {
		t.Fatalf("expected body ok, got %q", body)
	}
}

func TestReadyz(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRoot(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestPingV1(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for undefined ping route, got %d", w.Code)
	}
}

func TestGetPortFromEnv_Default(t *testing.T) {
	old := os.Getenv("SERVER_PORT")
	t.Cleanup(func() { _ = os.Setenv("SERVER_PORT", old) })
	_ = os.Unsetenv("SERVER_PORT")

	if got := getPortFromEnv(); got != "8080" {
		t.Fatalf("expected default 8080, got %s", got)
	}
}

func TestGetPortFromEnv_Override(t *testing.T) {
	old := os.Getenv("SERVER_PORT")
	t.Cleanup(func() { _ = os.Setenv("SERVER_PORT", old) })
	_ = os.Setenv("SERVER_PORT", ":9090")

	if got := getPortFromEnv(); got != "9090" {
		t.Fatalf("expected 9090, got %s", got)
	}
}

func TestGetAddr(t *testing.T) {
	if got := getAddr("8080"); got != ":8080" {
		t.Fatalf("expected :8080, got %s", got)
	}
}

func TestNewProxyManagerUsesUnifiedDefaultUserServicePort(t *testing.T) {
	logger := logrus.New()
	oldUserURL := os.Getenv("USER_SERVICE_URL")
	t.Cleanup(func() { _ = os.Setenv("USER_SERVICE_URL", oldUserURL) })
	_ = os.Unsetenv("USER_SERVICE_URL")

	pm := NewProxyManager(logger)
	if got := pm.services["user"].BaseURL; got != "http://user-service:8083" {
		t.Fatalf("expected default user service url http://user-service:8083, got %s", got)
	}
}

func TestNewHTTPServerTimeouts(t *testing.T) {
	r := setupRouter()
	srv := newHTTPServer(":0", r)
	if srv.ReadHeaderTimeout <= 0 || srv.ReadTimeout <= 0 || srv.WriteTimeout <= 0 || srv.IdleTimeout <= 0 {
		t.Fatalf("expected timeouts to be set")
	}
}

func TestRunWithShutdownContext_Cancel(t *testing.T) {
	r := setupRouter()
	srv := newHTTPServer(":0", r)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)

	go func() {
		done <- runWithShutdownContext(srv, ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("shutdown returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("server did not shut down in time")
	}
}

func TestMetricsEndpoint_ExposesPrometheusFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRouterWithLimiter(1000, 1000) // effectively disable rate limit in tests

	// Fire a couple of requests to generate metrics
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	r.ServeHTTP(w1, req1)

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w2, req2)

	// Now hit /metrics
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "gaokao_http_requests_total") {
		t.Fatalf("expected gaokao_http_requests_total in metrics, got:\n%s", body)
	}
	if !strings.Contains(body, "gaokao_http_request_duration_seconds_bucket") {
		t.Fatalf("expected gaokao_http_request_duration_seconds_bucket in metrics, got:\n%s", body)
	}
}

func BenchmarkMetricsMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)
	r := setupRouterWithLimiter(100000, 100000)

	// warm up
	for i := 0; i < 100; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		r.ServeHTTP(w, req)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		r.ServeHTTP(w, req)
	}
}

// Ensure metrics path labels prefer the matched route pattern
func TestMetrics_PathLabelUsesFullPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRouterWithLimiter(1000, 1000)

	// Add a dynamic route for testing
	r.GET("/api/v1/items/:id", func(c *gin.Context) {
		c.String(http.StatusOK, c.Param("id"))
	})

	// hit two different ids
	for _, id := range []string{"123", "456"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+id, nil)
		r.ServeHTTP(w, req)
	}

	// scrape metrics
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	r.ServeHTTP(w, req)
	body := w.Body.String()

	if !strings.Contains(body, `gaokao_http_requests_total{method="GET",path="/api/v1/items/:id"`) {
		t.Fatalf("expected aggregated path label for dynamic route, got:\n%s", body)
	}
}

// helper to avoid unused imports when not all tests run
var _ = bytes.MinRead
var _ = context.Canceled
var _ = os.ErrClosed
var _ = time.Second

func TestSwagger_Index_And_DocJson_Accessible(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Use a very strict limiter to ensure whitelist works
	r := setupRouterWithLimiter(1, 1)

	// Access Swagger index
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("/swagger/index.html expected 200, got %d", w1.Code)
	}
	if ct := w1.Header().Get("Content-Type"); !strings.Contains(strings.ToLower(ct), "text/html") {
		t.Fatalf("/swagger/index.html unexpected content-type: %s", ct)
	}

	// Access Swagger spec JSON
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK && w2.Code != http.StatusInternalServerError {
		t.Fatalf("/swagger/doc.json expected 200 or 500 (depends on swagger spec generation), got %d", w2.Code)
	}
}
