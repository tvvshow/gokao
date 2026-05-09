package database

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/oktetopython/gaokao/services/user-service/internal/config"
	"github.com/oktetopython/gaokao/services/user-service/internal/models"

	shareddb "github.com/oktetopython/gaokao/pkg/database"
)

// Initialize 初始化数据库连接：连接 + 池配置走共享层，AutoMigrate 与 seed 由本服务负责。
func Initialize(cfg *config.Config) (*gorm.DB, error) {
	db, err := shareddb.OpenGorm(cfg.DatabaseConfig, shareddb.GormOpenOptions{
		Production: cfg.Environment == "production",
	})
	if err != nil {
		return nil, err
	}

	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	if err := seedDefaultData(db); err != nil {
		return nil, fmt.Errorf("failed to seed default data: %w", err)
	}

	return db, nil
}

// InitializeRedis 初始化 Redis 连接（含 ping 验证）。
func InitializeRedis(cfg *config.Config) (*redis.Client, error) {
	return shareddb.OpenRedis(cfg.RedisConfig, 0)
}

// autoMigrate 自动迁移数据库表
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.UserRole{},
		&models.RolePermission{},
		&models.LoginAttempt{},
		&models.AuditLog{},
		&models.RefreshToken{},
		&models.DeviceFingerprint{},
		&models.DeviceLicense{},
		&models.MembershipOrder{},
		&models.UserSession{},
	)
}

// seedDefaultData 初始化默认数据
func seedDefaultData(db *gorm.DB) error {
	// 创建默认权限
	permissions := []models.Permission{
		// 用户管理权限
		{Name: "user:read", Description: "查看用户信息"},
		{Name: "user:write", Description: "修改用户信息"},
		{Name: "user:delete", Description: "删除用户"},
		{Name: "user:verify", Description: "验证用户身份"},

		// 角色权限管理
		{Name: "role:read", Description: "查看角色信息"},
		{Name: "role:write", Description: "修改角色信息"},
		{Name: "role:delete", Description: "删除角色"},
		{Name: "permission:manage", Description: "管理权限"},

		// 会员管理权限
		{Name: "membership:read", Description: "查看会员信息"},
		{Name: "membership:write", Description: "修改会员信息"},
		{Name: "membership:upgrade", Description: "会员升级"},
		{Name: "membership:order", Description: "创建会员订单"},
		{Name: "order:read", Description: "查看订单信息"},
		{Name: "order:process", Description: "处理订单"},
		{Name: "order:refund", Description: "订单退款"},

		// 设备管理权限
		{Name: "device:read", Description: "查看设备信息"},
		{Name: "device:manage", Description: "管理设备绑定"},
		{Name: "device:trust", Description: "设置信任设备"},
		{Name: "device:revoke", Description: "撤销设备授权"},

		// 会话管理权限
		{Name: "session:read", Description: "查看会话信息"},
		{Name: "session:manage", Description: "管理用户会话"},
		{Name: "session:revoke", Description: "撤销会话"},

		// 审计日志权限
		{Name: "audit:read", Description: "查看审计日志"},
		{Name: "audit:export", Description: "导出审计日志"},

		// 系统管理权限
		{Name: "system:monitor", Description: "系统监控"},
		{Name: "system:stats", Description: "查看系统统计"},
		{Name: "admin:all", Description: "管理员全部权限"},
	}

	for _, perm := range permissions {
		var existingPerm models.Permission
		if err := db.Where("name = ?", perm.Name).First(&existingPerm).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&perm).Error; err != nil {
					return fmt.Errorf("failed to create permission %s: %w", perm.Name, err)
				}
			}
		}
	}

	// 创建默认角色
	roles := []models.Role{
		{Name: "admin", Description: "系统管理员", IsSystem: true},
		{Name: "user", Description: "普通用户", IsSystem: true},
		{Name: "basic", Description: "基础会员", IsSystem: true},
		{Name: "premium", Description: "高级会员", IsSystem: true},
		{Name: "enterprise", Description: "企业会员", IsSystem: true},
		{Name: "moderator", Description: "内容审核员", IsSystem: false},
	}

	for _, role := range roles {
		var existingRole models.Role
		if err := db.Where("name = ?", role.Name).First(&existingRole).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&role).Error; err != nil {
					return fmt.Errorf("failed to create role %s: %w", role.Name, err)
				}
			}
		}
	}

	// 为管理员角色分配所有权限
	var adminRole models.Role
	if err := db.Where("name = ?", "admin").First(&adminRole).Error; err == nil {
		var allPermissions []models.Permission
		db.Find(&allPermissions)

		for _, perm := range allPermissions {
			var existingRolePerm models.RolePermission
			if err := db.Where("role_id = ? AND permission_id = ?", adminRole.ID, perm.ID).First(&existingRolePerm).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					rolePerm := models.RolePermission{
						RoleID:       adminRole.ID,
						PermissionID: perm.ID,
					}
					db.Create(&rolePerm)
				}
			}
		}
	}

	// 为不同角色分配权限
	rolePermissionMap := map[string][]string{
		"user": {
			"user:read",
			"device:read",
			"session:read",
		},
		"basic": {
			"user:read", "user:write",
			"device:read", "device:manage",
			"session:read", "session:manage",
			"membership:read", "membership:order",
			"order:read",
		},
		"premium": {
			"user:read", "user:write",
			"device:read", "device:manage", "device:trust",
			"session:read", "session:manage",
			"membership:read", "membership:write", "membership:order", "membership:upgrade",
			"order:read",
			"audit:read",
		},
		"enterprise": {
			"user:read", "user:write", "user:verify",
			"device:read", "device:manage", "device:trust", "device:revoke",
			"session:read", "session:manage", "session:revoke",
			"membership:read", "membership:write", "membership:order", "membership:upgrade",
			"order:read", "order:process",
			"audit:read", "audit:export",
			"system:stats",
		},
		"moderator": {
			"user:read", "user:verify",
			"audit:read",
			"system:monitor",
		},
	}

	for roleName, permNames := range rolePermissionMap {
		var role models.Role
		if err := db.Where("name = ?", roleName).First(&role).Error; err == nil {
			for _, permName := range permNames {
				var perm models.Permission
				if err := db.Where("name = ?", permName).First(&perm).Error; err == nil {
					var existingRolePerm models.RolePermission
					if err := db.Where("role_id = ? AND permission_id = ?", role.ID, perm.ID).First(&existingRolePerm).Error; err != nil {
						if err == gorm.ErrRecordNotFound {
							rolePerm := models.RolePermission{
								RoleID:       role.ID,
								PermissionID: perm.ID,
							}
							db.Create(&rolePerm)
						}
					}
				}
			}
		}
	}

	return nil
}
