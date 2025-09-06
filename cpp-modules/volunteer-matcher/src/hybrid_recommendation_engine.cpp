/**
 * @file hybrid_recommendation_engine.cpp
 * @brief 高考志愿填报系统 - 混合推荐引擎实现
 * @author 高考志愿填报系统开发团队
 * @version 1.0.0
 * @date 2025-01-18
 * 
 * 混合推荐引擎，融合传统算法和AI推荐，提供：
 * - 多层融合策略
 * - 动态权重调整
 * - 结果多样性保证
 * - 实时性能监控
 */

#include "volunteer_matcher.h"
#include <algorithm>
#include <cmath>
#include <fstream>
#include <iomanip>
#include <numeric>
#include <thread>
#include <future>
#include <json/json.h>

namespace volunteer_matcher {

/**
 * @brief 融合配置参数
 */
struct FusionConfig {
    // 权重配置
    double traditional_weight = 0.6;       ///< 传统算法权重
    double ai_weight = 0.4;                ///< AI推荐权重
    
    // 多样性配置
    double diversity_factor = 0.15;        ///< 多样性因子
    int max_same_city_ratio = 40;          ///< 同城市最大比例(%)
    int max_same_level_ratio = 60;         ///< 同层次最大比例(%)
    
    // 冲稳保策略
    double rush_ratio = 0.3;               ///< 冲刺比例
    double stable_ratio = 0.5;             ///< 稳妥比例
    double safe_ratio = 0.2;               ///< 保底比例
    
    // 性能配置
    int max_candidates = 500;              ///< 最大候选数
    double score_threshold = 0.1;          ///< 最低分数阈值
    bool enable_parallel = true;           ///< 是否启用并行
    
    // 自适应配置
    bool enable_adaptive_weights = true;   ///< 启用自适应权重
    double confidence_threshold = 0.8;     ///< 置信度阈值
};

/**
 * @brief 推荐候选项
 */
struct RecommendationCandidate {
    VolunteerRecommendation traditional_rec; ///< 传统推荐
    AIRecommendation ai_rec;                 ///< AI推荐
    
    double fusion_score;                     ///< 融合得分
    double confidence;                       ///< 综合置信度
    std::string fusion_reason;               ///< 融合理由
    
    // 多样性指标
    int diversity_penalty;                   ///< 多样性惩罚
    bool is_unique_choice;                   ///< 是否独特选择
};

/**
 * @brief 混合推荐引擎内部实现
 */
class HybridRecommendationEngine::HybridEngineImpl {
public:
    // 核心组件
    std::shared_ptr<VolunteerMatcher> traditional_matcher_;
    std::shared_ptr<AIRecommendationEngine> ai_engine_;
    std::unique_ptr<FeatureExtractor> feature_extractor_;
    
    // 配置和缓存
    FusionConfig config_;
    Json::Value global_config_;
    
    // 性能监控
    struct HybridStats {
        std::atomic<uint64_t> total_requests{0};
        std::atomic<uint64_t> successful_fusions{0};
        std::atomic<double> avg_fusion_time{0.0};
        std::atomic<double> traditional_weight_avg{0.0};
        std::atomic<double> ai_weight_avg{0.0};
        std::atomic<uint64_t> adaptive_adjustments{0};
        std::chrono::system_clock::time_point last_reset_time;
    };
    
    HybridStats stats_;
    mutable std::mutex stats_mutex_;
    
    // 缓存系统
    struct CacheEntry {
        std::vector<RecommendationCandidate> candidates;
        std::chrono::system_clock::time_point created_time;
        double cache_score;
    };
    
    std::unordered_map<std::string, CacheEntry> recommendation_cache_;
    mutable std::shared_mutex cache_mutex_;
    int cache_max_size_ = 1000;
    std::chrono::minutes cache_ttl_{30};
    
    HybridEngineImpl() {
        stats_.last_reset_time = std::chrono::system_clock::now();
        feature_extractor_ = std::make_unique<FeatureExtractor>();
    }
    
    ~HybridEngineImpl() = default;
    
