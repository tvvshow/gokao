package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func TestLoggerPropagatesRequestAndTraceID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	r := gin.New()
	r.Use(Logger(logger))
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Request-ID", "req-data-1")
	req.Header.Set("X-Trace-ID", "trace-data-1")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if got := w.Header().Get("X-Request-ID"); got != "req-data-1" {
		t.Fatalf("expected request id propagation, got %q", got)
	}
	if got := w.Header().Get("X-Trace-ID"); got != "trace-data-1" {
		t.Fatalf("expected trace id propagation, got %q", got)
	}
}

func TestCORSAllowsTraceIDHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS())
	r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Headers"); got == "" || !strings.Contains(got, "X-Trace-ID") {
		t.Fatalf("expected CORS allow headers to include X-Trace-ID, got %q", got)
	}
	if got := w.Header().Get("Access-Control-Expose-Headers"); got == "" || !strings.Contains(got, "X-Trace-ID") {
		t.Fatalf("expected CORS expose headers to include X-Trace-ID, got %q", got)
	}
}
