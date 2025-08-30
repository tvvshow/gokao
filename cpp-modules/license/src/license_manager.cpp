#include "../include/license_manager.h"
#include "../include/crypto_utils.h"
#include <json/json.h>
#include <sqlite3.h>
#include <sstream>
#include <iomanip>
#include <fstream>
#include <thread>
#include <mutex>
#include <unordered_map>

namespace gaokaohub {
namespace license {

// 许可证管理器实现类
class LicenseManager::Impl {
private:
    std::string private_key_;
    std::string public_key_;
    std::string aes_key_;
    std::string license_db_path_;
    sqlite3* db_;
    std::mutex db_mutex_;
    std::unordered_map<std::string, std::string> license_cache_;
    std::mutex cache_mutex_;

public:
    Impl(const std::string& private_key, const std::string& public_key,
         const std::string& aes_key, const std::string& license_db_path)
        : private_key_(private_key), public_key_(public_key), 
          aes_key_(aes_key), license_db_path_(license_db_path), db_(nullptr) {
        
        if (license_db_path_.empty()) {
            license_db_path_ = "licenses.db";
        }
        
        InitializeDatabase();
    }

    ~Impl() {
        if (db_) {
            sqlite3_close(db_);
        }
    }

    std::string GenerateLicense(const std::string& user_id,
                               const std::string& device_id,
                               const std::string& device_fingerprint,
                               const std::string& plan_code,
                               std::time_t expire_time,
                               int max_bind_count) {
        try {
            // 创建许可证数据
            Json::Value license_data;
            license_data["user_id"] = user_id;
            license_data["device_id"] = device_id;
            license_data["device_fingerprint"] = device_fingerprint;
            license_data["plan_code"] = plan_code;
            license_data["issue_time"] = static_cast<Json::Int64>(std::time(nullptr));
            license_data["expire_time"] = static_cast<Json::Int64>(expire_time);
            license_data["max_bind_count"] = max_bind_count;
            license_data["uuid"] = crypto::CryptoUtils::GenerateUUID();

            // 序列化为JSON字符串
            Json::StreamWriterBuilder builder;
            builder["indentation"] = "";
            std::string json_str = Json::writeString(builder, license_data);

            // AES加密许可证数据
            std::string encrypted_data = crypto::CryptoUtils::AESEncrypt(json_str, aes_key_);

            // RSA签名
            std::string data_hash = crypto::CryptoUtils::SHA256(encrypted_data);
            std::string signature = crypto::CryptoUtils::RSASign(data_hash, private_key_);

            // 构建最终许可证
            Json::Value final_license;
            final_license["data"] = crypto::CryptoUtils::Base64Encode(encrypted_data);
            final_license["signature"] = crypto::CryptoUtils::Base64Encode(signature);
            final_license["version"] = "1.0";

            std::string license_json = Json::writeString(builder, final_license);
            std::string license_key = crypto::CryptoUtils::Base64Encode(license_json);

            // 存储到数据库
            StoreLicenseToDb(license_key, user_id, device_id, device_fingerprint, 
                           plan_code, expire_time, max_bind_count, encrypted_data, signature);

            return license_key;

        } catch (const std::exception& e) {
            throw LicenseException("Failed to generate license: " + std::string(e.what()));
        }
    }

    LicenseStatus ValidateLicense(const std::string& license_key,
                                const std::string& device_id,
                                const std::string& device_fingerprint) {
        try {
            // 解析许可证
            auto license_info = ParseLicenseInternal(license_key);
            if (!license_info) {
                return LicenseStatus::INVALID;
            }

            // 检查是否过期
            std::time_t current_time = std::time(nullptr);
            if (current_time > license_info->expire_time) {
                return LicenseStatus::EXPIRED;
            }

            // 检查是否被撤销
            if (IsLicenseRevokedInDb(license_key)) {
                return LicenseStatus::REVOKED;
            }

            // 检查设备是否匹配
            if (!IsDeviceAuthorized(license_key, device_id, device_fingerprint)) {
                return LicenseStatus::DEVICE_MISMATCH;
            }

            return LicenseStatus::VALID;

        } catch (const std::exception&) {
            return LicenseStatus::INVALID;
        }
    }

