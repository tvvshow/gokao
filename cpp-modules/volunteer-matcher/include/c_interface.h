/**
 * @file c_interface.h
 * @brief 高考志愿填报系统 - C接口头文件
 * @author 高考志愿填报系统开发团队
 * @version 1.0.0
 * @date 2025-01-18
 * 
 * 为Go CGO提供的C风格接口，包含：
 * - 基础志愿匹配接口
 * - 混合推荐引擎接口
 * - AI功能接口
 * - 内存管理接口
 */

#ifndef C_INTERFACE_H
#define C_INTERFACE_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>
#include <stdbool.h>

// ================================
// 基础数据结构定义
// ================================

/**
 * @brief 错误码定义
 */
typedef enum {
    C_SUCCESS = 0,
    C_ERROR_INVALID_HANDLE = -1,
    C_ERROR_INITIALIZATION_FAILED = -2,
    C_ERROR_INVALID_PARAMETER = -3,
    C_ERROR_FILE_NOT_FOUND = -4,
    C_ERROR_PARSING_FAILED = -5,
    C_ERROR_MEMORY_ALLOCATION = -6,
    C_ERROR_ALGORITHM_FAILED = -7,
    C_ERROR_UNKNOWN = -999
} ErrorCode;

/**
 * @brief 结果结构体
 */
typedef struct {
    int error_code;
    char* message;
    char* data;
} CResult;

/**
 * @brief 性能统计结构体
 */
typedef struct {
    uint64_t total_requests;
    uint64_t successful_requests;
    double avg_response_time;
    double max_response_time;
    uint64_t memory_usage;
} CPerformanceStats;

/**
 * @brief 学生信息结构体
 */
typedef struct {
    char* student_id;
    char* name;
    int total_score;
    int ranking;
    char* province;
    char* subject_combination;
    
    // 单科成绩
    int chinese_score;
    int math_score;
    int english_score;
    int physics_score;
    int chemistry_score;
    int biology_score;
    int politics_score;
    int history_score;
    int geography_score;
    
    // 偏好设置
    char** preferred_cities;
    int preferred_cities_count;
    char** preferred_majors;
    int preferred_majors_count;
    char** avoided_majors;
    int avoided_majors_count;
    
    double city_weight;
    double major_weight;
    double school_ranking_weight;
    
    // 特殊情况
    int is_minority;
    int has_sports_specialty;
    int has_art_specialty;
} C_Student;

/**
 * @brief 志愿推荐结构体
 */
typedef struct {
    char* university_id;
    char* university_name;
    char* major_id;
    char* major_name;
    
    double admission_probability;
    char* risk_level;
    int score_gap;
    int ranking_gap;
    
    double match_score;
    char* recommendation_reason;
    char** risk_factors;
    int risk_factors_count;
} C_VolunteerRecommendation;

/**
 * @brief 志愿方案结构体
 */
typedef struct {
    char* student_id;
    C_VolunteerRecommendation* recommendations;
    int recommendations_count;
    
    int total_volunteers;
    int rush_count;
    int stable_count;
    int safe_count;
    
    double overall_risk_score;
    char* plan_quality;
    char** optimization_suggestions;
    int optimization_suggestions_count;
    
    long long generated_time;
} C_VolunteerPlan;

/**
 * @brief 特征向量结构体
 */
typedef struct {
    double* features;
    int features_count;
    char** feature_names;
    int feature_names_count;
    double target;
} C_FeatureVector;

/**
 * @brief AI推荐结果结构体
 */
typedef struct {
    char* student_id;
    char* university_id;
    char* major_id;
    double ai_score;
    double confidence;
    char* model_version;
    char* explanation;
    long long generated_time;
} C_AIRecommendation;

// ================================
// 基础志愿匹配接口
// ================================

/**
 * @brief 志愿匹配器句柄
 */
typedef struct VolunteerMatcherHandle VolunteerMatcherHandle;

/**
 * @brief 创建志愿匹配器
 * @return 匹配器句柄，失败返回NULL
 */
VolunteerMatcherHandle* CreateVolunteerMatcher();

/**
 * @brief 销毁志愿匹配器
 * @param handle 匹配器句柄
 */
void DestroyVolunteerMatcher(VolunteerMatcherHandle* handle);

/**
 * @brief 初始化志愿匹配器
 * @param handle 匹配器句柄
 * @param config_path 配置文件路径
 * @return 结果
 */
CResult* InitializeVolunteerMatcher(VolunteerMatcherHandle* handle, const char* config_path);

/**
 * @brief 加载大学数据
 * @param handle 匹配器句柄
 * @param universities_file 大学数据文件路径
 * @return 结果
 */
CResult* LoadUniversities(VolunteerMatcherHandle* handle, const char* universities_file);

/**
 * @brief 加载专业数据
 * @param handle 匹配器句柄
 * @param majors_file 专业数据文件路径
 * @return 结果
 */
CResult* LoadMajors(VolunteerMatcherHandle* handle, const char* majors_file);

