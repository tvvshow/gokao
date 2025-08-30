/**
 * @file volunteer_matcher.h
 * @brief 高考志愿填报系统 - 志愿匹配算法引擎核心头文件
 * @author 高考志愿填报系统开发团队
 * @version 1.0.0
 * @date 2025-01-18
 * 
 * 高性能志愿匹配算法引擎，支持：
 * - 冲稳保策略生成
 * - 个性化推荐算法
 * - 多目标优化
 * - 实时性能监控
 */

#ifndef VOLUNTEER_MATCHER_H
#define VOLUNTEER_MATCHER_H

#include <vector>
#include <string>
#include <unordered_map>
#include <memory>
#include <chrono>
#include <atomic>
#include <mutex>
#include <thread>

namespace volunteer_matcher {

/**
 * @brief 学生信息结构
 */
struct Student {
    std::string student_id;          ///< 学生ID
    std::string name;                ///< 姓名
    int total_score;                 ///< 总分
    int ranking;                     ///< 省排名
    std::string province;            ///< 省份
    std::string subject_combination; ///< 选科组合 (3+1+2 或传统文理)
    
    // 单科成绩
    int chinese_score;               ///< 语文
    int math_score;                  ///< 数学
    int english_score;               ///< 英语
    int physics_score;               ///< 物理
    int chemistry_score;             ///< 化学
    int biology_score;               ///< 生物
    int politics_score;              ///< 政治
    int history_score;               ///< 历史
    int geography_score;             ///< 地理
    
    // 学生偏好
    std::vector<std::string> preferred_cities;    ///< 偏好城市
    std::vector<std::string> preferred_majors;    ///< 偏好专业
    std::vector<std::string> avoided_majors;      ///< 不考虑专业
    double city_weight;                           ///< 城市权重 (0-1)
    double major_weight;                          ///< 专业权重 (0-1)
    double school_ranking_weight;                 ///< 学校排名权重 (0-1)
    
    // 特殊情况
    bool is_minority;                ///< 是否少数民族
    bool has_sports_specialty;       ///< 是否体育特长生
    bool has_art_specialty;          ///< 是否艺术特长生
};

/**
 * @brief 大学信息结构
 */
struct University {
    std::string university_id;       ///< 大学ID
    std::string name;                ///< 大学名称
    std::string province;            ///< 所在省份
    std::string city;                ///< 所在城市
    std::string level;               ///< 层次 (985/211/双一流/普通本科)
    int ranking;                     ///< 排名
    
    // 历年分数线数据 (最近5年)
    std::vector<int> historical_scores;          ///< 历年最低分
    std::vector<int> historical_rankings;        ///< 历年最低排名
    std::unordered_map<std::string, std::vector<int>> major_scores; ///< 专业分数线
    
    // 招生计划
    int total_enrollment;            ///< 总招生计划
    std::unordered_map<std::string, int> major_enrollment; ///< 专业招生计划
    
    // 学校特色
    std::vector<std::string> strong_majors;      ///< 优势专业
    double employment_rate;                      ///< 就业率
    double graduate_salary;                      ///< 毕业生平均薪资
};

/**
 * @brief 专业信息结构
 */
struct Major {
    std::string major_id;            ///< 专业ID
    std::string name;                ///< 专业名称
    std::string category;            ///< 专业类别
    std::string subject_requirements; ///< 选科要求
    
    // 就业相关
    double employment_rate;          ///< 就业率
    double salary_level;             ///< 薪资水平
    std::vector<std::string> career_directions; ///< 职业方向
    
    // 学习难度
    double difficulty_level;         ///< 难度系数
    bool requires_postgraduate;      ///< 是否需要深造
};

/**
 * @brief 志愿推荐结果
 */
struct VolunteerRecommendation {
    std::string university_id;       ///< 大学ID
    std::string university_name;     ///< 大学名称
    std::string major_id;            ///< 专业ID
    std::string major_name;          ///< 专业名称
    
    double admission_probability;    ///< 录取概率 (0-1)
    std::string risk_level;          ///< 风险等级 (冲/稳/保)
    int score_gap;                   ///< 分数差距
    int ranking_gap;                 ///< 排名差距
    
