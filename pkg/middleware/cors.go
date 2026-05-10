package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSConfig CORS 配置
type CORSConfig struct {
	AllowOrigins     []string // 允许的源，["*"] 表示全部允许
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int // 预检缓存秒数，默认 3600
}

// DefaultCORSConfig 返回开发环境默认配置（允许所有源）
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Authorization", "X-Request-ID", "X-Trace-ID"},
		ExposeHeaders: []string{"X-Request-ID", "X-Trace-ID"},
		MaxAge:        3600,
	}
}

// CORS 返回标准 CORS 中间件。
// 使用 DefaultCORSConfig() 作为快速启动，或自定义 CORSConfig。
func CORS(cfg CORSConfig) gin.HandlerFunc {
	if cfg.MaxAge == 0 {
		cfg.MaxAge = 3600
	}
	if len(cfg.AllowMethods) == 0 {
		cfg.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}
	if len(cfg.AllowHeaders) == 0 {
		cfg.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "X-Request-ID", "X-Trace-ID"}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			// 非 CORS 请求，直接放行
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
			c.Next()
			return
		}

		// 判断 origin 是否被允许
		allowed := false
		for _, o := range cfg.AllowOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if !allowed {
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
			c.Next()
			return
		}

		// 设置 CORS 响应头
		if len(cfg.AllowOrigins) == 1 && cfg.AllowOrigins[0] == "*" {
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}

		c.Header("Access-Control-Allow-Methods", joinStrings(cfg.AllowMethods))
		c.Header("Access-Control-Allow-Headers", joinStrings(cfg.AllowHeaders))

		if len(cfg.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", joinStrings(cfg.ExposeHeaders))
		}
		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Max-Age", itoa(cfg.MaxAge))
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func joinStrings(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	result := ss[0]
	for i := 1; i < len(ss); i++ {
		result += ", " + ss[i]
	}
	return result
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
