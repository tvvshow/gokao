package handlers

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/tvvshow/gokao/pkg/response"
	"github.com/tvvshow/gokao/services/user-service/internal/models"
	"github.com/tvvshow/gokao/services/user-service/internal/services"
)

// RoleHandler 角色处理器
type RoleHandler struct {
	roleService *services.RoleService
}

// NewRoleHandler 创建角色处理器实例
func NewRoleHandler(roleService *services.RoleService) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
	}
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"omitempty,max=200"`
	IsSystem    bool   `json:"is_system"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=50"`
	Description string `json:"description" binding:"omitempty,max=200"`
	IsSystem    *bool  `json:"is_system"`
}

// CreateRole 创建角色
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid_request", "Invalid request parameters", err.Error())
		return
	}

	role := &models.Role{
		Name:        req.Name,
		Description: req.Description,
		IsSystem:    req.IsSystem,
	}

	if err := h.roleService.CreateRole(role); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "already exists") {
			response.Conflict(c, "role_name_exists", "Role name already exists", nil)
			return
		}
		response.InternalError(c, "role_create_failed", "Failed to create role", nil)
		return
	}

	response.Created(c, role)
}

// GetRole 获取角色信息
func (h *RoleHandler) GetRole(c *gin.Context) {
	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid_role_id", "Invalid role ID format", nil)
		return
	}

	role, err := h.roleService.GetRoleByID(uint(roleID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "role_not_found", "Role not found")
			return
		}
		response.InternalError(c, "role_fetch_failed", "Failed to get role", nil)
		return
	}

	response.OK(c, role)
}

// UpdateRole 更新角色信息
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid_role_id", "Invalid role ID format", nil)
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid_request", "Invalid request parameters", err.Error())
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.IsSystem != nil {
		updates["is_system"] = *req.IsSystem
	}

	if len(updates) == 0 {
		response.BadRequest(c, "no_fields", "No valid fields to update", nil)
		return
	}

	if err := h.roleService.UpdateRole(uint(roleID), updates); err != nil {
		switch {
		case strings.Contains(err.Error(), "not found"):
			response.NotFound(c, "role_not_found", "Role not found")
		case strings.Contains(err.Error(), "duplicate"), strings.Contains(err.Error(), "already exists"):
			response.Conflict(c, "role_name_exists", "Role name already exists", nil)
		default:
			response.InternalError(c, "role_update_failed", "Failed to update role", nil)
		}
		return
	}

	response.OKWithMessage(c, nil, "Role updated successfully")
}

// DeleteRole 删除角色
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid_role_id", "Invalid role ID format", nil)
		return
	}

	if err := h.roleService.DeleteRole(uint(roleID)); err != nil {
		switch {
		case strings.Contains(err.Error(), "not found"):
			response.NotFound(c, "role_not_found", "Role not found")
		case strings.Contains(err.Error(), "in use"), strings.Contains(err.Error(), "assigned"):
			response.Conflict(c, "role_in_use", "Role is currently in use and cannot be deleted", nil)
		default:
			response.InternalError(c, "role_delete_failed", "Failed to delete role", nil)
		}
		return
	}

	response.OKWithMessage(c, nil, "Role deleted successfully")
}