    double match_score;              ///< 匹配度得分
    std::string recommendation_reason; ///< 推荐理由
    std::vector<std::string> risk_factors; ///< 风险因素
};

/**
 * @brief 志愿填报方案
 */
struct VolunteerPlan {
    std::string student_id;          ///< 学生ID
    std::vector<VolunteerRecommendation> recommendations; ///< 推荐列表
    
    // 方案统计
    int total_volunteers;            ///< 总志愿数
    int rush_count;                  ///< 冲刺志愿数
    int stable_count;                ///< 稳妥志愿数
    int safe_count;                  ///< 保底志愿数
    
    double overall_risk_score;       ///< 整体风险评分
    std::string plan_quality;        ///< 方案质量评级
    std::vector<std::string> optimization_suggestions; ///< 优化建议
    
    std::chrono::system_clock::time_point generated_time; ///< 生成时间
};

/**
 * @brief 性能监控统计
 */
struct PerformanceStats {
    std::atomic<uint64_t> total_requests{0};     ///< 总请求数
    std::atomic<uint64_t> successful_requests{0}; ///< 成功请求数
    std::atomic<double> avg_response_time{0.0};   ///< 平均响应时间(ms)
    std::atomic<double> max_response_time{0.0};   ///< 最大响应时间(ms)
    std::atomic<uint64_t> memory_usage{0};        ///< 内存使用量(bytes)
    
    std::chrono::system_clock::time_point last_reset_time; ///< 上次重置时间
};

/**
 * @brief 志愿匹配算法引擎主类
 * 
 * 高性能、线程安全的志愿匹配算法引擎
 * 支持多种匹配策略和实时性能监控
 */
class VolunteerMatcher {
public:
    /**
     * @brief 构造函数
     */
    VolunteerMatcher();
    
    /**
     * @brief 析构函数
     */
    ~VolunteerMatcher();
    
    /**
     * @brief 初始化算法引擎
     * @param config_path 配置文件路径
     * @return 是否初始化成功
     */
    bool Initialize(const std::string& config_path);
    
    /**
     * @brief 加载大学数据
     * @param universities_file 大学数据文件路径
     * @return 加载的大学数量
     */
    int LoadUniversities(const std::string& universities_file);
    
    /**
     * @brief 加载专业数据
     * @param majors_file 专业数据文件路径
     * @return 加载的专业数量
     */
    int LoadMajors(const std::string& majors_file);
    
    /**
     * @brief 加载历史录取数据
     * @param historical_data_file 历史数据文件路径
     * @return 加载的记录数量
     */
    int LoadHistoricalData(const std::string& historical_data_file);
    
    /**
     * @brief 生成志愿填报方案
     * @param student 学生信息
     * @param max_volunteers 最大志愿数量
     * @return 志愿填报方案
     */
    VolunteerPlan GenerateVolunteerPlan(const Student& student, int max_volunteers = 96);
    
    /**
     * @brief 批量生成志愿方案
     * @param students 学生列表
     * @param max_volunteers 最大志愿数量
     * @return 方案列表
     */
    std::vector<VolunteerPlan> BatchGenerateVolunteerPlans(
        const std::vector<Student>& students, 
        int max_volunteers = 96);
    
    /**
     * @brief 获取推荐原因详情
     * @param recommendation 推荐结果
     * @return 详细推荐原因
     */
    std::string GetRecommendationDetails(const VolunteerRecommendation& recommendation);
    
    /**
     * @brief 优化志愿方案
     * @param plan 原方案
     * @param optimization_target 优化目标 ("safety"|"probability"|"preference")
     * @return 优化后的方案
     */
    VolunteerPlan OptimizeVolunteerPlan(
        const VolunteerPlan& plan, 
        const std::string& optimization_target);
    
    /**
     * @brief 获取性能统计
     * @return 性能统计信息
     */
    PerformanceStats GetPerformanceStats() const;
    
    /**
     * @brief 重置性能统计
     */
    void ResetPerformanceStats();
    
    /**
     * @brief 设置日志级别
     * @param level 日志级别 ("DEBUG"|"INFO"|"WARNING"|"ERROR")
     */
    void SetLogLevel(const std::string& level);
    
    /**
     * @brief 热更新数据
     * @param data_type 数据类型 ("universities"|"majors"|"historical")
     * @param file_path 数据文件路径
     * @return 是否更新成功
     */
    bool HotUpdateData(const std::string& data_type, const std::string& file_path);
    
