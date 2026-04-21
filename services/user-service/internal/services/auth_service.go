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
	db     *gorm.DB
	redis  *redis.Client
	config *config.Config
}

// NewAuthService 创建认证服务实例
func NewAuthService(db *gorm.DB, redis *redis.Client, config *config.Config) *AuthService {
	return &AuthService{
		db:     db,
		redis:  redis,
		config: config,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	IP       string `json:"ip"`
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
	var user models.User
	if err := s.db.Preload("Roles").Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 记录失败的登录尝试
			s.recordFailedLogin(req.Username, req.IP, "user not found")
			return nil, errors.New("invalid username or password")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 检查用户状态
	if user.Status != "active" {
		s.recordFailedLogin(req.Username, req.IP, "inactive user")
		return nil, errors.New("user account is inactive")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.recordFailedLogin(req.Username, req.IP, "invalid password")
		return nil, errors.New("invalid username or password")
	}

	// 生成访问令牌
	accessToken, err := s.generateAccessToken(user.ID, user.Roles)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 生成刷新令牌
	refreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// 存储刷新令牌
	if err := s.storeRefreshToken(user.ID, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// 清除失败的登录尝试记录
	s.clearFailedLoginAttempts(req.Username, req.IP)

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

	// 构建响应
	response := &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.config.JWTExpiration),
		User:         userInfo,
	}

	// 记录登录日志
	s.logAuthEvent(user.ID, "login", req.IP, "Login successful", "success")

	return response, nil
}

// RefreshToken 刷新访问令牌
func (s *AuthService) RefreshToken(refreshToken string) (*LoginResponse, error) {
	// 验证刷新令牌
	userID, err := s.validateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired refresh token: %w", err)
	}

	// 获取用户
	var user models.User
	if err := s.db.Preload("Roles").Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, errors.New("user account is inactive")
	}

	// 生成新的访问令牌
	accessToken, err := s.generateAccessToken(user.ID, user.Roles)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 生成新的刷新令牌
	newRefreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// 存储新的刷新令牌
	if err := s.storeRefreshToken(user.ID, newRefreshToken); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// 删除旧的刷新令牌
	s.deleteRefreshToken(refreshToken)

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

	// 构建响应
	response := &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Add(s.config.JWTExpiration),
		User:         userInfo,
	}

	return response, nil
}

// Logout 用户登出
func (s *AuthService) Logout(userID uuid.UUID, refreshToken string) error {
	// 删除刷新令牌
	if refreshToken != "" {
		s.deleteRefreshToken(refreshToken)
	}

	// 记录登出日志
	s.logAuthEvent(userID, "logout", "", "User logged out", "success")

	return nil
}

// GetUserPermissions 获取用户权限
func (s *AuthService) GetUserPermissions(userID uuid.UUID) ([]string, error) {
	// 获取用户角色
	var user models.User
	if err := s.db.Preload("Roles").Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 获取角色权限
	var permissions []string
	permissionMap := make(map[string]bool)

	for _, role := range user.Roles {
		// 获取角色权限关联
		var rolePermissions []models.RolePermission
		if err := s.db.Where("role_id = ?", role.ID).Find(&rolePermissions).Error; err != nil {
			return nil, fmt.Errorf("failed to get role permissions: %w", err)
		}

		// 获取权限详情
		for _, rp := range rolePermissions {
			var permission models.Permission
			if err := s.db.Where("id = ?", rp.PermissionID).First(&permission).Error; err != nil {
				continue
			}
			if !permissionMap[permission.Name] {
				permissions = append(permissions, permission.Name)
				permissionMap[permission.Name] = true
			}
		}
	}

	return permissions, nil
}

