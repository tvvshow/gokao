package services

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/tvvshow/gokao/services/user-service/internal/config"
	"github.com/tvvshow/gokao/services/user-service/internal/models"
)

// UserService 用户服务
type UserService struct {
	db     *gorm.DB
	redis  *redis.Client
	config *config.Config
}

// NewUserService 创建用户服务实例
func NewUserService(db *gorm.DB, redis *redis.Client, config *config.Config) *UserService {
	return &UserService{
		db:     db,
		redis:  redis,
		config: config,
	}
}

// CreateUser 创建用户
func (s *UserService) CreateUser(user *models.User) error {
	// 检查用户名是否已存在
	var existingUser models.User
	if err := s.db.Where("username = ? OR email = ?", user.Username, user.Email).First(&existingUser).Error; err == nil {
		if existingUser.Username == user.Username {
			return errors.New("username already exists")
		}
		if existingUser.Email == user.Email {
			return errors.New("email already exists")
		}
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), s.config.BcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = string(hashedPassword)

	// 设置默认状态
	if user.Status == "" {
		user.Status = "active"
	}

	// 创建用户
	if err := s.db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// 分配默认角色
	var defaultRole models.Role
	if err := s.db.Where("name = ?", "user").First(&defaultRole).Error; err == nil {
		userRole := models.UserRole{
			UserID: user.ID,
			RoleID: defaultRole.ID,
		}
		s.db.Create(&userRole)
	}

	// 记录审计日志
	s.logAudit(nil, "user.create", "user", user.ID.String(), fmt.Sprintf("User %s created", user.Username), "success", "")

	return nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Roles").Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Roles").Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Roles").Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(id uuid.UUID, updates map[string]interface{}) error {
	// 检查用户是否存在
	var user models.User
	if err := s.db.Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 如果更新密码，需要加密
	if password, ok := updates["password"]; ok {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password.(string)), s.config.BcryptCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		updates["password"] = string(hashedPassword)
	}

	// 更新用户
	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// 记录审计日志
	s.logAudit(&id, "user.update", "user", id.String(), fmt.Sprintf("User %s updated", user.Username), "success", "")

	return nil
}

// DeleteUser 删除用户（软删除）
func (s *UserService) DeleteUser(id uuid.UUID) error {
	// 检查用户是否存在
	var user models.User
	if err := s.db.Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 软删除用户
	if err := s.db.Delete(&user).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// 记录审计日志
	s.logAudit(&id, "user.delete", "user", id.String(), fmt.Sprintf("User %s deleted", user.Username), "success", "")

	return nil
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(page, pageSize int, filters map[string]interface{}) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// 构建查询
	query := s.db.Model(&models.User{}).Preload("Roles")

	// 应用过滤条件
	for key, value := range filters {
		switch key {
		case "status":
			query = query.Where("status = ?", value)
		case "search":
			searchTerm := fmt.Sprintf("%%%s%%", value)
			query = query.Where("username ILIKE ? OR email ILIKE ? OR nickname ILIKE ?", searchTerm, searchTerm, searchTerm)
		case "province":
			query = query.Where("province = ?", value)
		case "city":
			query = query.Where("city = ?", value)
		}
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(userID uuid.UUID, oldPassword, newPassword string) error {
	// 获取用户
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("old password is incorrect")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.config.BcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 更新密码
	if err := s.db.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// 记录审计日志
	s.logAudit(&userID, "user.change_password", "user", userID.String(), "Password changed", "success", "")

	return nil
}

// AssignRole 分配角色给用户
func (s *UserService) AssignRole(userID uuid.UUID, roleID uint) error {
	// 检查用户是否存在
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 检查角色是否存在
	var role models.Role
	if err := s.db.Where("id = ?", roleID).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role not found")
		}
		return fmt.Errorf("failed to get role: %w", err)
	}

	// 检查是否已经分配了该角色
	var existingUserRole models.UserRole
	if err := s.db.Where("user_id = ? AND role_id = ?", userID, roleID).First(&existingUserRole).Error; err == nil {
		return errors.New("role already assigned to user")
	}

	// 分配角色
	userRole := models.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	if err := s.db.Create(&userRole).Error; err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	// 记录审计日志
	s.logAudit(&userID, "user.assign_role", "user", userID.String(), fmt.Sprintf("Role %s assigned to user %s", role.Name, user.Username), "success", "")

	return nil
}

// RevokeRole 撤销用户角色
func (s *UserService) RevokeRole(userID uuid.UUID, roleID uint) error {
	// 检查用户角色是否存在
	var userRole models.UserRole
	if err := s.db.Where("user_id = ? AND role_id = ?", userID, roleID).First(&userRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user role not found")
		}
		return fmt.Errorf("failed to get user role: %w", err)
	}

	// 撤销角色
	if err := s.db.Delete(&userRole).Error; err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}

	// 记录审计日志
	s.logAudit(&userID, "user.revoke_role", "user", userID.String(), fmt.Sprintf("Role revoked from user"), "success", "")

	return nil
}

// logAudit 记录审计日志
func (s *UserService) logAudit(userID *uuid.UUID, action, resource, resourceID, details, status, ip string) {
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