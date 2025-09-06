/**
 * @file fusion_strategies.cpp
 * @brief 高考志愿填报系统 - 融合策略算法实现
 * @author 高考志愿填报系统开发团队
 * @version 1.0.0
 * @date 2025-01-18
 * 
 * 混合推荐引擎的融合策略算法集合，包括：
 * - 多层融合策略
 * - 动态权重调整
 * - 多样性控制
 * - 冲稳保平衡
 */

#include "volunteer_matcher.h"
#include <algorithm>
#include <cmath>
#include <numeric>
#include <unordered_set>
#include <random>

namespace volunteer_matcher {

/**
 * @brief 融合策略工具类
 */
class FusionStrategies {
public:
    /**
     * @brief Weighted Score Fusion - 加权分数融合
     * @param traditional_score 传统算法分数
     * @param ai_score AI推荐分数
     * @param traditional_weight 传统算法权重
     * @param ai_weight AI权重
     * @param confidence_factor 置信度因子
     * @return 融合分数
     */
    static double WeightedScoreFusion(
        double traditional_score, 
        double ai_score,
        double traditional_weight,
        double ai_weight,
        double confidence_factor = 1.0) {
        
        // 归一化权重
        double total_weight = traditional_weight + ai_weight;
        if (total_weight > 0) {
            traditional_weight /= total_weight;
            ai_weight /= total_weight;
        }
        
        // 基础加权融合
        double fusion_score = traditional_weight * traditional_score + ai_weight * ai_score;
        
        // 应用置信度调整
        fusion_score *= confidence_factor;
        
        return std::max(0.0, std::min(1.0, fusion_score));
    }
    
    /**
     * @brief Bayesian Fusion - 贝叶斯融合
     * @param traditional_prob 传统算法概率
     * @param ai_prob AI推荐概率
     * @param traditional_confidence 传统算法置信度
     * @param ai_confidence AI置信度
     * @return 融合概率
     */
    static double BayesianFusion(
        double traditional_prob,
        double ai_prob,
        double traditional_confidence,
        double ai_confidence) {
        
        // 计算加权平均
        double total_confidence = traditional_confidence + ai_confidence;
        if (total_confidence <= 0) return 0.5; // 默认概率
        
        double weighted_prob = (traditional_prob * traditional_confidence + 
                               ai_prob * ai_confidence) / total_confidence;
        
        // 贝叶斯调整 - 考虑不确定性
        double uncertainty = 1.0 - total_confidence / 2.0; // 假设最大置信度为2
        double adjusted_prob = weighted_prob * (1.0 - uncertainty) + 0.5 * uncertainty;
        
        return std::max(0.0, std::min(1.0, adjusted_prob));
    }
    
    /**
     * @brief Rank Fusion - 排名融合（Borda Count）
     * @param traditional_ranks 传统算法排名列表
     * @param ai_ranks AI推荐排名列表
     * @param weight_ratio 权重比例 (traditional:ai)
     * @return 融合排名分数
     */
    static std::vector<double> RankFusion(
        const std::vector<int>& traditional_ranks,
        const std::vector<int>& ai_ranks,
        double weight_ratio = 1.0) {
        
        size_t size = std::max(traditional_ranks.size(), ai_ranks.size());
        std::vector<double> fusion_scores(size, 0.0);
        
        // Borda count with weights
        for (size_t i = 0; i < size; ++i) {
            double traditional_score = 0.0;
            double ai_score = 0.0;
            
            if (i < traditional_ranks.size()) {
                traditional_score = size - traditional_ranks[i];
            }
            
            if (i < ai_ranks.size()) {
                ai_score = size - ai_ranks[i];
            }
            
            // 加权融合
            fusion_scores[i] = weight_ratio * traditional_score + ai_score;
        }
        
        return fusion_scores;
    }
    
    /**
     * @brief Ensemble Voting - 集成投票策略
     * @param traditional_recommendations 传统推荐列表
     * @param ai_recommendations AI推荐列表
     * @param voting_threshold 投票阈值
     * @return 投票结果
     */
    static std::vector<std::pair<std::string, double>> EnsembleVoting(
        const std::vector<VolunteerRecommendation>& traditional_recommendations,
        const std::vector<AIRecommendation>& ai_recommendations,
        double voting_threshold = 0.5) {
        
        std::unordered_map<std::string, double> vote_scores;
        
        // 传统算法投票
        for (const auto& rec : traditional_recommendations) {
            std::string key = rec.university_id + "_" + rec.major_id;
            vote_scores[key] += rec.admission_probability;
        }
        
        // AI推荐投票
        for (const auto& rec : ai_recommendations) {
            std::string key = rec.university_id + "_" + rec.major_id;
            vote_scores[key] += rec.ai_score;
        }
        
        // 筛选和排序
        std::vector<std::pair<std::string, double>> results;
        for (const auto& pair : vote_scores) {
            if (pair.second >= voting_threshold) {
                results.push_back(pair);
            }
        }
        
        std::sort(results.begin(), results.end(),
                 [](const auto& a, const auto& b) { return a.second > b.second; });
        
        return results;
    }
    
