/**
 * @file university_filter.h
 * @brief 高考志愿填报系统 - 院校筛选算法
 * @author 高考志愿填报系统开发团队
 * @version 1.0.0
 * @date 2025-01-18
 * 
 * 多维度院校筛选算法：
 * - 分数线筛选
 * - 地理位置筛选
 * - 专业匹配筛选
 * - 个性化偏好筛选
 * - 就业质量筛选
 */

#ifndef UNIVERSITY_FILTER_H
#define UNIVERSITY_FILTER_H

#include <vector>
#include <string>
#include <unordered_map>
#include <unordered_set>
#include <functional>
#include <memory>
#include <optional>

namespace volunteer_matcher {

// 前向声明
struct Student;
struct University;
struct Major;

/**
 * @brief 筛选条件结构
 */
struct FilterCriteria {
    // 分数相关
    int min_score = 0;                      ///< 最低分数要求
    int max_score = 750;                    ///< 最高分数要求
    int score_tolerance = 20;               ///< 分数容忍度
    bool use_ranking = true;                ///< 是否使用排名筛选
    
    // 地理位置
    std::vector<std::string> allowed_provinces; ///< 允许的省份
    std::vector<std::string> preferred_cities;  ///< 偏好城市
    std::vector<std::string> excluded_cities;   ///< 排除城市
    bool prefer_hometown = false;           ///< 是否偏好家乡
    
    // 学校层次
    std::vector<std::string> school_levels; ///< 学校层次 (985/211/双一流)
    int min_ranking = 0;                    ///< 最低排名要求
    int max_ranking = 1000;                 ///< 最高排名要求
    
    // 专业相关
    std::vector<std::string> target_majors; ///< 目标专业
    std::vector<std::string> major_categories; ///< 专业类别
    bool strict_major_match = false;        ///< 严格专业匹配
    
    // 就业相关
    double min_employment_rate = 0.0;       ///< 最低就业率
    double min_salary_level = 0.0;          ///< 最低薪资水平
    bool consider_career_prospects = true;   ///< 考虑职业前景
    
    // 特殊要求
    bool only_public_universities = false;  ///< 仅公办大学
    bool allow_independent_colleges = true;  ///< 允许独立学院
    bool allow_joint_programs = true;       ///< 允许合作办学
    
    // 招生计划
    int min_enrollment = 1;                 ///< 最小招生计划
    bool consider_enrollment_trend = true;   ///< 考虑招生趋势
};

/**
 * @brief 筛选结果
 */
struct FilterResult {
    std::vector<std::string> university_ids; ///< 符合条件的大学ID列表
    std::unordered_map<std::string, double> match_scores; ///< 匹配度得分
    std::unordered_map<std::string, std::vector<std::string>> filter_reasons; ///< 筛选原因
    
    int total_candidates;                   ///< 候选大学总数
    int filtered_count;                     ///< 筛选后数量
    double filter_ratio;                    ///< 筛选比例
    
    std::chrono::system_clock::time_point filter_time; ///< 筛选时间
};

/**
 * @brief 筛选统计信息
 */
struct FilterStats {
    std::unordered_map<std::string, int> filter_type_counts; ///< 各筛选条件命中数
    std::unordered_map<std::string, int> province_distribution; ///< 省份分布
    std::unordered_map<std::string, int> level_distribution; ///< 层次分布
    std::unordered_map<std::string, int> major_distribution; ///< 专业分布
    
    double avg_match_score;                 ///< 平均匹配度
    double max_match_score;                 ///< 最高匹配度
    double min_match_score;                 ///< 最低匹配度
};

/**
 * @brief 院校筛选器类
 * 
 * 高性能、多维度的院校筛选算法
 * 支持复杂筛选条件和个性化偏好
 */
class UniversityFilter {
public:
    /**
     * @brief 构造函数
     */
    UniversityFilter();
    
    /**
     * @brief 析构函数
     */
    ~UniversityFilter();
    
    /**
     * @brief 初始化筛选器
     * @param config_path 配置文件路径
     * @return 是否初始化成功
     */
    bool Initialize(const std::string& config_path);
    
    /**
     * @brief 设置大学数据
     * @param universities 大学列表
     * @return 设置的大学数量
     */
    int SetUniversities(const std::vector<University>& universities);
    