/**
 * @brief 加载历史数据
 * @param handle 匹配器句柄
 * @param historical_data_file 历史数据文件路径
 * @return 结果
 */
CResult* LoadHistoricalData(VolunteerMatcherHandle* handle, const char* historical_data_file);

/**
 * @brief 生成志愿填报方案
 * @param handle 匹配器句柄
 * @param student_json 学生信息JSON字符串
 * @param max_volunteers 最大志愿数量
 * @return 结果，包含志愿方案JSON
 */
CResult* GenerateVolunteerPlan(VolunteerMatcherHandle* handle, const char* student_json, int max_volunteers);

/**
 * @brief 批量生成志愿方案
 * @param handle 匹配器句柄
 * @param students_json 学生信息数组JSON字符串
 * @param max_volunteers 最大志愿数量
 * @return 结果，包含志愿方案数组JSON
 */
CResult* BatchGenerateVolunteerPlans(VolunteerMatcherHandle* handle, const char* students_json, int max_volunteers);

/**
 * @brief 优化志愿方案
 * @param handle 匹配器句柄
 * @param plan_json 志愿方案JSON字符串
 * @param optimization_target 优化目标
 * @return 结果，包含优化后的方案JSON
 */
CResult* OptimizeVolunteerPlan(VolunteerMatcherHandle* handle, const char* plan_json, const char* optimization_target);

/**
 * @brief 获取性能统计
 * @param handle 匹配器句柄
 * @param stats 输出性能统计
 * @return 结果
 */
CResult* GetPerformanceStats(VolunteerMatcherHandle* handle, CPerformanceStats* stats);

/**
 * @brief 重置性能统计
 * @param handle 匹配器句柄
 * @return 结果
 */
CResult* ResetPerformanceStats(VolunteerMatcherHandle* handle);

/**
 * @brief 热更新数据
 * @param handle 匹配器句柄
 * @param data_type 数据类型
 * @param file_path 文件路径
 * @return 结果
 */
CResult* HotUpdateData(VolunteerMatcherHandle* handle, const char* data_type, const char* file_path);

/**
 * @brief 获取引擎状态
 * @param handle 匹配器句柄
 * @return 结果，包含状态信息JSON
 */
CResult* GetEngineStatus(VolunteerMatcherHandle* handle);

/**
 * @brief 设置日志级别
 * @param handle 匹配器句柄
 * @param level 日志级别
 * @return 结果
 */
CResult* SetLogLevel(VolunteerMatcherHandle* handle, const char* level);

// ================================
// AI功能接口
// ================================

/**
 * @brief 特征提取器句柄
 */
typedef struct FeatureExtractorHandle FeatureExtractorHandle;

/**
 * @brief 创建特征提取器
 * @return 特征提取器句柄，失败返回NULL
 */
FeatureExtractorHandle* CreateFeatureExtractor();

/**
 * @brief 销毁特征提取器
 * @param extractor 特征提取器句柄
 */
void DestroyFeatureExtractor(FeatureExtractorHandle* extractor);

/**
 * @brief 初始化特征提取器
 * @param extractor 特征提取器句柄
 * @param config_path 配置文件路径
 * @return 结果
 */
CResult* InitializeFeatureExtractor(FeatureExtractorHandle* extractor, const char* config_path);

/**
 * @brief 提取学生特征
 * @param extractor 特征提取器句柄
 * @param student_data 学生数据
 * @param feature_vector 输出特征向量
 * @return 结果
 */
CResult* ExtractStudentFeatures(FeatureExtractorHandle* extractor, const C_Student* student_data, C_FeatureVector* feature_vector);

/**
 * @brief AI推荐引擎句柄
 */
typedef struct AIRecommendationEngineHandle AIRecommendationEngineHandle;

/**
 * @brief 创建AI推荐引擎
 * @return AI引擎句柄，失败返回NULL
 */
AIRecommendationEngineHandle* CreateAIRecommendationEngine();

/**
 * @brief 销毁AI推荐引擎
 * @param engine AI引擎句柄
 */
void DestroyAIRecommendationEngine(AIRecommendationEngineHandle* engine);

/**
 * @brief 初始化AI推荐引擎
 * @param engine AI引擎句柄
 * @param model_path 模型文件路径
 * @param config_path 配置文件路径
 * @return 结果
 */
CResult* InitializeAIRecommendationEngine(AIRecommendationEngineHandle* engine, const char* model_path, const char* config_path);

/**
 * @brief AI推荐生成
 * @param engine AI引擎句柄
 * @param student_data 学生数据
 * @param universities_data 大学数据数组
 * @param university_count 大学数量
 * @param recommendations 输出推荐结果
 * @return 推荐数量
 */
int GenerateAIRecommendations(
    AIRecommendationEngineHandle* engine,
    const C_Student* student_data,
    const void* universities_data,
    int university_count,
    C_AIRecommendation** recommendations);

// ================================
// 混合推荐引擎接口
// ================================

