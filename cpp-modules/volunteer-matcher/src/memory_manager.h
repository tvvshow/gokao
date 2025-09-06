#ifndef MEMORY_MANAGER_H
#define MEMORY_MANAGER_H

#include <memory>
#include <vector>
#include <unordered_map>
#include <mutex>
#include <atomic>
#include <functional>
#include "volunteer_matcher.h"

namespace volunteer_matcher {

/**
 * @brief 内存管理策略
 */
enum class MemoryManagementStrategy {
    DEFAULT,           // 默认策略
    POOL,              // 对象池
    CACHE,             // 缓存策略
    LAZY,              // 懒加载
    EAGER              // 预加载
};

/**
 * @brief 内存使用统计
 */
struct MemoryUsageStats {
    std::atomic<size_t> total_allocated{0};      // 总分配内存
    std::atomic<size_t> current_usage{0};        // 当前使用内存
    std::atomic<size_t> peak_usage{0};           // 峰值内存使用
    std::atomic<size_t> allocation_count{0};     // 分配次数
    std::atomic<size_t> deallocation_count{0};    // 释放次数
    std::atomic<size_t> leak_count{0};            // 内存泄漏计数
    
    // 按类型统计
    std::unordered_map<std::string, size_t> type_usage;
    
    void UpdatePeak() {
        size_t current = current_usage.load();
        size_t peak = peak_usage.load();
        while (current > peak && !peak_usage.compare_exchange_weak(peak, current)) {
            // CAS循环直到成功更新
        }
    }
};

/**
 * @brief 智能内存管理器
 * 
 * 提供RAII内存管理和对象池功能
 */
class MemoryManager {
public:
    static MemoryManager& GetInstance();
    
    ~MemoryManager();
    
    // 禁用拷贝和移动
    MemoryManager(const MemoryManager&) = delete;
    MemoryManager& operator=(const MemoryManager&) = delete;
    MemoryManager(MemoryManager&&) = delete;
    MemoryManager& operator=(MemoryManager&&) = delete;
    
    /**
     * @brief 配置内存管理策略
     */
    void Configure(MemoryManagementStrategy strategy, size_t pool_size = 1000);
    
    /**
     * @brief 分配内存（带类型跟踪）
     */
    template<typename T, typename... Args>
    std::shared_ptr<T> Create(Args&&... args);
    
    /**
     * @brief 创建对象（使用对象池）
     */
    template<typename T, typename... Args>
    std::shared_ptr<T> CreatePooled(Args&&... args);
    
    /**
     * @brief 预加载对象到缓存
     */
    template<typename T>
    void Preload(size_t count, std::function<T*()> factory);
    
    /**
     * @brief 清理内存
     */
    void Cleanup();
    
    /**
     * @brief 获取内存统计
     */
    MemoryUsageStats GetStats() const;
    
    /**
     * @brief 检查内存泄漏
     */
    bool CheckForLeaks() const;
    
    /**
     * @brief 设置内存限制
     */
    void SetMemoryLimit(size_t limit_bytes);
    
    /**
     * @brief 注册自定义清理器
     */
    void RegisterCleanup(std::function<void()> cleanup_func);
    
private:
    MemoryManager();
    
    // 对象池实现
    template<typename T>
    class ObjectPool {
    public:
        ObjectPool(size_t initial_size = 100);
        ~ObjectPool();
        
        std::shared_ptr<T> Acquire();
        void Release(std::unique_ptr<T> object);
        size_t Size() const;
        void Resize(size_t new_size);
        
    private:
        std::vector<std::unique_ptr<T>> pool_;
        std::mutex mutex_;
        size_t max_size_;
    };
    
    // 内存跟踪器
    class AllocationTracker {
    public:
        void* Allocate(size_t size, const std::string& type_name);
        void Deallocate(void* ptr, size_t size, const std::string& type_name);
        
        MemoryUsageStats GetStats() const;
        
    private:
        mutable std::mutex mutex_;
        MemoryUsageStats stats_;
        std::unordered_map<void*, std::pair<size_t, std::string>> allocations_;
    };
    
    MemoryManagementStrategy strategy_;
    mutable std::mutex mutex_;
    AllocationTracker tracker_;
    std::vector<std::function<void()>> cleanup_registry_;
    size_t memory_limit_;
    
