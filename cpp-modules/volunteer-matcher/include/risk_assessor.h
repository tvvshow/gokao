/**
 * @file risk_assessor.h
 * @brief 高考志愿填报系统 - 风险评估算法
 * @author 高考志愿填报系统开发团队
 * @version 1.0.0
 * @date 2025-01-18
 * 
 * 智能风险评估算法：
 * - 录取风险分析
 * - 志愿方案风险评估
 * - 风险分散策略
 * - 风险预警系统
 * - 风险缓解建议
 */

#ifndef RISK_ASSESSOR_H
#define RISK_ASSESSOR_H

#include <vector>
#include <string>
#include <unordered_map>
#include <memory>
#include <chrono>

namespace volunteer_matcher {

// 前向声明
struct Student;
struct University;
struct Major;
struct VolunteerPlan;
struct VolunteerRecommendation;

/**
 * @brief 风险类型枚举
 */
enum class RiskType {
    ADMISSION_RISK,     ///< 录取风险
    SCORE_VOLATILITY,   ///< 分数波动风险
    ENROLLMENT_CHANGE,  ///< 招生计划变化风险
    POLICY_CHANGE,      ///< 政策变化风险
    COMPETITION_RISK,   ///< 竞争激烈度风险
    MAJOR_MISMATCH,     ///< 专业不匹配风险
    LOCATION_RISK,      ///< 地理位置风险
    EMPLOYMENT_RISK     ///< 就业风险
};

/**
 * @brief 风险等级枚举
 */
enum class RiskLevel {
    VERY_LOW = 1,       ///< 极低风险
    LOW = 2,            ///< 低风险
    MEDIUM = 3,         ///< 中等风险
    HIGH = 4,           ///< 高风险
    VERY_HIGH = 5       ///< 极高风险
};

/**
 * @brief 风险因素
 */
struct RiskFactor {
    RiskType type;                          ///< 风险类型
    RiskLevel level;                        ///< 风险等级
    double probability;                     ///< 发生概率 (0-1)
    double impact;                          ///< 影响程度 (0-1)
    double severity;                        ///< 风险严重度
    std::string description;                ///< 风险描述
    std::vector<std::string> causes;        ///< 风险原因
    std::vector<std::string> consequences;  ///< 潜在后果
};

/**
 * @brief 风险评估结果
 */
struct RiskAssessment {
    std::string target_id;                  ///< 评估目标ID (大学/专业/方案)
    double overall_risk_score;              ///< 总体风险分数 (0-100)
    RiskLevel overall_risk_level;           ///< 总体风险等级
    
    std::vector<RiskFactor> risk_factors;   ///< 具体风险因素
    std::unordered_map<RiskType, double> risk_type_scores; ///< 各类型风险得分
    
    double confidence;                      ///< 评估置信度 (0-1)
    std::string assessment_summary;         ///< 评估摘要
    std::chrono::system_clock::time_point assessment_time; ///< 评估时间
};

/**
 * @brief 志愿方案风险评估
 */
struct PlanRiskAssessment {
    std::string plan_id;                    ///< 方案ID
    double overall_plan_risk;               ///< 方案整体风险
    
    // 各志愿风险
    std::vector<RiskAssessment> volunteer_risks; ///< 各志愿风险评估
    
    // 风险分布
    int very_high_risk_count;               ///< 极高风险志愿数
    int high_risk_count;                    ///< 高风险志愿数
    int medium_risk_count;                  ///< 中等风险志愿数
    int low_risk_count;                     ///< 低风险志愿数
    int very_low_risk_count;                ///< 极低风险志愿数
    
    // 风险指标
    double risk_concentration;              ///< 风险集中度
    double risk_diversity;                  ///< 风险分散度
    double expected_admission_rate;         ///< 预期录取率
    
    std::vector<std::string> risk_warnings; ///< 风险预警
    std::vector<std::string> optimization_suggestions; ///< 优化建议
};

/**
 * @brief 风险缓解策略
 */
struct RiskMitigationStrategy {
    RiskType target_risk_type;              ///< 目标风险类型
    std::string strategy_name;              ///< 策略名称
    std::string description;                ///< 策略描述
    
    std::vector<std::string> action_items;  ///< 具体行动项
    double effectiveness;                   ///< 有效性评估 (0-1)
    double implementation_difficulty;       ///< 实施难度 (0-1)
    double cost;                           ///< 成本估算 (0-1)
    