/**
 * @brief 混合推荐引擎句柄
 */
typedef struct HybridRecommendationEngineHandle HybridRecommendationEngineHandle;

/**
 * @brief 创建混合推荐引擎
 * @return 混合引擎句柄，失败返回NULL
 */
HybridRecommendationEngineHandle* CreateHybridRecommendationEngine();

/**
 * @brief 销毁混合推荐引擎
 * @param engine 混合引擎句柄
 */
void DestroyHybridRecommendationEngine(HybridRecommendationEngineHandle* engine);

/**
 * @brief 初始化混合引擎
 * @param engine 混合引擎句柄
 * @param traditional_matcher 传统匹配器句柄
 * @param ai_engine AI引擎句柄
 * @param config_path 配置文件路径
 * @return 结果码，0表示成功
 */
int InitializeHybridEngine(
    HybridRecommendationEngineHandle* engine,
    VolunteerMatcherHandle* traditional_matcher,
    AIRecommendationEngineHandle* ai_engine,
    const char* config_path);

/**
 * @brief 生成混合推荐方案
 * @param engine 混合引擎句柄
 * @param student 学生信息
 * @param max_volunteers 最大志愿数量
 * @return 志愿方案，失败返回NULL
 */
C_VolunteerPlan* GenerateHybridPlan(
    HybridRecommendationEngineHandle* engine,
    C_Student* student,
    int max_volunteers);

/**
 * @brief 设置融合权重
 * @param engine 混合引擎句柄
 * @param traditional_weight 传统算法权重
 * @param ai_weight AI算法权重
 * @return 结果码，0表示成功
 */
int SetFusionWeights(
    HybridRecommendationEngineHandle* engine,
    double traditional_weight,
    double ai_weight);

/**
 * @brief 获取混合推荐解释
 * @param engine 混合引擎句柄
 * @param recommendation 推荐结果
 * @return 解释字符串，失败返回NULL
 */
char* GetHybridExplanation(
    HybridRecommendationEngineHandle* engine,
    C_VolunteerRecommendation* recommendation);

/**
 * @brief 获取混合引擎统计信息
 * @param engine 混合引擎句柄
 * @return 统计信息JSON字符串，失败返回NULL
 */
char* GetHybridStats(HybridRecommendationEngineHandle* engine);

// ================================
// 内存管理接口
// ================================

/**
 * @brief 释放C结果对象
 * @param result C结果对象指针
 */
void FreeCResult(CResult* result);

/**
 * @brief 释放C字符串
 * @param str C字符串指针
 */
void FreeCString(char* str);

/**
 * @brief 释放C学生对象
 * @param student C学生对象指针
 */
void FreeCStudent(C_Student* student);

/**
 * @brief 释放C志愿推荐对象
 * @param recommendation C志愿推荐对象指针
 */
void FreeCVolunteerRecommendation(C_VolunteerRecommendation* recommendation);

/**
 * @brief 释放C志愿方案对象
 * @param plan C志愿方案对象指针
 */
void FreeCVolunteerPlan(C_VolunteerPlan* plan);

/**
 * @brief 释放C特征向量对象
 * @param feature_vector C特征向量对象指针
 */
void FreeCFeatureVector(C_FeatureVector* feature_vector);

/**
 * @brief 释放C AI推荐结果对象
 * @param recommendation C AI推荐结果对象指针
 */
void FreeCAPIRecommendation(C_AIRecommendation* recommendation);

/**
 * @brief 释放C AI推荐结果数组
 * @param recommendations C AI推荐结果数组指针
 * @param count 数组大小
 */
void FreeCAPIRecommendationArray(C_AIRecommendation* recommendations, int count);

// ================================
// 工具函数接口
// ================================

/**
 * @brief 获取库版本信息
 * @return 版本字符串
 */
const char* GetLibraryVersion();

/**
 * @brief 获取构建信息
 * @return 构建信息JSON字符串
 */
const char* GetBuildInfo();

/**
 * @brief 设置全局日志处理器
 * @param log_handler 日志处理函数指针
 */
void SetGlobalLogHandler(void (*log_handler)(const char* level, const char* message));

/**
 * @brief 获取最后一次错误信息
 * @return 错误信息字符串
 */
const char* GetLastError();

/**
 * @brief 清除最后一次错误信息
 */
void ClearLastError();

/**
 * @brief 内存使用统计
 * @return 内存使用字节数
 */
uint64_t GetMemoryUsage();

/**
 * @brief 触发垃圾回收
 */
void TriggerGarbageCollection();

/**
 * @brief 设置内存限制
 * @param limit_bytes 内存限制字节数
 * @return 是否设置成功
 */
bool SetMemoryLimit(uint64_t limit_bytes);

/**
 * @brief 验证数据完整性
 * @param data 数据指针
 * @param size 数据大小
 * @return 验证结果
 */
bool ValidateDataIntegrity(const void* data, size_t size);

#ifdef __cplusplus
}
#endif

#endif // C_INTERFACE_H