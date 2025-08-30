package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"user-service/internal/config"
	"user-service/internal/models"
)

// AuthService 认证服务
type AuthService struct {
	db          *gorm.DB
	redis       *redis.Client
	config      *config.Config
	userService *UserService
}

// NewAuthService 创建认证服务实例
func NewAuthService(db *gorm.DB, redis *redis.Client, config *config.Config, userService *UserService) *AuthService {
	return &AuthService{
		db:          db,
		redis:       redis,
		config:      config,
		userService: userService,
	}
}

// JWTClaims JWT声明
type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Roles    []string  `json:"roles"`
	jwt.RegisteredClaims
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	IP       string `json:"-"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         UserInfo  `json:"user"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Nickname string    `json:"nickname"`
	Avatar   string    `json:"avatar"`
	Roles    []string  `json:"roles"`
}

// Login 用户登录
func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
	// 检查登录尝试次数
	if err := s.checkLoginAttempts(req.Username, req.IP); err != nil {
		return nil, err
	}

	// 获取用户
	user, err := s.userService.GetUserByUsername(req.Username)
	if err != nil {
		// 记录失败的登录尝试
		s.recordLoginAttempt(req.Username, req.IP, false, "user not found")
		return nil, errors.New("invalid username or password")
	}

	// 检查用户状态
	if user.Status != "active" {
		s.recordLoginAttempt(req.Username, req.IP, false, "user inactive")
		return nil, errors.New("user account is inactive")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.recordLoginAttempt(req.Username, req.IP, false, "invalid password")
		return nil, errors.New("invalid username or password")
	}

	// 生成令牌
	accessToken, refreshToken, expiresAt, err := s.generateTokens(user)
	if err != nil {
		s.recordLoginAttempt(req.Username, req.IP, false, "token generation failed")
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// 记录成功的登录尝试
	s.recordLoginAttempt(req.Username, req.IP, true, "login successful")

	// 清除失败的登录尝试记录
	s.clearLoginAttempts(req.Username, req.IP)

	// 更新最后登录时间
	s.userService.UpdateUser(user.ID, map[string]interface{}{
		"last_login_at": time.Now(),
		"last_login_ip": req.IP,
	})

	// 构建用户信息
	roles := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = role.Name
	}

	userInfo := UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Roles:    roles,
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User:         userInfo,
	}, nil
}

// RefreshToken 刷新令牌
func (s *AuthService) RefreshToken(refreshToken string) (*LoginResponse, error) {
	// 验证刷新令牌
	var tokenRecord models.RefreshToken
	if err := s.db.Where("token = ? AND expires_at > ? AND is_revoked = false", refreshToken, time.Now()).First(&tokenRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid or expired refresh token")
		}
		return nil, fmt.Errorf("failed to validate refresh token: %w", err)
	}

	// 获取用户
	user, err := s.userService.GetUserByID(tokenRecord.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, errors.New("user account is inactive")
	}

	// 撤销旧的刷新令牌
	s.db.Model(&tokenRecord).Update("is_revoked", true)

	// 生成新的令牌
	accessToken, newRefreshToken, expiresAt, err := s.generateTokens(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// 构建用户信息
	roles := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = role.Name
	}

	userInfo := UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Roles:    roles,
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
		User:         userInfo,
	}, nil
}

// Logout 用户登出
func (s *AuthService) Logout(userID uuid.UUID, refreshToken string) error {
	// 撤销刷新令牌
	if refreshToken != "" {
		s.db.Model(&models.RefreshToken{}).Where("token = ? AND user_id = ?", refreshToken, userID).Update("is_revoked", true)
	}

	// 将访问令牌加入黑名单（通过Redis）
	// 这里可以根据需要实现令牌黑名单机制

	// 记录审计日志
	s.userService.logAudit(&userID, "auth.logout", "user", userID.String(), "User logged out", "success", "")

	return nil
}

// ValidateToken 验证访问令牌
func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// 验证令牌
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		// 检查令牌是否在黑名单中（可选）
		// 这里可以根据需要实现令牌黑名单检查

		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// generateTokens 生成访问令牌和刷新令牌
func (s *AuthService) generateTokens(user *models.User) (string, string, time.Time, error) {
	// 构建角色列表
	roles := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = role.Name
	}

	// 生成访问令牌
	accessTokenExpiry := time.Now().Add(s.config.JWTExpiration)
	accessClaims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessTokenExpiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "gaokao-user-service",
			Subject:   user.ID.String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to sign access token: %w", err)
	}

	// 生成刷新令牌
	refreshTokenString := uuid.New().String()
	refreshTokenExpiry := time.Now().Add(s.config.RefreshExpiration)

	// 保存刷新令牌到数据库
	refreshToken := models.RefreshToken{
		Token:     refreshTokenString,
		UserID:    user.ID,
		ExpiresAt: refreshTokenExpiry,
		IsRevoked: false,
	}

	if err := s.db.Create(&refreshToken).Error; err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return accessTokenString, refreshTokenString, accessTokenExpiry, nil
}