    // 对象池实例
    std::unordered_map<std::string, std::shared_ptr<void>> object_pools_;
};

// 模板方法实现

template<typename T, typename... Args>
std::shared_ptr<T> MemoryManager::Create(Args&&... args) {
    // 使用make_shared并跟踪内存
    auto ptr = std::make_shared<T>(std::forward<Args>(args)...);
    
    // 跟踪内存分配
    size_t size = sizeof(T);
    void* raw_ptr = ptr.get();
    tracker_.Allocate(size, typeid(T).name());
    
    // 设置自定义删除器来跟踪释放
    return std::shared_ptr<T>(ptr.get(), 
        [this, size, raw_ptr, type_name = typeid(T).name()](T*) {
            tracker_.Deallocate(raw_ptr, size, type_name);
            // 实际对象由shared_ptr管理，这里只做跟踪
        });
}

template<typename T, typename... Args>
std::shared_ptr<T> MemoryManager::CreatePooled(Args&&... args) {
    // 获取或创建对象池
    std::string type_name = typeid(T).name();
    
    std::lock_guard<std::mutex> lock(mutex_);
    
    auto it = object_pools_.find(type_name);
    if (it == object_pools_.end()) {
        // 创建新的对象池
        auto pool = std::make_shared<ObjectPool<T>>(100);
        object_pools_[type_name] = pool;
        it = object_pools_.find(type_name);
    }
    
    // 从对象池获取对象
    auto pooled_ptr = std::static_pointer_cast<ObjectPool<T>>(it->second)->Acquire();
    if (!pooled_ptr) {
        // 对象池为空，创建新对象
        return Create<T>(std::forward<Args>(args)...);
    }
    
    // 使用placement new重新初始化对象
    new(pooled_ptr.get()) T(std::forward<Args>(args)...);
    
    return pooled_ptr;
}

template<typename T>
void MemoryManager::Preload(size_t count, std::function<T*()> factory) {
    std::vector<std::shared_ptr<T>> objects;
    objects.reserve(count);
    
    for (size_t i = 0; i < count; ++i) {
        objects.push_back(Create<T>([&factory]() { return factory(); }));
    }
    
    // 对象会在离开作用域时自动管理
}

// ObjectPool 模板实现

template<typename T>
MemoryManager::ObjectPool<T>::ObjectPool(size_t initial_size) 
    : max_size_(initial_size * 2) {
    pool_.reserve(initial_size);
    for (size_t i = 0; i < initial_size; ++i) {
        pool_.push_back(std::make_unique<T>());
    }
}

template<typename T>
MemoryManager::ObjectPool<T>::~ObjectPool() {
    pool_.clear();
}

template<typename T>
std::shared_ptr<T> MemoryManager::ObjectPool<T>::Acquire() {
    std::lock_guard<std::mutex> lock(mutex_);
    
    if (pool_.empty()) {
        return nullptr;
    }
    
    auto ptr = std::move(pool_.back());
    pool_.pop_back();
    
    // 使用自定义删除器返回对象
    return std::shared_ptr<T>(ptr.release(), 
        [this](T* obj) {
            // 对象返回到池中
            std::unique_ptr<T> unique_obj(obj);
            Release(std::move(unique_obj));
        });
}

template<typename T>
void MemoryManager::ObjectPool<T>::Release(std::unique_ptr<T> object) {
    std::lock_guard<std::mutex> lock(mutex_);
    
    if (pool_.size() < max_size_) {
        pool_.push_back(std::move(object));
    }
    // 如果池已满，对象会被自动销毁
}

template<typename T>
size_t MemoryManager::ObjectPool<T>::Size() const {
    std::lock_guard<std::mutex> lock(mutex_);
    return pool_.size();
}

template<typename T>
void MemoryManager::ObjectPool<T>::Resize(size_t new_size) {
    std::lock_guard<std::mutex> lock(mutex_);
    max_size_ = new_size;
    
    // 如果当前大小超过新限制，移除多余对象
    if (pool_.size() > new_size) {
        pool_.resize(new_size);
    }
}

/**
 * @brief RAII内存管理包装器
 */
template<typename T>
class ScopedMemory {
public:
    template<typename... Args>
    explicit ScopedMemory(Args&&... args)
        : object_(MemoryManager::GetInstance().Create<T>(std::forward<Args>(args)...)) {}
    
    T* operator->() const { return object_.get(); }
    T& operator*() const { return *object_; }
    T* get() const { return object_.get(); }
    
    // 显式释放内存
    void Release() { object_.reset(); }
    
private:
    std::shared_ptr<T> object_;
};

/**
 * @brief 智能资源句柄（类似unique_ptr但支持自定义清理）
 */
template<typename T, typename Cleanup = std::function<void(T*)>>
class ResourceHandle {
public:
    ResourceHandle(T* resource = nullptr, Cleanup cleanup = {})
        : resource_(resource), cleanup_(std::move(cleanup)) {}
    
    ~ResourceHandle() { Reset(); }
    
    // 禁用拷贝
    ResourceHandle(const ResourceHandle&) = delete;
    ResourceHandle& operator=(const ResourceHandle&) = delete;
    
    // 支持移动
    ResourceHandle(ResourceHandle&& other) noexcept
        : resource_(other.resource_), cleanup_(std::move(other.cleanup_)) {
        other.resource_ = nullptr;
    }
    
    ResourceHandle& operator=(ResourceHandle&& other) noexcept {
        if (this != &other) {
            Reset();
            resource_ = other.resource_;
            cleanup_ = std::move(other.cleanup_);
            other.resource_ = nullptr;
        }
        return *this;
    }
    
    T* get() const { return resource_; }
    T* operator->() const { return resource_; }
    T& operator*() const { return *resource_; }
    
    explicit operator bool() const { return resource_ != nullptr; }
    
    void Reset(T* new_resource = nullptr) {
        if (resource_ && cleanup_) {
            cleanup_(resource_);
        }
        resource_ = new_resource;
    }
    
    T* Release() {
        T* released = resource_;
        resource_ = nullptr;
        return released;
    }
    
private:
    T* resource_;
    Cleanup cleanup_;
};

} // namespace volunteer_matcher

#endif // MEMORY_MANAGER_H