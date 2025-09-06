#ifndef EXCEPTIONS_H
#define EXCEPTIONS_H

#include <exception>
#include <string>
#include <memory>
#include <system_error>
#include <fmt/format.h>

namespace volunteer_matcher {

/**
 * @brief 基础异常类
 */
class BaseException : public std::exception {
public:
    explicit BaseException(const std::string& message, 
                         const std::string& component = "",
                         int error_code = 0)
        : message_(message), component_(component), error_code_(error_code) {
        full_message_ = FormatMessage();
    }
    
    const char* what() const noexcept override {
        return full_message_.c_str();
    }
    
    const std::string& GetMessage() const { return message_; }
    const std::string& GetComponent() const { return component_; }
    int GetErrorCode() const { return error_code_; }
    
    virtual std::string GetErrorType() const { return "BaseException"; }
    
protected:
    virtual std::string FormatMessage() const {
        if (component_.empty()) {
            return fmt::format("{} [Error: {}]", message_, error_code_);
        }
        return fmt::format("{}: {} [Error: {}]", component_, message_, error_code_);
    }
    
    std::string message_;
    std::string component_;
    int error_code_;
    std::string full_message_;
};

/**
 * @brief 配置异常
 */
class ConfigurationException : public BaseException {
public:
    explicit ConfigurationException(const std::string& message,
                                  const std::string& config_key = "",
                                  int error_code = 1001)
        : BaseException(message, "Configuration", error_code), config_key_(config_key) {}
    
    std::string GetErrorType() const override { return "ConfigurationException"; }
    
    const std::string& GetConfigKey() const { return config_key_; }
    
protected:
    std::string FormatMessage() const override {
        if (config_key_.empty()) {
            return BaseException::FormatMessage();
        }
        return fmt::format("{}: {} (Key: {}) [Error: {}]", 
                          GetComponent(), message_, config_key_, error_code_);
    }
    
private:
    std::string config_key_;
};

/**
 * @brief 数据加载异常
 */
class DataLoadException : public BaseException {
public:
    explicit DataLoadException(const std::string& message,
                             const std::string& data_source = "",
                             int error_code = 2001)
        : BaseException(message, "DataLoad", error_code), data_source_(data_source) {}
    
    std::string GetErrorType() const override { return "DataLoadException"; }
    
    const std::string& GetDataSource() const { return data_source_; }
    
protected:
    std::string FormatMessage() const override {
        if (data_source_.empty()) {
            return BaseException::FormatMessage();
        }
        return fmt::format("{}: {} (Source: {}) [Error: {}]", 
                          GetComponent(), message_, data_source_, error_code_);
    }
    
private:
    std::string data_source_;
};

/**
 * @brief 算法异常
 */
class AlgorithmException : public BaseException {
public:
    explicit AlgorithmException(const std::string& message,
                              const std::string& algorithm_name = "",
                              int error_code = 3001)
        : BaseException(message, "Algorithm", error_code), algorithm_name_(algorithm_name) {}
    
    std::string GetErrorType() const override { return "AlgorithmException"; }
    
    const std::string& GetAlgorithmName() const { return algorithm_name_; }
    
protected:
    std::string FormatMessage() const override {
        if (algorithm_name_.empty()) {
            return BaseException::FormatMessage();
        }
        return fmt::format("{}: {} (Algorithm: {}) [Error: {}]", 
                          GetComponent(), message_, algorithm_name_, error_code_);
    }
    
private:
    std::string algorithm_name_;
};

/**
 * @brief 内存异常
 */
class MemoryException : public BaseException {
public:
    explicit MemoryException(const std::string& message,
                           size_t allocated_size = 0,
                           int error_code = 4001)
        : BaseException(message, "Memory", error_code), allocated_size_(allocated_size) {}
    
    std::string GetErrorType() const override { return "MemoryException"; }
    
    size_t GetAllocatedSize() const { return allocated_size_; }
    
protected:
    std::string FormatMessage() const override {
        if (allocated_size_ == 0) {
            return BaseException::FormatMessage();
        }
        return fmt::format("{}: {} (Size: {} bytes) [Error: {}]", 
                          GetComponent(), message_, allocated_size_, error_code_);
    }
    
private:
    size_t allocated_size_;
};

/**
 * @brief 线程/并发异常
 */
class ConcurrencyException : public BaseException {
public:
    explicit ConcurrencyException(const std::string& message,
                                const std::string& thread_id = "",
                                int error_code = 5001)
        : BaseException(message, "Concurrency", error_code), thread_id_(thread_id) {}
    
    std::string GetErrorType() const override { return "ConcurrencyException"; }
    