// ListRolesQuery 角色列表查询参数
type ListRolesQuery struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"page_size,default=20" binding:"min=1,max=100"`
	Search   string `form:"search"`
	IsSystem *bool  `form:"is_system"`
}

// ListRoles 获取角色列表
func (h *RoleHandler) ListRoles(c *gin.Context) {
	var query ListRolesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "invalid_query", "Invalid query parameters", err.Error())
		return
	}

	filters := make(map[string]interface{})
	if query.Search != "" {
		filters["search"] = query.Search
	}
	if query.IsSystem != nil {
		filters["is_system"] = *query.IsSystem
	}

	roles, total, err := h.roleService.ListRoles(query.Page, query.PageSize, filters)
	if err != nil {
		response.InternalError(c, "role_list_failed", "Failed to get role list", nil)
		return
	}

	totalPages := (int(total) + query.PageSize - 1) / query.PageSize
	response.OK(c, gin.H{
		"roles": roles,
		"pagination": gin.H{
			"page":        query.Page,
			"page_size":   query.PageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// AssignPermissionRequest 分配权限请求
type AssignPermissionRequest struct {
	PermissionID uint `json:"permission_id" binding:"required"`
}

// AssignPermission 为角色分配权限
func (h *RoleHandler) AssignPermission(c *gin.Context) {
	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid_role_id", "Invalid role ID format", nil)
		return
	}

	var req AssignPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid_request", "Invalid request parameters", err.Error())
		return
	}

	if err := h.roleService.AssignPermission(uint(roleID), req.PermissionID); err != nil {
		switch {
		case strings.Contains(err.Error(), "not found"):
			response.NotFound(c, "not_found", err.Error())
		case strings.Contains(err.Error(), "already assigned"):
			response.Conflict(c, "permission_already_assigned", err.Error(), nil)
		default:
			response.InternalError(c, "permission_assign_failed", "Failed to assign permission", nil)
		}
		return
	}

	response.OKWithMessage(c, nil, "Permission assigned successfully")
}

// RevokePermission 撤销角色权限
func (h *RoleHandler) RevokePermission(c *gin.Context) {
	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid_role_id", "Invalid role ID format", nil)
		return
	}

	permissionIDStr := c.Param("permission_id")
	permissionID, err := strconv.ParseUint(permissionIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid_permission_id", "Invalid permission ID format", nil)
		return
	}

	if err := h.roleService.RevokePermission(uint(roleID), uint(permissionID)); err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "not_found", err.Error())
			return
		}
		response.InternalError(c, "permission_revoke_failed", "Failed to revoke permission", nil)
		return
	}

	response.OKWithMessage(c, nil, "Permission revoked successfully")
}

// GetRolePermissions 获取角色权限
func (h *RoleHandler) GetRolePermissions(c *gin.Context) {
	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid_role_id", "Invalid role ID format", nil)
		return
	}

	permissions, err := h.roleService.GetRolePermissions(uint(roleID))
	if err != nil {
		response.InternalError(c, "role_permissions_failed", "Failed to get role permissions", nil)
		return
	}

	response.OK(c, gin.H{"permissions": permissions})
}

// CreatePermissionRequest 创建权限请求
type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Resource    string `json:"resource" binding:"required,min=2,max=50"`
	Action      string `json:"action" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"omitempty,max=200"`
}

// CreatePermission 创建权限
func (h *RoleHandler) CreatePermission(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid_request", "Invalid request parameters", err.Error())
		return
	}

	permission := &models.Permission{
		Name:        req.Name,
		Resource:    req.Resource,
		Action:      req.Action,
		Description: req.Description,
	}

	if err := h.roleService.CreatePermission(permission); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "already exists") {
			response.Conflict(c, "permission_exists", "Permission already exists", nil)
			return
		}
		response.InternalError(c, "permission_create_failed", "Failed to create permission", nil)
		return
	}

	response.Created(c, permission)
}

// ListPermissionsQuery 权限列表查询参数
type ListPermissionsQuery struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"page_size,default=20" binding:"min=1,max=100"`
	Search   string `form:"search"`
	Resource string `form:"resource"`
	Action   string `form:"action"`
}

// ListPermissions 获取权限列表
func (h *RoleHandler) ListPermissions(c *gin.Context) {
	var query ListPermissionsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "invalid_query", "Invalid query parameters", err.Error())
		return
	}

	filters := make(map[string]interface{})
	if query.Search != "" {
		filters["search"] = query.Search
	}
	if query.Resource != "" {
		filters["resource"] = query.Resource
	}
	if query.Action != "" {
		filters["action"] = query.Action
	}

	permissions, total, err := h.roleService.ListPermissions(query.Page, query.PageSize, filters)
	if err != nil {
		response.InternalError(c, "permission_list_failed", "Failed to get permission list", nil)
		return
	}

	totalPages := (int(total) + query.PageSize - 1) / query.PageSize
	response.OK(c, gin.H{
		"permissions": permissions,
		"pagination": gin.H{
			"page":        query.Page,
			"page_size":   query.PageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// AssignPermissionsRequest 批量分配权限请求
type AssignPermissionsRequest struct {
	PermissionIDs []uint `json:"permission_ids" binding:"required"`
}

// AssignPermissions 批量分配角色权限
func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid_role_id", "Invalid role ID format", nil)
		return
	}

	var req AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid_request", err.Error(), nil)
		return
	}

	for _, permissionID := range req.PermissionIDs {
		if err := h.roleService.AssignPermission(uint(roleID), permissionID); err != nil {
			response.InternalError(c, "permission_assign_failed", "Failed to assign permission", nil)
			return
		}
	}

	response.OKWithMessage(c, nil, "Permissions assigned successfully")
}
