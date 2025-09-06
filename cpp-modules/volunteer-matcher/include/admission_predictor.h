/**
 * @file admission_predictor.h
 * @brief 高考志愿填报系统 - 录取概率预测算法
 * @author 高考志愿填报系统开发团队
 * @version 1.0.0
 * @date 2025-01-18
 * 
 * 基于机器学习的录取概率预测引擎：
 * - XGBoost回归模型
 * - 特征工程和数据预处理
 * - 模型训练和预测
 * - 预测结果解释
 */

#ifndef ADMISSION_PREDICTOR_H
#define ADMISSION_PREDICTOR_H

#include <vector>
#include <string>
#include <unordered_map>
#include <memory>
#include <functional>

namespace volunteer_matcher {

// 前向声明
struct Student;
struct University;
struct Major;

<<<<<<< HEAD
// FeatureVector已在volunteer_matcher.h中定义
// 使用前向声明
struct FeatureVector;
=======
/**
 * @brief 特征向量结构
 */
struct FeatureVector {
    std::vector<double> features;           ///< 特征值
    std::vector<std::string> feature_names; ///< 特征名称
    double target;                          ///< 目标值(训练时使用)
};
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc

/**
 * @brief 预测结果
 */
struct PredictionResult {
    double probability;                     ///< 录取概率 (0-1)
    double confidence;                      ///< 置信度 (0-1)
    std::vector<std::pair<std::string, double>> feature_importance; ///< 特征重要性
    std::string explanation;                ///< 预测解释
};

/**
 * @brief 模型训练参数
 */
struct TrainingParams {
    int max_depth = 6;                      ///< 最大树深度
    double learning_rate = 0.1;             ///< 学习率
    int n_estimators = 100;                 ///< 估计器数量
    double subsample = 0.8;                 ///< 子样本比例
    double colsample_bytree = 0.8;          ///< 特征子采样比例
    double reg_alpha = 0.0;                 ///< L1正则化
    double reg_lambda = 1.0;                ///< L2正则化
    int random_state = 42;                  ///< 随机种子
    bool early_stopping = true;             ///< 早停
    int early_stopping_rounds = 10;         ///< 早停轮数
};

/**
 * @brief 模型评估结果
 */
struct ModelEvaluation {
    double accuracy;                        ///< 准确率
    double precision;                       ///< 精确率
    double recall;                          ///< 召回率
    double f1_score;                        ///< F1分数
    double auc_score;                       ///< AUC分数
    double rmse;                            ///< 均方根误差
    double mae;                             ///< 平均绝对误差
    
    std::vector<std::pair<std::string, double>> confusion_matrix; ///< 混淆矩阵
    std::vector<std::pair<double, double>> roc_curve; ///< ROC曲线数据
};

/**
 * @brief 录取概率预测器类
 * 
 * 使用XGBoost算法进行录取概率预测
 * 支持特征工程、模型训练、预测和评估
 */
class AdmissionPredictor {
public:
    /**
     * @brief 构造函数
     */
    AdmissionPredictor();
    
    /**
     * @brief 析构函数
     */
    ~AdmissionPredictor();
    
    /**
     * @brief 初始化预测器
     * @param config_path 配置文件路径
     * @return 是否初始化成功
     */
    bool Initialize(const std::string& config_path);
    
    /**
     * @brief 加载预训练模型
     * @param model_path 模型文件路径
     * @return 是否加载成功
     */
    bool LoadModel(const std::string& model_path);
    
    /**
     * @brief 保存模型
     * @param model_path 保存路径
     * @return 是否保存成功
     */
    bool SaveModel(const std::string& model_path) const;
    
    /**
     * @brief 特征工程 - 提取学生特征
     * @param student 学生信息
     * @param university 大学信息
     * @param major 专业信息
     * @return 特征向量
     */
    FeatureVector ExtractFeatures(
        const Student& student,
        const University& university,
        const Major& major) const;
    
    /**
     * @brief 批量特征提取
     * @param students 学生列表
     * @param universities 大学列表
     * @param majors 专业列表
     * @return 特征向量列表
     */
    std::vector<FeatureVector> BatchExtractFeatures(
        const std::vector<Student>& students,
        const std::vector<University>& universities,
        const std::vector<Major>& majors) const;
    
    /**
     * @brief 训练模型
     * @param training_data 训练数据
     * @param validation_data 验证数据
     * @param params 训练参数
     * @return 是否训练成功
     */
    bool TrainModel(
        const std::vector<FeatureVector>& training_data,
        const std::vector<FeatureVector>& validation_data,
        const TrainingParams& params = TrainingParams());
    
    /**
     * @brief 预测录取概率
     * @param student 学生信息
     * @param university 大学信息
     * @param major 专业信息
     * @return 预测结果
     */
    PredictionResult PredictAdmissionProbability(
        const Student& student,
        const University& university,
        const Major& major) const;
    