    /**
     * @brief 设置专业数据
     * @param majors 专业列表
     * @return 设置的专业数量
     */
    int SetMajors(const std::vector<Major>& majors);
    
    /**
     * @brief 基础筛选 - 根据学生基本信息筛选
     * @param student 学生信息
     * @param criteria 筛选条件
     * @return 筛选结果
     */
    FilterResult BasicFilter(const Student& student, const FilterCriteria& criteria) const;
    
    /**
     * @brief 高级筛选 - 包含个性化偏好
     * @param student 学生信息
     * @param criteria 筛选条件
     * @return 筛选结果
     */
    FilterResult AdvancedFilter(const Student& student, const FilterCriteria& criteria) const;
    
    /**
     * @brief 智能筛选 - 基于机器学习的筛选
     * @param student 学生信息
     * @param max_candidates 最大候选数量
     * @return 筛选结果
     */
    FilterResult IntelligentFilter(const Student& student, int max_candidates = 200) const;
    
    /**
     * @brief 分数线筛选
     * @param student 学生信息
     * @param score_range 分数范围 (lower, upper)
     * @return 符合条件的大学ID列表
     */
    std::vector<std::string> FilterByScore(
        const Student& student,
        const std::pair<int, int>& score_range) const;
    
    /**
     * @brief 地理位置筛选
     * @param criteria 地理位置条件
     * @return 符合条件的大学ID列表
     */
    std::vector<std::string> FilterByLocation(const FilterCriteria& criteria) const;
    
    /**
     * @brief 专业筛选
     * @param target_majors 目标专业列表
     * @param student_subjects 学生选科
     * @return 符合条件的大学-专业组合
     */
    std::vector<std::pair<std::string, std::string>> FilterByMajor(
        const std::vector<std::string>& target_majors,
        const std::vector<std::string>& student_subjects) const;
    
    /**
     * @brief 学校层次筛选
     * @param levels 目标层次列表
     * @param ranking_range 排名范围
     * @return 符合条件的大学ID列表
     */
    std::vector<std::string> FilterByLevel(
        const std::vector<std::string>& levels,
        const std::pair<int, int>& ranking_range) const;
    
    /**
     * @brief 就业质量筛选
     * @param min_employment_rate 最低就业率
     * @param min_salary 最低薪资
     * @return 符合条件的大学-专业组合
     */
    std::vector<std::pair<std::string, std::string>> FilterByEmployment(
        double min_employment_rate,
        double min_salary) const;
    
    /**
     * @brief 招生计划筛选
     * @param student 学生信息
     * @param min_enrollment 最小招生数
     * @return 符合条件的大学ID列表
     */
    std::vector<std::string> FilterByEnrollment(
        const Student& student,
        int min_enrollment) const;
    
    /**
     * @brief 组合筛选 - 多个条件的逻辑组合
     * @param filters 筛选器列表
     * @param logic_operator 逻辑运算符 ("AND"|"OR"|"XOR")
     * @return 筛选结果
     */
    FilterResult CombineFilters(
        const std::vector<FilterResult>& filters,
        const std::string& logic_operator = "AND") const;
    
    /**
     * @brief 计算匹配度得分
     * @param student 学生信息
     * @param university_id 大学ID
     * @param criteria 筛选条件
     * @return 匹配度得分 (0-100)
     */
    double CalculateMatchScore(
        const Student& student,
        const std::string& university_id,
        const FilterCriteria& criteria) const;
    
    /**
     * @brief 批量计算匹配度
     * @param student 学生信息
     * @param university_ids 大学ID列表
     * @param criteria 筛选条件
     * @return 匹配度得分映射
     */
    std::unordered_map<std::string, double> BatchCalculateMatchScores(
        const Student& student,
        const std::vector<std::string>& university_ids,
        const FilterCriteria& criteria) const;
    
    /**
     * @brief 获取筛选统计
     * @param filter_result 筛选结果
     * @return 统计信息
     */
    FilterStats GetFilterStats(const FilterResult& filter_result) const;
    
    /**
     * @brief 生成筛选报告
     * @param filter_result 筛选结果
     * @param criteria 筛选条件
     * @return 筛选报告JSON字符串
     */
    std::string GenerateFilterReport(
        const FilterResult& filter_result,
        const FilterCriteria& criteria) const;
    