    bool Initialize(
        std::shared_ptr<VolunteerMatcher> traditional_matcher,
        std::shared_ptr<AIRecommendationEngine> ai_engine,
        const std::string& config_path) {
        
        traditional_matcher_ = traditional_matcher;
        ai_engine_ = ai_engine;
        
        // 初始化特征提取器
        if (!feature_extractor_->Initialize("")) {
            return false;
        }
        
        // 加载配置
        if (!config_path.empty()) {
            if (!LoadConfig(config_path)) {
                return false;
            }
        }
        
        return traditional_matcher_ && ai_engine_;
    }
    
    bool LoadConfig(const std::string& config_path) {
        try {
            std::ifstream config_file(config_path);
            if (!config_file.is_open()) {
                return false;
            }
            
            Json::Reader reader;
            if (!reader.parse(config_file, global_config_)) {
                return false;
            }
            
            // 解析混合引擎配置
            if (global_config_.isMember("hybrid_engine")) {
                const Json::Value& hybrid_config = global_config_["hybrid_engine"];
                
                config_.traditional_weight = hybrid_config.get("traditional_weight", 0.6).asDouble();
                config_.ai_weight = hybrid_config.get("ai_weight", 0.4).asDouble();
                config_.diversity_factor = hybrid_config.get("diversity_factor", 0.15).asDouble();
                config_.max_same_city_ratio = hybrid_config.get("max_same_city_ratio", 40).asInt();
                config_.max_same_level_ratio = hybrid_config.get("max_same_level_ratio", 60).asInt();
                config_.rush_ratio = hybrid_config.get("rush_ratio", 0.3).asDouble();
                config_.stable_ratio = hybrid_config.get("stable_ratio", 0.5).asDouble();
                config_.safe_ratio = hybrid_config.get("safe_ratio", 0.2).asDouble();
                config_.max_candidates = hybrid_config.get("max_candidates", 500).asInt();
                config_.score_threshold = hybrid_config.get("score_threshold", 0.1).asDouble();
                config_.enable_parallel = hybrid_config.get("enable_parallel", true).asBool();
                config_.enable_adaptive_weights = hybrid_config.get("enable_adaptive_weights", true).asBool();
                config_.confidence_threshold = hybrid_config.get("confidence_threshold", 0.8).asDouble();
            }
            
            // 解析缓存配置
            if (global_config_.isMember("cache")) {
                const Json::Value& cache_config = global_config_["cache"];
                cache_max_size_ = cache_config.get("max_size", 1000).asInt();
                int ttl_minutes = cache_config.get("ttl_minutes", 30).asInt();
                cache_ttl_ = std::chrono::minutes(ttl_minutes);
            }
            
            return true;
        } catch (const std::exception& e) {
            return false;
        }
    }
    
    VolunteerPlan GenerateHybridPlan(const Student& student, int max_volunteers) {
        auto start_time = std::chrono::high_resolution_clock::now();
        
        VolunteerPlan plan;
        plan.student_id = student.student_id;
        plan.total_volunteers = max_volunteers;
        plan.generated_time = std::chrono::system_clock::now();
        
        try {
            // 检查缓存
            std::string cache_key = GenerateCacheKey(student);
            auto cached_candidates = GetCachedRecommendations(cache_key);
            
            std::vector<RecommendationCandidate> candidates;
            
            if (!cached_candidates.empty()) {
                candidates = cached_candidates;
            } else {
                // 生成推荐候选
                candidates = GenerateRecommendationCandidates(student);
                
                // 缓存结果
                CacheRecommendations(cache_key, candidates);
            }
            
            // 应用融合策略
            auto fused_recommendations = ApplyFusionStrategy(candidates, max_volunteers);
            
            // 应用冲稳保策略
            plan.recommendations = ApplyRushStableSafeStrategy(fused_recommendations, max_volunteers);
            
            // 计算统计信息
            CalculateHybridPlanStatistics(plan);
            
            // 生成优化建议
            plan.optimization_suggestions = GenerateHybridOptimizationSuggestions(plan);
            
            stats_.successful_fusions++;
            
        } catch (const std::exception& e) {
            // 错误处理：返回传统算法结果作为fallback
            plan = traditional_matcher_->GenerateVolunteerPlan(student, max_volunteers);
        }
        
        auto end_time = std::chrono::high_resolution_clock::now();
        double fusion_time = std::chrono::duration<double, std::milli>(end_time - start_time).count();
        UpdateFusionStats(fusion_time, true);
        
        return plan;
    }
    