    std::unique_ptr<LicenseInfo> ParseLicense(const std::string& license_key) {
        return ParseLicenseInternal(license_key);
    }

    bool BindDevice(const std::string& license_key,
                   const std::string& device_id,
                   const std::string& device_fingerprint) {
        try {
            std::lock_guard<std::mutex> lock(db_mutex_);

            // 检查许可证是否有效
            auto license_info = ParseLicenseInternal(license_key);
            if (!license_info) {
                return false;
            }

            // 检查是否已达到最大绑定数
            int current_bind_count = GetDeviceBindCountFromDb(license_key);
            if (current_bind_count >= license_info->max_bind_count) {
                return false;
            }

            // 检查设备是否已绑定
            if (IsDeviceBoundInDb(license_key, device_id)) {
                // 更新设备指纹
                return UpdateDeviceBindingInDb(license_key, device_id, device_fingerprint);
            }

            // 添加新的设备绑定
            return AddDeviceBindingToDb(license_key, device_id, device_fingerprint);

        } catch (const std::exception&) {
            return false;
        }
    }

    bool UnbindDevice(const std::string& license_key, const std::string& device_id) {
        try {
            std::lock_guard<std::mutex> lock(db_mutex_);
            return RemoveDeviceBindingFromDb(license_key, device_id);
        } catch (const std::exception&) {
            return false;
        }
    }

    std::vector<DeviceBindInfo> GetBoundDevices(const std::string& license_key) {
        try {
            std::lock_guard<std::mutex> lock(db_mutex_);
            return GetDeviceBindingsFromDb(license_key);
        } catch (const std::exception&) {
            return {};
        }
    }

    bool RevokeLicense(const std::string& license_key) {
        try {
            std::lock_guard<std::mutex> lock(db_mutex_);
            return SetLicenseRevokedInDb(license_key, true);
        } catch (const std::exception&) {
            return false;
        }
    }

    bool IsLicenseExpired(const std::string& license_key) {
        try {
            auto license_info = ParseLicenseInternal(license_key);
            if (!license_info) {
                return true;
            }

            std::time_t current_time = std::time(nullptr);
            return current_time > license_info->expire_time;

        } catch (const std::exception&) {
            return true;
        }
    }

    long GetRemainingTime(const std::string& license_key) {
        try {
            auto license_info = ParseLicenseInternal(license_key);
            if (!license_info) {
                return -1;
            }

            std::time_t current_time = std::time(nullptr);
            return static_cast<long>(license_info->expire_time - current_time);

        } catch (const std::exception&) {
            return -1;
        }
    }

private:
    void InitializeDatabase() {
        std::lock_guard<std::mutex> lock(db_mutex_);

        int rc = sqlite3_open(license_db_path_.c_str(), &db_);
        if (rc != SQLITE_OK) {
            throw LicenseException("Failed to open license database: " + 
                                 std::string(sqlite3_errmsg(db_)));
        }

        // 创建许可证表
        const char* create_licenses_sql = R"(
            CREATE TABLE IF NOT EXISTS licenses (
                license_key TEXT PRIMARY KEY,
                user_id TEXT NOT NULL,
                device_id TEXT NOT NULL,
                device_fingerprint TEXT NOT NULL,
                plan_code TEXT NOT NULL,
                issue_time INTEGER NOT NULL,
                expire_time INTEGER NOT NULL,
                max_bind_count INTEGER NOT NULL,
                encrypted_data TEXT NOT NULL,
                signature TEXT NOT NULL,
                is_revoked INTEGER DEFAULT 0,
                created_at INTEGER DEFAULT (strftime('%s', 'now'))
            )
        )";

        // 创建设备绑定表
        const char* create_device_bindings_sql = R"(
            CREATE TABLE IF NOT EXISTS device_bindings (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                license_key TEXT NOT NULL,
                device_id TEXT NOT NULL,
                device_fingerprint TEXT NOT NULL,
                bind_time INTEGER DEFAULT (strftime('%s', 'now')),
                is_active INTEGER DEFAULT 1,
                FOREIGN KEY (license_key) REFERENCES licenses (license_key),
                UNIQUE(license_key, device_id)
            )
        )";

