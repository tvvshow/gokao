package handlers

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tvvshow/gokao/pkg/response"
	"github.com/tvvshow/gokao/services/user-service/internal/models"
	"github.com/tvvshow/gokao/services/user-service/internal/services"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *services.AuthService
	userService *services.UserService
}

// NewAuthHandler 创建认证处理器实例
func NewAuthHandler(authService *services.AuthService, userService *services.UserService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Nickname string `json:"nickname" binding:"max=50"`
	Phone    string `json:"phone" binding:"max=20"`
	Province string `json:"province" binding:"max=50"`
	City     string `json:"city" binding:"max=50"`
	Gender   string `json:"gender" binding:"oneof=male female other"`
	Birthday string `json:"birthday" binding:"omitempty,datetime=2006-01-02"`
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户账户
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册信息"
// @Success 201 {object} map[string]interface{} "注册成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 409 {object} map[string]interface{} "用户名或邮箱已存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid_request", "Invalid request parameters", err.Error())
		return
	}

	if err := h.validatePassword(req.Password); err != nil {
		response.BadRequest(c, "password_complexity_failed", "Password does not meet complexity requirements", err.Error())
		return
	}

	user := &models.User{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		Nickname: req.Nickname,
		Phone:    req.Phone,
		Province: req.Province,
		City:     req.City,
		Gender:   req.Gender,
		Status:   "active",
	}

	if req.Birthday != "" {
		if birthday, err := time.Parse("2006-01-02", req.Birthday); err == nil {
			user.Birthday = &birthday
		}
	}

	if err := h.userService.CreateUser(user); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			response.Conflict(c, "user_exists", err.Error(), nil)
			return
		}
		response.InternalError(c, "user_creation_failed", "Failed to create user", nil)
		return
	}

	response.CreatedWithMessage(c, gin.H{"user_id": user.ID}, "User registered successfully")
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body services.LoginRequest true "登录信息"
// @Success 200 {object} services.LoginResponse "登录成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "用户名或密码错误"
// @Failure 423 {object} map[string]interface{} "账户被锁定"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid_request", "Invalid request parameters", err.Error())
		return
	}

	req.IP = c.ClientIP()

	loginResp, err := h.authService.Login(&req)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "too many login attempts"):
			response.Locked(c, "account_locked", err.Error())
		case strings.Contains(err.Error(), "invalid username or password"), strings.Contains(err.Error(), "inactive"):
			response.Unauthorized(c, "invalid_credentials", err.Error())
		default:
			response.InternalError(c, "login_failed", "Login failed", nil)
		}
		return
	}

	// frontend api/user.ts 已通过 isWrappedResponse 自动兼容标准 {success,data} 格式；
	// 同时保留 access_token / refresh_token / expires_at 字段供尚未迁移的旧客户端取用。
	response.OKWithMessage(c, gin.H{
		"token":         loginResp.AccessToken,
		"refreshToken":  loginResp.RefreshToken,
		"expiresAt":     loginResp.ExpiresAt,
		"user":          loginResp.User,
		"access_token":  loginResp.AccessToken,
		"refresh_token": loginResp.RefreshToken,
		"expires_at":    loginResp.ExpiresAt,
	}, "登录成功")
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshToken 刷新访问令牌
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "刷新令牌"
// @Success 200 {object} services.LoginResponse "刷新成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "刷新令牌无效或过期"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid_request", "Invalid request parameters", err.Error())
		return
	}

	refreshResp, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		if strings.Contains(err.Error(), "invalid or expired") {
			response.Unauthorized(c, "refresh_token_invalid", err.Error())
			return
		}
		response.InternalError(c, "refresh_failed", "Failed to refresh token", nil)
		return
	}

	// 同 Login：保留新旧字段以兼容尚未迁移的客户端。
	response.OKWithMessage(c, gin.H{
		"token":         refreshResp.AccessToken,
		"refreshToken":  refreshResp.RefreshToken,
		"expiresAt":     refreshResp.ExpiresAt,
		"user":          refreshResp.User,
		"access_token":  refreshResp.AccessToken,
		"refresh_token": refreshResp.RefreshToken,
		"expires_at":    refreshResp.ExpiresAt,
	}, "刷新成功")
}