    std::string priority;                   ///< 优先级 ("HIGH"|"MEDIUM"|"LOW")
    std::chrono::duration<int> time_to_implement; ///< 实施时间
};

/**
 * @brief 风险评估器类
 * 
 * 智能风险评估系统，提供全方位的风险分析和缓解建议
 */
class RiskAssessor {
public:
    /**
     * @brief 构造函数
     */
    RiskAssessor();
    
    /**
     * @brief 析构函数
     */
    ~RiskAssessor();
    
    /**
     * @brief 初始化风险评估器
     * @param config_path 配置文件路径
     * @return 是否初始化成功
     */
    bool Initialize(const std::string& config_path);
    
    /**
     * @brief 设置历史数据
     * @param historical_data_path 历史数据文件路径
     * @return 是否设置成功
     */
    bool SetHistoricalData(const std::string& historical_data_path);
    
    /**
     * @brief 评估单个志愿的风险
     * @param student 学生信息
     * @param university 大学信息
     * @param major 专业信息
     * @return 风险评估结果
     */
    RiskAssessment AssessVolunteerRisk(
        const Student& student,
        const University& university,
        const Major& major) const;
    
    /**
     * @brief 评估志愿方案的整体风险
     * @param student 学生信息
     * @param plan 志愿方案
     * @return 方案风险评估结果
     */
    PlanRiskAssessment AssessPlanRisk(
        const Student& student,
        const VolunteerPlan& plan) const;
    
    /**
     * @brief 批量评估志愿风险
     * @param student 学生信息
     * @param recommendations 推荐列表
     * @return 风险评估结果列表
     */
    std::vector<RiskAssessment> BatchAssessVolunteerRisks(
        const Student& student,
        const std::vector<VolunteerRecommendation>& recommendations) const;
    
    /**
     * @brief 录取风险评估
     * @param student 学生信息
     * @param university 大学信息
     * @param major 专业信息
     * @return 录取风险评估
     */
    RiskAssessment AssessAdmissionRisk(
        const Student& student,
        const University& university,
        const Major& major) const;
    
    /**
     * @brief 分数波动风险评估
     * @param student 学生信息
     * @param university 大学信息
     * @return 分数波动风险
     */
    RiskAssessment AssessScoreVolatilityRisk(
        const Student& student,
        const University& university) const;
    
    /**
     * @brief 招生计划变化风险评估
     * @param university 大学信息
     * @param major 专业信息
     * @return 招生计划风险
     */
    RiskAssessment AssessEnrollmentChangeRisk(
        const University& university,
        const Major& major) const;
    
    /**
     * @brief 竞争激烈度风险评估
     * @param student 学生信息
     * @param university 大学信息
     * @param major 专业信息
     * @return 竞争风险
     */
    RiskAssessment AssessCompetitionRisk(
        const Student& student,
        const University& university,
        const Major& major) const;
    
    /**
     * @brief 专业匹配风险评估
     * @param student 学生信息
     * @param major 专业信息
     * @return 专业匹配风险
     */
    RiskAssessment AssessMajorMismatchRisk(
        const Student& student,
        const Major& major) const;
    
    /**
     * @brief 就业风险评估
     * @param major 专业信息
     * @param university 大学信息
     * @return 就业风险
     */
    RiskAssessment AssessEmploymentRisk(
        const Major& major,
        const University& university) const;
    
    /**
     * @brief 计算风险分散度
     * @param plan 志愿方案
     * @return 风险分散度得分 (0-1)
     */
    double CalculateRiskDiversification(const VolunteerPlan& plan) const;
    
    /**
     * @brief 风险预警检查
     * @param plan_assessment 方案风险评估
     * @return 预警信息列表
     */
    std::vector<std::string> CheckRiskWarnings(const PlanRiskAssessment& plan_assessment) const;
    
    /**
     * @brief 生成风险缓解策略
     * @param risk_assessment 风险评估结果
     * @return 缓解策略列表
     */
    std::vector<RiskMitigationStrategy> GenerateMitigationStrategies(
        const RiskAssessment& risk_assessment) const;
    
    /**
     * @brief 方案风险优化建议
     * @param plan_assessment 方案风险评估
     * @return 优化建议列表
     */
    std::vector<std::string> GenerateOptimizationSuggestions(
        const PlanRiskAssessment& plan_assessment) const;
    
    /**
     * @brief 风险容忍度评估
     * @param student 学生信息
     * @return 风险容忍度得分 (0-1)
     */
    double AssessRiskTolerance(const Student& student) const;
    
    /**
     * @brief 历史风险分析
     * @param university 大学信息
     * @param major 专业信息
     * @param years 分析年数
     * @return 历史风险报告
     */
    std::string AnalyzeHistoricalRisk(
        const University& university,
        const Major& major,
        int years = 5) const;
    