    /**
     * @brief Dynamic Weight Adjustment - 动态权重调整
     * @param historical_performance 历史性能数据
     * @param current_confidence 当前置信度
     * @param base_traditional_weight 基础传统权重
     * @param base_ai_weight 基础AI权重
     * @return 调整后的权重对
     */
    static std::pair<double, double> DynamicWeightAdjustment(
        const std::vector<double>& historical_performance,
        double current_confidence,
        double base_traditional_weight = 0.6,
        double base_ai_weight = 0.4) {
        
        double traditional_weight = base_traditional_weight;
        double ai_weight = base_ai_weight;
        
        // 基于历史性能调整
        if (!historical_performance.empty()) {
            double avg_performance = std::accumulate(historical_performance.begin(),
                                                   historical_performance.end(), 0.0) /
                                   historical_performance.size();
            
            if (avg_performance > 0.8) {
                // 历史表现好，增加AI权重
                ai_weight *= 1.2;
                traditional_weight *= 0.9;
            } else if (avg_performance < 0.5) {
                // 历史表现差，增加传统算法权重
                traditional_weight *= 1.2;
                ai_weight *= 0.9;
            }
        }
        
        // 基于当前置信度调整
        if (current_confidence > 0.8) {
            ai_weight *= (1.0 + current_confidence - 0.8);
        } else if (current_confidence < 0.5) {
            traditional_weight *= (1.5 - current_confidence);
        }
        
        // 归一化权重
        double total = traditional_weight + ai_weight;
        if (total > 0) {
            traditional_weight /= total;
            ai_weight /= total;
        }
        
        return {traditional_weight, ai_weight};
    }
    
    /**
     * @brief Diversity Control - 多样性控制
     * @param recommendations 推荐列表
     * @param max_same_category_ratio 同类别最大比例
     * @param diversity_penalty_factor 多样性惩罚因子
     * @return 多样性调整后的推荐列表
     */
    static std::vector<VolunteerRecommendation> DiversityControl(
        std::vector<VolunteerRecommendation> recommendations,
        double max_same_category_ratio = 0.4,
        double diversity_penalty_factor = 0.2) {
        
        if (recommendations.empty()) return recommendations;
        
        // 统计类别分布
        std::unordered_map<std::string, int> city_count;
        std::unordered_map<std::string, int> level_count;
        std::unordered_map<std::string, int> major_category_count;
        
        // 这里需要从大学和专业信息中获取城市、层次、专业类别
        // 简化实现，使用university_id的前缀作为类别标识
        for (const auto& rec : recommendations) {
            std::string city = rec.university_id.substr(0, 2); // 假设前两位表示城市
            std::string level = rec.university_id.substr(2, 1); // 假设第三位表示层次
            std::string major_cat = rec.major_id.substr(0, 2); // 假设前两位表示专业类别
            
            city_count[city]++;
            level_count[level]++;
            major_category_count[major_cat]++;
        }
        
        int total_count = static_cast<int>(recommendations.size());
        int max_same_count = static_cast<int>(total_count * max_same_category_ratio);
        
        // 应用多样性惩罚
        for (auto& rec : recommendations) {
            std::string city = rec.university_id.substr(0, 2);
            std::string level = rec.university_id.substr(2, 1);
            std::string major_cat = rec.major_id.substr(0, 2);
            
            double penalty = 0.0;
            
            // 计算城市集中度惩罚
            if (city_count[city] > max_same_count) {
                penalty += diversity_penalty_factor * 
                          (static_cast<double>(city_count[city]) / total_count - max_same_category_ratio);
            }
            
            // 计算层次集中度惩罚
            if (level_count[level] > max_same_count) {
                penalty += diversity_penalty_factor * 
                          (static_cast<double>(level_count[level]) / total_count - max_same_category_ratio);
            }
            
            // 计算专业类别集中度惩罚
            if (major_category_count[major_cat] > max_same_count) {
                penalty += diversity_penalty_factor * 
                          (static_cast<double>(major_category_count[major_cat]) / total_count - max_same_category_ratio);
            }
            
            // 应用惩罚到匹配分数
            rec.match_score *= (1.0 - penalty);
            rec.match_score = std::max(0.0, rec.match_score);
        }
        
        // 重新排序
        std::sort(recommendations.begin(), recommendations.end(),
                 [](const VolunteerRecommendation& a, const VolunteerRecommendation& b) {
                     return a.match_score > b.match_score;
                 });
        
        return recommendations;
    }
    
