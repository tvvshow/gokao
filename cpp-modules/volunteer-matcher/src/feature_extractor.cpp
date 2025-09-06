/**
 * @file feature_extractor.cpp
 * @brief 高考志愿填报系统 - AI特征工程模块实现
 * @author 高考志愿填报系统开发团队
 * @version 1.0.0
 * @date 2025-01-18
 */

#include "volunteer_matcher.h"
#include <algorithm>
#include <cmath>
#include <numeric>
#include <unordered_set>
#include <unordered_map>
#include <sstream>
#include <fstream>
#include <json/json.h>
#include <functional>
#include <shared_mutex>

namespace volunteer_matcher {

/**
 * @brief 特征提取器内部实现类
 */
class FeatureExtractor::FeatureExtractorImpl {
public:
    // 特征缓存
    std::unordered_map<std::string, FeatureVector> feature_cache_;
    mutable std::shared_mutex cache_mutex_;
    
    // 特征统计信息
    struct FeatureStats {
        std::unordered_map<std::string, double> mean_values;
        std::unordered_map<std::string, double> std_values;
        std::unordered_map<std::string, double> min_values;
        std::unordered_map<std::string, double> max_values;
        int total_features_extracted = 0;
        std::chrono::system_clock::time_point last_update;
    };
    
    FeatureStats stats_;
    mutable std::mutex stats_mutex_;
    
    // 配置参数
    Json::Value config_;
    bool normalized_features_ = true;
    double cache_expire_hours_ = 24.0;
    
    FeatureExtractorImpl() {
        stats_.last_update = std::chrono::system_clock::now();
    }
    
    /**
     * @brief 生成缓存键
     */
    std::string GenerateCacheKey(const std::string& type, const std::string& id) {
        return type + "_" + id;
    }
    
    /**
     * @brief 检查缓存是否有效
     */
    bool IsCacheValid(const FeatureVector& cached_features) {
        auto now = std::chrono::system_clock::now();
        auto duration = std::chrono::duration_cast<std::chrono::hours>(
            now - cached_features.created_time);
        return duration.count() < cache_expire_hours_;
    }
    
    /**
     * @brief 计算特征哈希
     */
    double ComputeFeatureHash(const FeatureVector& features) {
        std::hash<std::string> hasher;
        std::string feature_str;
        
        for (double val : features.numerical_features) {
            feature_str += std::to_string(val) + ",";
        }
        for (int val : features.categorical_features) {
            feature_str += std::to_string(val) + ",";
        }
        
        return static_cast<double>(hasher(feature_str));
    }
    
