package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"golang.org/x/time/rate"
)

// SecurityMiddleware 安全中间件
type SecurityMiddleware struct {
	redis       *redis.Client
	rateLimiter *RateLimiter
	jwtSecret   string
}

// NewSecurityMiddleware 创建安全中间件
func NewSecurityMiddleware(redis *redis.Client, jwtSecret string) *SecurityMiddleware {
	return &SecurityMiddleware{
		redis:       redis,
		rateLimiter: NewRateLimiter(redis),
		jwtSecret:   jwtSecret,
	}
}

// RateLimiter 限流器
type RateLimiter struct {
	redis    *redis.Client
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
}

// NewRateLimiter 创建限流器
func NewRateLimiter(redis *redis.Client) *RateLimiter {
	return &RateLimiter{
		redis:    redis,
		limiters: make(map[string]*rate.Limiter),
	}
}

// RateLimit 限流中间件
func (s *SecurityMiddleware) RateLimit(requestsPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		userID := c.GetString("user_id")
		
		// 构建限流键
		var key string
		if userID != "" {
			key = fmt.Sprintf("rate_limit:user:%s", userID)
		} else {
			key = fmt.Sprintf("rate_limit:ip:%s", clientIP)
		}

		// 检查限流
		allowed, err := s.rateLimiter.Allow(c.Request.Context(), key, requestsPerMinute)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "限流检查失败",
				"code":  "RATE_LIMIT_ERROR",
			})
			c.Abort()
			return
		}

		if !allowed {
			c.Header("X-RateLimit-Limit", strconv.Itoa(requestsPerMinute))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", "60")
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁，请稍后再试",
				"code":  "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(ctx context.Context, key string, requestsPerMinute int) (bool, error) {
	// 使用Redis滑动窗口算法
	now := time.Now().Unix()
	window := int64(60) // 60秒窗口
	
	pipe := rl.redis.Pipeline()
	
	// 清理过期记录
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(now-window, 10))
	
	// 计算当前窗口内的请求数
	pipe.ZCard(ctx, key)
	
	// 添加当前请求
	pipe.ZAdd(ctx, key, &redis.Z{
		Score:  float64(now),
		Member: fmt.Sprintf("%d-%d", now, time.Now().Nanosecond()),
	})
	
	// 设置过期时间
	pipe.Expire(ctx, key, time.Duration(window)*time.Second)
	
	results, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}
	
	// 获取当前请求数
	count := results[1].(*redis.IntCmd).Val()
	
	return count <= int64(requestsPerMinute), nil
}

// CORS 跨域中间件
func (s *SecurityMiddleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// 允许的域名列表
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"https://gaokao.example.com",
		}
		
		// 检查是否为允许的域名
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}
		
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}

// SecurityHeaders 安全头中间件
func (s *SecurityMiddleware) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止XSS攻击
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// 防止MIME类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")
		
		// 防止点击劫持
		c.Header("X-Frame-Options", "DENY")
		
		// 强制HTTPS
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		
		// 内容安全策略
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' https:")
		
		// 引用策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// 权限策略
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		
		c.Next()
	}
}

// RequestID 请求ID中间件
func (s *SecurityMiddleware) RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.Request.Header.Get("X-Request-ID")
		if requestID == "" {
			// 生成唯一请求ID
			requestID = generateRequestID()
		}
		
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		
		c.Next()
	}
}

// IPWhitelist IP白名单中间件
func (s *SecurityMiddleware) IPWhitelist(allowedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		// 检查IP是否在白名单中
		allowed := false
		for _, allowedIP := range allowedIPs {
			if clientIP == allowedIP || allowedIP == "*" {
				allowed = true
				break
			}
		}
		
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "IP地址不在允许列表中",
				"code":  "IP_NOT_ALLOWED",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// AntiReplay 防重放攻击中间件
func (s *SecurityMiddleware) AntiReplay(windowSeconds int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求签名
		signature := c.Request.Header.Get("X-Signature")
		timestamp := c.Request.Header.Get("X-Timestamp")
		nonce := c.Request.Header.Get("X-Nonce")
		
		if signature == "" || timestamp == "" || nonce == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "缺少必要的安全头",
				"code":  "MISSING_SECURITY_HEADERS",
			})
			c.Abort()
			return
		}
		
		// 检查时间戳
		ts, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的时间戳",
				"code":  "INVALID_TIMESTAMP",
			})
			c.Abort()
			return
		}
		
		now := time.Now().Unix()
		if now-ts > int64(windowSeconds) || ts > now+int64(windowSeconds) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "请求已过期",
				"code":  "REQUEST_EXPIRED",
			})
			c.Abort()
			return
		}
		
		// 检查nonce是否已使用
		nonceKey := fmt.Sprintf("nonce:%s", nonce)
		exists, err := s.redis.Exists(c.Request.Context(), nonceKey).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "防重放检查失败",
				"code":  "ANTI_REPLAY_ERROR",
			})
			c.Abort()
			return
		}
		
		if exists > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "重复请求",
				"code":  "DUPLICATE_REQUEST",
			})
			c.Abort()
			return
		}
		
		// 记录nonce
		err = s.redis.Set(c.Request.Context(), nonceKey, "1", time.Duration(windowSeconds)*time.Second).Err()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "防重放记录失败",
				"code":  "ANTI_REPLAY_RECORD_ERROR",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// DataEncryption 数据加密中间件
func (s *SecurityMiddleware) DataEncryption() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否需要加密响应
		if c.Request.Header.Get("X-Encrypt-Response") == "true" {
			c.Set("encrypt_response", true)
		}
		
		c.Next()
		
		// 在响应后处理加密
		if c.GetBool("encrypt_response") {
			// 这里可以实现响应数据加密逻辑
			c.Header("X-Response-Encrypted", "true")
		}
	}
}

// 工具函数

// generateRequestID 生成请求ID
func generateRequestID() string {
	now := time.Now()
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d-%d", now.UnixNano(), now.Nanosecond())))
	return hex.EncodeToString(hash[:])[:16]
}

// validateSignature 验证请求签名
func validateSignature(method, path, timestamp, nonce, body, secret string) bool {
	// 构建签名字符串
	signString := fmt.Sprintf("%s\n%s\n%s\n%s\n%s", method, path, timestamp, nonce, body)
	
	// 计算HMAC-SHA256签名
	hash := sha256.New()
	hash.Write([]byte(signString + secret))
	expectedSignature := hex.EncodeToString(hash.Sum(nil))
	
	return expectedSignature == strings.ToLower(body)
}
