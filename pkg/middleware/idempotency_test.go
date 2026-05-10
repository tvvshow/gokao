package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// fakeStore 是测试用的内存 IdempotencyStore，避免依赖真实 Redis。
type fakeStore struct {
	mu         sync.Mutex
	values     map[string]string
	expiry     map[string]time.Time
	setNXErr   error
	getErr     error
	setErr     error
	setNXCalls int
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		values: map[string]string{},
		expiry: map[string]time.Time{},
	}
}

func (s *fakeStore) SetNX(ctx context.Context, key, value string, ttl time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.setNXCalls++
	if s.setNXErr != nil {
		return false, s.setNXErr
	}
	if exp, ok := s.expiry[key]; ok && time.Now().Before(exp) {
		return false, nil
	}
	s.values[key] = value
	s.expiry[key] = time.Now().Add(ttl)
	return true, nil
}

func (s *fakeStore) Get(ctx context.Context, key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.getErr != nil {
		return "", s.getErr
	}
	if exp, ok := s.expiry[key]; ok && time.Now().After(exp) {
		delete(s.values, key)
		delete(s.expiry, key)
		return "", nil
	}
	return s.values[key], nil
}

func (s *fakeStore) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.setErr != nil {
		return s.setErr
	}
	s.values[key] = value
	s.expiry[key] = time.Now().Add(ttl)
	return nil
}

func newTestRouter(store IdempotencyStore, handler gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/pay", Idempotency(store, time.Minute), handler)
	return r
}

func performPOST(t *testing.T, r http.Handler, key string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/pay", bytes.NewBufferString(`{"amount":1}`))
	req.Header.Set("Content-Type", "application/json")
	if key != "" {
		req.Header.Set(IdempotencyHeader, key)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestIdempotency_NoHeaderPassesThrough(t *testing.T) {
	store := newFakeStore()
	calls := 0
	r := newTestRouter(store, func(c *gin.Context) {
		calls++
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := performPOST(t, r, "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if calls != 1 {
		t.Fatalf("handler should have run once, ran %d times", calls)
	}
	if store.setNXCalls != 0 {
		t.Fatalf("store should not have been touched, got %d SetNX calls", store.setNXCalls)
	}
}

func TestIdempotency_FirstRequestStoresResult(t *testing.T) {
	store := newFakeStore()
	r := newTestRouter(store, func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"order_no": "GK123"})
	})

	w := performPOST(t, r, "key-1")
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	if w.Header().Get("X-Idempotency-Status") != "stored" {
		t.Fatalf("expected stored header, got %q", w.Header().Get("X-Idempotency-Status"))
	}
	if _, ok := store.values["idem:result:key-1"]; !ok {
		t.Fatalf("expected result cached under idem:result:key-1; values=%v", store.values)
	}
}

func TestIdempotency_DuplicateReplaysCached(t *testing.T) {
	store := newFakeStore()
	calls := 0
	r := newTestRouter(store, func(c *gin.Context) {
		calls++
		c.JSON(http.StatusCreated, gin.H{"order_no": "GK456", "call": calls})
	})

	first := performPOST(t, r, "key-2")
	if first.Code != http.StatusCreated {
		t.Fatalf("first call: expected 201, got %d", first.Code)
	}

	second := performPOST(t, r, "key-2")
	if second.Code != http.StatusCreated {
		t.Fatalf("replay: expected 201, got %d", second.Code)
	}
	if second.Header().Get("X-Idempotency-Status") != "replayed" {
		t.Fatalf("expected replayed header, got %q", second.Header().Get("X-Idempotency-Status"))
	}
	if calls != 1 {
		t.Fatalf("handler should have executed exactly once, executed %d times", calls)
	}
	if !bytes.Equal(first.Body.Bytes(), second.Body.Bytes()) {
		t.Fatalf("replay body mismatch:\nfirst=%s\nsecond=%s", first.Body.String(), second.Body.String())
	}
}

func TestIdempotency_LockHeldButNoResultReturns409(t *testing.T) {
	store := newFakeStore()
	// 预置锁但不写结果：模拟首请求仍在执行。
	if _, err := store.SetNX(context.Background(), "idem:lock:key-3", "1", time.Minute); err != nil {
		t.Fatalf("preload lock: %v", err)
	}

	calls := 0
	r := newTestRouter(store, func(c *gin.Context) {
		calls++
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := performPOST(t, r, "key-3")
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict, got %d body=%s", w.Code, w.Body.String())
	}
	if w.Header().Get("X-Idempotency-Status") != "in-flight" {
		t.Fatalf("expected in-flight header, got %q", w.Header().Get("X-Idempotency-Status"))
	}
	if calls != 0 {
		t.Fatalf("handler must not run when lock contended; ran %d times", calls)
	}
}

func TestIdempotency_StoreErrorFailsOpen(t *testing.T) {
	store := newFakeStore()
	store.setNXErr = errors.New("redis down")
	calls := 0
	r := newTestRouter(store, func(c *gin.Context) {
		calls++
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := performPOST(t, r, "key-4")
	if w.Code != http.StatusOK {
		t.Fatalf("store error should fail open, got %d", w.Code)
	}
	if w.Header().Get("X-Idempotency-Status") != "store-error" {
		t.Fatalf("expected store-error header, got %q", w.Header().Get("X-Idempotency-Status"))
	}
	if calls != 1 {
		t.Fatalf("handler should still run when store is unavailable, got %d", calls)
	}
}

func TestIdempotency_Non2xxNotCached(t *testing.T) {
	store := newFakeStore()
	r := newTestRouter(store, func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad"})
	})

	w := performPOST(t, r, "key-5")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if _, ok := store.values["idem:result:key-5"]; ok {
		t.Fatalf("non-2xx response must not be cached; values=%v", store.values)
	}
}

func TestIdempotency_ReplayPreservesContentType(t *testing.T) {
	store := newFakeStore()
	r := newTestRouter(store, func(c *gin.Context) {
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.String(http.StatusCreated, `{"id":"abc"}`)
	})

	first := performPOST(t, r, "key-6")
	if first.Code != http.StatusCreated {
		t.Fatalf("first: expected 201, got %d", first.Code)
	}

	second := performPOST(t, r, "key-6")
	if got := second.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("replay content-type lost: %q", got)
	}

	var payload idempotencyPayload
	raw := store.values["idem:result:key-6"]
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("stored payload not valid JSON: %v", err)
	}
	if payload.Status != http.StatusCreated {
		t.Fatalf("stored status mismatch: %d", payload.Status)
	}
	body, _ := io.ReadAll(bytes.NewReader(payload.Body))
	if string(body) != `{"id":"abc"}` {
		t.Fatalf("stored body mismatch: %q", string(body))
	}
}
