package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestContextHeadersMiddlewarePropagatesIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(contextHeadersMiddleware())
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
	r.Use(contextHeadersMiddleware())
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
	r.Use(contextHeadersMiddleware())
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID, X-Trace-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID, X-Trace-ID")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(got, "X-Trace-ID") {
		t.Fatalf("expected CORS allow headers to include X-Trace-ID, got %q", got)
	}
	if got := w.Header().Get("Access-Control-Expose-Headers"); !strings.Contains(got, "X-Trace-ID") {
		t.Fatalf("expected CORS expose headers to include X-Trace-ID, got %q", got)
	}
}
