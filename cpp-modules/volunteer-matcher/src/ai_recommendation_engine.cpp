/**
 * @file ai_recommendation_engine.cpp
 * @brief 高考志愿填报系统 - AI推荐引擎实现 (ONNX Runtime)
 * @author 高考志愿填报系统开发团队
 * @version 1.0.0
 * @date 2025-01-18
 */

#include "volunteer_matcher.h"
#include <onnxruntime_cxx_api.h>
#include <algorithm>
#include <fstream>
#include <sstream>
#include <cmath>
#include <memory>
#include <thread>
#include <future>
#include <iomanip>
#include <numeric>
#include <json/json.h>

namespace volunteer_matcher {

/**
 * @brief AI推理引擎内部实现类
 */
class AIRecommendationEngine::AIEngineImpl {
public:
    // ONNX Runtime核心组件
    Ort::Env ort_env_;
    std::unique_ptr<Ort::Session> session_;
    std::unique_ptr<Ort::SessionOptions> session_options_;
    std::unique_ptr<Ort::MemoryInfo> memory_info_;
    
    // 模型信息
    std::string model_path_;
    std::string model_version_;
    std::vector<std::string> input_names_;
    std::vector<std::string> output_names_;
    std::vector<std::vector<int64_t>> input_shapes_;
    std::vector<std::vector<int64_t>> output_shapes_;
    
    // 内存池和性能优化
    struct MemoryPool {
        std::vector<float> input_buffer;
        std::vector<float> output_buffer;
        std::vector<Ort::Value> input_tensors;
        std::vector<Ort::Value> output_tensors;
        std::mutex mutex;
        bool in_use = false;
    };
    
    std::vector<std::unique_ptr<MemoryPool>> memory_pools_;
    std::atomic<int> pool_index_{0};
    
    // 推理统计
    struct InferenceStats {
        std::atomic<uint64_t> total_inferences{0};
        std::atomic<uint64_t> successful_inferences{0};
        std::atomic<double> avg_inference_time{0.0};
        std::atomic<double> max_inference_time{0.0};
        std::atomic<uint64_t> batch_inferences{0};
        std::chrono::system_clock::time_point last_reset_time;
    };
    
    InferenceStats stats_;
    mutable std::mutex stats_mutex_;
    
    // 配置参数
    Json::Value config_;
    bool enable_gpu_ = false;
    int max_batch_size_ = 64;
    double inference_timeout_ms_ = 10.0;
    int num_threads_ = std::thread::hardware_concurrency();
    
    AIEngineImpl() : ort_env_(ORT_LOGGING_LEVEL_WARNING, "VolunteerMatcherAI") {
        stats_.last_reset_time = std::chrono::system_clock::now();
        
        // 创建内存信息
        memory_info_ = std::make_unique<Ort::MemoryInfo>(
            Ort::MemoryInfo::CreateCpu(OrtArenaAllocator, OrtMemTypeDefault));
        
        // 初始化内存池
        InitializeMemoryPools();
    }
    
    ~AIEngineImpl() {
        // 释放会话资源
        session_.reset();
        session_options_.reset();
    }
    
    void InitializeMemoryPools() {
        int pool_count = std::min(8, num_threads_); // 限制内存池数量
        memory_pools_.reserve(pool_count);
        
        for (int i = 0; i < pool_count; ++i) {
            auto pool = std::make_unique<MemoryPool>();
            // 预分配缓冲区（假设最大特征数为1000）
            pool->input_buffer.reserve(1000 * max_batch_size_);
            pool->output_buffer.reserve(100 * max_batch_size_);
            memory_pools_.push_back(std::move(pool));
        }
    }
    
    MemoryPool* AcquireMemoryPool() {
        for (int attempts = 0; attempts < 100; ++attempts) {
            int idx = pool_index_.fetch_add(1) % memory_pools_.size();
            auto& pool = memory_pools_[idx];
            
            std::unique_lock<std::mutex> lock(pool->mutex, std::try_to_lock);
            if (lock.owns_lock() && !pool->in_use) {
                pool->in_use = true;
                return pool.get();
            }
        }
        return nullptr; // 所有池都被占用
    }
    
    void ReleaseMemoryPool(MemoryPool* pool) {
        if (pool) {
            std::lock_guard<std::mutex> lock(pool->mutex);
            pool->in_use = false;
            // 清理tensor（避免内存泄漏）
            pool->input_tensors.clear();
            pool->output_tensors.clear();
        }
    }
    