// LogoutRequest 登出请求
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Logout 用户登出
// @Summary 用户登出
// @Description 撤销用户的访问令牌和刷新令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LogoutRequest false "登出信息"
// @Success 200 {object} map[string]interface{} "登出成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthenticated", "User not authenticated")
		return
	}

	var req LogoutRequest
	_ = c.ShouldBindJSON(&req)

	if err := h.authService.Logout(userID.(uuid.UUID), req.RefreshToken); err != nil {
		response.InternalError(c, "logout_failed", "Failed to logout", nil)
		return
	}

	response.OKWithMessage(c, nil, "Logged out successfully")
}

// GetProfile 获取用户资料
// @Summary 获取用户资料
// @Description 获取当前登录用户的详细信息
// @Tags 认证
// @Produce json
// @Success 200 {object} services.UserInfo "用户信息"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 404 {object} map[string]interface{} "用户不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthenticated", "User not authenticated")
		return
	}

	user, err := h.userService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "user_not_found", "User not found")
			return
		}
		response.InternalError(c, "profile_fetch_failed", "Failed to get user profile", nil)
		return
	}

	roles := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = role.Name
	}

	userInfo := services.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Roles:    roles,
	}

	response.OK(c, userInfo)
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=100"`
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前用户的密码
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "密码信息"
// @Success 200 {object} map[string]interface{} "修改成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "旧密码错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthenticated", "User not authenticated")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid_request", "Invalid request parameters", err.Error())
		return
	}

	if err := h.validatePassword(req.NewPassword); err != nil {
		response.BadRequest(c, "password_complexity_failed", "New password does not meet complexity requirements", err.Error())
		return
	}

	if err := h.userService.ChangePassword(userID.(uuid.UUID), req.OldPassword, req.NewPassword); err != nil {
		if strings.Contains(err.Error(), "incorrect") {
			response.Unauthorized(c, "old_password_incorrect", err.Error())
			return
		}
		response.InternalError(c, "password_change_failed", "Failed to change password", nil)
		return
	}

	response.OKWithMessage(c, nil, "Password changed successfully")
}

// GetPermissions 获取用户权限
// @Summary 获取用户权限
// @Description 获取当前用户的所有权限列表
// @Tags 认证
// @Produce json
// @Success 200 {object} map[string]interface{} "权限列表"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/auth/permissions [get]
func (h *AuthHandler) GetPermissions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthenticated", "User not authenticated")
		return
	}

	permissions, err := h.authService.GetUserPermissions(userID.(uuid.UUID))
	if err != nil {
		response.InternalError(c, "permissions_fetch_failed", "Failed to get user permissions", nil)
		return
	}

	response.OK(c, gin.H{"permissions": permissions})
}

// validatePassword 验证密码复杂度
func (h *AuthHandler) validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	var requirements []string
	if !hasUpper {
		requirements = append(requirements, "at least one uppercase letter")
	}
	if !hasLower {
		requirements = append(requirements, "at least one lowercase letter")
	}
	if !hasDigit {
		requirements = append(requirements, "at least one digit")
	}
	if !hasSpecial {
		requirements = append(requirements, "at least one special character")
	}

	if len(requirements) > 0 {
		return fmt.Errorf("password must contain %s", strings.Join(requirements, ", "))
	}

	weakPasswords := []string{
		"password", "123456", "12345678", "123456789", "qwerty",
		"abc123", "password1", "admin", "welcome", "letmein",
	}

	for _, weak := range weakPasswords {
		if strings.ToLower(password) == weak {
			return fmt.Errorf("password is too common and easily guessable")
		}
	}

	if hasRepeatingPattern(password, 3) {
		return fmt.Errorf("password contains easily guessable patterns")
	}

	return nil
}

// hasRepeatingPattern 检查密码中是否有重复模式
func hasRepeatingPattern(s string, minLength int) bool {
	if len(s) < minLength*2 {
		return false
	}

	for length := minLength; length <= len(s)/2; length++ {
		for i := 0; i <= len(s)-length*2; i++ {
			sub1 := s[i : i+length]
			sub2 := s[i+length : i+length*2]

			if sub1 == sub2 {
				return true
			}
		}
	}

	return false
}
