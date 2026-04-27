package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestContextHeadersPropagatesRequestAndTraceID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ContextHeaders())
	r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Request-ID", "req-user-1")
	req.Header.Set("X-Trace-ID", "trace-user-1")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if got := w.Header().Get("X-Request-ID"); got != "req-user-1" {
		t.Fatalf("expected request id propagation, got %q", got)
	}
	if got := w.Header().Get("X-Trace-ID"); got != "trace-user-1" {
		t.Fatalf("expected trace id propagation, got %q", got)
	}
}

func TestContextHeadersGeneratesRequestAndTraceID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ContextHeaders())
	r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if got := w.Header().Get("X-Request-ID"); got == "" {
		t.Fatal("expected generated X-Request-ID")
	}
	if got := w.Header().Get("X-Trace-ID"); got == "" {
		t.Fatal("expected generated X-Trace-ID")
	}
}