    /**
     * @brief 更新特征统计
     */
    void UpdateFeatureStats(const FeatureVector& features) {
        std::lock_guard<std::mutex> lock(stats_mutex_);
        
        stats_.total_features_extracted++;
        stats_.last_update = std::chrono::system_clock::now();
        
        // 更新数值特征统计
        for (size_t i = 0; i < features.numerical_features.size() && 
                         i < features.feature_names.size(); ++i) {
            const std::string& name = features.feature_names[i];
            double value = features.numerical_features[i];
            
            // 计算均值和方差的增量更新
            if (stats_.mean_values.find(name) == stats_.mean_values.end()) {
                stats_.mean_values[name] = value;
                stats_.std_values[name] = 0.0;
                stats_.min_values[name] = value;
                stats_.max_values[name] = value;
            } else {
                // 增量均值计算
                double old_mean = stats_.mean_values[name];
                double new_mean = old_mean + (value - old_mean) / stats_.total_features_extracted;
                stats_.mean_values[name] = new_mean;
                
                // 更新最值
                stats_.min_values[name] = std::min(stats_.min_values[name], value);
                stats_.max_values[name] = std::max(stats_.max_values[name], value);
            }
        }
    }
};

// FeatureExtractor实现

FeatureExtractor::FeatureExtractor() 
    : pimpl_(std::make_unique<FeatureExtractorImpl>()) {}

FeatureExtractor::~FeatureExtractor() = default;

bool FeatureExtractor::Initialize(const std::string& config_path) {
    try {
        std::ifstream config_file(config_path);
        if (!config_file.is_open()) {
            // 使用默认配置
            Json::Value default_config;
            default_config["feature_extraction"]["normalized"] = true;
            default_config["feature_extraction"]["cache_expire_hours"] = 24.0;
            default_config["feature_extraction"]["max_cache_size"] = 10000;
            pimpl_->config_ = default_config;
            return true;
        }
        
        Json::Reader reader;
        if (!reader.parse(config_file, pimpl_->config_)) {
            return false;
        }
        
        // 读取配置参数
        if (pimpl_->config_.isMember("feature_extraction")) {
            const Json::Value& fe_config = pimpl_->config_["feature_extraction"];
            pimpl_->normalized_features_ = fe_config.get("normalized", true).asBool();
            pimpl_->cache_expire_hours_ = fe_config.get("cache_expire_hours", 24.0).asDouble();
        }
        
        return true;
    } catch (const std::exception& e) {
        return false;
    }
}

FeatureVector FeatureExtractor::ExtractStudentFeatures(const Student& student) {
    // 检查缓存
    std::string cache_key = pimpl_->GenerateCacheKey("student", student.student_id);
    
    {
        std::shared_lock<std::shared_mutex> lock(pimpl_->cache_mutex_);
        auto it = pimpl_->feature_cache_.find(cache_key);
        if (it != pimpl_->feature_cache_.end() && pimpl_->IsCacheValid(it->second)) {
            return it->second;
        }
    }
    
    FeatureVector features;
    features.student_id = student.student_id;
    features.created_time = std::chrono::system_clock::now();
    
    // 1. 基础数值特征
    features.numerical_features.clear();
    features.feature_names.clear();
    
    // 分数相关特征
    features.numerical_features.push_back(static_cast<double>(student.total_score));
    features.feature_names.push_back("total_score");
    
    features.numerical_features.push_back(static_cast<double>(student.ranking));
    features.feature_names.push_back("ranking");
    
    // 分数段特征（标准化分数）
    double normalized_score = student.total_score / 750.0; // 假设满分750
    features.numerical_features.push_back(normalized_score);
    features.feature_names.push_back("normalized_score");
    
    // 排名百分位特征（假设总考生100万）
    double ranking_percentile = 1.0 - (static_cast<double>(student.ranking) / 1000000.0);
    features.numerical_features.push_back(ranking_percentile);
    features.feature_names.push_back("ranking_percentile");
    
    // 单科成绩特征
    std::vector<int> subject_scores = {
        student.chinese_score, student.math_score, student.english_score,
        student.physics_score, student.chemistry_score, student.biology_score,
        student.politics_score, student.history_score, student.geography_score
    };
    
    std::vector<std::string> subject_names = {
        "chinese", "math", "english", "physics", "chemistry",
        "biology", "politics", "history", "geography"
    };
    
    for (size_t i = 0; i < subject_scores.size(); ++i) {
        if (subject_scores[i] > 0) { // 只添加有效成绩
            features.numerical_features.push_back(static_cast<double>(subject_scores[i]));
            features.feature_names.push_back(subject_names[i] + "_score");
        }
    }
    
    // 偏好权重特征
    features.numerical_features.push_back(student.city_weight);
    features.feature_names.push_back("city_weight");
    
    features.numerical_features.push_back(student.major_weight);
    features.feature_names.push_back("major_weight");
    
    features.numerical_features.push_back(student.school_ranking_weight);
    features.feature_names.push_back("school_ranking_weight");
    
    // 偏好数量特征
    features.numerical_features.push_back(static_cast<double>(student.preferred_cities.size()));
    features.feature_names.push_back("preferred_cities_count");
    
    features.numerical_features.push_back(static_cast<double>(student.preferred_majors.size()));
    features.feature_names.push_back("preferred_majors_count");
    
    features.numerical_features.push_back(static_cast<double>(student.avoided_majors.size()));
    features.feature_names.push_back("avoided_majors_count");
    
    // 2. 类别特征
    features.categorical_features.clear();
    
    // 省份编码（假设用省份名哈希）
    std::hash<std::string> hasher;
    int province_code = static_cast<int>(hasher(student.province) % 1000);
    features.categorical_features.push_back(province_code);
    
    // 选科组合编码
    int subject_combo_code = static_cast<int>(hasher(student.subject_combination) % 100);
    features.categorical_features.push_back(subject_combo_code);
    
    // 特殊情况编码
    features.categorical_features.push_back(student.is_minority ? 1 : 0);
    features.categorical_features.push_back(student.has_sports_specialty ? 1 : 0);
    features.categorical_features.push_back(student.has_art_specialty ? 1 : 0);
    
    // 3. 派生特征
    // 成绩稳定性（各科成绩标准差）
    std::vector<double> valid_scores;
    for (int score : subject_scores) {
        if (score > 0) valid_scores.push_back(static_cast<double>(score));
    }
    
    if (valid_scores.size() > 1) {
        double mean = std::accumulate(valid_scores.begin(), valid_scores.end(), 0.0) / valid_scores.size();
        double variance = 0.0;
        for (double score : valid_scores) {
            variance += (score - mean) * (score - mean);
        }
        double std_dev = std::sqrt(variance / valid_scores.size());
        features.numerical_features.push_back(std_dev);
        features.feature_names.push_back("score_stability");
    }
    
    // 优势科目特征（最高成绩与平均成绩的差值）
    if (!valid_scores.empty()) {
        double max_score = *std::max_element(valid_scores.begin(), valid_scores.end());
        double avg_score = std::accumulate(valid_scores.begin(), valid_scores.end(), 0.0) / valid_scores.size();
        features.numerical_features.push_back(max_score - avg_score);
        features.feature_names.push_back("advantage_subject_gap");
    }
    
    // 计算特征哈希
    features.feature_hash = pimpl_->ComputeFeatureHash(features);
    
    // 更新统计信息
    pimpl_->UpdateFeatureStats(features);
    
    // 缓存结果
    {
        std::unique_lock<std::shared_mutex> lock(pimpl_->cache_mutex_);
        pimpl_->feature_cache_[cache_key] = features;
    }
    
    return features;
}

FeatureVector FeatureExtractor::ExtractUniversityFeatures(const University& university) {
    // 检查缓存
    std::string cache_key = pimpl_->GenerateCacheKey("university", university.university_id);
    
    {
        std::shared_lock<std::shared_mutex> lock(pimpl_->cache_mutex_);
        auto it = pimpl_->feature_cache_.find(cache_key);
        if (it != pimpl_->feature_cache_.end() && pimpl_->IsCacheValid(it->second)) {
            return it->second;
        }
    }
    
    FeatureVector features;
    features.university_id = university.university_id;
    features.created_time = std::chrono::system_clock::now();
    
    // 1. 基础数值特征
    features.numerical_features.clear();
    features.feature_names.clear();
    
    // 学校排名特征
    features.numerical_features.push_back(static_cast<double>(university.ranking));
    features.feature_names.push_back("university_ranking");
    
    // 排名归一化（假设最大排名1000）
    double normalized_ranking = 1.0 - (static_cast<double>(university.ranking) / 1000.0);
    features.numerical_features.push_back(normalized_ranking);
    features.feature_names.push_back("normalized_ranking");
    
    // 招生规模特征
    features.numerical_features.push_back(static_cast<double>(university.total_enrollment));
    features.feature_names.push_back("total_enrollment");
    
    // 就业质量特征
    features.numerical_features.push_back(university.employment_rate);
    features.feature_names.push_back("employment_rate");
    
    features.numerical_features.push_back(university.graduate_salary);
    features.feature_names.push_back("graduate_salary");
    
    // 薪资归一化（假设最高薪资50万）
    double normalized_salary = university.graduate_salary / 500000.0;
    features.numerical_features.push_back(normalized_salary);
    features.feature_names.push_back("normalized_salary");
    
    // 历史分数线特征
    if (!university.historical_scores.empty()) {
        // 最新分数线
        features.numerical_features.push_back(static_cast<double>(university.historical_scores.back()));
        features.feature_names.push_back("latest_score");
        
        // 历史分数均值
        double avg_score = std::accumulate(university.historical_scores.begin(), 
                                         university.historical_scores.end(), 0.0) / 
                          university.historical_scores.size();
        features.numerical_features.push_back(avg_score);
        features.feature_names.push_back("avg_historical_score");
        
        // 分数线趋势（线性回归斜率）
        double trend = CalculateTrendCoefficient(university.historical_scores);
        features.numerical_features.push_back(trend);
        features.feature_names.push_back("score_trend");
        
        // 分数线稳定性
        if (university.historical_scores.size() > 1) {
            double variance = 0.0;
            for (int score : university.historical_scores) {
                variance += (score - avg_score) * (score - avg_score);
            }
            double stability = std::sqrt(variance / university.historical_scores.size());
            features.numerical_features.push_back(stability);
            features.feature_names.push_back("score_stability");
        }
    }
    
    // 专业数量特征
    features.numerical_features.push_back(static_cast<double>(university.strong_majors.size()));
    features.feature_names.push_back("strong_majors_count");
    
    // 2. 类别特征
    features.categorical_features.clear();
    
    // 省份编码
    std::hash<std::string> hasher;
    int province_code = static_cast<int>(hasher(university.province) % 1000);
    features.categorical_features.push_back(province_code);
    
    // 城市编码
    int city_code = static_cast<int>(hasher(university.city) % 1000);
    features.categorical_features.push_back(city_code);
    
    // 学校层次编码
    int level_code = 0;
    if (university.level == "985") level_code = 4;
    else if (university.level == "211") level_code = 3;
    else if (university.level == "双一流") level_code = 2;
    else if (university.level == "重点本科") level_code = 1;
    else level_code = 0;
    features.categorical_features.push_back(level_code);
    
    // 3. 竞争度特征
    // 录取难度（基于历史排名）
    if (!university.historical_rankings.empty()) {
        double avg_ranking = std::accumulate(university.historical_rankings.begin(),
                                           university.historical_rankings.end(), 0.0) /
                            university.historical_rankings.size();
        double difficulty = 1.0 / (avg_ranking + 1.0); // 排名越高难度越大
        features.numerical_features.push_back(difficulty);
        features.feature_names.push_back("admission_difficulty");
    }
    
    // 计算特征哈希
    features.feature_hash = pimpl_->ComputeFeatureHash(features);
    
    // 更新统计信息
    pimpl_->UpdateFeatureStats(features);
    
    // 缓存结果
    {
        std::unique_lock<std::shared_mutex> lock(pimpl_->cache_mutex_);
        pimpl_->feature_cache_[cache_key] = features;
    }
    
    return features;
}

FeatureVector FeatureExtractor::ExtractInteractionFeatures(
    const Student& student, const University& university) {
    
    // 交互特征不缓存，每次实时计算
    FeatureVector features;
    features.student_id = student.student_id;
    features.university_id = university.university_id;
    features.created_time = std::chrono::system_clock::now();
    
    features.numerical_features.clear();
    features.feature_names.clear();
    features.categorical_features.clear();
    
    // 1. 分数匹配特征
    if (!university.historical_scores.empty()) {
        int score_gap = student.total_score - university.historical_scores.back();
        features.numerical_features.push_back(static_cast<double>(score_gap));
        features.feature_names.push_back("score_gap");
        
        // 分数匹配度
        double score_match = 1.0 - std::abs(score_gap) / 100.0;
        score_match = std::max(0.0, std::min(1.0, score_match));
        features.numerical_features.push_back(score_match);
        features.feature_names.push_back("score_match");
    }
    
    // 2. 地理匹配特征
    bool same_province = (student.province == university.province);
    features.categorical_features.push_back(same_province ? 1 : 0);
    
    // 地理偏好匹配度
    double location_preference = 0.0;
    for (const auto& preferred_city : student.preferred_cities) {
        if (preferred_city == university.city) {
            location_preference = 1.0;
            break;
        }
    }
    features.numerical_features.push_back(location_preference);
    features.feature_names.push_back("location_preference_match");
    
    // 3. 专业匹配特征
    double major_preference = 0.0;
    int matched_majors = 0;
    for (const auto& preferred_major : student.preferred_majors) {
        for (const auto& strong_major : university.strong_majors) {
            if (preferred_major == strong_major) {
                matched_majors++;
            }
        }
    }
    
    if (!student.preferred_majors.empty()) {
        major_preference = static_cast<double>(matched_majors) / student.preferred_majors.size();
    }
    features.numerical_features.push_back(major_preference);
    features.feature_names.push_back("major_preference_match");
    
    // 4. 避开专业检查
    bool has_avoided_majors = false;
    for (const auto& avoided_major : student.avoided_majors) {
        for (const auto& strong_major : university.strong_majors) {
            if (avoided_major == strong_major) {
                has_avoided_majors = true;
                break;
            }
        }
        if (has_avoided_majors) break;
    }
    features.categorical_features.push_back(has_avoided_majors ? 1 : 0);
    
    // 5. 综合匹配度特征
    double comprehensive_match = 
        (same_province ? 0.2 : 0.0) +
        location_preference * 0.3 +
        major_preference * 0.5;
    features.numerical_features.push_back(comprehensive_match);
    features.feature_names.push_back("comprehensive_match");
    
    // 6. 竞争强度特征
    if (!university.historical_rankings.empty() && student.ranking > 0) {
        int ranking_gap = student.ranking - university.historical_rankings.back();
        features.numerical_features.push_back(static_cast<double>(ranking_gap));
        features.feature_names.push_back("ranking_gap");
        
        // 竞争优势
        double competitive_advantage = ranking_gap < 0 ? 1.0 : 
                                     (1.0 / (1.0 + std::abs(ranking_gap) / 10000.0));
        features.numerical_features.push_back(competitive_advantage);
        features.feature_names.push_back("competitive_advantage");
    }
    
    // 计算特征哈希
    features.feature_hash = pimpl_->ComputeFeatureHash(features);
    
    // 更新统计信息
    pimpl_->UpdateFeatureStats(features);
    
    return features;
}

std::vector<FeatureVector> FeatureExtractor::BatchExtractFeatures(
    const std::vector<Student>& students,
    const std::vector<University>& universities) {
    
    std::vector<FeatureVector> all_features;
    all_features.reserve(students.size() * (universities.size() + 1)); // +1 for student features
    
    // 并行处理学生特征
    for (const auto& student : students) {
        auto student_features = ExtractStudentFeatures(student);
        all_features.push_back(student_features);
        
        // 为每个学生与大学组合生成交互特征
        for (const auto& university : universities) {
            auto interaction_features = ExtractInteractionFeatures(student, university);
            all_features.push_back(interaction_features);
        }
    }
    
    // 提取大学特征（去重）
    std::unordered_set<std::string> processed_universities;
    for (const auto& university : universities) {
        if (processed_universities.find(university.university_id) == processed_universities.end()) {
            auto university_features = ExtractUniversityFeatures(university);
            all_features.push_back(university_features);
            processed_universities.insert(university.university_id);
        }
    }
    
    return all_features;
}

FeatureVector FeatureExtractor::NormalizeFeatures(const FeatureVector& features) {
    if (!pimpl_->normalized_features_) {
        return features; // 返回原始特征
    }
    
    FeatureVector normalized = features;
    
    std::lock_guard<std::mutex> lock(pimpl_->stats_mutex_);
    
    // 标准化数值特征 (Z-score标准化)
    for (size_t i = 0; i < features.numerical_features.size() && 
                     i < features.feature_names.size(); ++i) {
        const std::string& name = features.feature_names[i];
        double value = features.numerical_features[i];
        
        // 查找统计信息
        auto mean_it = pimpl_->stats_.mean_values.find(name);
        auto std_it = pimpl_->stats_.std_values.find(name);
        
        if (mean_it != pimpl_->stats_.mean_values.end() && 
            std_it != pimpl_->stats_.std_values.end()) {
            double mean = mean_it->second;
            double std_dev = std_it->second;
            
            if (std_dev > 1e-8) { // 避免除零
                normalized.numerical_features[i] = (value - mean) / std_dev;
            }
        }
    }
    
    // 重新计算哈希
    normalized.feature_hash = pimpl_->ComputeFeatureHash(normalized);
    
    return normalized;
}

std::string FeatureExtractor::GetFeatureStats() const {
    std::lock_guard<std::mutex> lock(pimpl_->stats_mutex_);
    
    Json::Value stats;
    stats["total_features_extracted"] = pimpl_->stats_.total_features_extracted;
    
    auto time_t = std::chrono::system_clock::to_time_t(pimpl_->stats_.last_update);
    stats["last_update"] = static_cast<int64_t>(time_t);
    
    // 特征统计信息
    Json::Value features_stats;
    for (const auto& pair : pimpl_->stats_.mean_values) {
        Json::Value feature_stat;
        feature_stat["mean"] = pair.second;
        feature_stat["std"] = pimpl_->stats_.std_values.at(pair.first);
        feature_stat["min"] = pimpl_->stats_.min_values.at(pair.first);
        feature_stat["max"] = pimpl_->stats_.max_values.at(pair.first);
        features_stats[pair.first] = feature_stat;
    }
    stats["features"] = features_stats;
    
    // 缓存统计
    {
        std::shared_lock<std::shared_mutex> cache_lock(pimpl_->cache_mutex_);
        stats["cache_size"] = static_cast<int>(pimpl_->feature_cache_.size());
    }
    
    Json::StreamWriterBuilder builder;
    return Json::writeString(builder, stats);
}

void FeatureExtractor::ClearFeatureCache() {
    std::unique_lock<std::shared_mutex> lock(pimpl_->cache_mutex_);
    pimpl_->feature_cache_.clear();
}

} // namespace volunteer_matcher