package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"user-service/internal/services"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *services.UserService
	roleService *services.RoleService
}

// NewUserHandler 创建用户处理器实例
func NewUserHandler(userService *services.UserService, roleService *services.RoleService) *UserHandler {
	return &UserHandler{
		userService: userService,
		roleService: roleService,
	}
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Nickname string `json:"nickname" binding:"omitempty,max=50"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"omitempty,max=20"`
	Province string `json:"province" binding:"omitempty,max=50"`
	City     string `json:"city" binding:"omitempty,max=50"`
	Gender   string `json:"gender" binding:"omitempty,oneof=male female other"`
	Birthday string `json:"birthday" binding:"omitempty,datetime=2006-01-02"`
	Avatar   string `json:"avatar" binding:"omitempty,url"`
	Status   string `json:"status" binding:"omitempty,oneof=active inactive suspended"`
}

// GetUser 获取用户信息
// @Summary 获取用户信息
// @Description 根据用户ID获取用户详细信息
// @Tags 用户管理
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} models.User "用户信息"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "用户不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	// 解析用户ID
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	// 获取用户信息
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user",
		})
		return
	}

	// 清除敏感信息
	user.Password = ""

	c.JSON(http.StatusOK, user)
}

// UpdateUser 更新用户信息
// @Summary 更新用户信息
// @Description 更新指定用户的信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Param request body UpdateUserRequest true "更新信息"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "用户不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	// 解析用户ID
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// 构建更新数据
	updates := make(map[string]interface{})
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Province != "" {
		updates["province"] = req.Province
	}
	if req.City != "" {
		updates["city"] = req.City
	}
	if req.Gender != "" {
		updates["gender"] = req.Gender
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.Birthday != "" {
		if birthday, err := time.Parse("2006-01-02", req.Birthday); err == nil {
			updates["birthday"] = birthday
		}
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No valid fields to update",
		})
		return
	}

	// 更新用户
	if err := h.userService.UpdateUser(userID, updates); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
	})
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 软删除指定用户
// @Tags 用户管理
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "用户不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// 解析用户ID
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	// 检查是否尝试删除自己
	currentUserID, exists := c.Get("user_id")
	if exists && currentUserID.(uuid.UUID) == userID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot delete your own account",
		})
		return
	}

	// 删除用户
	if err := h.userService.DeleteUser(userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// ListUsersQuery 用户列表查询参数
type ListUsersQuery struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"page_size,default=20" binding:"min=1,max=100"`
	Search   string `form:"search"`
	Status   string `form:"status" binding:"omitempty,oneof=active inactive suspended"`
	Province string `form:"province"`
	City     string `form:"city"`
}

// ListUsers 获取用户列表
// @Summary 获取用户列表
// @Description 分页获取用户列表，支持搜索和过滤
// @Tags 用户管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param search query string false "搜索关键词"
// @Param status query string false "用户状态" Enums(active, inactive, suspended)
// @Param province query string false "省份"
// @Param city query string false "城市"
// @Success 200 {object} map[string]interface{} "用户列表"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	var query ListUsersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	// 构建过滤条件
	filters := make(map[string]interface{})
	if query.Search != "" {
		filters["search"] = query.Search
	}
	if query.Status != "" {
		filters["status"] = query.Status
	}
	if query.Province != "" {
		filters["province"] = query.Province
	}
	if query.City != "" {
		filters["city"] = query.City
	}

	// 获取用户列表
	users, total, err := h.userService.ListUsers(query.Page, query.PageSize, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user list",
		})
		return
	}

	// 清除敏感信息
	for i := range users {
		users[i].Password = ""
	}

	// 计算分页信息
	totalPages := (int(total) + query.PageSize - 1) / query.PageSize

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"page":        query.Page,
			"page_size":   query.PageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// AssignRoleRequest 分配角色请求
type AssignRoleRequest struct {
	RoleID uint `json:"role_id" binding:"required"`
}

// AssignRole 为用户分配角色
// @Summary 为用户分配角色
// @Description 为指定用户分配角色
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Param request body AssignRoleRequest true "角色信息"
// @Success 200 {object} map[string]interface{} "分配成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "用户或角色不存在"
// @Failure 409 {object} map[string]interface{} "角色已分配"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/users/{id}/roles [post]
func (h *UserHandler) AssignRole(c *gin.Context) {
	// 解析用户ID
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// 分配角色
	if err := h.userService.AssignRole(userID, req.RoleID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		if strings.Contains(err.Error(), "already assigned") {
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to assign role",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Role assigned successfully",
	})
}

// RevokeRole 撤销用户角色
// @Summary 撤销用户角色
// @Description 撤销指定用户的角色
// @Tags 用户管理
// @Produce json
// @Param id path string true "用户ID"
// @Param role_id path int true "角色ID"
// @Success 200 {object} map[string]interface{} "撤销成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "用户角色不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/users/{id}/roles/{role_id} [delete]
func (h *UserHandler) RevokeRole(c *gin.Context) {
	// 解析用户ID
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	// 解析角色ID
	roleIDStr := c.Param("role_id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid role ID format",
		})
		return
	}

	// 撤销角色
	if err := h.userService.RevokeRole(userID, uint(roleID)); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to revoke role",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Role revoked successfully",
	})
}

// GetProfile 获取当前用户资料
// @Summary 获取当前用户资料
// @Description 获取当前登录用户的详细资料
// @Tags 用户管理
// @Accept json
// @Produce json
// @Success 200 {object} models.User
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	// 从JWT中获取用户ID
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
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile 更新当前用户资料
// @Summary 更新当前用户资料
// @Description 更新当前登录用户的资料信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body UpdateUserRequest true "用户信息"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// 解析请求体
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 构建更新数据
	updates := make(map[string]interface{})
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Province != "" {
		updates["province"] = req.Province
	}
	if req.City != "" {
		updates["city"] = req.City
	}
	if req.Gender != "" {
		updates["gender"] = req.Gender
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.Birthday != "" {
		if birthday, err := time.Parse("2006-01-02", req.Birthday); err == nil {
			updates["birthday"] = birthday
		}
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No valid fields to update",
		})
		return
	}

	// 更新用户
	if err := h.userService.UpdateUser(userID.(uuid.UUID), updates); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update user",
		})
		return
	}

	// 返回更新后的用户信息
	updatedUser, err := h.userService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get updated user",
		})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=100"`
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前登录用户的密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/users/change-password [post]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// 解析请求体
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 验证旧密码并更新新密码
	if err := h.userService.ChangePassword(userID.(uuid.UUID), req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// GetUserRoles 获取用户角色
// @Summary 获取用户角色
// @Description 获取指定用户的所有角色
// @Tags 用户管理
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} map[string]interface{} "用户角色列表"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/users/{id}/roles [get]
func (h *UserHandler) GetUserRoles(c *gin.Context) {
	// 解析用户ID
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	// 获取用户角色
	roles, err := h.roleService.GetUserRoles(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user roles",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"roles": roles,
	})
}