    /**
     * @brief 获取算法引擎状态
     * @return 状态信息 JSON 字符串
     */
    std::string GetEngineStatus() const;

private:
    // 内部实现类（PIMPL模式）
    class Impl;
    std::unique_ptr<Impl> pimpl_;
    
    // 性能监控
    mutable std::mutex stats_mutex_;
    PerformanceStats performance_stats_;
    
    // 线程安全保障
    mutable std::shared_mutex data_mutex_;
    std::atomic<bool> is_initialized_{false};
    
    /**
     * @brief 记录请求开始
     * @return 开始时间点
     */
    std::chrono::high_resolution_clock::time_point StartRequest();
    
    /**
     * @brief 记录请求结束
     * @param start_time 开始时间点
     * @param success 是否成功
     */
    void EndRequest(std::chrono::high_resolution_clock::time_point start_time, bool success);
};

/**
 * @brief 工具函数：从CSV行解析大学信息
 * @param csv_line CSV数据行
 * @return 大学信息结构
 */
University ParseUniversityFromCSV(const std::string& csv_line);

/**
 * @brief 工具函数：从CSV行解析专业信息
 * @param csv_line CSV数据行
 * @return 专业信息结构
 */
Major ParseMajorFromCSV(const std::string& csv_line);

/**
 * @brief 工具函数：构建筛选条件
 * @param student 学生信息
 * @return 筛选条件
 */
FilterCriteria BuildFilterCriteria(const Student& student);

/**
 * @brief 工具函数：为特定大学生成推荐
 * @param student 学生信息
 * @param university 大学信息
 * @return 推荐列表
 */
std::vector<VolunteerRecommendation> GenerateRecommendationsForUniversity(
    const Student& student, const University& university);

/**
 * @brief 工具函数：确定风险等级
 * @param risk_score 风险分数
 * @return 风险等级字符串
 */
std::string DetermineRiskLevel(double risk_score);

/**
 * @brief 工具函数：生成推荐理由
 * @param student 学生信息
 * @param university 大学信息
 * @param major 专业信息
 * @return 推荐理由
 */
std::string GenerateRecommendationReason(
    const Student& student, const University& university, const Major& major);

/**
 * @brief 工具函数：应用冲稳保策略
 * @param candidates 候选推荐列表
 * @param max_volunteers 最大志愿数
 * @return 最终推荐列表
 */
std::vector<VolunteerRecommendation> ApplyRushStableSafeStrategy(
    const std::vector<VolunteerRecommendation>& candidates, int max_volunteers);

/**
 * @brief 工具函数：计算方案统计
 * @param plan 志愿方案（输入输出参数）
 */
void CalculatePlanStatistics(VolunteerPlan& plan);

/**
 * @brief 工具函数：生成优化建议
 * @param plan 志愿方案
 * @return 优化建议列表
 */
std::vector<std::string> GenerateOptimizationSuggestions(const VolunteerPlan& plan);

/**
 * @brief 工具函数：安全性优化
 * @param plan 志愿方案（输入输出参数）
 */
void OptimizeForSafety(VolunteerPlan& plan);

/**
 * @brief 工具函数：概率优化
 * @param plan 志愿方案（输入输出参数）
 */
void OptimizeForProbability(VolunteerPlan& plan);

/**
 * @brief 工具函数：偏好优化
 * @param plan 志愿方案（输入输出参数）
 */
void OptimizeForPreference(VolunteerPlan& plan);

/**
 * @brief 工具函数：计算学生与大学的匹配度
 * @param student 学生信息
 * @param university 大学信息
 * @param major 专业信息
 * @return 匹配度得分 (0-100)
 */
double CalculateMatchScore(const Student& student, const University& university, const Major& major);

/**
 * @brief 工具函数：解析选科组合
 * @param combination 选科组合字符串
 * @return 选科列表
 */
std::vector<std::string> ParseSubjectCombination(const std::string& combination);

/**
 * @brief 工具函数：验证选科要求
 * @param student_subjects 学生选科
 * @param major_requirements 专业要求
 * @return 是否满足要求
 */
bool ValidateSubjectRequirements(
    const std::vector<std::string>& student_subjects,
    const std::string& major_requirements);

/**
 * @brief 工具函数：计算历史趋势
 * @param historical_data 历史数据
 * @return 趋势系数
 */
double CalculateTrendCoefficient(const std::vector<int>& historical_data);

/**
 * @brief AI特征向量结构
 * 
 * 为机器学习模型提供标准化特征输入
 */
struct FeatureVector {
    std::vector<double> numerical_features;     ///< 数值特征
    std::vector<int> categorical_features;      ///< 类别特征
    std::vector<std::string> feature_names;    ///< 特征名称
    double feature_hash;                        ///< 特征哈希（用于缓存）
    
