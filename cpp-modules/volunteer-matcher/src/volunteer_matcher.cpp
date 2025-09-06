/**
 * @file volunteer_matcher.cpp
 * @brief 高考志愿填报系统 - 志愿匹配算法引擎实现
 * @author 高考志愿填报系统开发团队
 * @version 1.0.0
 * @date 2025-01-18
 */

#include "volunteer_matcher.h"
#include "admission_predictor.h"
#include "university_filter.h"
#include "risk_assessor.h"

#include <algorithm>
#include <fstream>
#include <sstream>
#include <cmath>
#include <random>
#include <thread>
#include <future>
<<<<<<< HEAD
#include <shared_mutex>
=======
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
#include <iomanip>
#include <unordered_set>
#include <json/json.h>

namespace volunteer_matcher {

// PIMPL实现类
class VolunteerMatcher::Impl {
public:
    // 数据存储
    std::vector<University> universities_;
    std::vector<Major> majors_;
    std::unordered_map<std::string, University> university_map_;
    std::unordered_map<std::string, Major> major_map_;
    
    // 算法组件
    std::unique_ptr<AdmissionPredictor> admission_predictor_;
    std::unique_ptr<UniversityFilter> university_filter_;
    std::unique_ptr<RiskAssessor> risk_assessor_;
    
    // 配置参数
    Json::Value config_;
    std::string config_path_;
    
    // 线程池
    std::vector<std::thread> thread_pool_;
    std::atomic<bool> shutdown_{false};
    
    // 缓存系统
    std::unordered_map<std::string, VolunteerPlan> plan_cache_;
    mutable std::shared_mutex cache_mutex_;
    
    Impl() {
        admission_predictor_ = std::make_unique<AdmissionPredictor>();
        university_filter_ = std::make_unique<UniversityFilter>();
        risk_assessor_ = std::make_unique<RiskAssessor>();
    }
    
    ~Impl() {
        shutdown_ = true;
        for (auto& thread : thread_pool_) {
            if (thread.joinable()) {
                thread.join();
            }
        }
    }
    
    bool LoadConfig(const std::string& config_path) {
        std::ifstream config_file(config_path);
        if (!config_file.is_open()) {
            return false;
        }
        
        Json::Reader reader;
        if (!reader.parse(config_file, config_)) {
            return false;
        }
        
        config_path_ = config_path;
        return true;
    }
    