    /**
     * @brief 优化筛选条件
     * @param student 学生信息
     * @param target_count 目标数量
     * @param base_criteria 基础条件
     * @return 优化后的筛选条件
     */
    FilterCriteria OptimizeFilterCriteria(
        const Student& student,
        int target_count,
        const FilterCriteria& base_criteria) const;
    
    /**
     * @brief 设置自定义筛选器
     * @param name 筛选器名称
     * @param filter_func 筛选函数
     */
    void SetCustomFilter(
        const std::string& name,
        std::function<std::vector<std::string>(const Student&, const FilterCriteria&)> filter_func);
    
    /**
     * @brief 应用自定义筛选器
     * @param name 筛选器名称
     * @param student 学生信息
     * @param criteria 筛选条件
     * @return 筛选结果
     */
    std::vector<std::string> ApplyCustomFilter(
        const std::string& name,
        const Student& student,
        const FilterCriteria& criteria) const;
    
    /**
     * @brief 缓存筛选结果
     * @param key 缓存键
     * @param result 筛选结果
     */
    void CacheFilterResult(const std::string& key, const FilterResult& result);
    
    /**
     * @brief 获取缓存的筛选结果
     * @param key 缓存键
     * @return 筛选结果 (如果存在)
     */
    std::optional<FilterResult> GetCachedFilterResult(const std::string& key) const;
    
    /**
     * @brief 清空筛选缓存
     */
    void ClearFilterCache();
    
    /**
     * @brief 获取可用的筛选维度
     * @return 筛选维度列表
     */
    std::vector<std::string> GetAvailableFilterDimensions() const;
    
    /**
     * @brief 获取筛选器状态
     * @return 状态信息JSON字符串
     */
    std::string GetFilterStatus() const;

private:
    // 内部实现类
    class Impl;
    std::unique_ptr<Impl> pimpl_;
    
    /**
     * @brief 分数匹配计算
     * @param student_score 学生分数
     * @param university_scores 大学历年分数
     * @return 匹配度得分
     */
    double CalculateScoreMatch(int student_score, const std::vector<int>& university_scores) const;
    
    /**
     * @brief 地理偏好计算
     * @param student 学生信息
     * @param university 大学信息
     * @return 地理偏好得分
     */
    double CalculateLocationPreference(const Student& student, const University& university) const;
    
    /**
     * @brief 专业匹配计算
     * @param student_preferences 学生专业偏好
     * @param university_majors 大学专业列表
     * @return 专业匹配得分
     */
    double CalculateMajorMatch(
        const std::vector<std::string>& student_preferences,
        const std::vector<std::string>& university_majors) const;
    
    /**
     * @brief 选科要求验证
     * @param student_subjects 学生选科
     * @param required_subjects 专业要求选科
     * @return 是否满足要求
     */
    bool ValidateSubjectRequirements(
        const std::vector<std::string>& student_subjects,
        const std::string& required_subjects) const;
    
    /**
     * @brief 历史趋势分析
     * @param historical_data 历史数据
     * @return 趋势系数
     */
    double AnalyzeTrend(const std::vector<int>& historical_data) const;
};

/**
 * @brief 筛选条件构建器
 */
class FilterCriteriaBuilder {
public:
    FilterCriteriaBuilder();
    
    FilterCriteriaBuilder& SetScoreRange(int min_score, int max_score);
    FilterCriteriaBuilder& SetScoreTolerance(int tolerance);
    FilterCriteriaBuilder& AddAllowedProvince(const std::string& province);
    FilterCriteriaBuilder& AddPreferredCity(const std::string& city);
    FilterCriteriaBuilder& AddExcludedCity(const std::string& city);
    FilterCriteriaBuilder& SetSchoolLevels(const std::vector<std::string>& levels);
    FilterCriteriaBuilder& SetRankingRange(int min_ranking, int max_ranking);
    FilterCriteriaBuilder& AddTargetMajor(const std::string& major);
    FilterCriteriaBuilder& SetEmploymentRequirements(double min_rate, double min_salary);
    FilterCriteriaBuilder& SetPublicUniversitiesOnly(bool only_public);
    FilterCriteriaBuilder& SetMinEnrollment(int min_enrollment);
    
    FilterCriteria Build() const;

private:
    FilterCriteria criteria_;
};

} // namespace volunteer_matcher

#endif // UNIVERSITY_FILTER_H