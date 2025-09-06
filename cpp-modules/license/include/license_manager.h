#pragma once

#include <string>
#include <memory>
#include <vector>
#include <ctime>

namespace gaokaohub {
namespace license {

// 许可证状态枚举
enum class LicenseStatus {
    VALID,          // 有效
    EXPIRED,        // 已过期
    REVOKED,        // 已撤销
    INVALID,        // 无效
    DEVICE_MISMATCH // 设备不匹配
};

// 许可证信息结构
struct LicenseInfo {
    std::string license_key;        // 许可证密钥
    std::string user_id;           // 用户ID
    std::string device_id;         // 设备ID
    std::string device_fingerprint; // 设备指纹
    std::string plan_code;         // 套餐代码
    std::time_t issue_time;        // 签发时间
    std::time_t expire_time;       // 过期时间
    int max_bind_count;            // 最大绑定设备数
    int current_bind_count;        // 当前绑定设备数
    std::string encrypted_data;    // 加密数据
    std::string signature;         // 数字签名
};

// 设备绑定信息
struct DeviceBindInfo {
    std::string device_id;
    std::string device_fingerprint;
    std::time_t bind_time;
    bool is_active;
};

// 许可证管理器接口
class ILicenseManager {
public:
    virtual ~ILicenseManager() = default;

    // 生成许可证
    virtual std::string GenerateLicense(
        const std::string& user_id,
        const std::string& device_id,
        const std::string& device_fingerprint,
        const std::string& plan_code,
        std::time_t expire_time,
        int max_bind_count = 3
    ) = 0;

    // 验证许可证
    virtual LicenseStatus ValidateLicense(
        const std::string& license_key,
        const std::string& device_id,
        const std::string& device_fingerprint
    ) = 0;

    // 解析许可证信息
    virtual std::unique_ptr<LicenseInfo> ParseLicense(
        const std::string& license_key
    ) = 0;

    // 绑定设备
    virtual bool BindDevice(
        const std::string& license_key,
        const std::string& device_id,
        const std::string& device_fingerprint
    ) = 0;

    // 解绑设备
    virtual bool UnbindDevice(
        const std::string& license_key,
        const std::string& device_id
    ) = 0;

    // 获取绑定设备列表
    virtual std::vector<DeviceBindInfo> GetBoundDevices(
        const std::string& license_key
    ) = 0;

    // 撤销许可证
    virtual bool RevokeLicense(const std::string& license_key) = 0;

    // 检查许可证是否过期
    virtual bool IsLicenseExpired(const std::string& license_key) = 0;

    // 获取许可证剩余时间（秒）
    virtual long GetRemainingTime(const std::string& license_key) = 0;
};

// 许可证管理器实现
class LicenseManager : public ILicenseManager {
private:
    std::string private_key_;       // RSA私钥
    std::string public_key_;        // RSA公钥
    std::string aes_key_;           // AES密钥
    std::string license_db_path_;   // 许可证数据库路径

    // 内部实现类
    class Impl;
    std::unique_ptr<Impl> pImpl_;

public:
    // 构造函数
    explicit LicenseManager(
        const std::string& private_key,
        const std::string& public_key,
        const std::string& aes_key,
        const std::string& license_db_path = ""
    );

    // 析构函数
    ~LicenseManager();

    // 禁用拷贝构造和赋值
    LicenseManager(const LicenseManager&) = delete;
    LicenseManager& operator=(const LicenseManager&) = delete;

    // 实现接口方法
    std::string GenerateLicense(
        const std::string& user_id,
        const std::string& device_id,
        const std::string& device_fingerprint,
        const std::string& plan_code,
        std::time_t expire_time,
        int max_bind_count = 3
    ) override;

    LicenseStatus ValidateLicense(
        const std::string& license_key,
        const std::string& device_id,
        const std::string& device_fingerprint
    ) override;

    std::unique_ptr<LicenseInfo> ParseLicense(
        const std::string& license_key
    ) override;

    bool BindDevice(
        const std::string& license_key,
        const std::string& device_id,
        const std::string& device_fingerprint
    ) override;

    bool UnbindDevice(
        const std::string& license_key,
        const std::string& device_id
    ) override;

    std::vector<DeviceBindInfo> GetBoundDevices(
        const std::string& license_key
    ) override;

    bool RevokeLicense(const std::string& license_key) override;

    bool IsLicenseExpired(const std::string& license_key) override;

    long GetRemainingTime(const std::string& license_key) override;

    // 静态工厂方法
    static std::unique_ptr<LicenseManager> Create(
        const std::string& private_key,
        const std::string& public_key,
        const std::string& aes_key,
        const std::string& license_db_path = ""
    );
};

// 许可证错误类
class LicenseException : public std::exception {
private:
    std::string message_;
    int error_code_;

public:
    LicenseException(const std::string& message, int error_code = 0)
        : message_(message), error_code_(error_code) {}

    const char* what() const noexcept override {
        return message_.c_str();
    }

    int error_code() const noexcept {
        return error_code_;
    }
};

} // namespace license
} // namespace gaokaohub