    std::vector<RecommendationCandidate> GenerateRecommendationCandidates(const Student& student) {
        std::vector<RecommendationCandidate> candidates;
        
        // 并行获取传统推荐和AI推荐
        auto traditional_future = std::async(std::launch::async, [this, &student]() {
            return traditional_matcher_->GenerateVolunteerPlan(student, config_.max_candidates);
        });
        
        auto ai_future = std::async(std::launch::async, [this, &student]() {
            // 获取候选大学列表（简化实现，实际应从数据库获取）
            std::vector<University> universities; // TODO: 从数据源获取
            return ai_engine_->GenerateRecommendations(student, universities);
        });
        
        // 等待结果
        auto traditional_plan = traditional_future.get();
        auto ai_recommendations = ai_future.get();
        
        // 创建候选映射
        std::unordered_map<std::string, AIRecommendation> ai_map;
        for (const auto& ai_rec : ai_recommendations) {
            std::string key = ai_rec.university_id + "_" + ai_rec.major_id;
            ai_map[key] = ai_rec;
        }
        
        // 融合推荐
        for (const auto& trad_rec : traditional_plan.recommendations) {
            RecommendationCandidate candidate;
            candidate.traditional_rec = trad_rec;
            
            // 查找对应的AI推荐
            std::string key = trad_rec.university_id + "_" + trad_rec.major_id;
            auto ai_it = ai_map.find(key);
            
            if (ai_it != ai_map.end()) {
                candidate.ai_rec = ai_it->second;
                candidate.confidence = (trad_rec.admission_probability + ai_it->second.confidence) / 2.0;
            } else {
                // 创建默认AI推荐
                candidate.ai_rec.ai_score = 0.5;
                candidate.ai_rec.confidence = 0.3;
                candidate.confidence = trad_rec.admission_probability * 0.7; // 降低置信度
            }
            
            candidates.push_back(candidate);
        }
        
        // 添加AI独有的推荐
        for (const auto& ai_rec : ai_recommendations) {
            std::string key = ai_rec.university_id + "_" + ai_rec.major_id;
            bool found = false;
            
            for (const auto& candidate : candidates) {
                std::string cand_key = candidate.traditional_rec.university_id + "_" + 
                                     candidate.traditional_rec.major_id;
                if (key == cand_key) {
                    found = true;
                    break;
                }
            }
            
            if (!found && ai_rec.ai_score > config_.score_threshold) {
                RecommendationCandidate candidate;
                candidate.ai_rec = ai_rec;
                
                // 创建默认传统推荐
                candidate.traditional_rec.university_id = ai_rec.university_id;
                candidate.traditional_rec.major_id = ai_rec.major_id;
                candidate.traditional_rec.admission_probability = ai_rec.ai_score * 0.8; // 保守估计
                candidate.traditional_rec.match_score = ai_rec.ai_score * 100;
                candidate.traditional_rec.risk_level = "未知";
                
                candidate.confidence = ai_rec.confidence * 0.6; // AI独有推荐置信度较低
                candidate.is_unique_choice = true;
                
                candidates.push_back(candidate);
            }
        }
        
        return candidates;
    }
    
    std::vector<VolunteerRecommendation> ApplyFusionStrategy(
        std::vector<RecommendationCandidate>& candidates, 
        int max_volunteers) {
        
        // 自适应权重调整
        if (config_.enable_adaptive_weights) {
            AdjustAdaptiveWeights(candidates);
        }
        
        // 计算融合得分
        for (auto& candidate : candidates) {
            candidate.fusion_score = CalculateFusionScore(candidate);
        }
        
        // 应用多样性策略
        ApplyDiversityStrategy(candidates);
        
        // 排序候选项
        std::sort(candidates.begin(), candidates.end(), 
            [](const RecommendationCandidate& a, const RecommendationCandidate& b) {
                return a.fusion_score > b.fusion_score;
            });
        
        // 提取最终推荐
        std::vector<VolunteerRecommendation> recommendations;
        int count = std::min(max_volunteers, static_cast<int>(candidates.size()));
        
        for (int i = 0; i < count; ++i) {
            VolunteerRecommendation rec = candidates[i].traditional_rec;
            
            // 更新融合信息
            rec.match_score = candidates[i].fusion_score;
            rec.recommendation_reason = GenerateHybridReason(candidates[i]);
            
            recommendations.push_back(rec);
        }
        
        return recommendations;
    }
    
