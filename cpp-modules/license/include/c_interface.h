#pragma once

#ifdef __cplusplus
extern "C" {
#endif

// 错误码定义
#define LICENSE_SUCCESS 0
#define LICENSE_ERROR_INVALID_PARAM -1
#define LICENSE_ERROR_CRYPTO_FAILED -2
#define LICENSE_ERROR_EXPIRED -3
#define LICENSE_ERROR_REVOKED -4
#define LICENSE_ERROR_DEVICE_MISMATCH -5
#define LICENSE_ERROR_MAX_DEVICES -6
#define LICENSE_ERROR_NOT_FOUND -7
#define LICENSE_ERROR_INVALID_FORMAT -8
#define LICENSE_ERROR_INTERNAL -9

// 许可证句柄
typedef void* license_manager_t;

// 许可证信息结构体（C兼容）
typedef struct {
    char license_key[512];
    char user_id[64];
    char device_id[128];
    char device_fingerprint[256];
    char plan_code[32];
    long issue_time;
    long expire_time;
    int max_bind_count;
    int current_bind_count;
    char encrypted_data[1024];
    char signature[512];
} license_info_t;

// 设备绑定信息结构体（C兼容）
typedef struct {
    char device_id[128];
    char device_fingerprint[256];
    long bind_time;
    int is_active;
} device_bind_info_t;

// 许可证状态
typedef enum {
    LICENSE_STATUS_VALID = 0,
    LICENSE_STATUS_EXPIRED = 1,
    LICENSE_STATUS_REVOKED = 2,
    LICENSE_STATUS_INVALID = 3,
    LICENSE_STATUS_DEVICE_MISMATCH = 4
} license_status_t;

// 初始化许可证管理器
// 参数：private_key, public_key, aes_key, license_db_path
// 返回：许可证管理器句柄，失败返回NULL
license_manager_t license_manager_create(
    const char* private_key,
    const char* public_key, 
    const char* aes_key,
    const char* license_db_path
);

// 销毁许可证管理器
// 参数：manager - 许可证管理器句柄
void license_manager_destroy(license_manager_t manager);

// 生成许可证
// 参数：manager, user_id, device_id, device_fingerprint, plan_code, expire_time, max_bind_count
// 返回：生成的许可证密钥长度，失败返回负数错误码
// 输出：license_key - 生成的许可证密钥（调用者负责分配足够的缓冲区）
int license_generate(
    license_manager_t manager,
    const char* user_id,
    const char* device_id,
    const char* device_fingerprint,
    const char* plan_code,
    long expire_time,
    int max_bind_count,
    char* license_key,
    int license_key_size
);

// 验证许可证
// 参数：manager, license_key, device_id, device_fingerprint
// 返回：许可证状态
license_status_t license_validate(
    license_manager_t manager,
    const char* license_key,
    const char* device_id,
    const char* device_fingerprint
);

// 解析许可证信息
// 参数：manager, license_key
// 返回：成功返回0，失败返回负数错误码
// 输出：license_info - 许可证信息
int license_parse(
    license_manager_t manager,
    const char* license_key,
    license_info_t* license_info
);

// 绑定设备
// 参数：manager, license_key, device_id, device_fingerprint
// 返回：成功返回0，失败返回负数错误码
int license_bind_device(
    license_manager_t manager,
    const char* license_key,
    const char* device_id,
    const char* device_fingerprint
);

// 解绑设备
// 参数：manager, license_key, device_id
// 返回：成功返回0，失败返回负数错误码
int license_unbind_device(
    license_manager_t manager,
    const char* license_key,
    const char* device_id
);

// 获取绑定设备列表
// 参数：manager, license_key, devices, max_devices
// 返回：实际设备数量，失败返回负数错误码
// 输出：devices - 设备信息数组（调用者负责分配足够的缓冲区）
int license_get_bound_devices(
    license_manager_t manager,
    const char* license_key,
    device_bind_info_t* devices,
    int max_devices
);

// 撤销许可证
// 参数：manager, license_key
// 返回：成功返回0，失败返回负数错误码
int license_revoke(
    license_manager_t manager,
    const char* license_key
);

// 检查许可证是否过期
// 参数：manager, license_key
// 返回：1-已过期，0-未过期，负数-错误码
int license_is_expired(
    license_manager_t manager,
    const char* license_key
);

// 获取许可证剩余时间（秒）
// 参数：manager, license_key
// 返回：剩余时间（秒），负数表示错误码
long license_get_remaining_time(
    license_manager_t manager,
    const char* license_key
);

// 获取最后错误信息
// 返回：错误信息字符串
const char* license_get_last_error();

// 清除最后错误信息
void license_clear_last_error();

// 工具函数

// 生成RSA密钥对
// 返回：成功返回0，失败返回负数错误码
// 输出：private_key, public_key - RSA密钥对（调用者负责分配足够的缓冲区）
int crypto_generate_rsa_keypair(
    char* private_key,
    int private_key_size,
    char* public_key,
    int public_key_size
);

// 生成AES密钥
// 返回：成功返回0，失败返回负数错误码
// 输出：aes_key - AES密钥（调用者负责分配足够的缓冲区）
int crypto_generate_aes_key(
    char* aes_key,
    int aes_key_size
);

// RSA加密
// 参数：data, public_key
// 返回：加密数据长度，失败返回负数错误码
// 输出：encrypted_data - 加密后的数据（调用者负责分配足够的缓冲区）
int crypto_rsa_encrypt(
    const char* data,
    const char* public_key,
    char* encrypted_data,
    int encrypted_data_size
);

// RSA解密
// 参数：encrypted_data, private_key
// 返回：解密数据长度，失败返回负数错误码
// 输出：decrypted_data - 解密后的数据（调用者负责分配足够的缓冲区）
int crypto_rsa_decrypt(
    const char* encrypted_data,
    const char* private_key,
    char* decrypted_data,
    int decrypted_data_size
);

// RSA签名
// 参数：data, private_key
// 返回：签名数据长度，失败返回负数错误码
// 输出：signature - 签名数据（调用者负责分配足够的缓冲区）
int crypto_rsa_sign(
    const char* data,
    const char* private_key,
    char* signature,
    int signature_size
);

// RSA验签
// 参数：data, signature, public_key
// 返回：1-验签成功，0-验签失败，负数-错误码
int crypto_rsa_verify(
    const char* data,
    const char* signature,
    const char* public_key
);

// AES加密
// 参数：data, key, iv（可为NULL使用默认IV）
// 返回：加密数据长度，失败返回负数错误码
// 输出：encrypted_data - 加密后的数据（调用者负责分配足够的缓冲区）
int crypto_aes_encrypt(
    const char* data,
    const char* key,
    const char* iv,
    char* encrypted_data,
    int encrypted_data_size
);

// AES解密
// 参数：encrypted_data, key, iv（可为NULL使用默认IV）
// 返回：解密数据长度，失败返回负数错误码
// 输出：decrypted_data - 解密后的数据（调用者负责分配足够的缓冲区）
int crypto_aes_decrypt(
    const char* encrypted_data,
    const char* key,
    const char* iv,
    char* decrypted_data,
    int decrypted_data_size
);

// SHA256哈希
// 参数：data
// 返回：哈希数据长度，失败返回负数错误码
// 输出：hash - 哈希值（调用者负责分配足够的缓冲区）
int crypto_sha256(
    const char* data,
    char* hash,
    int hash_size
);

// Base64编码
// 参数：data
// 返回：编码数据长度，失败返回负数错误码
// 输出：encoded_data - 编码后的数据（调用者负责分配足够的缓冲区）
int crypto_base64_encode(
    const char* data,
    char* encoded_data,
    int encoded_data_size
);

// Base64解码
// 参数：encoded_data
// 返回：解码数据长度，失败返回负数错误码
// 输出：decoded_data - 解码后的数据（调用者负责分配足够的缓冲区）
int crypto_base64_decode(
    const char* encoded_data,
    char* decoded_data,
    int decoded_data_size
);

// 生成UUID
// 返回：成功返回0，失败返回负数错误码
// 输出：uuid - UUID字符串（调用者负责分配足够的缓冲区）
int crypto_generate_uuid(
    char* uuid,
    int uuid_size
);

#ifdef __cplusplus
}
#endif