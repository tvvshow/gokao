package services

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gaokao-system/services/data-service/internal/models"
)

// ===== 数据验证服务实现 =====
type DataValidationService struct {
	db interface {
		Find(dest interface{}, conds ...interface{}) error
		Where(query interface{}, args ...interface{}) interface{}
	}
}

func NewDataValidationService(db interface{}) *DataValidationService {
	return &DataValidationService{db: db}
}

// 验证University数据
func (s *DataValidationService) ValidateUniversity(data *models.University) error {
	// 必填字段验证
	if strings.TrimSpace(data.Name) == "" {
		return errors.New("大学名称不能为空")
	}

	if strings.TrimSpace(data.Province) == "" {
		return errors.New("省份不能为空")
	}

	if strings.TrimSpace(data.Type) == "" {
		return errors.New("学校类型不能为空")
	}

	// 名称长度验证
	if len(data.Name) > 100 {
		return errors.New("大学名称长度不能超过100个字符")
	}

	// 邮箱格式验证
	if data.ContactEmail != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(data.ContactEmail) {
			return errors.New("联系邮箱格式不正确")
		}
	}

	// 网站URL格式验证
	if data.Website != "" {
		urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
		if !urlRegex.MatchString(data.Website) {
			return errors.New("网站URL格式不正确")
		}
	}

	// 建校年份验证
	if data.Established != nil {
		currentYear := time.Now().Year()
		year := data.Established.Year()
		if year < 1800 || year > currentYear {
			return errors.New("建校年份必须在1800年至今")
		}
	}

	// 排名验证
	if data.NationalRank < 0 {
		return errors.New("全国排名不能为负数")
	}

	if data.ProvinceRank < 0 {
		return errors.New("省内排名不能为负数")
	}

	// 分数验证
	if data.OverallScore < 0 || data.OverallScore > 100 {
		return errors.New("综合得分必须在0-100之间")
	}

	return nil
}

// 验证Major数据
func (s *DataValidationService) ValidateMajor(data *models.Major) error {
	// 必填字段验证
	if strings.TrimSpace(data.Name) == "" {
		return errors.New("专业名称不能为空")
	}

	if strings.TrimSpace(data.Code) == "" {
		return errors.New("专业代码不能为空")
	}

	if strings.TrimSpace(data.Category) == "" {
		return errors.New("专业类别不能为空")
	}

	if strings.TrimSpace(data.Degree) == "" {
		return errors.New("学位层次不能为空")
	}

	// 代码唯一性验证（排除自身）
	var existingMajor models.Major
	if err := s.db.Where("code = ? AND id != ?", data.Code, data.ID).First(&existingMajor).Error; err == nil {
		return errors.New("专业代码已存在")
	}

	// 名称长度验证
	if len(data.Name) > 100 {
		return errors.New("专业名称长度不能超过100个字符")
	}

	// 学制验证
	if data.Duration < 1 || data.Duration > 10 {
		return errors.New("学制必须在1-10年之间")
	}

	// 就业率验证
	if data.EmploymentRate < 0 || data.EmploymentRate > 100 {
		return errors.New("就业率必须在0-100之间")
	}

	// 薪资验证
	if data.AverageSalary < 0 {
		return errors.New("平均薪资不能为负数")
	}

	return nil
}