        char* err_msg = nullptr;
        rc = sqlite3_exec(db_, create_licenses_sql, nullptr, nullptr, &err_msg);
        if (rc != SQLITE_OK) {
            std::string error = "Failed to create licenses table: " + 
                              std::string(err_msg ? err_msg : "unknown error");
            sqlite3_free(err_msg);
            throw LicenseException(error);
        }

        rc = sqlite3_exec(db_, create_device_bindings_sql, nullptr, nullptr, &err_msg);
        if (rc != SQLITE_OK) {
            std::string error = "Failed to create device_bindings table: " + 
                              std::string(err_msg ? err_msg : "unknown error");
            sqlite3_free(err_msg);
            throw LicenseException(error);
        }

        // 创建索引
        const char* create_indexes_sql[] = {
            "CREATE INDEX IF NOT EXISTS idx_licenses_user_id ON licenses(user_id)",
            "CREATE INDEX IF NOT EXISTS idx_licenses_expire_time ON licenses(expire_time)",
            "CREATE INDEX IF NOT EXISTS idx_device_bindings_license_key ON device_bindings(license_key)",
            "CREATE INDEX IF NOT EXISTS idx_device_bindings_device_id ON device_bindings(device_id)"
        };

        for (const char* sql : create_indexes_sql) {
            rc = sqlite3_exec(db_, sql, nullptr, nullptr, &err_msg);
            if (rc != SQLITE_OK) {
                sqlite3_free(err_msg);
                // 索引创建失败不抛异常，只记录日志
            }
        }
    }

    std::unique_ptr<LicenseInfo> ParseLicenseInternal(const std::string& license_key) {
        try {
            // 从缓存中查找
            {
                std::lock_guard<std::mutex> lock(cache_mutex_);
                auto it = license_cache_.find(license_key);
                if (it != license_cache_.end()) {
                    // 从缓存解析（这里简化处理，实际应该缓存解析后的对象）
                }
            }

            // Base64解码许可证
            std::string license_json = crypto::CryptoUtils::Base64Decode(license_key);

            // 解析JSON
            Json::Value final_license;
            Json::CharReaderBuilder builder;
            std::string errors;
            std::istringstream stream(license_json);
            
            if (!Json::parseFromStream(builder, stream, &final_license, &errors)) {
                return nullptr;
            }

            // 验证版本
            if (!final_license.isMember("version") || 
                final_license["version"].asString() != "1.0") {
                return nullptr;
            }

            // 提取数据和签名
            std::string encrypted_data = crypto::CryptoUtils::Base64Decode(
                final_license["data"].asString());
            std::string signature = crypto::CryptoUtils::Base64Decode(
                final_license["signature"].asString());

            // 验证签名
            std::string data_hash = crypto::CryptoUtils::SHA256(encrypted_data);
            if (!crypto::CryptoUtils::RSAVerify(data_hash, signature, public_key_)) {
                return nullptr;
            }

            // AES解密数据
            std::string decrypted_data = crypto::CryptoUtils::AESDecrypt(encrypted_data, aes_key_);

            // 解析许可证数据
            Json::Value license_data;
            std::istringstream data_stream(decrypted_data);
            if (!Json::parseFromStream(builder, data_stream, &license_data, &errors)) {
                return nullptr;
            }

            // 创建许可证信息对象
            auto info = std::make_unique<LicenseInfo>();
            info->license_key = license_key;
            info->user_id = license_data["user_id"].asString();
            info->device_id = license_data["device_id"].asString();
            info->device_fingerprint = license_data["device_fingerprint"].asString();
            info->plan_code = license_data["plan_code"].asString();
            info->issue_time = license_data["issue_time"].asInt64();
            info->expire_time = license_data["expire_time"].asInt64();
            info->max_bind_count = license_data["max_bind_count"].asInt();
            info->encrypted_data = crypto::CryptoUtils::Base64Encode(encrypted_data);
            info->signature = crypto::CryptoUtils::Base64Encode(signature);

            // 获取当前绑定数量
            info->current_bind_count = GetDeviceBindCountFromDb(license_key);

            // 添加到缓存
            {
                std::lock_guard<std::mutex> lock(cache_mutex_);
                license_cache_[license_key] = license_json;
            }

            return info;

        } catch (const std::exception&) {
            return nullptr;
        }
    }

