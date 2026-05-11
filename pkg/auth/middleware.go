package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/tvvshow/gokao/pkg/response"
)

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	jwtSecret string
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: jwtSecret,
	}
}

// RequireAuth JWT认证中间件
func (a *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.AbortWithError(c, http.StatusUnauthorized, "MISSING_TOKEN", "Authorization header is required", nil)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			response.AbortWithError(c, http.StatusUnauthorized, "INVALID_TOKEN_FORMAT", "Authorization header must start with 'Bearer '", nil)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			response.AbortWithError(c, http.StatusUnauthorized, "MISSING_TOKEN", "JWT token is required", nil)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(a.jwtSecret), nil
		})

		if err != nil {
			response.AbortWithError(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid JWT token", nil)
			return
		}

		if !token.Valid {
			response.AbortWithError(c, http.StatusUnauthorized, "INVALID_TOKEN", "JWT token is not valid", nil)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if exp, ok := claims["exp"].(float64); ok {
				if time.Now().Unix() > int64(exp) {
					response.AbortWithError(c, http.StatusUnauthorized, "TOKEN_EXPIRED", "JWT token has expired", nil)
					return
				}
			}

			if userID, ok := claims["user_id"].(string); ok {
				c.Set("user_id", userID)
			}
			if username, ok := claims["username"].(string); ok {
				c.Set("username", username)
			}
			if role, ok := claims["role"].(string); ok {
				c.Set("role", role)
			}
		}

		c.Next()
	}
}

// OptionalAuth 可选认证中间件
func (a *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString != "" {
				token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, jwt.ErrSignatureInvalid
					}
					return []byte(a.jwtSecret), nil
				})

				if err == nil && token.Valid {
					if claims, ok := token.Claims.(jwt.MapClaims); ok {
						if userID, ok := claims["user_id"].(string); ok {
							c.Set("user_id", userID)
						}
						if username, ok := claims["username"].(string); ok {
							c.Set("username", username)
						}
						if role, ok := claims["role"].(string); ok {
							c.Set("role", role)
						}
					}
				}
			}
		}

		c.Next()
	}
}
