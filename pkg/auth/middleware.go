package auth

import (
    "net/http"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
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
        // 获取Authorization头
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error":   "missing_authorization",
                "message": "Authorization header is required",
                "code":    "MISSING_TOKEN",
            })
            c.Abort()
            return
        }

        // 检查Bearer前缀
        if !strings.HasPrefix(authHeader, "Bearer ") {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error":   "invalid_authorization_format", 
                "message": "Authorization header must start with 'Bearer '",
                "code":    "INVALID_TOKEN_FORMAT",
            })
            c.Abort()
            return
        }

        // 提取token
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error":   "missing_token",
                "message": "JWT token is required",
                "code":    "MISSING_TOKEN",
            })
            c.Abort()
            return
        }

        // 验证token
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, jwt.ErrSignatureInvalid
            }
            return []byte(a.jwtSecret), nil
        })

        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error":   "invalid_token",
                "message": "Invalid JWT token",
                "code":    "INVALID_TOKEN",
            })
            c.Abort()
            return
        }

        if !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error":   "token_invalid",
                "message": "JWT token is not valid", 
                "code":    "INVALID_TOKEN",
            })
            c.Abort()
            return
        }

        // 提取claims
        if claims, ok := token.Claims.(jwt.MapClaims); ok {
            // 检查过期时间
            if exp, ok := claims["exp"].(float64); ok {
                if time.Now().Unix() > int64(exp) {
                    c.JSON(http.StatusUnauthorized, gin.H{
                        "error":   "token_expired",
                        "message": "JWT token has expired",
                        "code":    "TOKEN_EXPIRED",
                    })
                    c.Abort()
                    return
                }
            }

            // 将用户信息存储到上下文
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