    // 数据库操作方法
    bool StoreLicenseToDb(const std::string& license_key, const std::string& user_id,
                         const std::string& device_id, const std::string& device_fingerprint,
                         const std::string& plan_code, std::time_t expire_time,
                         int max_bind_count, const std::string& encrypted_data,
                         const std::string& signature) {
        const char* sql = R"(
            INSERT OR REPLACE INTO licenses 
            (license_key, user_id, device_id, device_fingerprint, plan_code, 
             issue_time, expire_time, max_bind_count, encrypted_data, signature)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        )";

        sqlite3_stmt* stmt;
        int rc = sqlite3_prepare_v2(db_, sql, -1, &stmt, nullptr);
        if (rc != SQLITE_OK) {
            return false;
        }

        sqlite3_bind_text(stmt, 1, license_key.c_str(), -1, SQLITE_STATIC);
        sqlite3_bind_text(stmt, 2, user_id.c_str(), -1, SQLITE_STATIC);
        sqlite3_bind_text(stmt, 3, device_id.c_str(), -1, SQLITE_STATIC);
        sqlite3_bind_text(stmt, 4, device_fingerprint.c_str(), -1, SQLITE_STATIC);
        sqlite3_bind_text(stmt, 5, plan_code.c_str(), -1, SQLITE_STATIC);
        sqlite3_bind_int64(stmt, 6, std::time(nullptr));
        sqlite3_bind_int64(stmt, 7, expire_time);
        sqlite3_bind_int(stmt, 8, max_bind_count);
        sqlite3_bind_text(stmt, 9, encrypted_data.c_str(), -1, SQLITE_STATIC);
        sqlite3_bind_text(stmt, 10, signature.c_str(), -1, SQLITE_STATIC);

        rc = sqlite3_step(stmt);
        sqlite3_finalize(stmt);