    /**
     * @brief 批量预测
     * @param feature_vectors 特征向量列表
     * @return 预测结果列表
     */
    std::vector<PredictionResult> BatchPredict(
        const std::vector<FeatureVector>& feature_vectors) const;
    
    /**
     * @brief 评估模型性能
     * @param test_data 测试数据
     * @return 评估结果
     */
    ModelEvaluation EvaluateModel(const std::vector<FeatureVector>& test_data) const;
    
    /**
     * @brief 获取特征重要性
     * @return 特征重要性排序
     */
    std::vector<std::pair<std::string, double>> GetFeatureImportance() const;
    
    /**
     * @brief 交叉验证
     * @param data 数据集
     * @param k_folds 折数
     * @param params 训练参数
     * @return 交叉验证结果
     */
    ModelEvaluation CrossValidate(
        const std::vector<FeatureVector>& data,
        int k_folds = 5,
        const TrainingParams& params = TrainingParams()) const;
    
    /**
     * @brief 超参数优化
     * @param training_data 训练数据
     * @param validation_data 验证数据
     * @param param_grid 参数网格
     * @return 最优参数
     */
    TrainingParams HyperparameterTuning(
        const std::vector<FeatureVector>& training_data,
        const std::vector<FeatureVector>& validation_data,
        const std::unordered_map<std::string, std::vector<double>>& param_grid) const;
    
    /**
     * @brief 增量学习
     * @param new_data 新数据
     * @return 是否更新成功
     */
    bool IncrementalLearning(const std::vector<FeatureVector>& new_data);
    
    /**
     * @brief 生成预测解释
     * @param features 特征向量
     * @param prediction 预测结果
     * @return 解释文本
     */
    std::string GeneratePredictionExplanation(
        const FeatureVector& features,
        const PredictionResult& prediction) const;
    
    /**
     * @brief 设置特征选择器
     * @param selector 特征选择函数
     */
    void SetFeatureSelector(std::function<std::vector<int>(const std::vector<FeatureVector>&)> selector);
    
    /**
     * @brief 数据预处理
     * @param data 原始数据
     * @return 预处理后的数据
     */
    std::vector<FeatureVector> PreprocessData(const std::vector<FeatureVector>& data) const;
    
    /**
     * @brief 获取模型信息
     * @return 模型信息JSON字符串
     */
    std::string GetModelInfo() const;
    
    /**
     * @brief 模型是否已训练
     * @return 是否已训练
     */
    bool IsModelTrained() const;

private:
    // 内部实现类
    class Impl;
    std::unique_ptr<Impl> pimpl_;
    
    /**
     * @brief 标准化特征
     * @param features 原始特征
     * @return 标准化后的特征
     */
    std::vector<double> NormalizeFeatures(const std::vector<double>& features) const;
    
    /**
     * @brief 处理缺失值
     * @param features 特征向量
     * @return 处理后的特征
     */
    FeatureVector HandleMissingValues(const FeatureVector& features) const;
    
    /**
     * @brief 特征选择
     * @param data 数据集
     * @param top_k 选择前k个特征
     * @return 选择的特征索引
     */
    std::vector<int> SelectFeatures(const std::vector<FeatureVector>& data, int top_k) const;
    
    /**
     * @brief 计算模型准确度
     * @param predictions 预测结果
     * @param ground_truth 真实标签
     * @return 准确度
     */
    double CalculateAccuracy(
        const std::vector<double>& predictions,
        const std::vector<double>& ground_truth) const;
};

/**
 * @brief 特征工程工具类
 */
class FeatureEngineering {
public:
    /**
     * @brief 创建多项式特征
     * @param features 原始特征
     * @param degree 多项式度数
     * @return 多项式特征
     */
    static std::vector<double> CreatePolynomialFeatures(
        const std::vector<double>& features, int degree);
    
    /**
     * @brief 创建交互特征
     * @param features 原始特征
     * @return 交互特征
     */
    static std::vector<double> CreateInteractionFeatures(const std::vector<double>& features);
    
    /**
     * @brief 分箱处理
     * @param value 数值
     * @param bins 分箱边界
     * @return 分箱编码
     */
    static std::vector<double> BinEncoding(double value, const std::vector<double>& bins);
    
    /**
     * @brief 独热编码
     * @param category 类别
     * @param all_categories 所有类别
     * @return 独热编码向量
     */
    static std::vector<double> OneHotEncoding(
        const std::string& category, const std::vector<std::string>& all_categories);
    
    /**
     * @brief 标签编码
     * @param category 类别
     * @param category_map 类别映射
     * @return 编码值
     */
    static double LabelEncoding(
        const std::string& category, const std::unordered_map<std::string, int>& category_map);
    
    /**
     * @brief 特征缩放
     * @param features 原始特征
     * @param method 缩放方法 ("minmax"|"standard"|"robust")
     * @return 缩放后的特征
     */
    static std::vector<double> ScaleFeatures(
        const std::vector<double>& features, const std::string& method = "standard");
};

} // namespace volunteer_matcher

#endif // ADMISSION_PREDICTOR_H