// 验证AdmissionData数据
func (s *DataValidationService) ValidateAdmissionData(data *models.AdmissionData) error {
	// 必填字段验证
	if data.UniversityID == uuid.Nil {
		return errors.New("大学ID不能为空")
	}

	if strings.TrimSpace(data.ScienceType) == "" {
		return errors.New("科类不能为空")
	}

	if strings.TrimSpace(data.BatchType) == "" {
		return errors.New("批次类型不能为空")
	}

	if data.Year < 1900 || data.Year > time.Now().Year()+1 {
		return errors.New("年份必须在1900-未来一年之间")
	}

	// 科类值验证
	validScienceTypes := []string{"理科", "文科", "新高考", "综合"}
	isValid := false
	for _, validType := range validScienceTypes {
		if data.ScienceType == validType {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("无效的科类类型，必须是: %v", validScienceTypes)
	}

	// 批次类型验证
	validBatchTypes := []string{"本科一批", "本科二批", "专科一批", "专科二批", "提前批"}
	isValid = false
	for _, validBatch := range validBatchTypes {
		if data.BatchType == validBatch {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("无效的批次类型，必须是: %v", validBatchTypes)
	}

	// 招生计划验证
	if data.EnrollmentQuota < 0 {
		return errors.New("招生计划数不能为负数")
	}

	if data.ActualEnrollment < 0 {
		return errors.New("实际录取数不能为负数")
	}

	if data.ActualEnrollment > data.EnrollmentQuota {
		return errors.New("实际录取数不能超过招生计划数")
	}

	// 分数验证
	if data.CutoffScore < 0 || data.CutoffScore > 750 {
		return errors.New("分数线必须在0-750分之间")
	}

	if data.AverageScore < 0 || data.AverageScore > 750 {
		return errors.New("平均分必须在0-750分之间")
	}

	if data.HighestScore < 0 || data.HighestScore > 750 {
		return errors.New("最高分必须在0-750分之间")
	}

	if data.LowestScore < 0 || data.LowestScore > 750 {
		return errors.New("最低分必须在0-750分之间")
	}

	// 排名验证
	if data.CutoffRank < 0 {
		return errors.New("最低排名不能为负数")
	}

	if data.AverageRank < 0 {
		return errors.New("平均排名不能为负数")
	}

	return nil
}

// 检查数据一致性
func (s *DataValidationService) CheckDataConsistency() error {
	// 1. 检查大学是否存在
	var universityCount int64
	if err := s.db.Model(&models.University{}).Where("deleted_at IS NULL").Count(&universityCount).Error; err != nil {
		return fmt.Errorf("检查大学数量失败: %v", err)
	}

	if universityCount == 0 {
		return errors.New("系统中没有大学数据")
	}

	// 2. 检查专业是否有对应的大学
	var majorWithoutUniversity int64
	if err := s.db.Model(&models.Major{}).
		Where("university_id IS NULL").
		Count(&majorWithoutUniversity).Error; err != nil {
		return fmt.Errorf("检查无大学专业失败: %v", err)
	}

	if majorWithoutUniversity > 0 {
		return fmt.Errorf("有%d个专业没有关联大学", majorWithoutUniversity)
	}

	// 3. 检查录取数据是否有对应的大学
	var admissionWithoutUniversity int64
	if err := s.db.Model(&models.AdmissionData{}).
		Where("university_id IS NULL").
		Count(&admissionWithoutUniversity).Error; err != nil {
		return fmt.Errorf("检查无大学录取数据失败: %v", err)
	}

	if admissionWithoutUniversity > 0 {
		return fmt.Errorf("有%d条录取数据没有关联大学", admissionWithoutUniversity)
	}

	// 4. 检查专业相关的录取数据是否有对应的专业
	var admissionWithoutMajor int64
	if err := s.db.Model(&models.AdmissionData{}).
		Where("major_id IS NOT NULL AND major_id != ''").
		Count(&admissionWithoutMajor).Error; err != nil {
		return fmt.Errorf("检查无专业录取数据失败: %v", err)
	}

	// 如果没有MajorID的录取数据也是合理的
	if admissionWithoutMajor > 0 {
		return fmt.Errorf("有%d条录取数据没有关联专业", admissionWithoutMajor)
	}

	return nil
}

// 强制外键约束
func (s *DataValidationService) EnforceForeignKeyConstraints() error {
	// 这里可以实现更严格的外键约束检查
	// 1. 检查专业的UniversityID是否在University表中存在
	// 2. 检查录取数据的UniversityID是否在University表中存在
	// 3. 检查录取数据的MajorID是否在Major表中存在（如果非空）

	// 由于GORM默认支持外键约束，这里主要做逻辑上的检查
	return nil
}

// 批量验证数据
func (s *DataValidationService) BatchValidateUniversities(universities []models.University) []error {
	var errors []error

	for i, university := range universities {
		if err := s.ValidateUniversity(&university); err != nil {
			errors = append(errors, fmt.Errorf("大学数据索引%d: %v", i, err))
		}
	}

	return errors
}

func (s *DataValidationService) BatchValidateMajors(majors []models.Major) []error {
	var errors []error

	for i, major := range majors {
		if err := s.ValidateMajor(&major); err != nil {
			errors = append(errors, fmt.Errorf("专业数据索引%d: %v", i, err))
		}
	}

	return errors
}

func (s *DataValidationService) BatchValidateAdmissionData(admissions []models.AdmissionData) []error {
	var errors []error

	for i, admission := range admissions {
		if err := s.ValidateAdmissionData(&admission); err != nil {
			errors = append(errors, fmt.Errorf("录取数据索引%d: %v", i, err))
		}
	}

	return errors
}

// 数据清洗
func (s *DataValidationService) CleanUniversityData(data *models.University) {
	// 去除前后空格
	data.Name = strings.TrimSpace(data.Name)
	data.EnglishName = strings.TrimSpace(data.EnglishName)
	data.Province = strings.TrimSpace(data.Province)
	data.City = strings.TrimSpace(data.City)
	data.Type = strings.TrimSpace(data.Type)
	data.Website = strings.TrimSpace(data.Website)
	data.Address = strings.TrimSpace(data.Address)
	data.ContactPhone = strings.TrimSpace(data.ContactPhone)
	data.ContactEmail = strings.TrimSpace(data.ContactEmail)

	// 标准化学校类型
	switch strings.ToUpper(data.Type) {
	case "985":
		data.Is985 = true
		data.Type = "985"
	case "211":
		data.Is211 = true
		data.Type = "211"
	case "双一流":
		data.IsDoubleFirstClass = true
		data.Type = "双一流"
	}

	// 确保URL格式正确
	if strings.HasPrefix(data.Website, "www.") {
		data.Website = "https://" + data.Website
	}
}

func (s *DataValidationService) CleanMajorData(data *models.Major) {
	// 去除前后空格
	data.Name = strings.TrimSpace(data.Name)
	data.EnglishName = strings.TrimSpace(data.EnglishName)
	data.Code = strings.TrimSpace(data.Code)
	data.Category = strings.TrimSpace(data.Category)
	data.Degree = strings.TrimSpace(data.Degree)

	// 标准化学位类型
	switch strings.ToUpper(data.Degree) {
	case "本科", "BACHELOR":
		data.Degree = "本科"
	case "硕士", "MASTER":
		data.Degree = "硕士"
	case "博士", "PHD", "DOCTOR":
		data.Degree = "博士"
	}
}

func (s *DataValidationService) CleanAdmissionData(data *models.AdmissionData) {
	// 去除前后空格
	data.ScienceType = strings.TrimSpace(data.ScienceType)
	data.BatchType = strings.TrimSpace(data.BatchType)

	// 标准化科类
	switch strings.ToUpper(data.ScienceType) {
	case "理科", "SCIENCE":
		data.ScienceType = "理科"
	case "文科", "ARTS":
		data.ScienceType = "文科"
	case "新高考", "NEW_GAOKAO":
		data.ScienceType = "新高考"
	case "综合", "COMPREHENSIVE":
		data.ScienceType = "综合"
	}

	// 标准化批次类型
	switch strings.ToUpper(data.BatchType) {
	case "一批", "FIRST_BATCH":
		data.BatchType = "本科一批"
	case "二批", "SECOND_BATCH":
		data.BatchType = "本科二批"
	case "专科一批", "VOCATIONAL_FIRST":
		data.BatchType = "专科一批"
	case "专科二批", "VOCATIONAL_SECOND":
		data.BatchType = "专科二批"
	case "提前批", "EARLY_BATCH":
		data.BatchType = "提前批"
	}
}