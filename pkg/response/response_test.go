package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// performGet 发起请求并返回 (statusCode, decoded body)。
func performGet(handler gin.HandlerFunc, headers map[string]string) (int, APIResponse) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	c.Request = req
	handler(c)
	var body APIResponse
	_ = json.NewDecoder(strings.NewReader(rec.Body.String())).Decode(&body)
	return rec.Code, body
}

func TestOK_Success(t *testing.T) {
	status, body := performGet(func(c *gin.Context) {
		OK(c, gin.H{"id": 1})
	}, nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d", status)
	}
	if !body.Success {
		t.Fatal("expected success=true")
	}
	if body.Message != "操作成功" {
		t.Errorf("default message = %q", body.Message)
	}
	if body.Timestamp == 0 {
		t.Error("timestamp not set")
	}
}

func TestOKWithMessage_PicksUpRequestID(t *testing.T) {
	status, body := performGet(func(c *gin.Context) {
		OKWithMessage(c, gin.H{"x": true}, "done")
	}, map[string]string{RequestIDKey: "req-abc"})
	if status != http.StatusOK {
		t.Fatalf("status = %d", status)
	}
	if body.Message != "done" {
		t.Errorf("message = %q", body.Message)
	}
	if body.RequestID != "req-abc" {
		t.Errorf("request_id = %q", body.RequestID)
	}
}

func TestBadRequest_BuildsErrorInfo(t *testing.T) {
	status, body := performGet(func(c *gin.Context) {
		BadRequest(c, "INVALID_PARAM", "field x missing", map[string]any{"field": "x"})
	}, nil)
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d", status)
	}
	if body.Success {
		t.Fatal("expected success=false")
	}
	if body.Error == nil || body.Error.Code != "INVALID_PARAM" {
		t.Fatalf("error = %+v", body.Error)
	}
	if body.Error.Message != "field x missing" {
		t.Errorf("error.message = %q", body.Error.Message)
	}
}

func TestInternalError_DefaultsCleanly(t *testing.T) {
	status, body := performGet(func(c *gin.Context) {
		InternalError(c, "INTERNAL", "boom", nil)
	}, nil)
	if status != http.StatusInternalServerError {
		t.Fatalf("status = %d", status)
	}
	if body.Error == nil || body.Error.Details != nil {
		t.Errorf("details should be nil, got %+v", body.Error)
	}
}

func TestAbortWithError_AbortsChain(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	AbortWithError(c, http.StatusForbidden, "FORBIDDEN", "no access", nil)
	if !c.IsAborted() {
		t.Fatal("expected aborted")
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d", rec.Code)
	}
}