    const std::string& GetThreadId() const { return thread_id_; }
    
protected:
    std::string FormatMessage() const override {
        if (thread_id_.empty()) {
            return BaseException::FormatMessage();
        }
        return fmt::format("{}: {} (Thread: {}) [Error: {}]", 
                          GetComponent(), message_, thread_id_, error_code_);
    }
    
private:
    std::string thread_id_;
};

/**
 * @brief AI模型异常
 */
class AIModelException : public BaseException {
public:
    explicit AIModelException(const std::string& message,
                            const std::string& model_name = "",
                            int error_code = 6001)
        : BaseException(message, "AIModel", error_code), model_name_(model_name) {}
    
    std::string GetErrorType() const override { return "AIModelException"; }
    
    const std::string& GetModelName() const { return model_name_; }
    
protected:
    std::string FormatMessage() const override {
        if (model_name_.empty()) {
            return BaseException::FormatMessage();
        }
        return fmt::format("{}: {} (Model: {}) [Error: {}]", 
                          GetComponent(), message_, model_name_, error_code_);
    }
    
private:
    std::string model_name_;
};

/**
 * @brief 验证异常
 */
class ValidationException : public BaseException {
public:
    explicit ValidationException(const std::string& message,
                               const std::string& field_name = "",
                               int error_code = 7001)
        : BaseException(message, "Validation", error_code), field_name_(field_name) {}
    
    std::string GetErrorType() const override { return "ValidationException"; }
    
    const std::string& GetFieldName() const { return field_name_; }
    
protected:
    std::string FormatMessage() const override {
        if (field_name_.empty()) {
            return BaseException::FormatMessage();
        }
        return fmt::format("{}: {} (Field: {}) [Error: {}]", 
                          GetComponent(), message_, field_name_, error_code_);
    }
    
private:
    std::string field_name_;
};

/**
 * @brief 网络/IO异常
 */
class IOException : public BaseException {
public:
    explicit IOException(const std::string& message,
                       const std::string& file_path = "",
                       int error_code = 8001)
        : BaseException(message, "IO", error_code), file_path_(file_path) {}
    
    std::string GetErrorType() const override { return "IOException"; }
    
    const std::string& GetFilePath() const { return file_path_; }
    
protected:
    std::string FormatMessage() const override {
        if (file_path_.empty()) {
            return BaseException::FormatMessage();
        }
        return fmt::format("{}: {} (File: {}) [Error: {}]", 
                          GetComponent(), message_, file_path_, error_code_);
    }
    
private:
    std::string file_path_;
};

/**
 * @brief 超时异常
 */
class TimeoutException : public BaseException {
public:
    explicit TimeoutException(const std::string& message,
                            int timeout_ms = 0,
                            int error_code = 9001)
        : BaseException(message, "Timeout", error_code), timeout_ms_(timeout_ms) {}
    
    std::string GetErrorType() const override { return "TimeoutException"; }
    
    int GetTimeoutMs() const { return timeout_ms_; }
    
protected:
    std::string FormatMessage() const override {
        if (timeout_ms_ == 0) {
            return BaseException::FormatMessage();
        }
        return fmt::format("{}: {} (Timeout: {}ms) [Error: {}]", 
                          GetComponent(), message_, timeout_ms_, error_code_);
    }
    
private:
    int timeout_ms_;
};

/**
 * @brief 错误码定义
 */
namespace ErrorCodes {
    // 配置错误 (1000-1999)
    constexpr int CONFIG_FILE_NOT_FOUND = 1001;
    constexpr int CONFIG_PARSE_ERROR = 1002;
    constexpr int CONFIG_VALIDATION_ERROR = 1003;
    
    // 数据加载错误 (2000-2999)
    constexpr int DATA_FILE_NOT_FOUND = 2001;
    constexpr int DATA_FORMAT_ERROR = 2002;
    constexpr int DATA_VALIDATION_ERROR = 2003;
    constexpr int DATABASE_CONNECTION_ERROR = 2004;
    
    // 算法错误 (3000-3999)
    constexpr int ALGORITHM_INIT_ERROR = 3001;
    constexpr int ALGORITHM_RUNTIME_ERROR = 3002;
    constexpr int ALGORITHM_PARAMETER_ERROR = 3003;
    
    // 内存错误 (4000-4999)
    constexpr int MEMORY_ALLOCATION_FAILED = 4001;
    constexpr int MEMORY_LIMIT_EXCEEDED = 4002;
    constexpr int MEMORY_LEAK_DETECTED = 4003;
    
