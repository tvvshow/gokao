package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/tvvshow/gokao/pkg/response"
	"github.com/tvvshow/gokao/services/data-service/internal/database"
	"github.com/tvvshow/gokao/services/data-service/internal/services"
)

// MigrationHandler 数据库迁移处理器
type MigrationHandler struct {
	db      *database.DB
	service *services.MigrationService
	logger  *logrus.Logger
}

// NewMigrationHandler 创建新的迁移处理器
func NewMigrationHandler(db *database.DB, service *services.MigrationService, logger *logrus.Logger) *MigrationHandler {
	return &MigrationHandler{
		db:      db,
		service: service,
		logger:  logger,
	}
}

// ApplyMigrations 应用数据库迁移
func (h *MigrationHandler) ApplyMigrations(c *gin.Context) {
	if err := h.service.ApplyAllMigrations(); err != nil {
		h.logger.Errorf("应用数据库迁移失败: %v", err)
		response.InternalError(c, "migration_apply_failed", "应用数据库迁移失败", err.Error())
		return
	}

	response.OKWithMessage(c, nil, "数据库迁移应用成功")
}

// GetMigrationStatus 获取迁移状态
func (h *MigrationHandler) GetMigrationStatus(c *gin.Context) {
	status, err := h.service.GetMigrationStatus()
	if err != nil {
		h.logger.Errorf("获取迁移状态失败: %v", err)
		response.InternalError(c, "migration_status_failed", "获取迁移状态失败", err.Error())
		return
	}

	response.OK(c, gin.H{"migrations": status})
}

// RollbackMigration 回滚数据库迁移
func (h *MigrationHandler) RollbackMigration(c *gin.Context) {
	version := c.Param("version")
	if version == "" {
		response.BadRequest(c, "version_required", "版本参数不能为空", nil)
		return
	}

	if err := h.service.RollbackMigration(version); err != nil {
		h.logger.Errorf("回滚数据库迁移失败: %v", err)
		response.InternalError(c, "migration_rollback_failed", "回滚数据库迁移失败", err.Error())
		return
	}

	response.OKWithMessage(c, nil, "数据库迁移回滚成功")
}
