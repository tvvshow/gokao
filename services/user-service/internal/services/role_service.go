package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/tvvshow/gokao/services/user-service/internal/config"
	"github.com/tvvshow/gokao/services/user-service/internal/models"
)

// RoleService 角色服务
type RoleService struct {
	db     *gorm.DB
	redis  *redis.Client
	config *config.Config
}

// NewRoleService 创建角色服务实例
func NewRoleService(db *gorm.DB, redis *redis.Client, config *config.Config) *RoleService {
	return &RoleService{
		db:     db,
		redis:  redis,
		config: config,
	}
}

// CreateRole 创建角色
func (s *RoleService) CreateRole(role *models.Role) error {
	// 检查角色名是否已存在
	var existingRole models.Role
	if err := s.db.Where("name = ?", role.Name).First(&existingRole).Error; err == nil {
		return errors.New("role name already exists")
	}

	// 创建角色
	if err := s.db.Create(role).Error; err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	// 清除角色缓存
	s.clearRoleCache()

	// 记录审计日志
	s.logAudit(nil, "role.create", "role", fmt.Sprintf("%d", role.ID), fmt.Sprintf("Role %s created", role.Name), "success", "")

	return nil
}

// GetRoleByID 根据ID获取角色
func (s *RoleService) GetRoleByID(id uint) (*models.Role, error) {
	var role models.Role
	if err := s.db.Preload("Permissions").Where("id = ?", id).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	return &role, nil
}

// GetRoleByName 根据名称获取角色
func (s *RoleService) GetRoleByName(name string) (*models.Role, error) {
	var role models.Role
	if err := s.db.Preload("Permissions").Where("name = ?", name).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	return &role, nil
}

// UpdateRole 更新角色
func (s *RoleService) UpdateRole(id uint, updates map[string]interface{}) error {
	// 检查角色是否存在
	var role models.Role
	if err := s.db.Where("id = ?", id).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role not found")
		}
		return fmt.Errorf("failed to get role: %w", err)
	}

	// 如果更新名称，检查是否重复
	if name, ok := updates["name"]; ok {
		var existingRole models.Role
		if err := s.db.Where("name = ? AND id != ?", name, id).First(&existingRole).Error; err == nil {
			return errors.New("role name already exists")
		}
	}

	// 更新角色
	if err := s.db.Model(&role).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	// 清除角色缓存
	s.clearRoleCache()

	// 记录审计日志
	s.logAudit(nil, "role.update", "role", fmt.Sprintf("%d", id), fmt.Sprintf("Role %s updated", role.Name), "success", "")

	return nil
}

// DeleteRole 删除角色
func (s *RoleService) DeleteRole(id uint) error {
	// 检查角色是否存在
	var role models.Role
	if err := s.db.Where("id = ?", id).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role not found")
		}
		return fmt.Errorf("failed to get role: %w", err)
	}

	// 检查是否有用户使用该角色
	var userRoleCount int64
	if err := s.db.Model(&models.UserRole{}).Where("role_id = ?", id).Count(&userRoleCount).Error; err != nil {
		return fmt.Errorf("failed to check role usage: %w", err)
	}

	if userRoleCount > 0 {
		return errors.New("cannot delete role that is assigned to users")
	}

	// 删除角色权限关联
	if err := s.db.Where("role_id = ?", id).Delete(&models.RolePermission{}).Error; err != nil {
		return fmt.Errorf("failed to delete role permissions: %w", err)
	}

	// 删除角色
	if err := s.db.Delete(&role).Error; err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	// 清除角色缓存
	s.clearRoleCache()

	// 记录审计日志
	s.logAudit(nil, "role.delete", "role", fmt.Sprintf("%d", id), fmt.Sprintf("Role %s deleted", role.Name), "success", "")

	return nil
}

// ListRoles 获取角色列表
func (s *RoleService) ListRoles(page, pageSize int, filters map[string]interface{}) ([]models.Role, int64, error) {
	var roles []models.Role
	var total int64

	// 构建查询
	query := s.db.Model(&models.Role{}).Preload("Permissions")

	// 应用过滤条件
	for key, value := range filters {
		switch key {
		case "search":
			searchTerm := fmt.Sprintf("%%%s%%", value)
			query = query.Where("name ILIKE ? OR description ILIKE ?", searchTerm, searchTerm)
		case "status":
			query = query.Where("status = ?", value)
		}
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count roles: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&roles).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list roles: %w", err)
	}

	return roles, total, nil
}