    /**
     * @brief Risk-Return Optimization - 风险收益优化
     * @param recommendations 推荐列表
     * @param risk_tolerance 风险容忍度 (0-1)
     * @param target_rush_ratio 目标冲刺比例
     * @param target_stable_ratio 目标稳妥比例
     * @param target_safe_ratio 目标保底比例
     * @return 优化后的推荐列表
     */
    static std::vector<VolunteerRecommendation> RiskReturnOptimization(
        std::vector<VolunteerRecommendation> recommendations,
        double risk_tolerance = 0.6,
        double target_rush_ratio = 0.3,
        double target_stable_ratio = 0.5,
        double target_safe_ratio = 0.2) {
        
        if (recommendations.empty()) return recommendations;
        
        // 对推荐进行风险分类
        std::vector<VolunteerRecommendation> rush_recs;
        std::vector<VolunteerRecommendation> stable_recs;
        std::vector<VolunteerRecommendation> safe_recs;
        
        for (const auto& rec : recommendations) {
            if (rec.risk_level == "冲" || rec.admission_probability < 0.4) {
                rush_recs.push_back(rec);
            } else if (rec.risk_level == "稳" || rec.admission_probability < 0.7) {
                stable_recs.push_back(rec);
            } else {
                safe_recs.push_back(rec);
            }
        }
        
        int total_count = static_cast<int>(recommendations.size());
        int target_rush_count = static_cast<int>(total_count * target_rush_ratio);
        int target_stable_count = static_cast<int>(total_count * target_stable_ratio);
        int target_safe_count = total_count - target_rush_count - target_stable_count;
        
        // 排序各类别
        auto sort_by_score = [](const VolunteerRecommendation& a, const VolunteerRecommendation& b) {
            return a.match_score > b.match_score;
        };
        
        std::sort(rush_recs.begin(), rush_recs.end(), sort_by_score);
        std::sort(stable_recs.begin(), stable_recs.end(), sort_by_score);
        std::sort(safe_recs.begin(), safe_recs.end(), sort_by_score);
        
        // 根据风险容忍度调整比例
        if (risk_tolerance > 0.7) {
            // 高风险容忍度，增加冲刺比例
            target_rush_count = std::min(static_cast<int>(target_rush_count * 1.3), 
                                       static_cast<int>(rush_recs.size()));
            target_stable_count = std::max(1, total_count - target_rush_count - target_safe_count);
        } else if (risk_tolerance < 0.4) {
            // 低风险容忍度，增加保底比例
            target_safe_count = std::min(static_cast<int>(target_safe_count * 1.5), 
                                       static_cast<int>(safe_recs.size()));
            target_rush_count = std::max(1, total_count - target_stable_count - target_safe_count);
        }
        
        // 组合最终结果
        std::vector<VolunteerRecommendation> optimized_recs;
        
        // 添加冲刺志愿
        for (int i = 0; i < target_rush_count && i < static_cast<int>(rush_recs.size()); ++i) {
            optimized_recs.push_back(rush_recs[i]);
        }
        
        // 添加稳妥志愿
        for (int i = 0; i < target_stable_count && i < static_cast<int>(stable_recs.size()); ++i) {
            optimized_recs.push_back(stable_recs[i]);
        }
        
        // 添加保底志愿
        for (int i = 0; i < target_safe_count && i < static_cast<int>(safe_recs.size()); ++i) {
            optimized_recs.push_back(safe_recs[i]);
        }
        
        return optimized_recs;
    }
    
