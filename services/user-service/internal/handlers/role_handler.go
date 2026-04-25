package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/oktetopython/gaokao/services/user-service/internal/models"
	"github.com/oktetopython/gaokao/services/user-service/internal/services"
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
// @Summary 创建角色
// @Description 创建新的角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param request body CreateRoleRequest true "角色信息"
// @Success 201 {object} models.Role "创建的角色"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 409 {object} map[string]interface{} "角色名已存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/roles [post]
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// 创建角色
	role := &models.Role{
		Name:        req.Name,
		Description: req.Description,
		IsSystem:    req.IsSystem,
	}

	if err := h.roleService.CreateRole(role); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Role name already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create role",
		})
		return
	}

	c.JSON(http.StatusCreated, role)
}

// GetRole 获取角色信息
// @Summary 获取角色信息
// @Description 根据角色ID获取角色详细信息
// @Tags 角色管理
// @Produce json
// @Param id path int true "角色ID"
// @Success 200 {object} models.Role "角色信息"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "角色不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/roles/{id} [get]
func (h *RoleHandler) GetRole(c *gin.Context) {
	// 解析角色ID
	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid role ID format",
		})
		return
	}

	// 获取角色信息
	role, err := h.roleService.GetRoleByID(uint(roleID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Role not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get role",
		})
		return
	}

	c.JSON(http.StatusOK, role)
}

// UpdateRole 更新角色信息
// @Summary 更新角色信息
// @Description 更新指定角色的信息
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Param request body UpdateRoleRequest true "更新信息"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "角色不存在"
// @Failure 409 {object} map[string]interface{} "角色名已存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/roles/{id} [put]
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	// 解析角色ID
	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid role ID format",
		})
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// 构建更新数据
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No valid fields to update",
		})
		return
	}

	// 更新角色
	if err := h.roleService.UpdateRole(uint(roleID), updates); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Role not found",
			})
			return
		}
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Role name already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update role",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Role updated successfully",
	})
}

// DeleteRole 删除角色
// @Summary 删除角色
// @Description 删除指定角色
// @Tags 角色管理
// @Produce json
// @Param id path int true "角色ID"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "角色不存在"
// @Failure 409 {object} map[string]interface{} "角色正在使用中"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/roles/{id} [delete]
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	// 解析角色ID
	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid role ID format",
		})
		return
	}

	// 删除角色
	if err := h.roleService.DeleteRole(uint(roleID)); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Role not found",
			})
			return
		}
		if strings.Contains(err.Error(), "in use") || strings.Contains(err.Error(), "assigned") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Role is currently in use and cannot be deleted",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete role",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Role deleted successfully",
	})
}

