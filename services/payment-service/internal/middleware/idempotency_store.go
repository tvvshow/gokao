package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	pkgmiddleware "github.com/tvvshow/gokao/pkg/middleware"
)

// redisIdempotencyStore 把 go-redis/v8 客户端适配为 pkgmiddleware.IdempotencyStore。
// 之所以放在服务侧而非 pkg/middleware，是为了让 pkg/middleware 保持零 Redis SDK 依赖，
// 各服务可基于自身已选用的客户端版本（v8/v9/其他 KV）独立实现。
type redisIdempotencyStore struct {
	client *redis.Client
}

// NewRedisIdempotencyStore 用 redis v8 客户端构造幂等存储实现。
func NewRedisIdempotencyStore(client *redis.Client) pkgmiddleware.IdempotencyStore {
	return &redisIdempotencyStore{client: client}
}

func (s *redisIdempotencyStore) SetNX(ctx context.Context, key, value string, ttl time.Duration) (bool, error) {
	return s.client.SetNX(ctx, key, value, ttl).Result()
}

func (s *redisIdempotencyStore) Get(ctx context.Context, key string) (string, error) {
	val, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (s *redisIdempotencyStore) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return s.client.Set(ctx, key, value, ttl).Err()
}

// Idempotency 基于 X-Idempotency-Key 头的幂等中间件；用 Redis 抢锁 + 缓存首请求响应。
// ttl 应覆盖客户端的最大重试窗口，支付场景建议 24h。
func Idempotency(client *redis.Client, ttl time.Duration) gin.HandlerFunc {
	return pkgmiddleware.Idempotency(NewRedisIdempotencyStore(client), ttl)
}