    std::string GenerateCacheKey(const Student& student, int max_volunteers) {
        std::ostringstream oss;
        oss << student.student_id << "_" << student.total_score << "_" 
            << student.ranking << "_" << student.province << "_" << max_volunteers;
        return oss.str();
    }
};

// VolunteerMatcher实现

VolunteerMatcher::VolunteerMatcher() : pimpl_(std::make_unique<Impl>()) {}

VolunteerMatcher::~VolunteerMatcher() = default;

bool VolunteerMatcher::Initialize(const std::string& config_path) {
    auto start_time = StartRequest();
    
    try {
        if (!pimpl_->LoadConfig(config_path)) {
            EndRequest(start_time, false);
            return false;
        }
        
        // 初始化各算法组件
        bool success = true;
        success &= pimpl_->admission_predictor_->Initialize(config_path);
        success &= pimpl_->university_filter_->Initialize(config_path);
        success &= pimpl_->risk_assessor_->Initialize(config_path);
        
        if (success) {
            is_initialized_ = true;
        }
        
        EndRequest(start_time, success);
        return success;
    } catch (const std::exception& e) {
        EndRequest(start_time, false);
        return false;
    }
}

int VolunteerMatcher::LoadUniversities(const std::string& universities_file) {
    auto start_time = StartRequest();
    
    try {
        std::ifstream file(universities_file);
        if (!file.is_open()) {
            EndRequest(start_time, false);
            return 0;
        }
        
        std::string line;
        std::getline(file, line); // 跳过标题行
        
        pimpl_->universities_.clear();
        pimpl_->university_map_.clear();
        
        while (std::getline(file, line)) {
            University university = ParseUniversityFromCSV(line);
            if (!university.university_id.empty()) {
                pimpl_->universities_.push_back(university);
                pimpl_->university_map_[university.university_id] = university;
            }
        }
        
        // 设置到筛选器
        pimpl_->university_filter_->SetUniversities(pimpl_->universities_);
        
        int count = static_cast<int>(pimpl_->universities_.size());
        EndRequest(start_time, true);
        return count;
    } catch (const std::exception& e) {
        EndRequest(start_time, false);
        return 0;
    }
}

int VolunteerMatcher::LoadMajors(const std::string& majors_file) {
    auto start_time = StartRequest();
    
    try {
        std::ifstream file(majors_file);
        if (!file.is_open()) {
            EndRequest(start_time, false);
            return 0;
        }
        
        std::string line;
        std::getline(file, line); // 跳过标题行
        
        pimpl_->majors_.clear();
        pimpl_->major_map_.clear();
        
        while (std::getline(file, line)) {
            Major major = ParseMajorFromCSV(line);
            if (!major.major_id.empty()) {
                pimpl_->majors_.push_back(major);
                pimpl_->major_map_[major.major_id] = major;
            }
        }
        
        // 设置到筛选器
        pimpl_->university_filter_->SetMajors(pimpl_->majors_);
        
        int count = static_cast<int>(pimpl_->majors_.size());
        EndRequest(start_time, true);
        return count;
    } catch (const std::exception& e) {
        EndRequest(start_time, false);
        return 0;
    }
}

int VolunteerMatcher::LoadHistoricalData(const std::string& historical_data_file) {
    auto start_time = StartRequest();
    
    try {
        bool success = pimpl_->risk_assessor_->SetHistoricalData(historical_data_file);
        EndRequest(start_time, success);
        return success ? 1 : 0;
    } catch (const std::exception& e) {
        EndRequest(start_time, false);
        return 0;
    }
}

VolunteerPlan VolunteerMatcher::GenerateVolunteerPlan(const Student& student, int max_volunteers) {
    auto start_time = StartRequest();
    
    try {
        if (!is_initialized_) {
            EndRequest(start_time, false);
            return VolunteerPlan{};
        }
        
        // 检查缓存
        std::string cache_key = pimpl_->GenerateCacheKey(student, max_volunteers);
        {
            std::shared_lock<std::shared_mutex> lock(pimpl_->cache_mutex_);
            auto it = pimpl_->plan_cache_.find(cache_key);
            if (it != pimpl_->plan_cache_.end()) {
                EndRequest(start_time, true);
                return it->second;
            }
        }
        
        VolunteerPlan plan;
        plan.student_id = student.student_id;
        plan.generated_time = std::chrono::system_clock::now();
        
        // 第一步：智能筛选候选院校
<<<<<<< HEAD
        FilterCriteria criteria = this->BuildFilterCriteria(student);
=======
        FilterCriteria criteria = BuildFilterCriteria(student);
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
        auto filter_result = pimpl_->university_filter_->IntelligentFilter(student, max_volunteers * 3);
        
        if (filter_result.university_ids.empty()) {
            EndRequest(start_time, false);
            return plan;
        }
        
        // 第二步：生成推荐列表
        std::vector<VolunteerRecommendation> candidates;
        for (const auto& uni_id : filter_result.university_ids) {
            auto it = pimpl_->university_map_.find(uni_id);
            if (it != pimpl_->university_map_.end()) {
                auto recommendations = GenerateRecommendationsForUniversity(student, it->second);
                candidates.insert(candidates.end(), recommendations.begin(), recommendations.end());
            }
        }
        
        // 第三步：计算录取概率和风险评估
        for (auto& rec : candidates) {
            auto it_uni = pimpl_->university_map_.find(rec.university_id);
            auto it_major = pimpl_->major_map_.find(rec.major_id);
            
            if (it_uni != pimpl_->university_map_.end() && it_major != pimpl_->major_map_.end()) {
                // 录取概率预测
                auto prediction = pimpl_->admission_predictor_->PredictAdmissionProbability(
                    student, it_uni->second, it_major->second);
                rec.admission_probability = prediction.probability;
                
                // 风险评估
                auto risk_assessment = pimpl_->risk_assessor_->AssessVolunteerRisk(
                    student, it_uni->second, it_major->second);
                rec.risk_level = DetermineRiskLevel(risk_assessment.overall_risk_score);
                
                // 匹配度计算
                rec.match_score = CalculateMatchScore(student, it_uni->second, it_major->second);
                
                // 推荐理由生成
                rec.recommendation_reason = GenerateRecommendationReason(student, it_uni->second, it_major->second);
            }
        }
        
        // 第四步：排序和筛选
        std::sort(candidates.begin(), candidates.end(), 
                 [](const VolunteerRecommendation& a, const VolunteerRecommendation& b) {
                     return a.match_score > b.match_score;
                 });
        
        // 第五步：冲稳保策略分配
        plan.recommendations = ApplyRushStableSafeStrategy(candidates, max_volunteers);
        
        // 第六步：计算方案统计
        CalculatePlanStatistics(plan);
        
        // 第七步：生成优化建议
        plan.optimization_suggestions = GenerateOptimizationSuggestions(plan);
        
        // 缓存结果
        {
            std::unique_lock<std::shared_mutex> lock(pimpl_->cache_mutex_);
            pimpl_->plan_cache_[cache_key] = plan;
        }
        
        EndRequest(start_time, true);
        return plan;
    } catch (const std::exception& e) {
        EndRequest(start_time, false);
        return VolunteerPlan{};
    }
}

std::vector<VolunteerPlan> VolunteerMatcher::BatchGenerateVolunteerPlans(
    const std::vector<Student>& students, int max_volunteers) {
    
    std::vector<VolunteerPlan> plans;
    plans.reserve(students.size());
    
    // 使用线程池并行处理
    const int num_threads = std::min(static_cast<int>(students.size()), 
                                   static_cast<int>(std::thread::hardware_concurrency()));
<<<<<<< HEAD
    (void)num_threads; // 避免未使用变量警告
=======
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
    std::vector<std::future<VolunteerPlan>> futures;
    
    for (const auto& student : students) {
        auto future = std::async(std::launch::async, 
                                [this, &student, max_volunteers]() {
                                    return GenerateVolunteerPlan(student, max_volunteers);
                                });
        futures.push_back(std::move(future));
    }
    
    for (auto& future : futures) {
        plans.push_back(future.get());
    }
    
    return plans;
}

std::string VolunteerMatcher::GetRecommendationDetails(const VolunteerRecommendation& recommendation) {
    Json::Value details;
    details["university_name"] = recommendation.university_name;
    details["major_name"] = recommendation.major_name;
    details["admission_probability"] = recommendation.admission_probability;
    details["risk_level"] = recommendation.risk_level;
    details["match_score"] = recommendation.match_score;
    details["recommendation_reason"] = recommendation.recommendation_reason;
    
    Json::StreamWriterBuilder builder;
    return Json::writeString(builder, details);
}

VolunteerPlan VolunteerMatcher::OptimizeVolunteerPlan(
    const VolunteerPlan& plan, const std::string& optimization_target) {
    
    VolunteerPlan optimized_plan = plan;
    
    if (optimization_target == "safety") {
        // 安全优化：增加保底志愿比例
        OptimizeForSafety(optimized_plan);
    } else if (optimization_target == "probability") {
        // 概率优化：最大化整体录取概率
        OptimizeForProbability(optimized_plan);
    } else if (optimization_target == "preference") {
        // 偏好优化：最大化学生偏好匹配
        OptimizeForPreference(optimized_plan);
    }
    
    return optimized_plan;
}

<<<<<<< HEAD
void VolunteerMatcher::GetPerformanceStats(PerformanceStats& stats) const {
    std::lock_guard<std::mutex> lock(stats_mutex_);
    stats.total_requests.store(performance_stats_.total_requests.load());
    stats.successful_requests.store(performance_stats_.successful_requests.load());
    stats.avg_response_time.store(performance_stats_.avg_response_time.load());
    stats.max_response_time.store(performance_stats_.max_response_time.load());
    stats.memory_usage.store(performance_stats_.memory_usage.load());
    stats.last_reset_time = performance_stats_.last_reset_time;
=======
PerformanceStats VolunteerMatcher::GetPerformanceStats() const {
    std::lock_guard<std::mutex> lock(stats_mutex_);
    return performance_stats_;
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
}

void VolunteerMatcher::ResetPerformanceStats() {
    std::lock_guard<std::mutex> lock(stats_mutex_);
<<<<<<< HEAD
    performance_stats_.total_requests.store(0);
    performance_stats_.successful_requests.store(0);
    performance_stats_.avg_response_time.store(0.0);
    performance_stats_.max_response_time.store(0.0);
    performance_stats_.memory_usage.store(0);
=======
    performance_stats_ = PerformanceStats{};
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
    performance_stats_.last_reset_time = std::chrono::system_clock::now();
}

void VolunteerMatcher::SetLogLevel(const std::string& level) {
    // 实现日志级别设置
}

bool VolunteerMatcher::HotUpdateData(const std::string& data_type, const std::string& file_path) {
    if (data_type == "universities") {
        return LoadUniversities(file_path) > 0;
    } else if (data_type == "majors") {
        return LoadMajors(file_path) > 0;
    } else if (data_type == "historical") {
        return LoadHistoricalData(file_path) > 0;
    }
    return false;
}

std::string VolunteerMatcher::GetEngineStatus() const {
    Json::Value status;
    status["initialized"] = is_initialized_.load();
    status["universities_count"] = static_cast<int>(pimpl_->universities_.size());
    status["majors_count"] = static_cast<int>(pimpl_->majors_.size());
    status["cache_size"] = static_cast<int>(pimpl_->plan_cache_.size());
    
<<<<<<< HEAD
    PerformanceStats stats;
    GetPerformanceStats(stats);
=======
    auto stats = GetPerformanceStats();
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
    status["performance"]["total_requests"] = static_cast<int>(stats.total_requests.load());
    status["performance"]["successful_requests"] = static_cast<int>(stats.successful_requests.load());
    status["performance"]["avg_response_time"] = stats.avg_response_time.load();
    status["performance"]["memory_usage"] = static_cast<int>(stats.memory_usage.load());
    
    Json::StreamWriterBuilder builder;
    return Json::writeString(builder, status);
}

// 私有辅助方法实现

std::chrono::high_resolution_clock::time_point VolunteerMatcher::StartRequest() {
    performance_stats_.total_requests++;
    return std::chrono::high_resolution_clock::now();
}

void VolunteerMatcher::EndRequest(std::chrono::high_resolution_clock::time_point start_time, bool success) {
    auto end_time = std::chrono::high_resolution_clock::now();
    auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(end_time - start_time);
    double response_time = static_cast<double>(duration.count());
    
    if (success) {
        performance_stats_.successful_requests++;
    }
    
    // 更新平均响应时间
    double current_avg = performance_stats_.avg_response_time.load();
    double total_requests = static_cast<double>(performance_stats_.total_requests.load());
    double new_avg = (current_avg * (total_requests - 1) + response_time) / total_requests;
    performance_stats_.avg_response_time = new_avg;
    
    // 更新最大响应时间
    double current_max = performance_stats_.max_response_time.load();
    if (response_time > current_max) {
        performance_stats_.max_response_time = response_time;
    }
}

// 辅助函数实现

double CalculateMatchScore(const Student& student, const University& university, const Major& major) {
    double score = 0.0;
    
    // 分数匹配度 (30%)
    double score_match = 1.0 - std::abs(student.total_score - university.historical_scores.back()) / 100.0;
    score += score_match * 30.0;
    
    // 地理偏好匹配度 (20%)
    double location_match = 0.5; // 默认中等匹配
    for (const auto& city : student.preferred_cities) {
        if (city == university.city) {
            location_match = 1.0;
            break;
        }
    }
    score += location_match * 20.0;
    
    // 专业偏好匹配度 (30%)
    double major_match = 0.5; // 默认中等匹配
    for (const auto& preferred : student.preferred_majors) {
        if (preferred == major.name) {
            major_match = 1.0;
            break;
        }
    }
    score += major_match * 30.0;
    
    // 学校排名匹配度 (20%)
    double ranking_match = std::max(0.0, 1.0 - university.ranking / 1000.0);
    score += ranking_match * 20.0;
    
    return std::min(100.0, std::max(0.0, score));
}

std::vector<std::string> ParseSubjectCombination(const std::string& combination) {
    std::vector<std::string> subjects;
    std::stringstream ss(combination);
    std::string subject;
    
    while (std::getline(ss, subject, '+')) {
        // 去除空格
        subject.erase(std::remove_if(subject.begin(), subject.end(), ::isspace), subject.end());
        if (!subject.empty()) {
            subjects.push_back(subject);
        }
    }
    
    return subjects;
}

bool ValidateSubjectRequirements(
    const std::vector<std::string>& student_subjects,
    const std::string& major_requirements) {
    
    if (major_requirements.empty() || major_requirements == "不限") {
        return true;
    }
    
    auto required_subjects = ParseSubjectCombination(major_requirements);
    
    for (const auto& required : required_subjects) {
        bool found = false;
        for (const auto& student_subject : student_subjects) {
            if (student_subject == required) {
                found = true;
                break;
            }
        }
        if (!found) {
            return false;
        }
    }
    
    return true;
}

double CalculateTrendCoefficient(const std::vector<int>& historical_data) {
    if (historical_data.size() < 2) {
        return 1.0;
    }
    
    double sum_x = 0.0, sum_y = 0.0, sum_xy = 0.0, sum_x2 = 0.0;
    int n = static_cast<int>(historical_data.size());
    
    for (int i = 0; i < n; ++i) {
        double x = static_cast<double>(i);
        double y = static_cast<double>(historical_data[i]);
        
        sum_x += x;
        sum_y += y;
        sum_xy += x * y;
        sum_x2 += x * x;
    }
    
    double slope = (n * sum_xy - sum_x * sum_y) / (n * sum_x2 - sum_x * sum_x);
    return 1.0 + slope / 100.0; // 趋势系数
}

// 新增：缺失的核心算法实现

/**
 * @brief 从CSV行解析大学信息
 */
University ParseUniversityFromCSV(const std::string& csv_line) {
    University university{};
    
    try {
        std::vector<std::string> fields;
        std::stringstream ss(csv_line);
        std::string field;
        
        // 解析CSV字段
        while (std::getline(ss, field, ',')) {
            // 去除引号和空格
            field.erase(std::remove(field.begin(), field.end(), '\"'), field.end());
            field.erase(std::remove_if(field.begin(), field.end(), ::isspace), field.end());
            fields.push_back(field);
        }
        
        if (fields.size() >= 10) {
            university.university_id = fields[0];
            university.name = fields[1];
            university.province = fields[2];
            university.city = fields[3];
            university.level = fields[4];
            
            if (!fields[5].empty()) {
                university.ranking = std::stoi(fields[5]);
            }
            
            if (!fields[6].empty()) {
                university.total_enrollment = std::stoi(fields[6]);
            }
            
            if (!fields[7].empty()) {
                university.employment_rate = std::stod(fields[7]);
            }
            
            if (!fields[8].empty()) {
                university.graduate_salary = std::stod(fields[8]);
            }
            
            // 解析历年分数线 (分号分隔)
            if (fields.size() > 9 && !fields[9].empty()) {
                std::stringstream scores_ss(fields[9]);
                std::string score_str;
                while (std::getline(scores_ss, score_str, ';')) {
                    if (!score_str.empty()) {
                        university.historical_scores.push_back(std::stoi(score_str));
                    }
                }
            }
            
            // 解析历年排名 (分号分隔)
            if (fields.size() > 10 && !fields[10].empty()) {
                std::stringstream rankings_ss(fields[10]);
                std::string ranking_str;
                while (std::getline(rankings_ss, ranking_str, ';')) {
                    if (!ranking_str.empty()) {
                        university.historical_rankings.push_back(std::stoi(ranking_str));
                    }
                }
            }
            
            // 解析优势专业 (分号分隔)
            if (fields.size() > 11 && !fields[11].empty()) {
                std::stringstream majors_ss(fields[11]);
                std::string major_str;
                while (std::getline(majors_ss, major_str, ';')) {
                    if (!major_str.empty()) {
                        university.strong_majors.push_back(major_str);
                    }
                }
            }
        }
    } catch (const std::exception& e) {
        // 解析失败，返回空对象
        university = University{};
    }
    
    return university;
}

/**
 * @brief 从CSV行解析专业信息
 */
Major ParseMajorFromCSV(const std::string& csv_line) {
    Major major{};
    
    try {
        std::vector<std::string> fields;
        std::stringstream ss(csv_line);
        std::string field;
        
        // 解析CSV字段
        while (std::getline(ss, field, ',')) {
            // 去除引号和空格
            field.erase(std::remove(field.begin(), field.end(), '\"'), field.end());
            field.erase(std::remove_if(field.begin(), field.end(), ::isspace), field.end());
            fields.push_back(field);
        }
        
        if (fields.size() >= 8) {
            major.major_id = fields[0];
            major.name = fields[1];
            major.category = fields[2];
            major.subject_requirements = fields[3];
            
            if (!fields[4].empty()) {
                major.employment_rate = std::stod(fields[4]);
            }
            
            if (!fields[5].empty()) {
                major.salary_level = std::stod(fields[5]);
            }
            
            if (!fields[6].empty()) {
                major.difficulty_level = std::stod(fields[6]);
            }
            
            if (!fields[7].empty()) {
                major.requires_postgraduate = (fields[7] == "true" || fields[7] == "1");
            }
            
            // 解析职业方向 (分号分隔)
            if (fields.size() > 8 && !fields[8].empty()) {
                std::stringstream careers_ss(fields[8]);
                std::string career_str;
                while (std::getline(careers_ss, career_str, ';')) {
                    if (!career_str.empty()) {
                        major.career_directions.push_back(career_str);
                    }
                }
            }
        }
    } catch (const std::exception& e) {
        // 解析失败，返回空对象
        major = Major{};
    }
    
    return major;
}

/**
 * @brief 构建筛选条件
 */
<<<<<<< HEAD
FilterCriteria VolunteerMatcher::BuildFilterCriteria(const Student& student) {
=======
FilterCriteria BuildFilterCriteria(const Student& student) {
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
    FilterCriteria criteria;
    
    // 分数范围：基于学生分数±50分
    criteria.min_score = std::max(0, student.total_score - 50);
    criteria.max_score = std::min(750, student.total_score + 50);
    criteria.score_tolerance = 30;
    criteria.use_ranking = true;
    
    // 地理偏好
    criteria.preferred_cities = student.preferred_cities;
    criteria.prefer_hometown = true;
    
    // 专业偏好
    criteria.target_majors = student.preferred_majors;
    criteria.strict_major_match = false;
    
    // 就业要求
    criteria.min_employment_rate = 0.85;
    criteria.min_salary_level = 5000.0;
    criteria.consider_career_prospects = true;
    
    // 学校层次：根据学生分数确定
    if (student.total_score >= 650) {
        criteria.school_levels = {"985", "211", "双一流"};
        criteria.max_ranking = 100;
    } else if (student.total_score >= 600) {
        criteria.school_levels = {"211", "双一流", "重点本科"};
        criteria.max_ranking = 300;
    } else if (student.total_score >= 550) {
        criteria.school_levels = {"双一流", "重点本科", "普通本科"};
        criteria.max_ranking = 600;
    } else {
        criteria.school_levels = {"普通本科"};
        criteria.max_ranking = 1000;
    }
    
    // 招生计划
    criteria.min_enrollment = 5;  // 至少招5人以上的专业
    criteria.consider_enrollment_trend = true;
    
    // 特殊要求
    criteria.only_public_universities = false;
    criteria.allow_independent_colleges = student.total_score < 500;
    criteria.allow_joint_programs = true;
    
    return criteria;
}

/**
 * @brief 为特定大学生成推荐
 */
std::vector<VolunteerRecommendation> GenerateRecommendationsForUniversity(
    const Student& student, const University& university) {
    
    std::vector<VolunteerRecommendation> recommendations;
    
    // 为该大学的每个适合专业生成推荐
    for (const auto& strong_major : university.strong_majors) {
        // 检查选科要求
        auto student_subjects = ParseSubjectCombination(student.subject_combination);
        
        VolunteerRecommendation rec;
        rec.university_id = university.university_id;
        rec.university_name = university.name;
        rec.major_name = strong_major;
        rec.major_id = "major_" + strong_major;  // 简化处理
        
        // 计算分数差距
        if (!university.historical_scores.empty()) {
            rec.score_gap = student.total_score - university.historical_scores.back();
        }
        
        // 计算排名差距
        if (!university.historical_rankings.empty() && student.ranking > 0) {
            rec.ranking_gap = student.ranking - university.historical_rankings.back();
        }
        
        recommendations.push_back(rec);
    }
    
    // 如果没有强势专业信息，至少生成一个通用推荐
    if (recommendations.empty()) {
        VolunteerRecommendation rec;
        rec.university_id = university.university_id;
        rec.university_name = university.name;
        rec.major_name = "通用专业";
        rec.major_id = "general";
        
        if (!university.historical_scores.empty()) {
            rec.score_gap = student.total_score - university.historical_scores.back();
        }
        
        recommendations.push_back(rec);
    }
    
    return recommendations;
}

/**
 * @brief 确定风险等级
 */
std::string DetermineRiskLevel(double risk_score) {
    if (risk_score <= 0.3) {
        return "保";
    } else if (risk_score <= 0.6) {
        return "稳";
    } else {
        return "冲";
    }
}

/**
 * @brief 生成推荐理由
 */
std::string GenerateRecommendationReason(
    const Student& student, const University& university, const Major& major) {
    
    std::ostringstream reason;
    
    // 分数匹配分析
    if (!university.historical_scores.empty()) {
        int score_diff = student.total_score - university.historical_scores.back();
        if (score_diff >= 20) {
            reason << "分数有优势，录取概率高；";
        } else if (score_diff >= 0) {
            reason << "分数基本匹配，有录取机会；";
        } else {
            reason << "分数略低，需谨慎考虑；";
        }
    }
    
    // 地理位置分析
<<<<<<< HEAD
    bool location_match __attribute__((unused)) = false;
=======
    bool location_match = false;
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
    for (const auto& preferred_city : student.preferred_cities) {
        if (preferred_city == university.city) {
            reason << "符合地理位置偏好(" << preferred_city << ")；";
            location_match = true;
            break;
        }
    }
    
    // 专业匹配分析
<<<<<<< HEAD
    bool major_match __attribute__((unused)) = false;
=======
    bool major_match = false;
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
    for (const auto& preferred_major : student.preferred_majors) {
        if (preferred_major == major.name) {
            reason << "专业完全匹配(" << preferred_major << ")；";
            major_match = true;
            break;
        }
    }
    
    // 学校层次分析
    if (university.level == "985" || university.level == "211") {
        reason << "名校背景，就业优势明显；";
    }
    
    // 就业前景分析
    if (university.employment_rate > 0.9) {
        reason << "就业率优秀(" << std::fixed << std::setprecision(1) 
               << university.employment_rate * 100 << "%)；";
    }
    
    if (university.graduate_salary > 8000) {
        reason << "薪资水平较高(" << static_cast<int>(university.graduate_salary) << "元)；";
    }
    
    std::string result = reason.str();
    if (result.empty()) {
        result = "综合条件较为匹配，建议关注。";
    } else {
        // 去除最后的分号
<<<<<<< HEAD
        if (!result.empty() && result.back() == static_cast<char>(0xEF)) {
            // 处理UTF-8编码的中文分号
            if (result.size() >= 3 && 
                result.substr(result.size()-3) == "；") {
                result.erase(result.size()-3);
            }
        } else if (!result.empty() && result.back() == ';') {
=======
        if (result.back() == '；') {
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
            result.pop_back();
        }
        result += "。";
    }
    
    return result;
}

/**
 * @brief 应用冲稳保策略
 */
std::vector<VolunteerRecommendation> ApplyRushStableSafeStrategy(
    const std::vector<VolunteerRecommendation>& candidates, int max_volunteers) {
    
    if (candidates.empty() || max_volunteers <= 0) {
        return {};
    }
    
    // 按匹配度排序
    std::vector<VolunteerRecommendation> sorted_candidates = candidates;
    std::sort(sorted_candidates.begin(), sorted_candidates.end(),
              [](const VolunteerRecommendation& a, const VolunteerRecommendation& b) {
                  return a.match_score > b.match_score;
              });
    
    std::vector<VolunteerRecommendation> final_recommendations;
    
    // 冲稳保比例：30% : 50% : 20%
    int rush_count = std::max(1, static_cast<int>(max_volunteers * 0.3));
    int stable_count = std::max(1, static_cast<int>(max_volunteers * 0.5));
    int safe_count = max_volunteers - rush_count - stable_count;
    
    int current_count = 0;
    
    // 添加冲刺志愿（高分段院校）
    for (const auto& candidate : sorted_candidates) {
        if (current_count >= rush_count) break;
        if (candidate.score_gap < -10) {  // 分数低于录取线10分以上的作为冲刺
            VolunteerRecommendation rec = candidate;
            rec.risk_level = "冲";
            final_recommendations.push_back(rec);
            current_count++;
        }
    }
    
    // 添加稳妥志愿（匹配度高的院校）
    current_count = 0;
    for (const auto& candidate : sorted_candidates) {
        if (current_count >= stable_count) break;
        if (candidate.score_gap >= -10 && candidate.score_gap <= 10) {
            VolunteerRecommendation rec = candidate;
            rec.risk_level = "稳";
            final_recommendations.push_back(rec);
            current_count++;
        }
    }
    
    // 添加保底志愿（录取概率高的院校）
    current_count = 0;
    for (const auto& candidate : sorted_candidates) {
        if (current_count >= safe_count) break;
        if (candidate.score_gap > 10) {  // 分数高于录取线10分以上的作为保底
            VolunteerRecommendation rec = candidate;
            rec.risk_level = "保";
            final_recommendations.push_back(rec);
            current_count++;
        }
    }
    
    // 如果志愿数不够，从剩余候选中补充
    while (final_recommendations.size() < static_cast<size_t>(max_volunteers) && 
           final_recommendations.size() < sorted_candidates.size()) {
        for (const auto& candidate : sorted_candidates) {
            if (final_recommendations.size() >= static_cast<size_t>(max_volunteers)) break;
            
            // 检查是否已存在
            bool exists = false;
            for (const auto& existing : final_recommendations) {
                if (existing.university_id == candidate.university_id && 
                    existing.major_id == candidate.major_id) {
                    exists = true;
                    break;
                }
            }
            
            if (!exists) {
                VolunteerRecommendation rec = candidate;
                if (rec.risk_level.empty()) {
                    rec.risk_level = "稳";  // 默认为稳妥
                }
                final_recommendations.push_back(rec);
            }
        }
        break;
    }
    
    return final_recommendations;
}

/**
 * @brief 计算方案统计
 */
void CalculatePlanStatistics(VolunteerPlan& plan) {
    if (plan.recommendations.empty()) {
        return;
    }
    
    plan.total_volunteers = static_cast<int>(plan.recommendations.size());
    plan.rush_count = 0;
    plan.stable_count = 0;
    plan.safe_count = 0;
    
    double total_risk = 0.0;
    double total_match_score = 0.0;
    
    for (const auto& rec : plan.recommendations) {
        // 统计风险等级
        if (rec.risk_level == "冲") {
            plan.rush_count++;
        } else if (rec.risk_level == "稳") {
            plan.stable_count++;
        } else if (rec.risk_level == "保") {
            plan.safe_count++;
        }
        
        // 计算整体风险和匹配度
        total_risk += (1.0 - rec.admission_probability);
        total_match_score += rec.match_score;
    }
    
    // 计算整体风险评分
    plan.overall_risk_score = total_risk / plan.total_volunteers;
    
    // 确定方案质量等级
    double avg_match_score = total_match_score / plan.total_volunteers;
    if (avg_match_score >= 80.0) {
        plan.plan_quality = "优秀";
    } else if (avg_match_score >= 70.0) {
        plan.plan_quality = "良好";
    } else if (avg_match_score >= 60.0) {
        plan.plan_quality = "中等";
    } else {
        plan.plan_quality = "需优化";
    }
}

/**
 * @brief 生成优化建议
 */
std::vector<std::string> GenerateOptimizationSuggestions(const VolunteerPlan& plan) {
    std::vector<std::string> suggestions;
    
    if (plan.recommendations.empty()) {
        suggestions.push_back("无有效志愿推荐，请调整筛选条件");
        return suggestions;
    }
    
    // 分析冲稳保比例
    double rush_ratio = static_cast<double>(plan.rush_count) / plan.total_volunteers;
    double stable_ratio = static_cast<double>(plan.stable_count) / plan.total_volunteers;
    double safe_ratio = static_cast<double>(plan.safe_count) / plan.total_volunteers;
    
    if (rush_ratio > 0.4) {
        suggestions.push_back("冲刺志愿比例过高，建议增加稳妥志愿以提高录取保障");
    }
    
    if (safe_ratio < 0.15) {
        suggestions.push_back("保底志愿不足，建议增加录取概率较高的院校");
    }
    
    if (stable_ratio < 0.4) {
        suggestions.push_back("稳妥志愿较少，建议增加匹配度适中的院校");
    }
    
    // 分析整体风险
    if (plan.overall_risk_score > 0.7) {
        suggestions.push_back("整体方案风险偏高，建议降低预期或增加保底志愿");
    } else if (plan.overall_risk_score < 0.3) {
        suggestions.push_back("方案过于保守，可适当增加冲刺志愿以争取更好机会");
    }
    
    // 分析地理分布
    std::unordered_set<std::string> cities;
    std::unordered_set<std::string> provinces;
    for (const auto& rec : plan.recommendations) {
<<<<<<< HEAD
        (void)rec; // 避免未使用变量警告
=======
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
        // 这里需要从大学信息中获取城市和省份信息
        // 简化处理，假设已有映射
    }
    
    if (cities.size() == 1) {
        suggestions.push_back("志愿地理位置过于集中，建议考虑其他城市以增加选择面");
    }
    
    // 分析专业多样性
    std::unordered_set<std::string> major_categories;
    for (const auto& rec : plan.recommendations) {
        // 从专业名称推断类别（简化处理）
        if (rec.major_name.find("工") != std::string::npos) {
            major_categories.insert("工学");
        } else if (rec.major_name.find("理") != std::string::npos) {
            major_categories.insert("理学");
        } else if (rec.major_name.find("经济") != std::string::npos || rec.major_name.find("管理") != std::string::npos) {
            major_categories.insert("经管");
        }
        // 可以继续扩展其他类别
    }
    
    if (major_categories.size() == 1) {
        suggestions.push_back("专业选择相对单一，建议考虑相关交叉专业以增加录取机会");
    }
    
    // 如果没有发现问题，给出积极建议
    if (suggestions.empty()) {
        suggestions.push_back("志愿方案整体合理，建议关注各院校最新招生政策");
        suggestions.push_back("可根据个人兴趣和职业规划进一步调整专业选择");
    }
    
    return suggestions;
}

/**
 * @brief 优化策略函数
 */
void OptimizeForSafety(VolunteerPlan& plan) {
    // 增加保底志愿比例，降低整体风险
    for (auto& rec : plan.recommendations) {
        if (rec.risk_level == "冲") {
            // 将部分冲刺志愿调整为稳妥志愿
            if (rec.admission_probability < 0.3) {
                rec.risk_level = "稳";
                rec.admission_probability = std::min(0.8, rec.admission_probability + 0.3);
            }
        }
    }
    
    // 重新计算统计信息
    CalculatePlanStatistics(plan);
}

void OptimizeForProbability(VolunteerPlan& plan) {
    // 基于录取概率重新排序
    std::sort(plan.recommendations.begin(), plan.recommendations.end(),
              [](const VolunteerRecommendation& a, const VolunteerRecommendation& b) {
                  return a.admission_probability > b.admission_probability;
              });
    
    // 重新计算统计信息
    CalculatePlanStatistics(plan);
}

void OptimizeForPreference(VolunteerPlan& plan) {
    // 基于匹配度重新排序
    std::sort(plan.recommendations.begin(), plan.recommendations.end(),
              [](const VolunteerRecommendation& a, const VolunteerRecommendation& b) {
                  return a.match_score > b.match_score;
              });
    
    // 重新计算统计信息
    CalculatePlanStatistics(plan);
}

} // namespace volunteer_matcher