    /**
     * @brief 设置风险权重
     * @param risk_weights 风险类型权重映射
     */
    void SetRiskWeights(const std::unordered_map<RiskType, double>& risk_weights);
    
    /**
     * @brief 获取风险权重
     * @return 当前风险权重设置
     */
    std::unordered_map<RiskType, double> GetRiskWeights() const;
    
    /**
     * @brief 风险敏感性分析
     * @param base_assessment 基础风险评估
     * @param parameter_changes 参数变化
     * @return 敏感性分析结果
     */
    std::unordered_map<std::string, double> PerformSensitivityAnalysis(
        const RiskAssessment& base_assessment,
        const std::unordered_map<std::string, double>& parameter_changes) const;
    
    /**
     * @brief 压力测试
     * @param plan 志愿方案
     * @param stress_scenarios 压力测试场景
     * @return 压力测试结果
     */
    std::unordered_map<std::string, PlanRiskAssessment> PerformStressTest(
        const VolunteerPlan& plan,
        const std::vector<std::string>& stress_scenarios) const;
    
    /**
     * @brief 生成风险报告
     * @param plan_assessment 方案风险评估
     * @param format 报告格式 ("JSON"|"HTML"|"PDF")
     * @return 风险报告内容
     */
    std::string GenerateRiskReport(
        const PlanRiskAssessment& plan_assessment,
        const std::string& format = "JSON") const;
    
    /**
     * @brief 风险监控告警
     * @param assessment 风险评估
     * @param threshold 告警阈值
     * @return 是否需要告警
     */
    bool CheckRiskAlert(const RiskAssessment& assessment, double threshold = 0.7) const;
    
    /**
     * @brief 更新风险模型
     * @param new_data 新的风险数据
     * @return 是否更新成功
     */
    bool UpdateRiskModel(const std::vector<RiskAssessment>& new_data);
    
    /**
     * @brief 获取风险评估器状态
     * @return 状态信息JSON字符串
     */
    std::string GetAssessorStatus() const;

private:
    // 内部实现类
    class Impl;
    std::unique_ptr<Impl> pimpl_;
    
    /**
     * @brief 计算基础风险分数
     * @param probability 发生概率
     * @param impact 影响程度
     * @return 风险分数
     */
    double CalculateBaseRiskScore(double probability, double impact) const;
    
    /**
     * @brief 分析历史波动性
     * @param historical_scores 历史分数
     * @return 波动性指标
     */
    double AnalyzeVolatility(const std::vector<int>& historical_scores) const;
    
    /**
     * @brief 评估市场竞争
     * @param university 大学信息
     * @param major 专业信息
     * @return 竞争激烈度
     */
    double AssessMarketCompetition(const University& university, const Major& major) const;
    
    /**
     * @brief 计算风险相关性
     * @param risk_factors 风险因素列表
     * @return 风险相关性矩阵
     */
    std::vector<std::vector<double>> CalculateRiskCorrelation(
        const std::vector<RiskFactor>& risk_factors) const;
    
    /**
     * @brief 蒙特卡罗风险模拟
     * @param base_scenario 基础场景
     * @param simulation_count 模拟次数
     * @return 模拟结果分布
     */
    std::vector<double> MonteCarloRiskSimulation(
        const RiskAssessment& base_scenario,
        int simulation_count = 10000) const;
};

/**
 * @brief 风险等级转换工具
 */
class RiskLevelConverter {
public:
    /**
     * @brief 分数转风险等级
     * @param score 风险分数 (0-100)
     * @return 风险等级
     */
    static RiskLevel ScoreToLevel(double score);
    
    /**
     * @brief 风险等级转分数
     * @param level 风险等级
     * @return 风险分数
     */
    static double LevelToScore(RiskLevel level);
    
    /**
     * @brief 风险等级转字符串
     * @param level 风险等级
     * @return 风险等级字符串
     */
    static std::string LevelToString(RiskLevel level);
    
    /**
     * @brief 字符串转风险等级
     * @param level_str 风险等级字符串
     * @return 风险等级
     */
    static RiskLevel StringToLevel(const std::string& level_str);
    
    /**
     * @brief 风险类型转字符串
     * @param type 风险类型
     * @return 风险类型字符串
     */
    static std::string TypeToString(RiskType type);
    
    /**
     * @brief 字符串转风险类型
     * @param type_str 风险类型字符串
     * @return 风险类型
     */
    static RiskType StringToType(const std::string& type_str);
};

} // namespace volunteer_matcher

#endif // RISK_ASSESSOR_H