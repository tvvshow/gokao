/**
 * @file c_interface.h
 * @brief C接口定义，供Go语言通过CGO调用
 * @author 高考志愿填报系统开发团队
 * @version 1.0.0
 * @date 2025-01-18
 */

#ifndef C_INTERFACE_H
#define C_INTERFACE_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stddef.h>

/**
 * @brief 错误码定义(C风格)
 */
typedef enum {
    C_SUCCESS = 0,                  ///< 成功
    C_ERROR_INIT_FAILED = 1001,     ///< 初始化失败
    C_ERROR_INVALID_PARAM = 1002,   ///< 无效参数
    C_ERROR_MEMORY_ALLOC = 1003,    ///< 内存分配失败
    C_ERROR_HARDWARE_ACCESS = 1004, ///< 硬件访问失败
    C_ERROR_SYSTEM_INFO = 1005,     ///< 系统信息获取失败
    C_ERROR_ENCRYPTION = 1006,      ///< 加密操作失败
    C_ERROR_PERMISSION_DENIED = 1007, ///< 权限不足
    C_ERROR_PLATFORM_UNSUPPORTED = 1008, ///< 平台不支持
    C_ERROR_UNKNOWN = 9999          ///< 未知错误
} CErrorCode;

/**
 * @brief C风格设备指纹结构体
 */
typedef struct {
    char device_id[64];             ///< 设备唯一标识
    char device_type[32];           ///< 设备类型
    char cpu_id[64];                ///< CPU标识符
    char cpu_model[128];            ///< CPU型号
    unsigned int cpu_cores;         ///< CPU核心数
    unsigned long long total_memory; ///< 总内存大小(字节)
    char motherboard_serial[64];    ///< 主板序列号
    char os_type[32];               ///< 操作系统类型
    char os_version[64];            ///< 操作系统版本
    char hostname[64];              ///< 主机名
    char username[64];              ///< 用户名
    char screen_resolution[32];     ///< 屏幕分辨率
    char fingerprint_hash[128];     ///< 指纹哈希值
    unsigned int confidence_score;  ///< 置信度分数(0-100)
    char error_message[256];        ///< 错误信息
} CDeviceFingerprint;

/**
 * @brief C风格配置结构体
 */
typedef struct {
    int collect_sensitive_info;     ///< 是否采集敏感信息(1是/0否)
    int enable_encryption;          ///< 是否启用加密(1是/0否)
    int enable_signature;           ///< 是否启用签名(1是/0否)
    char encryption_key[64];        ///< 加密密钥
    int timeout_seconds;            ///< 超时时间(秒)
} CConfiguration;

/**
 * @brief C风格性能统计结构体
 */
typedef struct {
    unsigned long long collect_time_us;    ///< 采集耗时(微秒)
    unsigned long long hash_time_us;       ///< 哈希计算耗时(微秒)
    unsigned long long encryption_time_us; ///< 加密耗时(微秒)
    unsigned int total_calls;              ///< 总调用次数
    unsigned int success_calls;            ///< 成功调用次数
    unsigned int error_calls;              ///< 错误调用次数
} CPerformanceStats;

// =============================================================================
// 核心功能接口
// =============================================================================

/**
 * @brief 初始化设备指纹采集器
 * @param config_path 配置文件路径(可为NULL)
 * @return 错误码
 */
CErrorCode DeviceFingerprint_Initialize(const char* config_path);

/**
 * @brief 反初始化设备指纹采集器
 */
void DeviceFingerprint_Uninitialize(void);

/**
 * @brief 采集设备指纹
 * @param fingerprint 输出的设备指纹结构体指针
 * @return 错误码
 */
CErrorCode DeviceFingerprint_Collect(CDeviceFingerprint* fingerprint);

/**
 * @brief 快速采集设备指纹(无需初始化)
 * @param fingerprint 输出的设备指纹结构体指针
 * @return 错误码
 */
CErrorCode DeviceFingerprint_QuickCollect(CDeviceFingerprint* fingerprint);

// =============================================================================
// 配置和管理接口
// =============================================================================

/**
 * @brief 设置采集配置
 * @param config 配置结构体指针
 * @return 错误码
 */
CErrorCode DeviceFingerprint_SetConfiguration(const CConfiguration* config);

