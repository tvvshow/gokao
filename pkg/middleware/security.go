package middleware

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	JWTSecret        string
	RateLimitEnabled bool
	SecurityHeaders  bool
	JWTIssuer        string
	JWTAudience      string
}

// SecurityMiddleware 安全中间件
type SecurityMiddleware struct {
	config *SecurityConfig
	logger *logrus.Logger
}

// NewSecurityMiddleware 创建安全中间件
func NewSecurityMiddleware(config *SecurityConfig) *SecurityMiddleware {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	
	return &SecurityMiddleware{
		config: config,
		logger: logger,
	}
}

// JWTClaims 自定义JWT声明
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// TokenBlacklist token黑名单接口
type TokenBlacklist interface {
	IsBlacklisted(token string) bool
	AddToBlacklist(token string, expiry time.Duration) error
}

// InMemoryTokenBlacklist 内存中的token黑名单
type InMemoryTokenBlacklist struct {
	blacklistedTokens map[string]time.Time
}

// NewInMemoryTokenBlacklist 创建内存黑名单
func NewInMemoryTokenBlacklist() *InMemoryTokenBlacklist {
	return &InMemoryTokenBlacklist{
		blacklistedTokens: make(map[string]time.Time),
	}
}

// IsBlacklisted 检查token是否在黑名单中
func (b *InMemoryTokenBlacklist) IsBlacklisted(token string) bool {
	expiry, exists := b.blacklistedTokens[token]
	if !exists {
		return false
	}

	// 如果已过期，从黑名单中移除
	if time.Now().After(expiry) {
		delete(b.blacklistedTokens, token)
		return false
	}

	return true
}

// AddToBlacklist 添加token到黑名单
func (b *InMemoryTokenBlacklist) AddToBlacklist(token string, expiry time.Duration) error {
	b.blacklistedTokens[token] = time.Now().Add(expiry)
	return nil
}

// SecurityHeaders 安全头中间件
func (sm *SecurityMiddleware) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		if sm.config.SecurityHeaders {
			// 设置安全头
			c.Header("X-Content-Type-Options", "nosniff")
			c.Header("X-Frame-Options", "DENY")
			c.Header("X-XSS-Protection", "1; mode=block")
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			c.Header("Content-Security-Policy", "default-src 'self'")
		}
		c.Next()
	}
}

// CORS 跨域中间件
func (sm *SecurityMiddleware) CORS(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// 检查是否在允许的源列表中
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if allowedOrigin == origin || allowedOrigin == "*" {
				allowed = true
				break
			}
		}
		
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With, X-Request-ID")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "3600")
		}
		
		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}

// EnhancedJWTAuth 增强的JWT认证中间件
func (sm *SecurityMiddleware) EnhancedJWTAuth(blacklist TokenBlacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Token提取
		tokenString, err := sm.extractToken(c)
		if err != nil {
			sm.logger.WithError(err).Warn("Token extraction failed")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "认证信息缺失或格式错误",
			})
			c.Abort()
			return
		}

		// 2. Token黑名单检查
		if blacklist != nil && blacklist.IsBlacklisted(tokenString) {
			sm.logger.Warn("Token is blacklisted")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "token_revoked",
				"message": "认证令牌已失效",
			})
			c.Abort()
			return
		}

		// 3. Token验证
		claims, err := sm.validateToken(tokenString)
		if err != nil {
			sm.logger.WithError(err).Warn("Token validation failed")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_token",
				"message": "无效的认证令牌",
			})
			c.Abort()
			return
		}

		// 4. 权限检查
		if !sm.checkPermissions(claims, c.Request) {
			sm.logger.Warn("Permission denied")
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "permission_denied",
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		// 5. 存储用户信息到上下文
		sm.storeUserInfo(c, claims)

		c.Next()
	}
}

// extractToken 从请求中提取token
func (sm *SecurityMiddleware) extractToken(c *gin.Context) (string, error) {
	// 从Authorization头获取
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	// 检查Bearer前缀
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("invalid authorization format")
	}

	// 提取token
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" {
		return "", fmt.Errorf("empty token")
	}

	return tokenString, nil
}

// validateToken 验证JWT token
func (sm *SecurityMiddleware) validateToken(tokenString string) (*JWTClaims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(sm.config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证claims
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		// 验证发行者
		if sm.config.JWTIssuer != "" && claims.Issuer != sm.config.JWTIssuer {
			return nil, fmt.Errorf("invalid issuer")
		}

		// 验证受众
		if sm.config.JWTAudience != "" && !claims.VerifyAudience(sm.config.JWTAudience, true) {
			return nil, fmt.Errorf("invalid audience")
		}

		// 验证过期时间
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			return nil, fmt.Errorf("token expired")
		}

		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// checkPermissions 检查权限
func (sm *SecurityMiddleware) checkPermissions(claims *JWTClaims, req *http.Request) bool {
	// 这里可以根据实际业务需求实现更复杂的权限检查
	// 例如基于角色的访问控制(RBAC)或基于属性的访问控制(ABAC)

	// 示例：管理员可以访问所有资源
	if claims.Role == "admin" {
		return true
	}

	// 示例：普通用户只能访问GET请求（只读）
	if claims.Role == "user" && req.Method == "GET" {
		return true
	}

	// 其他角色或方法需要具体实现
	return false
}

// storeUserInfo 存储用户信息到上下文
func (sm *SecurityMiddleware) storeUserInfo(c *gin.Context, claims *JWTClaims) {
	c.Set("user_id", claims.UserID)
	c.Set("username", claims.Username)
	c.Set("role", claims.Role)
	c.Set("jwt_claims", claims)
}

// GenerateToken 生成JWT token
func (sm *SecurityMiddleware) GenerateToken(userID, username, role string, expiry time.Duration) (string, error) {
	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    sm.config.JWTIssuer,
			Audience:  jwt.ClaimStrings{sm.config.JWTAudience},
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(sm.config.JWTSecret))
}