        return rc == SQLITE_DONE;
    }

    bool IsLicenseRevokedInDb(const std::string& license_key) {
        const char* sql = "SELECT is_revoked FROM licenses WHERE license_key = ?";
        
        sqlite3_stmt* stmt;
        int rc = sqlite3_prepare_v2(db_, sql, -1, &stmt, nullptr);
        if (rc != SQLITE_OK) {
            return true; // 安全起见，查询失败则认为已撤销
        }

        sqlite3_bind_text(stmt, 1, license_key.c_str(), -1, SQLITE_STATIC);
        
        bool is_revoked = true;
        if (sqlite3_step(stmt) == SQLITE_ROW) {
            is_revoked = sqlite3_column_int(stmt, 0) != 0;
        }

        sqlite3_finalize(stmt);
        return is_revoked;
    }

    bool SetLicenseRevokedInDb(const std::string& license_key, bool revoked) {
        const char* sql = "UPDATE licenses SET is_revoked = ? WHERE license_key = ?";
        
        sqlite3_stmt* stmt;
        int rc = sqlite3_prepare_v2(db_, sql, -1, &stmt, nullptr);
        if (rc != SQLITE_OK) {
            return false;
        }

        sqlite3_bind_int(stmt, 1, revoked ? 1 : 0);
        sqlite3_bind_text(stmt, 2, license_key.c_str(), -1, SQLITE_STATIC);

        rc = sqlite3_step(stmt);
        sqlite3_finalize(stmt);

        return rc == SQLITE_DONE;
    }

    bool IsDeviceAuthorized(const std::string& license_key,
                           const std::string& device_id,
                           const std::string& device_fingerprint) {
        // 检查设备是否在绑定列表中
        const char* sql = R"(
            SELECT device_fingerprint FROM device_bindings 
            WHERE license_key = ? AND device_id = ? AND is_active = 1
        )";

        sqlite3_stmt* stmt;
        int rc = sqlite3_prepare_v2(db_, sql, -1, &stmt, nullptr);
        if (rc != SQLITE_OK) {
            return false;
        }

        sqlite3_bind_text(stmt, 1, license_key.c_str(), -1, SQLITE_STATIC);
        sqlite3_bind_text(stmt, 2, device_id.c_str(), -1, SQLITE_STATIC);

        bool authorized = false;
        if (sqlite3_step(stmt) == SQLITE_ROW) {
            const char* stored_fingerprint = reinterpret_cast<const char*>(
                sqlite3_column_text(stmt, 0));
            if (stored_fingerprint && device_fingerprint == stored_fingerprint) {
                authorized = true;
            }
        }

        sqlite3_finalize(stmt);
        return authorized;
    }

    int GetDeviceBindCountFromDb(const std::string& license_key) {
        const char* sql = R"(
            SELECT COUNT(*) FROM device_bindings 
            WHERE license_key = ? AND is_active = 1
        )";

        sqlite3_stmt* stmt;
        int rc = sqlite3_prepare_v2(db_, sql, -1, &stmt, nullptr);
        if (rc != SQLITE_OK) {
            return 0;
        }

        sqlite3_bind_text(stmt, 1, license_key.c_str(), -1, SQLITE_STATIC);

        int count = 0;
        if (sqlite3_step(stmt) == SQLITE_ROW) {
            count = sqlite3_column_int(stmt, 0);
        }

        sqlite3_finalize(stmt);
        return count;
    }

    bool IsDeviceBoundInDb(const std::string& license_key, const std::string& device_id) {
        const char* sql = R"(
            SELECT 1 FROM device_bindings 
            WHERE license_key = ? AND device_id = ? AND is_active = 1
        )";

        sqlite3_stmt* stmt;
        int rc = sqlite3_prepare_v2(db_, sql, -1, &stmt, nullptr);
        if (rc != SQLITE_OK) {
            return false;
        }

        sqlite3_bind_text(stmt, 1, license_key.c_str(), -1, SQLITE_STATIC);
        sqlite3_bind_text(stmt, 2, device_id.c_str(), -1, SQLITE_STATIC);

        bool bound = sqlite3_step(stmt) == SQLITE_ROW;
        sqlite3_finalize(stmt);

        return bound;
    }

    bool AddDeviceBindingToDb(const std::string& license_key,
                             const std::string& device_id,
                             const std::string& device_fingerprint) {
        const char* sql = R"(
            INSERT INTO device_bindings (license_key, device_id, device_fingerprint)
            VALUES (?, ?, ?)
        )";

        sqlite3_stmt* stmt;
        int rc = sqlite3_prepare_v2(db_, sql, -1, &stmt, nullptr);
        if (rc != SQLITE_OK) {
            return false;
        }

        sqlite3_bind_text(stmt, 1, license_key.c_str(), -1, SQLITE_STATIC);
        sqlite3_bind_text(stmt, 2, device_id.c_str(), -1, SQLITE_STATIC);
        sqlite3_bind_text(stmt, 3, device_fingerprint.c_str(), -1, SQLITE_STATIC);

        rc = sqlite3_step(stmt);
        sqlite3_finalize(stmt);

        return rc == SQLITE_DONE;
    }

    bool UpdateDeviceBindingInDb(const std::string& license_key,
                                const std::string& device_id,
                                const std::string& device_fingerprint) {
        const char* sql = R"(
            UPDATE device_bindings 
            SET device_fingerprint = ?, bind_time = strftime('%s', 'now')
            WHERE license_key = ? AND device_id = ?
        )";

        sqlite3_stmt* stmt;
        int rc = sqlite3_prepare_v2(db_, sql, -1, &stmt, nullptr);
        if (rc != SQLITE_OK) {
            return false;
        }

        sqlite3_bind_text(stmt, 1, device_fingerprint.c_str(), -1, SQLITE_STATIC);
        sqlite3_bind_text(stmt, 2, license_key.c_str(), -1, SQLITE_STATIC);
        sqlite3_bind_text(stmt, 3, device_id.c_str(), -1, SQLITE_STATIC);

        rc = sqlite3_step(stmt);
        sqlite3_finalize(stmt);

        return rc == SQLITE_DONE;
    }

    bool RemoveDeviceBindingFromDb(const std::string& license_key,
                                  const std::string& device_id) {
        const char* sql = R"(
            UPDATE device_bindings 
            SET is_active = 0
            WHERE license_key = ? AND device_id = ?
        )";

        sqlite3_stmt* stmt;
        int rc = sqlite3_prepare_v2(db_, sql, -1, &stmt, nullptr);
        if (rc != SQLITE_OK) {
            return false;
        }

        sqlite3_bind_text(stmt, 1, license_key.c_str(), -1, SQLITE_STATIC);
        sqlite3_bind_text(stmt, 2, device_id.c_str(), -1, SQLITE_STATIC);

        rc = sqlite3_step(stmt);
        sqlite3_finalize(stmt);

        return rc == SQLITE_DONE;
    }

    std::vector<DeviceBindInfo> GetDeviceBindingsFromDb(const std::string& license_key) {
        std::vector<DeviceBindInfo> devices;

        const char* sql = R"(
            SELECT device_id, device_fingerprint, bind_time, is_active
            FROM device_bindings 
            WHERE license_key = ?
            ORDER BY bind_time DESC
        )";

        sqlite3_stmt* stmt;
        int rc = sqlite3_prepare_v2(db_, sql, -1, &stmt, nullptr);
        if (rc != SQLITE_OK) {
            return devices;
        }

        sqlite3_bind_text(stmt, 1, license_key.c_str(), -1, SQLITE_STATIC);

        while (sqlite3_step(stmt) == SQLITE_ROW) {
            DeviceBindInfo info;
            info.device_id = reinterpret_cast<const char*>(sqlite3_column_text(stmt, 0));
            info.device_fingerprint = reinterpret_cast<const char*>(sqlite3_column_text(stmt, 1));
            info.bind_time = sqlite3_column_int64(stmt, 2);
            info.is_active = sqlite3_column_int(stmt, 3) != 0;
            devices.push_back(info);
        }

        sqlite3_finalize(stmt);
        return devices;
    }
};

