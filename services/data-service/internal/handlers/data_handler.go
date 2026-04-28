package handlers

import (
	"encoding/json"
	"github.com/oktetopython/gaokao/services/data-service/internal/database"
	"github.com/oktetopython/gaokao/services/data-service/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// DataHandler 数据处理处理器
type DataHandler struct {
	db                *database.DB
	processingService *services.DataProcessingService
	importService     *services.DataImportService
	logger            *logrus.Logger
}

// NewDataHandler 创建新的数据处理处理器
func NewDataHandler(db *database.DB, logger *logrus.Logger) *DataHandler {
	processingService := services.NewDataProcessingService(db, logger)
	importService := services.NewDataImportService(db, logger)

	return &DataHandler{
		db:                db,
		processingService: processingService,
		importService:     importService,
		logger:            logger,
	}
}

// ProcessDataRequest 数据处理请求结构
type ProcessDataRequest struct {
	Type string      `json:"type" binding:"required,oneof=universities majors admissions"`
	Data interface{} `json:"data" binding:"required"`
}

// ProcessData 处理数据
// @Summary 处理数据
// @Description 处理并存储标准化的JSON数据
// @Tags data
// @Accept json
// @Produce json
// @Param request body ProcessDataRequest true "数据处理请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /data/process [post]
func (h *DataHandler) ProcessData(c *gin.Context) {
	var req ProcessDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数验证失败",
			"message": err.Error(),
		})
		return
	}

	// 将数据转换为字节切片
	data, err := json.Marshal(req.Data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "数据格式错误",
			"message": err.Error(),
		})
		return
	}

	// 根据类型处理数据
	switch req.Type {
	case "universities":
		if err := h.processingService.ProcessUniversityData(data); err != nil {
			h.logger.Errorf("处理高校数据失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "处理高校数据失败",
				"message": err.Error(),
			})
			return
		}
	case "majors":
		if err := h.processingService.ProcessMajorData(data); err != nil {
			h.logger.Errorf("处理专业数据失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "处理专业数据失败",
				"message": err.Error(),
			})
			return
		}
	case "admissions":
		if err := h.processingService.ProcessAdmissionData(data); err != nil {
			h.logger.Errorf("处理录取数据失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "处理录取数据失败",
				"message": err.Error(),
			})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "不支持的数据类型",
			"message": "只支持 universities, majors, admissions 类型",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "数据处理成功",
	})
}

// ImportData 导入数据
// @Summary 导入数据
// @Description 通过上传文件导入数据
// @Tags data
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "JSON文件"
// @Param type formData string true "数据类型: universities, majors, admissions"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /data/import [post]
func (h *DataHandler) ImportData(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "获取上传文件失败",
			"message": err.Error(),
		})
		return
	}

	// 获取数据类型
	dataType := c.PostForm("type")
	if dataType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "数据类型不能为空",
			"message": "请指定数据类型: universities, majors, admissions",
		})
		return
	}

	// 验证文件
	if err := h.importService.ValidateFile(file); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "文件验证失败",
			"message": err.Error(),
		})
		return
	}

	// 打开文件
	fileContent, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "打开文件失败",
			"message": err.Error(),
		})
		return
	}
	defer fileContent.Close()

	// 导入数据
	if err := h.importService.ImportFromFile(fileContent, dataType); err != nil {
		h.logger.Errorf("导入数据失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "导入数据失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "数据导入成功",
	})
}
