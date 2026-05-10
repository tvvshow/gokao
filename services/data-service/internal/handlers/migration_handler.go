package handlers

import (
	"github.com/tvvshow/gokao/services/data-service/internal/database"
	"github.com/tvvshow/gokao/services/data-service/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
// @Summary 应用数据库迁移
// @Description 应用所有待处理的数据库迁移
// @Tags migrations
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /migrations/apply [post]
func (h *MigrationHandler) ApplyMigrations(c *gin.Context) {
	if err := h.service.ApplyAllMigrations(); err != nil {
		h.logger.Errorf("应用数据库迁移失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "应用数据库迁移失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "数据库迁移应用成功",
	})
}

// GetMigrationStatus 获取迁移状态
// @Summary 获取迁移状态
// @Description 获取所有数据库迁移的状态
// @Tags migrations
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /migrations/status [get]
func (h *MigrationHandler) GetMigrationStatus(c *gin.Context) {
	status, err := h.service.GetMigrationStatus()
	if err != nil {
		h.logger.Errorf("获取迁移状态失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取迁移状态失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"migrations": status,
	})
}

// RollbackMigration 回滚数据库迁移
// @Summary 回滚数据库迁移
// @Description 回滚指定版本的数据库迁移
// @Tags migrations
// @Accept json
// @Produce json
// @Param version path string true "迁移版本"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /migrations/rollback/{version} [post]
func (h *MigrationHandler) RollbackMigration(c *gin.Context) {
	version := c.Param("version")
	if version == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "版本参数不能为空",
		})
		return
	}

	if err := h.service.RollbackMigration(version); err != nil {
		h.logger.Errorf("回滚数据库迁移失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "回滚数据库迁移失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "数据库迁移回滚成功",
	})
}