// AssignPermission 为角色分配权限
func (s *RoleService) AssignPermission(roleID, permissionID uint) error {
	// 检查角色是否存在
	var role models.Role
	if err := s.db.Where("id = ?", roleID).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role not found")
		}
		return fmt.Errorf("failed to get role: %w", err)
	}

	// 检查权限是否存在
	var permission models.Permission
	if err := s.db.Where("id = ?", permissionID).First(&permission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("permission not found")
		}
		return fmt.Errorf("failed to get permission: %w", err)
	}

	// 检查是否已经分配了该权限
	var existingRolePermission models.RolePermission
	if err := s.db.Where("role_id = ? AND permission_id = ?", roleID, permissionID).First(&existingRolePermission).Error; err == nil {
		return errors.New("permission already assigned to role")
	}

	// 分配权限
	rolePermission := models.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
	}
	if err := s.db.Create(&rolePermission).Error; err != nil {
		return fmt.Errorf("failed to assign permission: %w", err)
	}

	// 清除相关缓存
	s.clearRoleCache()
	s.clearUserPermissionCache(roleID)

	// 记录审计日志
	s.logAudit(nil, "role.assign_permission", "role", fmt.Sprintf("%d", roleID), fmt.Sprintf("Permission %s assigned to role %s", permission.Name, role.Name), "success", "")

	return nil
}

// RevokePermission 撤销角色权限
func (s *RoleService) RevokePermission(roleID, permissionID uint) error {
	// 检查角色权限是否存在
	var rolePermission models.RolePermission
	if err := s.db.Where("role_id = ? AND permission_id = ?", roleID, permissionID).First(&rolePermission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role permission not found")
		}
		return fmt.Errorf("failed to get role permission: %w", err)
	}

	// 撤销权限
	if err := s.db.Delete(&rolePermission).Error; err != nil {
		return fmt.Errorf("failed to revoke permission: %w", err)
	}

	// 清除相关缓存
	s.clearRoleCache()
	s.clearUserPermissionCache(roleID)

	// 记录审计日志
	s.logAudit(nil, "role.revoke_permission", "role", fmt.Sprintf("%d", roleID), fmt.Sprintf("Permission revoked from role"), "success", "")

	return nil
}

// CreatePermission 创建权限
func (s *RoleService) CreatePermission(permission *models.Permission) error {
	// 检查权限名是否已存在
	var existingPermission models.Permission
	if err := s.db.Where("name = ?", permission.Name).First(&existingPermission).Error; err == nil {
		return errors.New("permission name already exists")
	}

	// 创建权限
	if err := s.db.Create(permission).Error; err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}

	// 清除权限缓存
	s.clearPermissionCache()

	// 记录审计日志
	s.logAudit(nil, "permission.create", "permission", fmt.Sprintf("%d", permission.ID), fmt.Sprintf("Permission %s created", permission.Name), "success", "")

	return nil
}

// GetPermissionByID 根据ID获取权限
func (s *RoleService) GetPermissionByID(id uint) (*models.Permission, error) {
	var permission models.Permission
	if err := s.db.Where("id = ?", id).First(&permission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("permission not found")
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	return &permission, nil
}

// UpdatePermission 更新权限
func (s *RoleService) UpdatePermission(id uint, updates map[string]interface{}) error {
	// 检查权限是否存在
	var permission models.Permission
	if err := s.db.Where("id = ?", id).First(&permission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("permission not found")
		}
		return fmt.Errorf("failed to get permission: %w", err)
	}

	// 如果更新名称，检查是否重复
	if name, ok := updates["name"]; ok {
		var existingPermission models.Permission
		if err := s.db.Where("name = ? AND id != ?", name, id).First(&existingPermission).Error; err == nil {
			return errors.New("permission name already exists")
		}
	}

	// 更新权限
	if err := s.db.Model(&permission).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update permission: %w", err)
	}

	// 清除权限缓存
	s.clearPermissionCache()

	// 记录审计日志
	s.logAudit(nil, "permission.update", "permission", fmt.Sprintf("%d", id), fmt.Sprintf("Permission %s updated", permission.Name), "success", "")

	return nil
}

// DeletePermission 删除权限
func (s *RoleService) DeletePermission(id uint) error {
	// 检查权限是否存在
	var permission models.Permission
	if err := s.db.Where("id = ?", id).First(&permission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("permission not found")
		}
		return fmt.Errorf("failed to get permission: %w", err)
	}

	// 检查是否有角色使用该权限
	var rolePermissionCount int64
	if err := s.db.Model(&models.RolePermission{}).Where("permission_id = ?", id).Count(&rolePermissionCount).Error; err != nil {
		return fmt.Errorf("failed to check permission usage: %w", err)
	}

	if rolePermissionCount > 0 {
		return errors.New("cannot delete permission that is assigned to roles")
	}

	// 删除权限
	if err := s.db.Delete(&permission).Error; err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	// 清除权限缓存
	s.clearPermissionCache()

	// 记录审计日志
	s.logAudit(nil, "permission.delete", "permission", fmt.Sprintf("%d", id), fmt.Sprintf("Permission %s deleted", permission.Name), "success", "")

	return nil
}