    double CalculateFusionScore(const RecommendationCandidate& candidate) {
        double traditional_score = candidate.traditional_rec.match_score / 100.0;
        double ai_score = candidate.ai_rec.ai_score;
        
        // 基础融合得分
        double fusion_score = config_.traditional_weight * traditional_score + 
                             config_.ai_weight * ai_score;
        
        // 置信度调整
        fusion_score *= candidate.confidence;
        
        // 多样性奖励
        if (candidate.is_unique_choice) {
            fusion_score *= (1.0 + config_.diversity_factor);
        }
        
        // 多样性惩罚
        fusion_score *= (1.0 - candidate.diversity_penalty * 0.01);
        
        return std::max(0.0, std::min(1.0, fusion_score));
    }
    
    void AdjustAdaptiveWeights(const std::vector<RecommendationCandidate>& candidates) {
        if (candidates.empty()) return;
        
        // 计算AI推荐的平均置信度
        double avg_ai_confidence = 0.0;
        int valid_ai_count = 0;
        
        for (const auto& candidate : candidates) {
            if (candidate.ai_rec.confidence > 0.1) {
                avg_ai_confidence += candidate.ai_rec.confidence;
                valid_ai_count++;
            }
        }
        
        if (valid_ai_count > 0) {
            avg_ai_confidence /= valid_ai_count;
            
            // 根据AI置信度调整权重
            if (avg_ai_confidence > config_.confidence_threshold) {
                // AI置信度高，增加AI权重
                config_.ai_weight = std::min(0.7, config_.ai_weight * 1.1);
                config_.traditional_weight = 1.0 - config_.ai_weight;
            } else if (avg_ai_confidence < 0.5) {
                // AI置信度低，增加传统算法权重
                config_.traditional_weight = std::min(0.8, config_.traditional_weight * 1.1);
                config_.ai_weight = 1.0 - config_.traditional_weight;
            }
            
            stats_.adaptive_adjustments++;
        }
        
        // 更新统计
        stats_.traditional_weight_avg = config_.traditional_weight;
        stats_.ai_weight_avg = config_.ai_weight;
    }
    
    void ApplyDiversityStrategy(std::vector<RecommendationCandidate>& candidates) {
        std::unordered_map<std::string, int> city_count;
        std::unordered_map<std::string, int> level_count;
        
        // 统计分布
        for (const auto& candidate : candidates) {
            // TODO: 从大学信息中获取城市和层次
            // city_count[candidate.traditional_rec.university_city]++;
            // level_count[candidate.traditional_rec.university_level]++;
        }
        
        // 计算多样性惩罚
        for (auto& candidate : candidates) {
            candidate.diversity_penalty = 0;
            
            // TODO: 实现具体的多样性惩罚逻辑
            // 基于城市和层次分布计算惩罚分数
        }
    }
    
    std::string GenerateHybridReason(const RecommendationCandidate& candidate) {
        std::ostringstream oss;
        
        oss << "混合推荐: ";
        oss << "传统算法匹配度" << std::fixed << std::setprecision(1) 
            << candidate.traditional_rec.match_score << "%";
        oss << ", AI推荐度" << std::fixed << std::setprecision(1) 
            << (candidate.ai_rec.ai_score * 100) << "%";
        oss << ", 综合置信度" << std::fixed << std::setprecision(1) 
            << (candidate.confidence * 100) << "%";
        
        if (candidate.is_unique_choice) {
            oss << " (AI特色推荐)";
        }
        
        return oss.str();
    }
    