// checkLoginAttempts 检查登录尝试次数
func (s *AuthService) checkLoginAttempts(username, ip string) error {
	if s.config.MaxLoginAttempts <= 0 {
		return nil // 未启用登录尝试限制
	}

	// 检查用户名的登录尝试
	usernameKey := fmt.Sprintf("login_attempts:username:%s", username)
	usernameAttempts, err := s.redis.Get(context.Background(), usernameKey).Int()
	if err != nil && err.Error() != "redis: nil" {
		logrus.WithError(err).Warn("Failed to get login attempts from Redis")
	}

	// 检查IP的登录尝试
	ipKey := fmt.Sprintf("login_attempts:ip:%s", ip)
	ipAttempts, err := s.redis.Get(context.Background(), ipKey).Int()
	if err != nil && err.Error() != "redis: nil" {
		logrus.WithError(err).Warn("Failed to get login attempts from Redis")
	}

	// 如果任一超过限制，则拒绝登录
	if usernameAttempts >= s.config.MaxLoginAttempts || ipAttempts >= s.config.MaxLoginAttempts {
		return errors.New("too many login attempts, please try again later")
	}

	return nil
}

// recordLoginAttempt 记录登录尝试
func (s *AuthService) recordLoginAttempt(username, ip string, success bool, reason string) {
	// 记录到数据库
	loginAttempt := models.LoginAttempt{
		Username:  username,
		IP:        ip,
		Success:   success,
		Reason:    reason,
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(&loginAttempt).Error; err != nil {
		logrus.WithError(err).Error("Failed to record login attempt")
	}

	// 如果登录失败，增加Redis中的计数
	if !success && s.config.MaxLoginAttempts > 0 {
		usernameKey := fmt.Sprintf("login_attempts:username:%s", username)
		ipKey := fmt.Sprintf("login_attempts:ip:%s", ip)
		lockoutDuration := s.config.LockoutDuration

		// 增加用户名计数
		s.redis.Incr(context.Background(), usernameKey)
		s.redis.Expire(context.Background(), usernameKey, lockoutDuration)

		// 增加IP计数
		s.redis.Incr(context.Background(), ipKey)
		s.redis.Expire(context.Background(), ipKey, lockoutDuration)
	}
}

// clearLoginAttempts 清除登录尝试记录
func (s *AuthService) clearLoginAttempts(username, ip string) {
	if s.config.MaxLoginAttempts <= 0 {
		return
	}

	usernameKey := fmt.Sprintf("login_attempts:username:%s", username)
	ipKey := fmt.Sprintf("login_attempts:ip:%s", ip)

	s.redis.Del(context.Background(), usernameKey, ipKey)
}

// GetUserPermissions 获取用户权限
func (s *AuthService) GetUserPermissions(userID uuid.UUID) ([]string, error) {
	// 从Redis缓存中获取权限
	cacheKey := fmt.Sprintf("user_permissions:%s", userID.String())
	cachedPermissions, err := s.redis.SMembers(context.Background(), cacheKey).Result()
	if err == nil && len(cachedPermissions) > 0 {
		return cachedPermissions, nil
	}

	// 从数据库获取权限
	var permissions []models.Permission
	query := `
		SELECT DISTINCT p.* FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = ?
	`
	if err := s.db.Raw(query, userID).Scan(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	// 构建权限列表
	permissionNames := make([]string, len(permissions))
	for i, permission := range permissions {
		permissionNames[i] = permission.Name
	}

	// 缓存权限到Redis（5分钟过期）
	if len(permissionNames) > 0 {
		s.redis.SAdd(context.Background(), cacheKey, permissionNames)
		s.redis.Expire(context.Background(), cacheKey, 5*time.Minute)
	}

	return permissionNames, nil
}

// HasPermission 检查用户是否有指定权限
func (s *AuthService) HasPermission(userID uuid.UUID, permission string) (bool, error) {
	permissions, err := s.GetUserPermissions(userID)
	if err != nil {
		return false, err
	}

	for _, p := range permissions {
		if p == permission {
			return true, nil
		}
	}

	return false, nil
}

// RevokeAllTokens 撤销用户的所有令牌
func (s *AuthService) RevokeAllTokens(userID uuid.UUID) error {
	// 撤销所有刷新令牌
	if err := s.db.Model(&models.RefreshToken{}).Where("user_id = ? AND is_revoked = false", userID).Update("is_revoked", true).Error; err != nil {
		return fmt.Errorf("failed to revoke refresh tokens: %w", err)
	}

	// 清除权限缓存
	cacheKey := fmt.Sprintf("user_permissions:%s", userID.String())
	s.redis.Del(context.Background(), cacheKey)

	// 记录审计日志
	s.userService.logAudit(&userID, "auth.revoke_all_tokens", "user", userID.String(), "All tokens revoked", "success", "")

	return nil
}