// JWTAuth JWT认证中间件（保持向后兼容）
func (sm *SecurityMiddleware) JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少认证信息"})
			c.Abort()
			return
		}
		
		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证格式错误"})
			c.Abort()
			return
		}
		
		// 提取token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		
		// 验证token（这里简化处理，实际应该验证JWT）
		if tokenString != sm.config.JWTSecret {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌"})
			c.Abort()
			return
		}
		
		// 将用户信息存储到上下文中
		c.Set("user_id", "test_user")
		c.Next()
	}
}

// InputValidationMiddleware 输入验证中间件
type InputValidationMiddleware struct {
	maxBodySize        int64
	logger             *logrus.Logger
	sqlInjectionRegex  *regexp.Regexp
	xssRegex          *regexp.Regexp
	allowedContentTypes []string
}

// InputValidationConfig 输入验证配置
type InputValidationConfig struct {
	MaxBodySize        int64
	Logger             *logrus.Logger
	AllowedContentTypes []string
}

// NewInputValidationMiddleware 创建输入验证中间件
func NewInputValidationMiddleware(config *InputValidationConfig) *InputValidationMiddleware {
	return &InputValidationMiddleware{
		maxBodySize:        config.MaxBodySize,
		logger:             config.Logger,
		allowedContentTypes: config.AllowedContentTypes,
		sqlInjectionRegex:  regexp.MustCompile(`(?i)(?:\b(?:select|insert|update|delete|drop|alter|union|exec|execute|truncate|declare|xp_cmdshell|;\s*--|/\*.*\*/)\b)`),
		xssRegex:          regexp.MustCompile(`(?i)(?:<script|javascript:|on\w+\s*=|eval\(|alert\(|document\.cookie|window\.location)`),
	}
}

// Middleware 输入验证中间件
func (m *InputValidationMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 请求体大小限制
		if !m.checkBodySize(c) {
			m.logger.Warn("Request body too large")
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":   "request_too_large",
				"message": "请求体过大",
			})
			c.Abort()
			return
		}

		// 2. 内容类型验证
		if !m.validateContentType(c) {
			m.logger.Warn("Invalid content type")
			c.JSON(http.StatusUnsupportedMediaType, gin.H{
				"error":   "unsupported_media_type",
				"message": "不支持的内容类型",
			})
			c.Abort()
			return
		}

		// 3. SQL注入检测
		if m.detectSQLInjection(c) {
			m.logger.Warn("SQL injection attempt detected")
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_input",
				"message": "输入包含非法字符",
			})
			c.Abort()
			return
		}

		// 4. XSS攻击检测
		if m.detectXSS(c) {
			m.logger.Warn("XSS attempt detected")
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_input",
				"message": "输入包含非法字符",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkBodySize 检查请求体大小
func (m *InputValidationMiddleware) checkBodySize(c *gin.Context) bool {
	if c.Request.ContentLength > m.maxBodySize {
		return false
	}
	return true
}

// validateContentType 验证内容类型
func (m *InputValidationMiddleware) validateContentType(c *gin.Context) bool {
	// 对于GET、HEAD、OPTIONS请求，不需要验证内容类型
	if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
		return true
	}

	// 如果没有设置允许的内容类型，允许所有类型
	if len(m.allowedContentTypes) == 0 {
		return true
	}

	contentType := c.Request.Header.Get("Content-Type")
	if contentType == "" {
		return false
	}

	// 检查是否在允许的内容类型列表中
	for _, allowedType := range m.allowedContentTypes {
		if strings.HasPrefix(contentType, allowedType) {
			return true
		}
	}

	return false
}

// detectSQLInjection 检测SQL注入
func (m *InputValidationMiddleware) detectSQLInjection(c *gin.Context) bool {
	// 检查URL参数
	for _, values := range c.Request.URL.Query() {
		for _, value := range values {
			if m.sqlInjectionRegex.MatchString(value) {
				return true
			}
		}
	}

	// 检查表单参数
	if c.Request.Form != nil {
		for _, values := range c.Request.Form {
			for _, value := range values {
				if m.sqlInjectionRegex.MatchString(value) {
					return true
				}
			}
		}
	}

	return false
}

// detectXSS 检测XSS攻击
func (m *InputValidationMiddleware) detectXSS(c *gin.Context) bool {
	// 检查URL参数
	for _, values := range c.Request.URL.Query() {
		for _, value := range values {
			if m.xssRegex.MatchString(value) {
				return true
			}
		}
	}

	// 检查表单参数
	if c.Request.Form != nil {
		for _, values := range c.Request.Form {
			for _, value := range values {
				if m.xssRegex.MatchString(value) {
					return true
				}
			}
		}
	}

	return false
}

// DefaultInputValidationMiddleware 默认输入验证中间件
func DefaultInputValidationMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	config := &InputValidationConfig{
		MaxBodySize:        10 * 1024 * 1024, // 10MB
		Logger:             logger,
		AllowedContentTypes: []string{"application/json", "application/x-www-form-urlencoded", "multipart/form-data"},
	}
	
	middleware := NewInputValidationMiddleware(config)
	return middleware.Middleware()
}