    void CalculateHybridPlanStatistics(VolunteerPlan& plan) {
        // 计算冲稳保分布
        plan.rush_count = 0;
        plan.stable_count = 0;
        plan.safe_count = 0;
        
        for (const auto& rec : plan.recommendations) {
            if (rec.risk_level == "冲") {
                plan.rush_count++;
            } else if (rec.risk_level == "稳") {
                plan.stable_count++;
            } else if (rec.risk_level == "保") {
                plan.safe_count++;
            }
        }
        
        // 计算整体风险评分
        double total_risk = 0.0;
        for (const auto& rec : plan.recommendations) {
            total_risk += rec.admission_probability;
        }
        
        plan.overall_risk_score = plan.recommendations.empty() ? 
            0.0 : total_risk / plan.recommendations.size();
        
        // 评估方案质量
        if (plan.overall_risk_score > 0.8) {
            plan.plan_quality = "优秀";
        } else if (plan.overall_risk_score > 0.6) {
            plan.plan_quality = "良好";
        } else if (plan.overall_risk_score > 0.4) {
            plan.plan_quality = "中等";
        } else {
            plan.plan_quality = "需优化";
        }
    }
    
    std::vector<std::string> GenerateHybridOptimizationSuggestions(const VolunteerPlan& plan) {
        std::vector<std::string> suggestions;
        
        // 冲稳保比例建议
        double rush_ratio = static_cast<double>(plan.rush_count) / plan.total_volunteers;
        double stable_ratio = static_cast<double>(plan.stable_count) / plan.total_volunteers;
        double safe_ratio = static_cast<double>(plan.safe_count) / plan.total_volunteers;
        
        if (rush_ratio > 0.4) {
            suggestions.push_back("冲刺志愿过多，建议增加稳妥志愿以提高录取概率");
        }
        
        if (safe_ratio < 0.15) {
            suggestions.push_back("保底志愿偏少，建议增加保底志愿以降低落榜风险");
        }
        
        if (plan.overall_risk_score < 0.5) {
            suggestions.push_back("整体方案较为保守，可适当增加冲刺志愿");
        }
        
        // 多样性建议
        suggestions.push_back("AI推荐已整合，建议关注推荐解释和置信度");
        
        return suggestions;
    }
    
    std::string GenerateCacheKey(const Student& student) {
        std::ostringstream oss;
        oss << student.student_id << "_" << student.total_score << "_" << student.ranking;
        
        // 添加偏好的hash
        std::hash<std::string> hasher;
        size_t pref_hash = 0;
        for (const auto& city : student.preferred_cities) {
            pref_hash ^= hasher(city);
        }
        for (const auto& major : student.preferred_majors) {
            pref_hash ^= hasher(major);
        }
        oss << "_" << pref_hash;
        
        return oss.str();
    }
    
    std::vector<RecommendationCandidate> GetCachedRecommendations(const std::string& cache_key) {
        std::shared_lock<std::shared_mutex> lock(cache_mutex_);
        
        auto it = recommendation_cache_.find(cache_key);
        if (it != recommendation_cache_.end()) {
            auto now = std::chrono::system_clock::now();
            auto age = std::chrono::duration_cast<std::chrono::minutes>(now - it->second.created_time);
            
            if (age < cache_ttl_) {
                return it->second.candidates;
            } else {
                // 缓存过期，需要在写锁下删除
                lock.unlock();
                std::unique_lock<std::shared_mutex> write_lock(cache_mutex_);
                recommendation_cache_.erase(it);
            }
        }
        
        return {};
    }
    
    void CacheRecommendations(const std::string& cache_key, 
                             const std::vector<RecommendationCandidate>& candidates) {
        std::unique_lock<std::shared_mutex> lock(cache_mutex_);
        
        // 清理过期缓存
        if (recommendation_cache_.size() >= cache_max_size_) {
            CleanExpiredCache();
        }
        
        CacheEntry entry;
        entry.candidates = candidates;
        entry.created_time = std::chrono::system_clock::now();
        entry.cache_score = CalculateCacheScore(candidates);
        
        recommendation_cache_[cache_key] = entry;
    }
    
    void CleanExpiredCache() {
        auto now = std::chrono::system_clock::now();
        auto it = recommendation_cache_.begin();
        
        while (it != recommendation_cache_.end()) {
            auto age = std::chrono::duration_cast<std::chrono::minutes>(now - it->second.created_time);
            if (age >= cache_ttl_) {
                it = recommendation_cache_.erase(it);
            } else {
                ++it;
            }
        }
    }
    