/**
 * @brief 获取当前配置
 * @param config 输出的配置结构体指针
 * @return 错误码
 */
CErrorCode DeviceFingerprint_GetConfiguration(CConfiguration* config);

// =============================================================================
// 哈希和比较接口
// =============================================================================

/**
 * @brief 生成设备指纹哈希
 * @param fingerprint 设备指纹结构体指针
 * @param hash_buffer 输出的哈希缓冲区
 * @param buffer_size 缓冲区大小
 * @return 错误码
 */
CErrorCode DeviceFingerprint_GenerateHash(const CDeviceFingerprint* fingerprint,
                                         char* hash_buffer,
                                         size_t buffer_size);

/**
 * @brief 比较两个设备指纹
 * @param fingerprint1 第一个设备指纹
 * @param fingerprint2 第二个设备指纹
 * @param similarity_score 输出的相似度分数(0.0-1.0)
 * @param is_same_device 输出是否为同一设备(1表示是，0表示否)
 * @return 错误码
 */
CErrorCode DeviceFingerprint_Compare(const CDeviceFingerprint* fingerprint1,
                                    const CDeviceFingerprint* fingerprint2,
                                    double* similarity_score,
                                    int* is_same_device);

/**
 * @brief 验证设备指纹
 * @param fingerprint 设备指纹结构体指针
 * @param reference_hash 参考哈希值
 * @param is_valid 输出验证结果(1表示有效，0表示无效)
 * @return 错误码
 */
CErrorCode DeviceFingerprint_Validate(const CDeviceFingerprint* fingerprint,
                                     const char* reference_hash,
                                     int* is_valid);

// =============================================================================
// 加密和签名接口
// =============================================================================

/**
 * @brief 加密设备指纹数据
 * @param data 待加密数据
 * @param data_size 数据大小
 * @param key 加密密钥
 * @param encrypted_buffer 输出的加密缓冲区
 * @param buffer_size 缓冲区大小
 * @param actual_size 实际加密数据大小
 * @return 错误码
 */
CErrorCode DeviceFingerprint_Encrypt(const char* data,
                                    size_t data_size,
                                    const char* key,
                                    char* encrypted_buffer,
                                    size_t buffer_size,
                                    size_t* actual_size);

/**
 * @brief 解密设备指纹数据
 * @param encrypted_data 加密数据
 * @param data_size 数据大小
 * @param key 解密密钥
 * @param decrypted_buffer 输出的解密缓冲区
 * @param buffer_size 缓冲区大小
 * @param actual_size 实际解密数据大小
 * @return 错误码
 */
CErrorCode DeviceFingerprint_Decrypt(const char* encrypted_data,
                                    size_t data_size,
                                    const char* key,
                                    char* decrypted_buffer,
                                    size_t buffer_size,
                                    size_t* actual_size);

/**
 * @brief 生成数字签名
 * @param data 待签名数据
 * @param data_size 数据大小
 * @param private_key 私钥
 * @param signature_buffer 输出的签名缓冲区
 * @param buffer_size 缓冲区大小
 * @param actual_size 实际签名大小
 * @return 错误码
 */
CErrorCode DeviceFingerprint_Sign(const char* data,
                                 size_t data_size,
                                 const char* private_key,
                                 char* signature_buffer,
                                 size_t buffer_size,
                                 size_t* actual_size);

/**
 * @brief 验证数字签名
 * @param data 原始数据
 * @param data_size 数据大小
 * @param signature 数字签名
 * @param signature_size 签名大小
 * @param public_key 公钥
 * @param is_valid 输出验证结果(1表示有效，0表示无效)
 * @return 错误码
 */
CErrorCode DeviceFingerprint_VerifySignature(const char* data,
                                            size_t data_size,
                                            const char* signature,
                                            size_t signature_size,
                                            const char* public_key,
                                            int* is_valid);

// =============================================================================
// 许可证验证接口
// =============================================================================

/**
 * @brief 验证许可证
 * @param license_data 许可证数据
 * @param device_id 设备ID
 * @param is_valid 输出验证结果(1表示有效，0表示无效)
 * @param expires_at 输出许可证过期时间(Unix时间戳)
 * @return 错误码
 */
CErrorCode DeviceFingerprint_ValidateLicense(const char* license_data,
                                            const char* device_id,
                                            int* is_valid,
                                            long long* expires_at);