// ListRolesQuery 角色列表查询参数
type ListRolesQuery struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"page_size,default=20" binding:"min=1,max=100"`
	Search   string `form:"search"`
	IsSystem *bool  `form:"is_system"`
}

// ListRoles 获取角色列表
// @Summary 获取角色列表
// @Description 分页获取角色列表，支持搜索和过滤
// @Tags 角色管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param search query string false "搜索关键词"
// @Param is_system query bool false "是否系统角色"
// @Success 200 {object} map[string]interface{} "角色列表"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/roles [get]
func (h *RoleHandler) ListRoles(c *gin.Context) {
	var query ListRolesQuery
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
	if query.IsSystem != nil {
		filters["is_system"] = *query.IsSystem
	}

	// 获取角色列表
	roles, total, err := h.roleService.ListRoles(query.Page, query.PageSize, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get role list",
		})
		return
	}

	// 计算分页信息
	totalPages := (int(total) + query.PageSize - 1) / query.PageSize

	c.JSON(http.StatusOK, gin.H{
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
// @Summary 为角色分配权限
// @Description 为指定角色分配权限
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Param request body AssignPermissionRequest true "权限信息"
// @Success 200 {object} map[string]interface{} "分配成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "角色或权限不存在"
// @Failure 409 {object} map[string]interface{} "权限已分配"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/roles/{id}/permissions [post]
func (h *RoleHandler) AssignPermission(c *gin.Context) {
	// 解析角色ID
	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid role ID format",
		})
		return
	}

	var req AssignPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// 分配权限
	if err := h.roleService.AssignPermission(uint(roleID), req.PermissionID); err != nil {
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
			"error": "Failed to assign permission",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Permission assigned successfully",
	})
}

// RevokePermission 撤销角色权限
// @Summary 撤销角色权限
// @Description 撤销指定角色的权限
// @Tags 角色管理
// @Produce json
// @Param id path int true "角色ID"
// @Param permission_id path int true "权限ID"
// @Success 200 {object} map[string]interface{} "撤销成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "角色权限不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/roles/{id}/permissions/{permission_id} [delete]
func (h *RoleHandler) RevokePermission(c *gin.Context) {
	// 解析角色ID
	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid role ID format",
		})
		return
	}

	// 解析权限ID
	permissionIDStr := c.Param("permission_id")
	permissionID, err := strconv.ParseUint(permissionIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid permission ID format",
		})
		return
	}

	// 撤销权限
	if err := h.roleService.RevokePermission(uint(roleID), uint(permissionID)); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to revoke permission",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Permission revoked successfully",
	})
}

// GetRolePermissions 获取角色权限
// @Summary 获取角色权限
// @Description 获取指定角色的所有权限
// @Tags 角色管理
// @Produce json
// @Param id path int true "角色ID"
// @Success 200 {object} map[string]interface{} "角色权限列表"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/roles/{id}/permissions [get]
func (h *RoleHandler) GetRolePermissions(c *gin.Context) {
	// 解析角色ID
	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid role ID format",
		})
		return
	}

	// 获取角色权限
	permissions, err := h.roleService.GetRolePermissions(uint(roleID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get role permissions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"permissions": permissions,
	})
}

// CreatePermissionRequest 创建权限请求
type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Resource    string `json:"resource" binding:"required,min=2,max=50"`
	Action      string `json:"action" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"omitempty,max=200"`
}

// CreatePermission 创建权限
// @Summary 创建权限
// @Description 创建新的权限
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param request body CreatePermissionRequest true "权限信息"
// @Success 201 {object} models.Permission "创建的权限"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 409 {object} map[string]interface{} "权限已存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/permissions [post]
func (h *RoleHandler) CreatePermission(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// 创建权限
	permission := &models.Permission{
		Name:        req.Name,
		Resource:    req.Resource,
		Action:      req.Action,
		Description: req.Description,
	}

	if err := h.roleService.CreatePermission(permission); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Permission already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create permission",
		})
		return
	}

	c.JSON(http.StatusCreated, permission)
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
// @Summary 获取权限列表
// @Description 分页获取权限列表，支持搜索和过滤
// @Tags 权限管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param search query string false "搜索关键词"
// @Param resource query string false "资源"
// @Param action query string false "操作"
// @Success 200 {object} map[string]interface{} "权限列表"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/permissions [get]
func (h *RoleHandler) ListPermissions(c *gin.Context) {
	var query ListPermissionsQuery
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
	if query.Resource != "" {
		filters["resource"] = query.Resource
	}
	if query.Action != "" {
		filters["action"] = query.Action
	}

	// 获取权限列表
	permissions, total, err := h.roleService.ListPermissions(query.Page, query.PageSize, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get permission list",
		})
		return
	}

	// 计算分页信息
	totalPages := (int(total) + query.PageSize - 1) / query.PageSize

	c.JSON(http.StatusOK, gin.H{
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
// @Summary 批量分配角色权限
// @Description 为指定角色批量分配权限
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Param request body AssignPermissionsRequest true "权限ID列表"
// @Success 200 {object} map[string]interface{} "分配成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "角色不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Security BearerAuth
// @Router /api/v1/roles/{id}/permissions [post]
func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	// 解析角色ID
	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid role ID format",
		})
		return
	}

	// 解析请求体
	var req AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 批量分配权限
	for _, permissionID := range req.PermissionIDs {
		if err := h.roleService.AssignPermission(uint(roleID), permissionID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to assign permission",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Permissions assigned successfully",
	})
}
