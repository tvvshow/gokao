package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// IdempotencyHeader 是客户端用来声明幂等键的 HTTP 头部。
const IdempotencyHeader = "X-Idempotency-Key"

// IdempotencyStore 抽象幂等所需的最小存储语义。各服务用自己已有的 Redis/任意 KV 客户端实现，
// pkg/middleware 不引入具体 SDK 依赖。
type IdempotencyStore interface {
	// SetNX 原子地"键不存在则写入"。返回 true 表示当前调用是首请求。
	SetNX(ctx context.Context, key, value string, ttl time.Duration) (bool, error)
	// Get 读取键值；键不存在时返回 ("", nil)，区分于真实错误。
	Get(ctx context.Context, key string) (string, error)
	// Set 无条件写入。
	Set(ctx context.Context, key, value string, ttl time.Duration) error
}

type idempotencyPayload struct {
	Status      int               `json:"status"`
	ContentType string            `json:"content_type,omitempty"`
	Body        []byte            `json:"body,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

type idempotencyResponseRecorder struct {
	gin.ResponseWriter
	buf *bytes.Buffer
}

func (w *idempotencyResponseRecorder) Write(b []byte) (int, error) {
	w.buf.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *idempotencyResponseRecorder) WriteString(s string) (int, error) {
	w.buf.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// Idempotency 返回基于 X-Idempotency-Key 头的幂等中间件。
//
// 语义：
//   - 请求未携带 header：放行（不强制启用，保持向后兼容）。
//   - 首次请求获得锁：执行 handler；handler 返回 2xx 时把 (status, body) 缓存 ttl。
//   - 同 key 重放且已有缓存结果：原样回放缓存的状态码与响应体，handler 不再执行。
//   - 同 key 重放但首请求仍在处理：返回 409 Conflict，提示客户端稍后重试。
//   - 存储层报错（Redis 抖动等）：不阻断业务，原路放行并通过响应头 X-Idempotency-Status 标记。
//
// ttl 建议 ≥ 客户端最大重试窗口（支付场景 24h 是常用值）。
func Idempotency(store IdempotencyStore, ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader(IdempotencyHeader)
		if key == "" {
			c.Next()
			return
		}

		lockKey := "idem:lock:" + key
		resultKey := "idem:result:" + key
		ctx := c.Request.Context()

		acquired, err := store.SetNX(ctx, lockKey, "1", ttl)
		if err != nil {
			c.Header("X-Idempotency-Status", "store-error")
			c.Next()
			return
		}

		if !acquired {
			cached, getErr := store.Get(ctx, resultKey)
			if getErr == nil && cached != "" {
				var payload idempotencyPayload
				if json.Unmarshal([]byte(cached), &payload) == nil {
					for k, v := range payload.Headers {
						c.Header(k, v)
					}
					c.Header("X-Idempotency-Status", "replayed")
					contentType := payload.ContentType
					if contentType == "" {
						contentType = "application/json; charset=utf-8"
					}
					c.Data(payload.Status, contentType, payload.Body)
					c.Abort()
					return
				}
			}
			c.Header("X-Idempotency-Status", "in-flight")
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"error":   "idempotency_in_flight",
				"message": "another request with the same Idempotency-Key is being processed",
			})
			return
		}

		recorder := &idempotencyResponseRecorder{
			ResponseWriter: c.Writer,
			buf:            &bytes.Buffer{},
		}
		c.Writer = recorder
		c.Header("X-Idempotency-Status", "stored")

		c.Next()

		status := recorder.Status()
		if status < http.StatusOK || status >= http.StatusMultipleChoices {
			// 非 2xx 不缓存：留给客户端用同 key 重试纠错；锁本身仍在，避免并发重复落库。
			return
		}

		payload := idempotencyPayload{
			Status:      status,
			ContentType: recorder.Header().Get("Content-Type"),
			Body:        recorder.buf.Bytes(),
		}
		raw, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			return
		}
		_ = store.Set(ctx, resultKey, string(raw), ttl)
	}
}
