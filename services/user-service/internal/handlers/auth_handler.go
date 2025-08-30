package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"user-service/internal/models"
	"user-service/internal/services"
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// 创建用户模型
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

	// 解析生日
	if req.Birthday != "" {
		if birthday, err := time.Parse("2006-01-02", req.Birthday); err == nil {
			user.Birthday = &birthday
		}
	}

	// 创建用户
	if err := h.userService.CreateUser(user); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create user",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user_id": user.ID,
	})
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// 获取客户端IP
	req.IP = c.ClientIP()

	// 执行登录
	response, err := h.authService.Login(&req)
	if err != nil {
		if strings.Contains(err.Error(), "too many login attempts") {
			c.JSON(http.StatusLocked, gin.H{
				"error": err.Error(),
			})
			return
		}
		if strings.Contains(err.Error(), "invalid username or password") || strings.Contains(err.Error(), "inactive") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Login failed",
		})
		return
	}

	c.JSON(http.StatusOK, response)
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// 刷新令牌
	response, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		if strings.Contains(err.Error(), "invalid or expired") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to refresh token",
		})
		return
	}

	c.JSON(http.StatusOK, response)
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
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req LogoutRequest
	c.ShouldBindJSON(&req)

	// 执行登出
	if err := h.authService.Logout(userID.(uuid.UUID), req.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to logout",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
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
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// 获取用户信息
	user, err := h.userService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user profile",
		})
		return
	}

	// 构建用户信息
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

	c.JSON(http.StatusOK, userInfo)
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
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// 修改密码
	if err := h.userService.ChangePassword(userID.(uuid.UUID), req.OldPassword, req.NewPassword); err != nil {
		if strings.Contains(err.Error(), "incorrect") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to change password",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
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
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// 获取用户权限
	permissions, err := h.authService.GetUserPermissions(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user permissions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"permissions": permissions,
	})
}