// generateAccessToken 生成访问令牌
func (s *AuthService) generateAccessToken(userID uuid.UUID, roles []models.Role) (string, error) {
	// 构建角色名称列表
	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	// 创建JWT声明
	claims := &models.JWTClaims{
		UserID:   userID.String(),
		Username: "", // 在实际应用中应该从用户对象获取用户名
		Role:     "", // 在实际应用中应该从角色列表中获取主要角色
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.JWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	signedToken, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// generateRefreshToken 生成刷新令牌
func (s *AuthService) generateRefreshToken(userID uuid.UUID) (string, error) {
	// 生成UUID作为刷新令牌
	token := uuid.New().String()

	// 添加前缀
	refreshToken := "rt_" + token

	return refreshToken, nil
}

// storeRefreshToken 存储刷新令牌
func (s *AuthService) storeRefreshToken(userID uuid.UUID, refreshToken string) error {
	// 存储到Redis，设置过期时间
	ctx := context.Background()
	key := fmt.Sprintf("refresh_token:%s", refreshToken)
	value := userID.String()
	expiration := s.config.RefreshExpiration

	if err := s.redis.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("failed to store refresh token in Redis: %w", err)
	}

	return nil
}

// validateRefreshToken 验证刷新令牌
func (s *AuthService) validateRefreshToken(refreshToken string) (uuid.UUID, error) {
	// 从Redis获取用户ID
	ctx := context.Background()
	key := fmt.Sprintf("refresh_token:%s", refreshToken)

	userIDStr, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return uuid.Nil, errors.New("refresh token not found or expired")
		}
		return uuid.Nil, fmt.Errorf("failed to get refresh token from Redis: %w", err)
	}

	// 解析用户ID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse user ID: %w", err)
	}

	return userID, nil
}

// deleteRefreshToken 删除刷新令牌
func (s *AuthService) deleteRefreshToken(refreshToken string) {
	// 从Redis删除
	ctx := context.Background()
	key := fmt.Sprintf("refresh_token:%s", refreshToken)

	s.redis.Del(ctx, key)
}

// checkLoginAttempts 检查登录尝试次数
func (s *AuthService) checkLoginAttempts(username, ip string) error {
	// 检查用户名尝试次数
	ctx := context.Background()
	usernameKey := fmt.Sprintf("login_attempts:username:%s", username)
	usernameAttempts, err := s.redis.Get(ctx, usernameKey).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		logrus.WithError(err).Warn("Failed to get username login attempts")
	}

	if usernameAttempts >= s.config.MaxLoginAttempts {
		return errors.New("too many login attempts for this username")
	}

	// 检查IP尝试次数
	ipKey := fmt.Sprintf("login_attempts:ip:%s", ip)
	ipAttempts, err := s.redis.Get(ctx, ipKey).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		logrus.WithError(err).Warn("Failed to get IP login attempts")
	}

	if ipAttempts >= s.config.MaxLoginAttempts {
		return errors.New("too many login attempts from this IP")
	}

	return nil
}

// recordFailedLogin 记录失败的登录尝试
func (s *AuthService) recordFailedLogin(username, ip, reason string) {
	ctx := context.Background()
	expiration := s.config.LockoutDuration

	// 记录用户名尝试次数
	usernameKey := fmt.Sprintf("login_attempts:username:%s", username)
	s.redis.Incr(ctx, usernameKey)
	s.redis.Expire(ctx, usernameKey, expiration)

	// 记录IP尝试次数
	ipKey := fmt.Sprintf("login_attempts:ip:%s", ip)
	s.redis.Incr(ctx, ipKey)
	s.redis.Expire(ctx, ipKey, expiration)

	// 记录认证事件
	s.logAuthEvent(uuid.Nil, "login_failed", ip, fmt.Sprintf("Login failed for user %s: %s", username, reason), "failed")
}

// clearFailedLoginAttempts 清除失败的登录尝试记录
func (s *AuthService) clearFailedLoginAttempts(username, ip string) {
	ctx := context.Background()

	// 清除用户名尝试记录
	usernameKey := fmt.Sprintf("login_attempts:username:%s", username)
	s.redis.Del(ctx, usernameKey)

	// 清除IP尝试记录
	ipKey := fmt.Sprintf("login_attempts:ip:%s", ip)
	s.redis.Del(ctx, ipKey)
}

// logAuthEvent 记录认证事件
func (s *AuthService) logAuthEvent(userID uuid.UUID, action, ip, details, status string) {
	if !s.config.EnableAudit {
		return
	}

	auditLog := models.AuditLog{
		UserID:  &userID,
		Action:  action,
		IP:      ip,
		Details: details,
		Status:  status,
	}

	if err := s.db.Create(&auditLog).Error; err != nil {
		logrus.WithError(err).Error("Failed to create auth log")
	}
}