    // 元数据
    std::string student_id;                     ///< 关联学生ID
    std::string university_id;                  ///< 关联大学ID（如果适用）
    std::chrono::system_clock::time_point created_time; ///< 创建时间
};

/**
 * @brief AI推荐结果
 * 
 * 机器学习模型输出的推荐结果
 */
struct AIRecommendation {
    std::string student_id;                     ///< 学生ID
    std::string university_id;                  ///< 大学ID
    std::string major_id;                       ///< 专业ID
    
    double ai_score;                            ///< AI推荐分数 (0-1)
    double confidence;                          ///< 置信度 (0-1)
    std::string model_version;                  ///< 使用的模型版本
    
    // 解释性特征
    std::vector<std::pair<std::string, double>> feature_importance; ///< 特征重要性
    std::string explanation;                    ///< 推荐解释
    
    std::chrono::system_clock::time_point generated_time; ///< 生成时间
};

/**
 * @brief 特征提取器类
 * 
 * 高性能特征工程模块，为AI模型提供标准化特征
 */
class FeatureExtractor {
public:
    /**
     * @brief 构造函数
     */
    FeatureExtractor();
    
    /**
     * @brief 析构函数
     */
    ~FeatureExtractor();
    
    /**
     * @brief 初始化特征提取器
     * @param config_path 配置文件路径
     * @return 是否初始化成功
     */
    bool Initialize(const std::string& config_path);
    
    /**
     * @brief 提取学生特征向量
     * @param student 学生信息
     * @return 特征向量
     */
    FeatureVector ExtractStudentFeatures(const Student& student);
    
    /**
     * @brief 提取大学特征向量
     * @param university 大学信息
     * @return 特征向量
     */
    FeatureVector ExtractUniversityFeatures(const University& university);
    
    /**
     * @brief 提取学生-大学交互特征
     * @param student 学生信息
     * @param university 大学信息
     * @return 交互特征向量
     */
    FeatureVector ExtractInteractionFeatures(const Student& student, const University& university);
    
    /**
     * @brief 批量特征提取
     * @param students 学生列表
     * @param universities 大学列表
     * @return 特征向量列表
     */
    std::vector<FeatureVector> BatchExtractFeatures(
        const std::vector<Student>& students,
        const std::vector<University>& universities);
    
    /**
     * @brief 特征标准化
     * @param features 原始特征向量
     * @return 标准化后的特征向量
     */
    FeatureVector NormalizeFeatures(const FeatureVector& features);
    
    /**
     * @brief 获取特征统计信息
     * @return JSON格式的特征统计
     */
    std::string GetFeatureStats() const;
    
    /**
     * @brief 清空特征缓存
     */
    void ClearFeatureCache();

private:
    class FeatureExtractorImpl;
    std::unique_ptr<FeatureExtractorImpl> pimpl_;
};

/**
 * @brief AI推荐引擎类
 * 
 * 集成机器学习模型的推荐引擎
 */
class AIRecommendationEngine {
public:
    /**
     * @brief 构造函数
     */
    AIRecommendationEngine();
    
    /**
     * @brief 析构函数
     */
    ~AIRecommendationEngine();
    
    /**
     * @brief 初始化AI引擎
     * @param model_path 模型文件路径
     * @param config_path 配置文件路径
     * @return 是否初始化成功
     */
    bool Initialize(const std::string& model_path, const std::string& config_path);
    
    /**
     * @brief 加载AI模型
     * @param model_path 模型文件路径
     * @return 是否加载成功
     */
    bool LoadModel(const std::string& model_path);
    
    /**
     * @brief 单个学生AI推荐
     * @param student 学生信息
     * @param universities 候选大学列表
     * @return AI推荐结果列表
     */
    std::vector<AIRecommendation> GenerateRecommendations(
        const Student& student, 
        const std::vector<University>& universities);
    