    /**
     * @brief Conflict Resolution - 冲突解决
     * @param traditional_rec 传统推荐
     * @param ai_rec AI推荐
     * @param resolution_strategy 解决策略 ("average"|"max"|"confident"|"weighted")
     * @return 解决后的推荐
     */
    static VolunteerRecommendation ConflictResolution(
        const VolunteerRecommendation& traditional_rec,
        const AIRecommendation& ai_rec,
        const std::string& resolution_strategy = "weighted") {
        
        VolunteerRecommendation resolved_rec = traditional_rec;
        
        if (resolution_strategy == "average") {
            // 平均策略
            resolved_rec.admission_probability = 
                (traditional_rec.admission_probability + ai_rec.ai_score) / 2.0;
            resolved_rec.match_score = 
                (traditional_rec.match_score + ai_rec.ai_score * 100) / 2.0;
                
        } else if (resolution_strategy == "max") {
            // 最大值策略
            if (ai_rec.ai_score > traditional_rec.admission_probability) {
                resolved_rec.admission_probability = ai_rec.ai_score;
                resolved_rec.match_score = ai_rec.ai_score * 100;
            }
            
        } else if (resolution_strategy == "confident") {
            // 置信度策略 - 选择更有信心的推荐
            double traditional_confidence = traditional_rec.admission_probability;
            double ai_confidence = ai_rec.confidence;
            
            if (ai_confidence > traditional_confidence) {
                resolved_rec.admission_probability = ai_rec.ai_score;
                resolved_rec.match_score = ai_rec.ai_score * 100;
                resolved_rec.recommendation_reason = ai_rec.explanation;
            }
            
        } else { // weighted (default)
            // 加权策略 - 基于置信度加权
            double traditional_weight = traditional_rec.admission_probability;
            double ai_weight = ai_rec.confidence;
            double total_weight = traditional_weight + ai_weight;
            
            if (total_weight > 0) {
                resolved_rec.admission_probability = 
                    (traditional_rec.admission_probability * traditional_weight + 
                     ai_rec.ai_score * ai_weight) / total_weight;
                resolved_rec.match_score = 
                    (traditional_rec.match_score * traditional_weight + 
                     ai_rec.ai_score * 100 * ai_weight) / total_weight;
            }
        }
        
        return resolved_rec;
    }
    
    /**
     * @brief Calculate Fusion Quality - 计算融合质量
     * @param fusion_results 融合结果
     * @param ground_truth 真实结果（如有）
     * @return 质量评分
     */
    static double CalculateFusionQuality(
        const std::vector<VolunteerRecommendation>& fusion_results,
        const std::vector<VolunteerRecommendation>& ground_truth = {}) {
        
        if (fusion_results.empty()) return 0.0;
        
        double quality_score = 0.0;
        
        // 计算内在质量指标
        // 1. 分数分布的合理性
        std::vector<double> scores;
        for (const auto& rec : fusion_results) {
            scores.push_back(rec.match_score);
        }
        
        // 计算分数的方差 - 好的推荐应该有适度的分散性
        double mean_score = std::accumulate(scores.begin(), scores.end(), 0.0) / scores.size();
        double variance = 0.0;
        for (double score : scores) {
            variance += (score - mean_score) * (score - mean_score);
        }
        variance /= scores.size();
        
        // 方差在合理范围内得高分
        double variance_score = 1.0 - std::min(1.0, std::abs(variance - 400) / 400); // 假设理想方差为400
        
        // 2. 风险分布的合理性
        int rush_count = 0, stable_count = 0, safe_count = 0;
        for (const auto& rec : fusion_results) {
            if (rec.risk_level == "冲") rush_count++;
            else if (rec.risk_level == "稳") stable_count++;
            else if (rec.risk_level == "保") safe_count++;
        }
        
        double total = static_cast<double>(fusion_results.size());
        double rush_ratio = rush_count / total;
        double stable_ratio = stable_count / total;
        double safe_ratio = safe_count / total;
        
        // 理想的冲稳保比例是 3:5:2
        double ideal_rush = 0.3, ideal_stable = 0.5, ideal_safe = 0.2;
        double ratio_score = 1.0 - (std::abs(rush_ratio - ideal_rush) + 
                                   std::abs(stable_ratio - ideal_stable) + 
                                   std::abs(safe_ratio - ideal_safe)) / 3.0;
        
        // 3. 多样性评分
        std::unordered_set<std::string> unique_universities;
        std::unordered_set<std::string> unique_cities;
        for (const auto& rec : fusion_results) {
            unique_universities.insert(rec.university_id);
            unique_cities.insert(rec.university_id.substr(0, 2)); // 简化的城市标识
        }
        
        double diversity_score = static_cast<double>(unique_universities.size()) / fusion_results.size();
        
        // 综合质量评分
        quality_score = 0.4 * variance_score + 0.3 * ratio_score + 0.3 * diversity_score;
        
        // 如果有真实结果，计算准确性
        if (!ground_truth.empty()) {
            // 计算overlap
            std::unordered_set<std::string> fusion_set;
            std::unordered_set<std::string> truth_set;
            
            for (const auto& rec : fusion_results) {
                fusion_set.insert(rec.university_id + "_" + rec.major_id);
            }
            for (const auto& rec : ground_truth) {
                truth_set.insert(rec.university_id + "_" + rec.major_id);
            }
            
            int intersection_count = 0;
            for (const auto& item : fusion_set) {
                if (truth_set.count(item)) {
                    intersection_count++;
                }
            }
            
            double accuracy = static_cast<double>(intersection_count) / 
                            std::max(fusion_set.size(), truth_set.size());
            
            // 结合准确性调整质量评分
            quality_score = 0.7 * quality_score + 0.3 * accuracy;
        }
        
        return std::max(0.0, std::min(1.0, quality_score));
    }
};

} // namespace volunteer_matcher