    bool LoadModel(const std::string& model_path) {
        try {
            // 创建会话选项
            session_options_ = std::make_unique<Ort::SessionOptions>();
            session_options_->SetIntraOpNumThreads(num_threads_);
            session_options_->SetGraphOptimizationLevel(GraphOptimizationLevel::ORT_ENABLE_ALL);
            
            // GPU支持（如果启用）
            if (enable_gpu_) {
                try {
                    Ort::ThrowOnError(OrtSessionOptionsAppendExecutionProvider_CUDA(*session_options_, 0));
                } catch (const std::exception& e) {
                    // GPU不可用，回退到CPU
                    enable_gpu_ = false;
                }
            }
            
            // 加载模型
#ifdef _WIN32
            std::wstring wide_path(model_path.begin(), model_path.end());
            session_ = std::make_unique<Ort::Session>(ort_env_, wide_path.c_str(), *session_options_);
#else
            session_ = std::make_unique<Ort::Session>(ort_env_, model_path.c_str(), *session_options_);
#endif
            
            model_path_ = model_path;
            
            // 获取输入输出信息
            ExtractModelMetadata();
            
            return true;
        } catch (const std::exception& e) {
            return false;
        }
    }
    
    void ExtractModelMetadata() {
        if (!session_) return;
        
        Ort::AllocatorWithDefaultOptions allocator;
        
        // 获取输入信息
        size_t num_inputs = session_->GetInputCount();
        input_names_.clear();
        input_shapes_.clear();
        input_names_.reserve(num_inputs);
        input_shapes_.reserve(num_inputs);
        
        for (size_t i = 0; i < num_inputs; ++i) {
            auto input_name = session_->GetInputNameAllocated(i, allocator);
            input_names_.emplace_back(input_name.get());
            
            auto type_info = session_->GetInputTypeInfo(i);
            auto tensor_info = type_info.GetTensorTypeAndShapeInfo();
            input_shapes_.push_back(tensor_info.GetShape());
        }
        
        // 获取输出信息
        size_t num_outputs = session_->GetOutputCount();
        output_names_.clear();
        output_shapes_.clear();
        output_names_.reserve(num_outputs);
        output_shapes_.reserve(num_outputs);
        
        for (size_t i = 0; i < num_outputs; ++i) {
            auto output_name = session_->GetOutputNameAllocated(i, allocator);
            output_names_.emplace_back(output_name.get());
            
            auto type_info = session_->GetOutputTypeInfo(i);
            auto tensor_info = type_info.GetTensorTypeAndShapeInfo();
            output_shapes_.push_back(tensor_info.GetShape());
        }
    }
    
    std::vector<float> PrepareInputTensor(const FeatureVector& features) {
        std::vector<float> input_data;
        
        // 合并数值特征和类别特征
        input_data.reserve(features.numerical_features.size() + features.categorical_features.size());
        
        // 添加数值特征
        for (double val : features.numerical_features) {
            input_data.push_back(static_cast<float>(val));
        }
        
        // 添加类别特征（转换为float）
        for (int val : features.categorical_features) {
            input_data.push_back(static_cast<float>(val));
        }
        
        // 如果特征数量不匹配模型期望，进行padding或截断
        if (!input_shapes_.empty() && input_shapes_[0].size() > 1) {
            size_t expected_size = input_shapes_[0][1]; // 假设第二维是特征维度
            if (input_data.size() < expected_size) {
                input_data.resize(expected_size, 0.0f); // padding with zeros
            } else if (input_data.size() > expected_size) {
                input_data.resize(expected_size); // 截断
            }
        }
        
        return input_data;
    }
    
    AIRecommendation CreateRecommendationFromOutput(
        const std::vector<float>& output_data,
        const std::string& student_id,
        const std::string& university_id,
        const std::string& major_id) {
        
        AIRecommendation recommendation;
        recommendation.student_id = student_id;
        recommendation.university_id = university_id;
        recommendation.major_id = major_id;
        recommendation.model_version = model_version_;
        recommendation.generated_time = std::chrono::system_clock::now();
        
        // 解析模型输出
        if (output_data.size() >= 2) {
            recommendation.ai_score = std::max(0.0, std::min(1.0, static_cast<double>(output_data[0])));
            recommendation.confidence = std::max(0.0, std::min(1.0, static_cast<double>(output_data[1])));
        } else if (output_data.size() >= 1) {
            recommendation.ai_score = std::max(0.0, std::min(1.0, static_cast<double>(output_data[0])));
            recommendation.confidence = 0.8; // 默认置信度
        }
        
        // 生成特征重要性（简化版本）
        recommendation.feature_importance.push_back({"score_match", 0.25});
        recommendation.feature_importance.push_back({"location_preference", 0.20});
        recommendation.feature_importance.push_back({"major_preference", 0.30});
        recommendation.feature_importance.push_back({"ranking_advantage", 0.25});
        
        // 生成解释
        recommendation.explanation = GenerateExplanation(recommendation);
        
        return recommendation;
    }
    