    /**
     * @brief 批量AI推荐
     * @param students 学生列表
     * @param universities 大学列表
     * @return 批量推荐结果
     */
    std::vector<std::vector<AIRecommendation>> BatchGenerateRecommendations(
        const std::vector<Student>& students,
        const std::vector<University>& universities);
    
    /**
     * @brief 预测录取概率
     * @param features 特征向量
     * @return 录取概率 (0-1)
     */
    double PredictAdmissionProbability(const FeatureVector& features);
    
    /**
     * @brief 获取模型信息
     * @return 模型信息JSON字符串
     */
    std::string GetModelInfo() const;
    
    /**
     * @brief 热更新模型
     * @param new_model_path 新模型文件路径
     * @return 是否更新成功
     */
    bool UpdateModel(const std::string& new_model_path);

private:
    class AIEngineImpl;
    std::unique_ptr<AIEngineImpl> pimpl_;
};

/**
 * @brief 混合推荐引擎类
 * 
 * 融合传统算法和AI推荐的混合引擎
 */
class HybridRecommendationEngine {
public:
    /**
     * @brief 构造函数
     */
    HybridRecommendationEngine();
    
    /**
     * @brief 析构函数
     */
    ~HybridRecommendationEngine();
    
    /**
     * @brief 初始化混合引擎
     * @param traditional_matcher 传统匹配器
     * @param ai_engine AI推荐引擎
     * @param config_path 配置文件路径
     * @return 是否初始化成功
     */
    bool Initialize(
        std::shared_ptr<VolunteerMatcher> traditional_matcher,
        std::shared_ptr<AIRecommendationEngine> ai_engine,
        const std::string& config_path);
    
    /**
     * @brief 混合推荐生成
     * @param student 学生信息
     * @param max_volunteers 最大推荐数量
     * @return 融合后的志愿方案
     */
    VolunteerPlan GenerateHybridPlan(const Student& student, int max_volunteers = 96);
    
    /**
     * @brief 设置融合权重
     * @param traditional_weight 传统算法权重
     * @param ai_weight AI推荐权重
     * @return 是否设置成功
     */
    bool SetFusionWeights(double traditional_weight, double ai_weight);
    
    /**
     * @brief 获取推荐解释
     * @param recommendation 推荐结果
     * @return 详细解释
     */
    std::string GetHybridExplanation(const VolunteerRecommendation& recommendation);

private:
    class HybridEngineImpl;
    std::unique_ptr<HybridEngineImpl> pimpl_;
};

// C接口扩展 - AI功能
extern "C" {

/**
 * @brief C接口：创建特征提取器
 * @return 特征提取器指针
 */
void* CreateFeatureExtractor();

/**
 * @brief C接口：销毁特征提取器
 * @param extractor 特征提取器指针
 */
void DestroyFeatureExtractor(void* extractor);

/**
 * @brief C接口：提取学生特征
 * @param extractor 特征提取器指针
 * @param student_data 学生数据指针
 * @param feature_vector 输出特征向量
 * @return 是否成功
 */
int ExtractStudentFeatures(void* extractor, const void* student_data, void* feature_vector);

/**
 * @brief C接口：创建AI推荐引擎
 * @return AI引擎指针
 */
void* CreateAIRecommendationEngine();

/**
 * @brief C接口：销毁AI推荐引擎
 * @param engine AI引擎指针
 */
void DestroyAIRecommendationEngine(void* engine);

/**
 * @brief C接口：AI推荐生成
 * @param engine AI引擎指针
 * @param student_data 学生数据
 * @param universities_data 大学数据数组
 * @param university_count 大学数量
 * @param recommendations 输出推荐结果
 * @return 推荐数量
 */
int GenerateAIRecommendations(
    void* engine,
    const void* student_data,
    const void* universities_data,
    int university_count,
    void* recommendations);

/**
 * @brief C接口：创建混合推荐引擎
 * @return 混合引擎指针
 */
void* CreateHybridRecommendationEngine();

/**
 * @brief C接口：销毁混合推荐引擎
 * @param engine 混合引擎指针
 */
void DestroyHybridRecommendationEngine(void* engine);

} // extern "C"

} // namespace volunteer_matcher

#endif // VOLUNTEER_MATCHER_H