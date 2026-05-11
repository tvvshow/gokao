package handlers

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/tvvshow/gokao/pkg/response"
	"github.com/tvvshow/gokao/services/data-service/internal/database"
	"github.com/tvvshow/gokao/services/data-service/internal/services"
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
func (h *DataHandler) ProcessData(c *gin.Context) {
	var req ProcessDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid_request", "请求参数验证失败", err.Error())
		return
	}

	data, err := json.Marshal(req.Data)
	if err != nil {
		response.BadRequest(c, "data_format_invalid", "数据格式错误", err.Error())
		return
	}

	switch req.Type {
	case "universities":
		if err := h.processingService.ProcessUniversityData(data); err != nil {
			h.logger.Errorf("处理高校数据失败: %v", err)
			response.InternalError(c, "process_universities_failed", "处理高校数据失败", err.Error())
			return
		}
	case "majors":
		if err := h.processingService.ProcessMajorData(data); err != nil {
			h.logger.Errorf("处理专业数据失败: %v", err)
			response.InternalError(c, "process_majors_failed", "处理专业数据失败", err.Error())
			return
		}
	case "admissions":
		if err := h.processingService.ProcessAdmissionData(data); err != nil {
			h.logger.Errorf("处理录取数据失败: %v", err)
			response.InternalError(c, "process_admissions_failed", "处理录取数据失败", err.Error())
			return
		}
	default:
		response.BadRequest(c, "unsupported_data_type", "不支持的数据类型", "只支持 universities, majors, admissions 类型")
		return
	}

	response.OKWithMessage(c, nil, "数据处理成功")
}

// ImportData 导入数据
func (h *DataHandler) ImportData(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file_required", "获取上传文件失败", err.Error())
		return
	}

	dataType := c.PostForm("type")
	if dataType == "" {
		response.BadRequest(c, "data_type_required", "数据类型不能为空", "请指定数据类型: universities, majors, admissions")
		return
	}

	if err := h.importService.ValidateFile(file); err != nil {
		response.BadRequest(c, "file_validation_failed", "文件验证失败", err.Error())
		return
	}

	fileContent, err := file.Open()
	if err != nil {
		response.InternalError(c, "file_open_failed", "打开文件失败", err.Error())
		return
	}
	defer fileContent.Close()

	if err := h.importService.ImportFromFile(fileContent, dataType); err != nil {
		h.logger.Errorf("导入数据失败: %v", err)
		response.InternalError(c, "import_failed", "导入数据失败", err.Error())
		return
	}

	response.OKWithMessage(c, nil, "数据导入成功")
}