    std::string GenerateExplanation(const AIRecommendation& rec) {
        std::ostringstream oss;
        oss << "AI分析: 推荐度" << std::fixed << std::setprecision(1) << (rec.ai_score * 100) << "%";
        
        if (rec.confidence > 0.8) {
            oss << "，高置信度推荐";
        } else if (rec.confidence > 0.6) {
            oss << "，中等置信度";
        } else {
            oss << "，低置信度，需谨慎考虑";
        }
        
        // 基于特征重要性生成解释
        if (!rec.feature_importance.empty()) {
            auto max_feature = std::max_element(
                rec.feature_importance.begin(),
                rec.feature_importance.end(),
                [](const auto& a, const auto& b) { return a.second < b.second; }
            );
            
            if (max_feature != rec.feature_importance.end()) {
                oss << "，主要基于" << max_feature->first << "匹配";
            }
        }
        
        return oss.str();
    }
    
    void UpdateInferenceStats(double inference_time, bool success) {
        std::lock_guard<std::mutex> lock(stats_mutex_);
        
        stats_.total_inferences++;
        if (success) {
            stats_.successful_inferences++;
        }
        
        // 更新平均推理时间
        double current_avg = stats_.avg_inference_time.load();
        uint64_t total = stats_.total_inferences.load();
        double new_avg = (current_avg * (total - 1) + inference_time) / total;
        stats_.avg_inference_time = new_avg;
        
        // 更新最大推理时间
        double current_max = stats_.max_inference_time.load();
        if (inference_time > current_max) {
            stats_.max_inference_time = inference_time;
        }
    }
};

// AIRecommendationEngine实现

AIRecommendationEngine::AIRecommendationEngine()
    : pimpl_(std::make_unique<AIEngineImpl>()) {}

AIRecommendationEngine::~AIRecommendationEngine() = default;

bool AIRecommendationEngine::Initialize(const std::string& model_path, const std::string& config_path) {
    try {
        // 加载配置文件
        if (!config_path.empty()) {
            std::ifstream config_file(config_path);
            if (config_file.is_open()) {
                Json::Reader reader;
                Json::Value root;
                if (reader.parse(config_file, root)) {
                    pimpl_->config_ = root;
                    
                    // 读取AI配置
                    if (root.isMember("ai_engine")) {
                        const Json::Value& ai_config = root["ai_engine"];
                        pimpl_->enable_gpu_ = ai_config.get("enable_gpu", false).asBool();
                        pimpl_->max_batch_size_ = ai_config.get("max_batch_size", 64).asInt();
                        pimpl_->inference_timeout_ms_ = ai_config.get("inference_timeout_ms", 10.0).asDouble();
                        pimpl_->num_threads_ = ai_config.get("num_threads", 
                            static_cast<int>(std::thread::hardware_concurrency())).asInt();
                    }
                }
            }
        }
        
        // 加载模型
        return LoadModel(model_path);
    } catch (const std::exception& e) {
        return false;
    }
}

bool AIRecommendationEngine::LoadModel(const std::string& model_path) {
    return pimpl_->LoadModel(model_path);
}

std::vector<AIRecommendation> AIRecommendationEngine::GenerateRecommendations(
    const Student& student, 
    const std::vector<University>& universities) {
    
    std::vector<AIRecommendation> recommendations;
    
    if (!pimpl_->session_) {
        return recommendations; // 模型未加载
    }
    
    auto start_time = std::chrono::high_resolution_clock::now();
    
    try {
        // 获取内存池
        auto* pool = pimpl_->AcquireMemoryPool();
        if (!pool) {
            return recommendations; // 无可用内存池
        }
        
        // RAII内存池管理
        struct PoolGuard {
            AIEngineImpl* impl;
            AIEngineImpl::MemoryPool* pool;
            PoolGuard(AIEngineImpl* i, AIEngineImpl::MemoryPool* p) : impl(i), pool(p) {}
            ~PoolGuard() { impl->ReleaseMemoryPool(pool); }
        } pool_guard(pimpl_.get(), pool);
        
        // 创建特征提取器
        FeatureExtractor extractor;
        extractor.Initialize("");
        
        // 为每个大学生成推荐
        for (const auto& university : universities) {
            // 提取交互特征
            FeatureVector features = extractor.ExtractInteractionFeatures(student, university);
            
            // 准备输入数据
            std::vector<float> input_data = pimpl_->PrepareInputTensor(features);
            
            // 创建输入tensor
            std::vector<int64_t> input_shape = {1, static_cast<int64_t>(input_data.size())};
            pool->input_tensors.clear();
            pool->input_tensors.emplace_back(Ort::Value::CreateTensor<float>(
                *pimpl_->memory_info_, input_data.data(), input_data.size(),
                input_shape.data(), input_shape.size()));
            
            // 准备输出tensor
            std::vector<int64_t> output_shape = {1, 2}; // 假设输出2个值：score和confidence
            pool->output_buffer.resize(2);
            pool->output_tensors.clear();
            pool->output_tensors.emplace_back(Ort::Value::CreateTensor<float>(
                *pimpl_->memory_info_, pool->output_buffer.data(), pool->output_buffer.size(),
                output_shape.data(), output_shape.size()));
            
            // 运行推理
            std::vector<const char*> input_names_cstr;
            std::vector<const char*> output_names_cstr;
            
            for (const auto& name : pimpl_->input_names_) {
                input_names_cstr.push_back(name.c_str());
            }
            for (const auto& name : pimpl_->output_names_) {
                output_names_cstr.push_back(name.c_str());
            }
            
            auto inference_start = std::chrono::high_resolution_clock::now();
            
            pimpl_->session_->Run(Ort::RunOptions{nullptr},
                                input_names_cstr.data(), pool->input_tensors.data(), pool->input_tensors.size(),
                                output_names_cstr.data(), pool->output_tensors.data(), pool->output_tensors.size());
            
            auto inference_end = std::chrono::high_resolution_clock::now();
            double inference_time = std::chrono::duration<double, std::milli>(inference_end - inference_start).count();
            
            // 处理输出
            AIRecommendation rec = pimpl_->CreateRecommendationFromOutput(
                pool->output_buffer, student.student_id, university.university_id, "default_major");
            
            recommendations.push_back(rec);
            
            // 更新统计
            pimpl_->UpdateInferenceStats(inference_time, true);
        }
        
    } catch (const std::exception& e) {
        auto end_time = std::chrono::high_resolution_clock::now();
        double total_time = std::chrono::duration<double, std::milli>(end_time - start_time).count();
        pimpl_->UpdateInferenceStats(total_time, false);
    }
    
    return recommendations;
}

std::vector<std::vector<AIRecommendation>> AIRecommendationEngine::BatchGenerateRecommendations(
    const std::vector<Student>& students,
    const std::vector<University>& universities) {
    
    std::vector<std::vector<AIRecommendation>> batch_results;
    batch_results.reserve(students.size());
    
    // 使用线程池并行处理
    std::vector<std::future<std::vector<AIRecommendation>>> futures;
    
    for (const auto& student : students) {
        auto future = std::async(std::launch::async, [this, &student, &universities]() {
            return GenerateRecommendations(student, universities);
        });
        futures.push_back(std::move(future));
    }
    
    // 收集结果
    for (auto& future : futures) {
        batch_results.push_back(future.get());
    }
    
    pimpl_->stats_.batch_inferences++;
    
    return batch_results;
}

double AIRecommendationEngine::PredictAdmissionProbability(const FeatureVector& features) {
    if (!pimpl_->session_) {
        return 0.5; // 默认概率
    }
    
    try {
        // 获取内存池
        auto* pool = pimpl_->AcquireMemoryPool();
        if (!pool) {
            return 0.5;
        }
        
        struct PoolGuard {
            AIEngineImpl* impl;
            AIEngineImpl::MemoryPool* pool;
            PoolGuard(AIEngineImpl* i, AIEngineImpl::MemoryPool* p) : impl(i), pool(p) {}
            ~PoolGuard() { impl->ReleaseMemoryPool(pool); }
        } pool_guard(pimpl_.get(), pool);
        
        // 准备输入数据
        std::vector<float> input_data = pimpl_->PrepareInputTensor(features);
        
        // 创建tensor并运行推理
        std::vector<int64_t> input_shape = {1, static_cast<int64_t>(input_data.size())};
        pool->input_tensors.clear();
        pool->input_tensors.emplace_back(Ort::Value::CreateTensor<float>(
            *pimpl_->memory_info_, input_data.data(), input_data.size(),
            input_shape.data(), input_shape.size()));
        
        std::vector<int64_t> output_shape = {1, 1};
        pool->output_buffer.resize(1);
        pool->output_tensors.clear();
        pool->output_tensors.emplace_back(Ort::Value::CreateTensor<float>(
            *pimpl_->memory_info_, pool->output_buffer.data(), pool->output_buffer.size(),
            output_shape.data(), output_shape.size()));
        
        std::vector<const char*> input_names_cstr;
        std::vector<const char*> output_names_cstr;
        
        for (const auto& name : pimpl_->input_names_) {
            input_names_cstr.push_back(name.c_str());
        }
        for (const auto& name : pimpl_->output_names_) {
            output_names_cstr.push_back(name.c_str());
        }
        
        pimpl_->session_->Run(Ort::RunOptions{nullptr},
                            input_names_cstr.data(), pool->input_tensors.data(), pool->input_tensors.size(),
                            output_names_cstr.data(), pool->output_tensors.data(), pool->output_tensors.size());
        
        return std::max(0.0, std::min(1.0, static_cast<double>(pool->output_buffer[0])));
        
    } catch (const std::exception& e) {
        return 0.5; // 推理失败时返回默认值
    }
}

std::string AIRecommendationEngine::GetModelInfo() const {
    Json::Value info;
    info["model_path"] = pimpl_->model_path_;
    info["model_version"] = pimpl_->model_version_;
    info["enable_gpu"] = pimpl_->enable_gpu_;
    info["num_threads"] = pimpl_->num_threads_;
    
    // 输入输出信息
    Json::Value inputs(Json::arrayValue);
    for (size_t i = 0; i < pimpl_->input_names_.size(); ++i) {
        Json::Value input_info;
        input_info["name"] = pimpl_->input_names_[i];
        
        Json::Value shape(Json::arrayValue);
        if (i < pimpl_->input_shapes_.size()) {
            for (auto dim : pimpl_->input_shapes_[i]) {
                shape.append(static_cast<int>(dim));
            }
        }
        input_info["shape"] = shape;
        inputs.append(input_info);
    }
    info["inputs"] = inputs;
    
    Json::Value outputs(Json::arrayValue);
    for (size_t i = 0; i < pimpl_->output_names_.size(); ++i) {
        Json::Value output_info;
        output_info["name"] = pimpl_->output_names_[i];
        
        Json::Value shape(Json::arrayValue);
        if (i < pimpl_->output_shapes_.size()) {
            for (auto dim : pimpl_->output_shapes_[i]) {
                shape.append(static_cast<int>(dim));
            }
        }
        output_info["shape"] = shape;
        outputs.append(output_info);
    }
    info["outputs"] = outputs;
    
    // 推理统计
    info["stats"]["total_inferences"] = static_cast<int>(pimpl_->stats_.total_inferences.load());
    info["stats"]["successful_inferences"] = static_cast<int>(pimpl_->stats_.successful_inferences.load());
    info["stats"]["avg_inference_time"] = pimpl_->stats_.avg_inference_time.load();
    info["stats"]["max_inference_time"] = pimpl_->stats_.max_inference_time.load();
    info["stats"]["batch_inferences"] = static_cast<int>(pimpl_->stats_.batch_inferences.load());
    
    Json::StreamWriterBuilder builder;
    return Json::writeString(builder, info);
}

bool AIRecommendationEngine::UpdateModel(const std::string& new_model_path) {
    try {
        // 创建新的实现实例
        auto new_impl = std::make_unique<AIEngineImpl>();
        
        // 复制配置
        new_impl->config_ = pimpl_->config_;
        new_impl->enable_gpu_ = pimpl_->enable_gpu_;
        new_impl->max_batch_size_ = pimpl_->max_batch_size_;
        new_impl->inference_timeout_ms_ = pimpl_->inference_timeout_ms_;
        new_impl->num_threads_ = pimpl_->num_threads_;
        
        // 加载新模型
        if (new_impl->LoadModel(new_model_path)) {
            // 原子替换
            pimpl_ = std::move(new_impl);
            return true;
        }
        
        return false;
    } catch (const std::exception& e) {
        return false;
    }
}

} // namespace volunteer_matcher