    double CalculateCacheScore(const std::vector<RecommendationCandidate>& candidates) {
        if (candidates.empty()) return 0.0;
        
        double total_score = 0.0;
        for (const auto& candidate : candidates) {
            total_score += candidate.fusion_score;
        }
        
        return total_score / candidates.size();
    }
    
    void UpdateFusionStats(double fusion_time, bool success) {
        std::lock_guard<std::mutex> lock(stats_mutex_);
        
        stats_.total_requests++;
        
        if (success) {
            // 更新平均融合时间
            uint64_t total = stats_.total_requests.load();
            double current_avg = stats_.avg_fusion_time.load();
            double new_avg = (current_avg * (total - 1) + fusion_time) / total;
            stats_.avg_fusion_time = new_avg;
        }
    }
    
    bool SetFusionWeights(double traditional_weight, double ai_weight) {
        if (traditional_weight < 0 || ai_weight < 0 || 
            std::abs(traditional_weight + ai_weight - 1.0) > 0.001) {
            return false;
        }
        
        config_.traditional_weight = traditional_weight;
        config_.ai_weight = ai_weight;
        
        return true;
    }
    
    std::string GetHybridExplanation(const VolunteerRecommendation& recommendation) {
        std::ostringstream oss;
        
        oss << "混合推荐解释:\n";
        oss << "• 匹配得分: " << std::fixed << std::setprecision(1) << recommendation.match_score << "%\n";
        oss << "• 录取概率: " << std::fixed << std::setprecision(1) << (recommendation.admission_probability * 100) << "%\n";
        oss << "• 风险等级: " << recommendation.risk_level << "\n";
        oss << "• 推荐理由: " << recommendation.recommendation_reason << "\n";
        
        if (!recommendation.risk_factors.empty()) {
            oss << "• 风险因素: ";
            for (size_t i = 0; i < recommendation.risk_factors.size(); ++i) {
                if (i > 0) oss << ", ";
                oss << recommendation.risk_factors[i];
            }
            oss << "\n";
        }
        
        return oss.str();
    }
    
    Json::Value GetHybridStats() const {
        std::lock_guard<std::mutex> lock(stats_mutex_);
        
        Json::Value stats;
        stats["total_requests"] = static_cast<int>(stats_.total_requests.load());
        stats["successful_fusions"] = static_cast<int>(stats_.successful_fusions.load());
        stats["avg_fusion_time_ms"] = stats_.avg_fusion_time.load();
        stats["traditional_weight_avg"] = stats_.traditional_weight_avg.load();
        stats["ai_weight_avg"] = stats_.ai_weight_avg.load();
        stats["adaptive_adjustments"] = static_cast<int>(stats_.adaptive_adjustments.load());
        
        // 缓存统计
        std::shared_lock<std::shared_mutex> cache_lock(cache_mutex_);
        stats["cache_size"] = static_cast<int>(recommendation_cache_.size());
        stats["cache_max_size"] = cache_max_size_;
        
        return stats;
    }
};

// HybridRecommendationEngine 实现

HybridRecommendationEngine::HybridRecommendationEngine()
    : pimpl_(std::make_unique<HybridEngineImpl>()) {}

HybridRecommendationEngine::~HybridRecommendationEngine() = default;

bool HybridRecommendationEngine::Initialize(
    std::shared_ptr<VolunteerMatcher> traditional_matcher,
    std::shared_ptr<AIRecommendationEngine> ai_engine,
    const std::string& config_path) {
    
    return pimpl_->Initialize(traditional_matcher, ai_engine, config_path);
}

VolunteerPlan HybridRecommendationEngine::GenerateHybridPlan(const Student& student, int max_volunteers) {
    return pimpl_->GenerateHybridPlan(student, max_volunteers);
}

bool HybridRecommendationEngine::SetFusionWeights(double traditional_weight, double ai_weight) {
    return pimpl_->SetFusionWeights(traditional_weight, ai_weight);
}

std::string HybridRecommendationEngine::GetHybridExplanation(const VolunteerRecommendation& recommendation) {
    return pimpl_->GetHybridExplanation(recommendation);
}

} // namespace volunteer_matcher