// LicenseManager 公共方法实现

LicenseManager::LicenseManager(const std::string& private_key,
                              const std::string& public_key,
                              const std::string& aes_key,
                              const std::string& license_db_path)
    : pImpl_(std::make_unique<Impl>(private_key, public_key, aes_key, license_db_path)) {
}

LicenseManager::~LicenseManager() = default;

std::string LicenseManager::GenerateLicense(const std::string& user_id,
                                           const std::string& device_id,
                                           const std::string& device_fingerprint,
                                           const std::string& plan_code,
                                           std::time_t expire_time,
                                           int max_bind_count) {
    return pImpl_->GenerateLicense(user_id, device_id, device_fingerprint, 
                                  plan_code, expire_time, max_bind_count);
}

LicenseStatus LicenseManager::ValidateLicense(const std::string& license_key,
                                             const std::string& device_id,
                                             const std::string& device_fingerprint) {
    return pImpl_->ValidateLicense(license_key, device_id, device_fingerprint);
}

std::unique_ptr<LicenseInfo> LicenseManager::ParseLicense(const std::string& license_key) {
    return pImpl_->ParseLicense(license_key);
}

bool LicenseManager::BindDevice(const std::string& license_key,
                               const std::string& device_id,
                               const std::string& device_fingerprint) {
    return pImpl_->BindDevice(license_key, device_id, device_fingerprint);
}

bool LicenseManager::UnbindDevice(const std::string& license_key,
                                 const std::string& device_id) {
    return pImpl_->UnbindDevice(license_key, device_id);
}

std::vector<DeviceBindInfo> LicenseManager::GetBoundDevices(const std::string& license_key) {
    return pImpl_->GetBoundDevices(license_key);
}

bool LicenseManager::RevokeLicense(const std::string& license_key) {
    return pImpl_->RevokeLicense(license_key);
}

bool LicenseManager::IsLicenseExpired(const std::string& license_key) {
    return pImpl_->IsLicenseExpired(license_key);
}

long LicenseManager::GetRemainingTime(const std::string& license_key) {
    return pImpl_->GetRemainingTime(license_key);
}

std::unique_ptr<LicenseManager> LicenseManager::Create(const std::string& private_key,
                                                      const std::string& public_key,
                                                      const std::string& aes_key,
                                                      const std::string& license_db_path) {
    return std::make_unique<LicenseManager>(private_key, public_key, aes_key, license_db_path);
}

} // namespace license
} // namespace gaokaohub