// ListPermissions 获取权限列表
func (s *RoleService) ListPermissions(page, pageSize int, filters map[string]interface{}) ([]models.Permission, int64, error) {
	var permissions []models.Permission
	var total int64

	// 构建查询
	query := s.db.Model(&models.Permission{})

	// 应用过滤条件
	for key, value := range filters {
		switch key {
		case "search":
			searchTerm := fmt.Sprintf("%%%s%%", value)
			query = query.Where("name ILIKE ? OR description ILIKE ?", searchTerm, searchTerm)
		case "resource":
			query = query.Where("resource = ?", value)
		}
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count permissions: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("resource, action").Find(&permissions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list permissions: %w", err)
	}

	return permissions, total, nil
}

// GetRolePermissions 获取角色的所有权限
func (s *RoleService) GetRolePermissions(roleID uint) ([]models.Permission, error) {
	var permissions []models.Permission
	query := `
		SELECT p.* FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = ?
		ORDER BY p.resource, p.action
	`
	if err := s.db.Raw(query, roleID).Scan(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	return permissions, nil
}

// GetUserRoles 获取用户角色 - 优化版本，添加Redis缓存
func (s *RoleService) GetUserRoles(userID uuid.UUID) ([]models.Role, error) {
	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("user_roles:%s", userID.String())
	var roles []models.Role
	
	// 检查Redis缓存
	cachedData, err := s.redis.Get(context.Background(), cacheKey).Result()
	if err == nil {
		// 缓存命中，解析JSON数据
		if err := json.Unmarshal([]byte(cachedData), &roles); err == nil {
			return roles, nil
		}
	}
	
	// 缓存未命中，从数据库查询
	query := `
		SELECT r.* FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = ?
		ORDER BY r.name
	`
	if err := s.db.Raw(query, userID).Scan(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	
	// 将结果缓存到Redis（缓存30分钟）
	if rolesJSON, err := json.Marshal(roles); err == nil {
		s.redis.Set(context.Background(), cacheKey, rolesJSON, 30*time.Minute)
	}
	
	return roles, nil
}

// clearRoleCache 清除角色缓存
func (s *RoleService) clearRoleCache() {
	// 清除所有角色相关的缓存
	pattern := "role:*"
	keys, err := s.redis.Keys(context.Background(), pattern).Result()
	if err != nil {
		logrus.WithError(err).Warn("Failed to get role cache keys")
		return
	}

	if len(keys) > 0 {
		s.redis.Del(context.Background(), keys...)
	}
}

// clearPermissionCache 清除权限缓存
func (s *RoleService) clearPermissionCache() {
	// 清除所有权限相关的缓存
	pattern := "permission:*"
	keys, err := s.redis.Keys(context.Background(), pattern).Result()
	if err != nil {
		logrus.WithError(err).Warn("Failed to get permission cache keys")
		return
	}

	if len(keys) > 0 {
		s.redis.Del(context.Background(), keys...)
	}
}

// GetUserPermissions 获取用户所有权限 - 带缓存优化
func (s *RoleService) GetUserPermissions(userID uuid.UUID) ([]models.Permission, error) {
	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("user_permissions:%s", userID.String())
	var permissions []models.Permission
	
	// 检查Redis缓存
	cachedData, err := s.redis.Get(context.Background(), cacheKey).Result()
	if err == nil {
		// 缓存命中，解析JSON数据
		if err := json.Unmarshal([]byte(cachedData), &permissions); err == nil {
			return permissions, nil
		}
	}
	
	// 缓存未命中，从数据库查询
	query := `
		SELECT DISTINCT p.* FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = ?
		ORDER BY p.resource, p.action
	`
	if err := s.db.Raw(query, userID).Scan(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	
	// 将结果缓存到Redis（缓存30分钟）
	if permissionsJSON, err := json.Marshal(permissions); err == nil {
		s.redis.Set(context.Background(), cacheKey, permissionsJSON, 30*time.Minute)
	}
	
	return permissions, nil
}

// HasPermission 检查用户是否有特定权限 - 带缓存优化
func (s *RoleService) HasPermission(userID uuid.UUID, permissionName string) (bool, error) {
	permissions, err := s.GetUserPermissions(userID)
	if err != nil {
		return false, err
	}
	
	for _, perm := range permissions {
		if perm.Name == permissionName {
			return true, nil
		}
	}
	
	return false, nil
}

// clearUserPermissionCache 清除用户权限缓存
func (s *RoleService) clearUserPermissionCache(roleID uint) {
	// 获取使用该角色的所有用户
	var userIDs []uuid.UUID
	if err := s.db.Model(&models.UserRole{}).Where("role_id = ?", roleID).Pluck("user_id", &userIDs).Error; err != nil {
		logrus.WithError(err).Warn("Failed to get users for role")
		return
	}

	// 清除这些用户的权限和角色缓存
	for _, userID := range userIDs {
		userIDStr := userID.String()
		s.redis.Del(context.Background(), 
			fmt.Sprintf("user_permissions:%s", userIDStr),
			fmt.Sprintf("user_roles:%s", userIDStr),
		)
	}
}

// logAudit 记录审计日志
func (s *RoleService) logAudit(userID *uuid.UUID, action, resource, resourceID, details, status, ip string) {
	if !s.config.EnableAudit {
		return
	}

	auditLog := models.AuditLog{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Details:    details,
		Status:     status,
		IP:         ip,
	}

	if err := s.db.Create(&auditLog).Error; err != nil {
		logrus.WithError(err).Error("Failed to create audit log")
	}
}
