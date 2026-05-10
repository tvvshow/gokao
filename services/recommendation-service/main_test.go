package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	pkgmiddleware "github.com/tvvshow/gokao/pkg/middleware"
)

func TestContextHeadersMiddlewarePropagatesIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(pkgmiddleware.RequestID(), pkgmiddleware.TraceID())
	r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Request-ID", "req-rec-1")
	req.Header.Set("X-Trace-ID", "trace-rec-1")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if got := w.Header().Get("X-Request-ID"); got != "req-rec-1" {
		t.Fatalf("expected request id propagation, got %q", got)
	}
	if got := w.Header().Get("X-Trace-ID"); got != "trace-rec-1" {
		t.Fatalf("expected trace id propagation, got %q", got)
	}
}

func TestContextHeadersMiddlewareGeneratesIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(pkgmiddleware.RequestID(), pkgmiddleware.TraceID())
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

func TestRecommendationCORSAllowsTraceHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(pkgmiddleware.RequestID(), pkgmiddleware.TraceID())
	r.Use(pkgmiddleware.CORS(pkgmiddleware.DefaultCORSConfig()))
	r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(got, "X-Trace-ID") {
		t.Fatalf("expected CORS allow headers to include X-Trace-ID, got %q", got)
	}
	if got := w.Header().Get("Access-Control-Expose-Headers"); !strings.Contains(got, "X-Trace-ID") {
		t.Fatalf("expected CORS expose headers to include X-Trace-ID, got %q", got)
	}
}