/**
 * @brief 生成设备许可证
 * @param device_id 设备ID
 * @param expires_at 过期时间(Unix时间戳)
 * @param private_key 私钥
 * @param license_buffer 输出的许可证缓冲区
 * @param buffer_size 缓冲区大小
 * @return 错误码
 */
CErrorCode DeviceFingerprint_GenerateLicense(const char* device_id,
                                            long long expires_at,
                                            const char* private_key,
                                            char* license_buffer,
                                            size_t buffer_size);

// =============================================================================
// 序列化接口
// =============================================================================

/**
 * @brief 序列化设备指纹为JSON字符串
 * @param fingerprint 设备指纹结构体指针
 * @param json_buffer 输出的JSON缓冲区
 * @param buffer_size 缓冲区大小
 * @return 错误码
 */
CErrorCode DeviceFingerprint_SerializeToJson(const CDeviceFingerprint* fingerprint,
                                            char* json_buffer,
                                            size_t buffer_size);

/**
 * @brief 从JSON字符串反序列化设备指纹
 * @param json_data JSON字符串
 * @param fingerprint 输出的设备指纹结构体指针
 * @return 错误码
 */
CErrorCode DeviceFingerprint_DeserializeFromJson(const char* json_data,
                                                CDeviceFingerprint* fingerprint);

// =============================================================================
// 安全检测接口
// =============================================================================

/**
 * @brief 检查是否存在调试器
 * @param is_debugger_present 输出结果(1表示存在，0表示不存在)
 * @return 错误码
 */
CErrorCode DeviceFingerprint_IsDebuggerPresent(int* is_debugger_present);

/**
 * @brief 检查是否在虚拟机中运行
 * @param is_virtual_machine 输出结果(1表示是，0表示否)
 * @return 错误码
 */
CErrorCode DeviceFingerprint_IsVirtualMachine(int* is_virtual_machine);

/**
 * @brief 检查运行环境安全性
 * @param security_level 输出安全级别(0-100)
 * @param risk_factors 输出风险因素描述缓冲区
 * @param buffer_size 缓冲区大小
 * @return 错误码
 */
CErrorCode DeviceFingerprint_CheckSecurity(int* security_level,
                                          char* risk_factors,
                                          size_t buffer_size);

// =============================================================================
// 性能监控接口
// =============================================================================

/**
 * @brief 获取性能统计信息
 * @param stats 输出的性能统计结构体指针
 * @return 错误码
 */
CErrorCode DeviceFingerprint_GetPerformanceStats(CPerformanceStats* stats);

/**
 * @brief 重置性能统计信息
 * @return 错误码
 */
CErrorCode DeviceFingerprint_ResetPerformanceStats(void);

/**
 * @brief 启用/禁用性能监控
 * @param enable 是否启用(1启用/0禁用)
 * @return 错误码
 */
CErrorCode DeviceFingerprint_SetPerformanceMonitoring(int enable);

// =============================================================================
// 工具函数接口
// =============================================================================

/**
 * @brief 获取库版本
 * @param version_buffer 输出的版本缓冲区
 * @param buffer_size 缓冲区大小
 * @return 错误码
 */
CErrorCode DeviceFingerprint_GetVersion(char* version_buffer, size_t buffer_size);

/**
 * @brief 获取错误描述
 * @param error_code 错误码
 * @param error_buffer 输出的错误描述缓冲区
 * @param buffer_size 缓冲区大小
 * @return 错误码
 */
CErrorCode DeviceFingerprint_GetErrorDescription(CErrorCode error_code,
                                                char* error_buffer,
                                                size_t buffer_size);

/**
 * @brief 检查库初始化状态
 * @param is_initialized 输出初始化状态(1已初始化/0未初始化)
 * @return 错误码
 */
CErrorCode DeviceFingerprint_IsInitialized(int* is_initialized);

/**
 * @brief 获取支持的平台列表
 * @param platforms_buffer 输出的平台列表缓冲区(逗号分隔)
 * @param buffer_size 缓冲区大小
 * @return 错误码
 */
CErrorCode DeviceFingerprint_GetSupportedPlatforms(char* platforms_buffer,
                                                   size_t buffer_size);

#ifdef __cplusplus
}
#endif

#endif // C_INTERFACE_H