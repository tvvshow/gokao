package handlers

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tvvshow/gokao/pkg/response"
	"github.com/tvvshow/gokao/services/user-service/internal/services"
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
func (h *UserHandler) GetUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.BadRequest(c, "invalid_user_id", "Invalid user ID format", nil)
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "user_not_found", "User not found")
			return
		}
		response.InternalError(c, "user_fetch_failed", "Failed to get user", nil)
		return
	}

	user.Password = ""
	response.OK(c, user)
}

// UpdateUser 更新用户信息
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.BadRequest(c, "invalid_user_id", "Invalid user ID format", nil)
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid_request", "Invalid request parameters", err.Error())
		return
	}

	updates := buildUserUpdates(&req)
	if len(updates) == 0 {
		response.BadRequest(c, "no_fields", "No valid fields to update", nil)
		return
	}

	if err := h.userService.UpdateUser(userID, updates); err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "user_not_found", "User not found")
			return
		}
		response.InternalError(c, "user_update_failed", "Failed to update user", nil)
		return
	}

	response.OKWithMessage(c, nil, "User updated successfully")
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.BadRequest(c, "invalid_user_id", "Invalid user ID format", nil)
		return
	}

	currentUserID, exists := c.Get("user_id")
	if exists && currentUserID.(uuid.UUID) == userID {
		response.BadRequest(c, "self_delete_forbidden", "Cannot delete your own account", nil)
		return
	}

	if err := h.userService.DeleteUser(userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "user_not_found", "User not found")
			return
		}
		response.InternalError(c, "user_delete_failed", "Failed to delete user", nil)
		return
	}

	response.OKWithMessage(c, nil, "User deleted successfully")
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
func (h *UserHandler) ListUsers(c *gin.Context) {
	var query ListUsersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "invalid_query", "Invalid query parameters", err.Error())
		return
	}

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

	users, total, err := h.userService.ListUsers(query.Page, query.PageSize, filters)
	if err != nil {
		response.InternalError(c, "user_list_failed", "Failed to get user list", nil)
		return
	}

	for i := range users {
		users[i].Password = ""
	}

	totalPages := (int(total) + query.PageSize - 1) / query.PageSize
	response.OK(c, gin.H{
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
func (h *UserHandler) AssignRole(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.BadRequest(c, "invalid_user_id", "Invalid user ID format", nil)
		return
	}

	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid_request", "Invalid request parameters", err.Error())
		return
	}

	if err := h.userService.AssignRole(userID, req.RoleID); err != nil {
		switch {
		case strings.Contains(err.Error(), "not found"):
			response.NotFound(c, "not_found", err.Error())
		case strings.Contains(err.Error(), "already assigned"):
			response.Conflict(c, "role_already_assigned", err.Error(), nil)
		default:
			response.InternalError(c, "role_assign_failed", "Failed to assign role", nil)
		}
		return
	}

	response.OKWithMessage(c, nil, "Role assigned successfully")
}

// RevokeRole 撤销用户角色
func (h *UserHandler) RevokeRole(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.BadRequest(c, "invalid_user_id", "Invalid user ID format", nil)
		return
	}

	roleIDStr := c.Param("role_id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid_role_id", "Invalid role ID format", nil)
		return
	}

	if err := h.userService.RevokeRole(userID, uint(roleID)); err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "not_found", err.Error())
			return
		}
		response.InternalError(c, "role_revoke_failed", "Failed to revoke role", nil)
		return
	}

	response.OKWithMessage(c, nil, "Role revoked successfully")
}

// GetProfile 获取当前用户资料
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthenticated", "User not authenticated")
		return
	}

	user, err := h.userService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		response.NotFound(c, "user_not_found", "User not found")
		return
	}

	response.OK(c, user)
}

// UpdateProfile 更新当前用户资料
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthenticated", "User not authenticated")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid_request", err.Error(), nil)
		return
	}

	updates := buildUserUpdates(&req)
	if len(updates) == 0 {
		response.BadRequest(c, "no_fields", "No valid fields to update", nil)
		return
	}

	if err := h.userService.UpdateUser(userID.(uuid.UUID), updates); err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "user_not_found", "User not found")
			return
		}
		response.InternalError(c, "user_update_failed", "Failed to update user", nil)
		return
	}

	updatedUser, err := h.userService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		response.InternalError(c, "user_fetch_failed", "Failed to get updated user", nil)
		return
	}

	response.OK(c, updatedUser)
}

// UserChangePasswordRequest 修改密码请求
type UserChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=100"`
}

// ChangePassword 修改密码
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthenticated", "User not authenticated")
		return
	}

	var req UserChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid_request", err.Error(), nil)
		return
	}

	if err := h.userService.ChangePassword(userID.(uuid.UUID), req.OldPassword, req.NewPassword); err != nil {
		response.BadRequest(c, "password_change_failed", err.Error(), nil)
		return
	}

	response.OKWithMessage(c, nil, "Password changed successfully")
}

// GetUserRoles 获取用户角色
func (h *UserHandler) GetUserRoles(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.BadRequest(c, "invalid_user_id", "Invalid user ID format", nil)
		return
	}

	roles, err := h.roleService.GetUserRoles(userID)
	if err != nil {
		response.InternalError(c, "user_roles_failed", "Failed to get user roles", nil)
		return
	}

	response.OK(c, gin.H{"roles": roles})
}

// GetMembership 获取当前用户会员信息
func (h *UserHandler) GetMembership(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthenticated", "User not authenticated")
		return
	}

	user, err := h.userService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		response.NotFound(c, "user_not_found", "User not found")
		return
	}

	response.OK(c, gin.H{
		"membership_level":  user.MembershipLevel,
		"membership_expiry": user.MembershipExpiry,
		"max_devices":       user.MaxDevices,
		"trial_used":        user.TrialUsed,
		"trial_expiry":      user.TrialExpiry,
	})
}

// buildUserUpdates 把 UpdateUserRequest 翻译成可直接喂给 service.UpdateUser 的 map。
// 抽到一处避免 UpdateUser/UpdateProfile 两处重复。
func buildUserUpdates(req *UpdateUserRequest) map[string]interface{} {
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
	return updates
}
