#pragma once

#include <string>
#include <vector>
#include <memory>

namespace gaokaohub {
namespace license {
namespace crypto {

// RSA密钥对结构
struct RSAKeyPair {
    std::string private_key;
    std::string public_key;
};

// 加密工具类
class CryptoUtils {
public:
    // RSA相关操作
    
    // 生成RSA密钥对（2048位）
    static RSAKeyPair GenerateRSAKeyPair();
    
    // RSA公钥加密
    static std::string RSAEncrypt(
        const std::string& data,
        const std::string& public_key
    );
    
    // RSA私钥解密
    static std::string RSADecrypt(
        const std::string& encrypted_data,
        const std::string& private_key
    );
    
    // RSA私钥签名
    static std::string RSASign(
        const std::string& data,
        const std::string& private_key
    );
    
    // RSA公钥验签
    static bool RSAVerify(
        const std::string& data,
        const std::string& signature,
        const std::string& public_key
    );

    // AES相关操作
    
    // 生成AES密钥（256位）
    static std::string GenerateAESKey();
    
    // AES加密（AES-256-CBC）
    static std::string AESEncrypt(
        const std::string& data,
        const std::string& key,
        const std::string& iv = ""
    );
    
    // AES解密（AES-256-CBC）
    static std::string AESDecrypt(
        const std::string& encrypted_data,
        const std::string& key,
        const std::string& iv = ""
    );

    // 哈希相关操作
    
    // SHA256哈希
    static std::string SHA256(const std::string& data);
    
    // SHA512哈希
    static std::string SHA512(const std::string& data);
    
    // MD5哈希（仅用于非安全场景）
    static std::string MD5(const std::string& data);

    // 编码相关操作
    
    // Base64编码
    static std::string Base64Encode(const std::string& data);
    static std::string Base64Encode(const std::vector<unsigned char>& data);
    
    // Base64解码
    static std::string Base64Decode(const std::string& encoded_data);
    static std::vector<unsigned char> Base64DecodeToBytes(const std::string& encoded_data);
    
    // 十六进制编码
    static std::string HexEncode(const std::string& data);
    static std::string HexEncode(const std::vector<unsigned char>& data);
    
    // 十六进制解码
    static std::string HexDecode(const std::string& hex_data);
    static std::vector<unsigned char> HexDecodeToBytes(const std::string& hex_data);

    // 随机数生成
    
    // 生成随机字节
    static std::vector<unsigned char> GenerateRandomBytes(size_t length);
    
    // 生成随机字符串（包含字母数字）
    static std::string GenerateRandomString(size_t length);
    
    // 生成UUID
    static std::string GenerateUUID();

    // 密钥派生
    
    // PBKDF2密钥派生
    static std::string PBKDF2(
        const std::string& password,
        const std::string& salt,
        int iterations = 10000,
        size_t key_length = 32
    );

    // 安全比较（防止时序攻击）
    static bool SecureCompare(const std::string& a, const std::string& b);

private:
    // 私有构造函数，防止实例化
    CryptoUtils() = default;
    
    // 内部辅助方法
    static std::string BytesToString(const std::vector<unsigned char>& bytes);
    static std::vector<unsigned char> StringToBytes(const std::string& str);
};

// 密钥管理器
class KeyManager {
private:
    std::string master_key_;
    std::string key_store_path_;

public:
    explicit KeyManager(const std::string& master_key, const std::string& key_store_path = "");
    
    // 存储加密密钥
    bool StoreKey(const std::string& key_id, const std::string& key_data);
    
    // 加载解密密钥
    std::string LoadKey(const std::string& key_id);
    
    // 删除密钥
    bool DeleteKey(const std::string& key_id);
    
    // 列出所有密钥ID
    std::vector<std::string> ListKeys();
    
    // 验证密钥完整性
    bool VerifyKeyIntegrity(const std::string& key_id);
};

// 安全字符串类（自动清零内存）
class SecureString {
private:
    std::vector<char> data_;

public:
    SecureString();
    explicit SecureString(const std::string& str);
    explicit SecureString(const char* str);
    SecureString(const SecureString& other);
    SecureString& operator=(const SecureString& other);
    SecureString(SecureString&& other) noexcept;
    SecureString& operator=(SecureString&& other) noexcept;
    ~SecureString();

    // 基本操作
    size_t size() const;
    bool empty() const;
    void clear();
    
    // 数据访问
    const char* c_str() const;
    std::string str() const;
    
    // 赋值操作
    void assign(const std::string& str);
    void assign(const char* str);
    
    // 比较操作
    bool operator==(const SecureString& other) const;
    bool operator!=(const SecureString& other) const;

private:
    void secure_zero();
};

} // namespace crypto
} // namespace license
} // namespace gaokaohub