    // 并发错误 (5000-5999)
    constexpr int THREAD_CREATION_FAILED = 5001;
    constexpr int DEADLOCK_DETECTED = 5002;
    constexpr int RACE_CONDITION_DETECTED = 5003;
    
    // AI模型错误 (6000-6999)
    constexpr int MODEL_LOAD_FAILED = 6001;
    constexpr int MODEL_PREDICTION_ERROR = 6002;
    constexpr int MODEL_VERSION_MISMATCH = 6003;
    
    // 验证错误 (7000-7999)
    constexpr int INPUT_VALIDATION_FAILED = 7001;
    constexpr int DATA_VALIDATION_FAILED = 7002;
    constexpr int BUSINESS_RULE_VIOLATION = 7003;
    
    // IO错误 (8000-8999)
    constexpr int FILE_READ_ERROR = 8001;
    constexpr int FILE_WRITE_ERROR = 8002;
    constexpr int NETWORK_ERROR = 8003;
    
    // 超时错误 (9000-9999)
    constexpr int OPERATION_TIMEOUT = 9001;
    constexpr int REQUEST_TIMEOUT = 9002;
    constexpr int DATABASE_TIMEOUT = 9003;
}

/**
 * @brief 异常工具函数
 */
namespace ExceptionUtils {
    
    /**
     * @brief 安全执行函数，捕获异常并转换为错误码
     */
    template<typename Func, typename... Args>
    auto SafeExecute(Func&& func, Args&&... args) 
        -> std::pair<int, std::optional<decltype(func(std::forward<Args>(args)...)>> {
        
        try {
            auto result = func(std::forward<Args>(args)...);
            return {0, std::move(result)};
        } catch (const BaseException& e) {
            return {e.GetErrorCode(), std::nullopt};
        } catch (const std::exception& e) {
            return {-1, std::nullopt}; // 通用错误
        } catch (...) {
            return {-2, std::nullopt}; // 未知错误
        }
    }
    
    /**
     * @brief 重新抛出异常并添加上下文信息
     */
    template<typename ExceptionT, typename... Args>
    void RethrowWithContext(const std::string& context, Args&&... args) {
        try {
            throw;
        } catch (const BaseException& e) {
            throw ExceptionT(
                fmt::format("{}: {}", context, e.GetMessage()),
                e.GetComponent(),
                e.GetErrorCode()
            );
        } catch (const std::exception& e) {
            throw ExceptionT(
                fmt::format("{}: {}", context, e.what()),
                "",
                -1
            );
        } catch (...) {
            throw ExceptionT(
                fmt::format("{}: Unknown error", context),
                "",
                -2
            );
        }
    }
    
    /**
     * @brief 检查条件，不满足时抛出异常
     */
    template<typename ExceptionT, typename... Args>
    void CheckCondition(bool condition, const std::string& message, Args&&... args) {
        if (!condition) {
            throw ExceptionT(fmt::format(message, std::forward<Args>(args)...));
        }
    }
    
    /**
     * @brief 创建异常指针（用于跨线程异常传递）
     */
    std::unique_ptr<BaseException> MakeExceptionPtr(const std::exception& e);
    
    /**
     * @brief 从异常指针重新抛出异常
     */
    void RethrowFromPtr(const std::unique_ptr<BaseException>& exception_ptr);
    
} // namespace ExceptionUtils

/**
 * @brief 异常安全资源管理
 */
template<typename Resource>
class ExceptionSafeResource {
public:
    template<typename... Args>
    explicit ExceptionSafeResource(Args&&... args) {
        resource_ = std::make_unique<Resource>(std::forward<Args>(args)...);
    }
    
    ~ExceptionSafeResource() {
        if (resource_) {
            try {
                // 安全清理资源
                resource_->Cleanup();
            } catch (...) {
                // 析构函数中不能抛出异常
                // 可以记录日志但不能传播异常
            }
        }
    }
    
    Resource* operator->() const { return resource_.get(); }
    Resource& operator*() const { return *resource_; }
    
    // 禁用拷贝
    ExceptionSafeResource(const ExceptionSafeResource&) = delete;
    ExceptionSafeResource& operator=(const ExceptionSafeResource&) = delete;
    
    // 支持移动
    ExceptionSafeResource(ExceptionSafeResource&& other) noexcept
        : resource_(std::move(other.resource_)) {}
    
    ExceptionSafeResource& operator=(ExceptionSafeResource&& other) noexcept {
        if (this != &other) {
            resource_ = std::move(other.resource_);
        }
        return *this;
    }
    
private:
    std::unique_ptr<Resource> resource_;
};

} // namespace volunteer_matcher

#endif // EXCEPTIONS_H