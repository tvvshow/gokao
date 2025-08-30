/**
 * @file crypto_utils.h
 * @brief 加密工具类头文件
 * @author 高考志愿填报系统开发团队
 * @version 1.0.0
 * @date 2025-01-18
 * 
 * 提供AES对称加密、RSA非对称加密、数字签名、哈希计算等密码学功能，
 * 用于保护设备指纹数据的安全性和完整性。
 */

#ifndef CRYPTO_UTILS_H
#define CRYPTO_UTILS_H

#include <string>
#include <vector>
#include <memory>

namespace gaokao {
namespace crypto {

/**
 * @brief 加密算法类型枚举
 */
enum class CryptoAlgorithm {
    AES_256_CBC,        ///< AES-256-CBC
    AES_256_GCM,        ///< AES-256-GCM
    AES_256_CTR,        ///< AES-256-CTR
    RSA_2048,           ///< RSA-2048
    RSA_4096,           ///< RSA-4096
    ECC_P256,           ///< ECC P-256
    ECC_P384            ///< ECC P-384
};

/**
 * @brief 哈希算法类型枚举
 */
enum class HashAlgorithm {
    SHA256,             ///< SHA-256
    SHA384,             ///< SHA-384
    SHA512,             ///< SHA-512
    SHA3_256,           ///< SHA3-256
    BLAKE2B             ///< BLAKE2b
};

/**
 * @brief 加密错误码
 */
enum class CryptoError {
    SUCCESS = 0,                ///< 成功
    INVALID_PARAMETER = 2001,   ///< 无效参数
    KEY_GENERATION_FAILED = 2002, ///< 密钥生成失败
    ENCRYPTION_FAILED = 2003,   ///< 加密失败
    DECRYPTION_FAILED = 2004,   ///< 解密失败
    SIGNATURE_FAILED = 2005,    ///< 签名失败
    VERIFICATION_FAILED = 2006, ///< 验证失败
    HASH_FAILED = 2007,         ///< 哈希计算失败
    KEY_INVALID = 2008,         ///< 无效密钥
    ALGORITHM_NOT_SUPPORTED = 2009, ///< 算法不支持
    MEMORY_ERROR = 2010,        ///< 内存错误
    RANDOM_GENERATION_FAILED = 2011 ///< 随机数生成失败
};

/**
 * @brief 密钥对结构体
 */
struct KeyPair {
    std::vector<uint8_t> public_key;   ///< 公钥
    std::vector<uint8_t> private_key;  ///< 私钥
    CryptoAlgorithm algorithm;          ///< 算法类型
    uint32_t key_size;                  ///< 密钥长度
    
    KeyPair() : algorithm(CryptoAlgorithm::RSA_2048), key_size(0) {}
};

/**
 * @brief 加密结果结构体
 */
struct EncryptionResult {
    std::vector<uint8_t> encrypted_data;   ///< 加密数据
    std::vector<uint8_t> iv;               ///< 初始化向量
    std::vector<uint8_t> tag;              ///< 认证标签(GCM模式)
    std::vector<uint8_t> salt;             ///< 盐值
    CryptoAlgorithm algorithm;              ///< 使用的算法
    
    EncryptionResult() : algorithm(CryptoAlgorithm::AES_256_CBC) {}
};

/**
 * @brief 签名结果结构体
 */
struct SignatureResult {
    std::vector<uint8_t> signature;        ///< 数字签名
    HashAlgorithm hash_algorithm;          ///< 哈希算法
    CryptoAlgorithm sign_algorithm;        ///< 签名算法
    
    SignatureResult() : hash_algorithm(HashAlgorithm::SHA256),
                       sign_algorithm(CryptoAlgorithm::RSA_2048) {}
};

/**
 * @brief AES加密工具类
 */
class AESCipher {
public:
    /**
     * @brief 构造函数
     */
    AESCipher();
    
    /**
     * @brief 析构函数
     */
    ~AESCipher();
    
    /**
     * @brief 生成AES密钥
     * @param key_size 密钥长度(位)
     * @param key 输出的密钥
     * @return 错误码
     */
    CryptoError GenerateKey(uint32_t key_size, std::vector<uint8_t>& key);
    
    /**
     * @brief AES加密
     * @param plaintext 明文数据
     * @param key 加密密钥
     * @param algorithm 加密算法
     * @param result 加密结果
     * @return 错误码
     */
    CryptoError Encrypt(const std::vector<uint8_t>& plaintext,
                       const std::vector<uint8_t>& key,
                       CryptoAlgorithm algorithm,
                       EncryptionResult& result);
    
    /**
     * @brief AES解密
     * @param result 加密结果
     * @param key 解密密钥
     * @param plaintext 输出的明文数据
     * @return 错误码
     */
    CryptoError Decrypt(const EncryptionResult& result,
                       const std::vector<uint8_t>& key,
                       std::vector<uint8_t>& plaintext);
    
    /**
     * @brief 字符串加密(便利函数)
     * @param plaintext 明文字符串
     * @param password 密码
     * @param encrypted_base64 输出的Base64编码密文
     * @return 错误码
     */
    CryptoError EncryptString(const std::string& plaintext,
                             const std::string& password,
                             std::string& encrypted_base64);
    
    /**
     * @brief 字符串解密(便利函数)
     * @param encrypted_base64 Base64编码密文
     * @param password 密码
     * @param plaintext 输出的明文字符串
     * @return 错误码
     */
    CryptoError DecryptString(const std::string& encrypted_base64,
                             const std::string& password,
                             std::string& plaintext);

private:
    class Impl;
    std::unique_ptr<Impl> pimpl_;
    
    // 禁用拷贝构造和赋值
    AESCipher(const AESCipher&) = delete;
    AESCipher& operator=(const AESCipher&) = delete;
};

/**
 * @brief RSA加密工具类
 */
class RSACipher {
public:
    /**
     * @brief 构造函数
     */
    RSACipher();
    
    /**
     * @brief 析构函数
     */
    ~RSACipher();
    
    /**
     * @brief 生成RSA密钥对
     * @param key_size 密钥长度(位)
     * @param key_pair 输出的密钥对
     * @return 错误码
     */
    CryptoError GenerateKeyPair(uint32_t key_size, KeyPair& key_pair);
    
    /**
     * @brief RSA公钥加密
     * @param plaintext 明文数据
     * @param public_key 公钥
     * @param encrypted_data 输出的密文数据
     * @return 错误码
     */
    CryptoError PublicKeyEncrypt(const std::vector<uint8_t>& plaintext,
                                const std::vector<uint8_t>& public_key,
                                std::vector<uint8_t>& encrypted_data);
    
    /**
     * @brief RSA私钥解密
     * @param encrypted_data 密文数据
     * @param private_key 私钥
     * @param plaintext 输出的明文数据
     * @return 错误码
     */
    CryptoError PrivateKeyDecrypt(const std::vector<uint8_t>& encrypted_data,
                                 const std::vector<uint8_t>& private_key,
                                 std::vector<uint8_t>& plaintext);
    
    /**
     * @brief RSA私钥签名
     * @param data 待签名数据
     * @param private_key 私钥
     * @param hash_algorithm 哈希算法
     * @param result 签名结果
     * @return 错误码
     */
    CryptoError Sign(const std::vector<uint8_t>& data,
                    const std::vector<uint8_t>& private_key,
                    HashAlgorithm hash_algorithm,
                    SignatureResult& result);
    
    /**
     * @brief RSA公钥验证签名
     * @param data 原始数据
     * @param signature 签名结果
     * @param public_key 公钥
     * @return 是否验证通过
     */
    bool VerifySignature(const std::vector<uint8_t>& data,
                        const SignatureResult& signature,
                        const std::vector<uint8_t>& public_key);
    
    /**
     * @brief 导入PEM格式密钥
     * @param pem_key PEM格式密钥字符串
     * @param key 输出的密钥数据
     * @return 错误码
     */
    CryptoError ImportPEMKey(const std::string& pem_key,
                            std::vector<uint8_t>& key);
    
    /**
     * @brief 导出PEM格式密钥
     * @param key 密钥数据
     * @param is_private 是否为私钥
     * @param pem_key 输出的PEM格式密钥字符串
     * @return 错误码
     */
    CryptoError ExportPEMKey(const std::vector<uint8_t>& key,
                            bool is_private,
                            std::string& pem_key);

private:
    class Impl;
    std::unique_ptr<Impl> pimpl_;
    
    // 禁用拷贝构造和赋值
    RSACipher(const RSACipher&) = delete;
    RSACipher& operator=(const RSACipher&) = delete;
};

/**
 * @brief 哈希工具类
 */
class HashUtils {
public:
    /**
     * @brief 计算数据哈希
     * @param data 输入数据
     * @param algorithm 哈希算法
     * @param hash 输出的哈希值
     * @return 错误码
     */
    static CryptoError CalculateHash(const std::vector<uint8_t>& data,
                                    HashAlgorithm algorithm,
                                    std::vector<uint8_t>& hash);
    
    /**
     * @brief 计算字符串哈希
     * @param data 输入字符串
     * @param algorithm 哈希算法
     * @param hash_hex 输出的十六进制哈希字符串
     * @return 错误码
     */
    static CryptoError CalculateStringHash(const std::string& data,
                                          HashAlgorithm algorithm,
                                          std::string& hash_hex);
    
    /**
     * @brief 计算HMAC
     * @param data 输入数据
     * @param key 密钥
     * @param algorithm 哈希算法
     * @param hmac 输出的HMAC值
     * @return 错误码
     */
    static CryptoError CalculateHMAC(const std::vector<uint8_t>& data,
                                    const std::vector<uint8_t>& key,
                                    HashAlgorithm algorithm,
                                    std::vector<uint8_t>& hmac);
    
    /**
     * @brief 生成随机盐值
     * @param size 盐值长度
     * @param salt 输出的盐值
     * @return 错误码
     */
    static CryptoError GenerateSalt(uint32_t size, std::vector<uint8_t>& salt);
    
    /**
     * @brief PBKDF2密钥派生
     * @param password 密码
     * @param salt 盐值
     * @param iterations 迭代次数
     * @param key_length 输出密钥长度
     * @param algorithm 哈希算法
     * @param derived_key 输出的派生密钥
     * @return 错误码
     */
    static CryptoError PBKDF2(const std::string& password,
                             const std::vector<uint8_t>& salt,
                             uint32_t iterations,
                             uint32_t key_length,
                             HashAlgorithm algorithm,
                             std::vector<uint8_t>& derived_key);
};

/**
 * @brief 编码工具类
 */
class EncodingUtils {
public:
    /**
     * @brief Base64编码
     * @param data 输入数据
     * @return Base64编码字符串
     */
    static std::string Base64Encode(const std::vector<uint8_t>& data);
    
    /**
     * @brief Base64解码
     * @param encoded Base64编码字符串
     * @param data 输出的原始数据
     * @return 是否解码成功
     */
    static bool Base64Decode(const std::string& encoded, 
                            std::vector<uint8_t>& data);
    
    /**
     * @brief 十六进制编码
     * @param data 输入数据
     * @param uppercase 是否使用大写字母
     * @return 十六进制字符串
     */
    static std::string HexEncode(const std::vector<uint8_t>& data,
                                bool uppercase = false);
    
    /**
     * @brief 十六进制解码
     * @param hex_string 十六进制字符串
     * @param data 输出的原始数据
     * @return 是否解码成功
     */
    static bool HexDecode(const std::string& hex_string,
                         std::vector<uint8_t>& data);
    
    /**
     * @brief 字符串转字节数组
     * @param str 输入字符串
     * @return 字节数组
     */
    static std::vector<uint8_t> StringToBytes(const std::string& str);
    
    /**
     * @brief 字节数组转字符串
     * @param data 字节数组
     * @return 字符串
     */
    static std::string BytesToString(const std::vector<uint8_t>& data);
};

/**
 * @brief 随机数生成器
 */
class RandomGenerator {
public:
    /**
     * @brief 生成随机字节数组
     * @param size 字节数组长度
     * @param random_bytes 输出的随机字节数组
     * @return 错误码
     */
    static CryptoError GenerateRandomBytes(uint32_t size,
                                          std::vector<uint8_t>& random_bytes);
    
    /**
     * @brief 生成随机字符串
     * @param length 字符串长度
     * @param charset 字符集
     * @return 随机字符串
     */
    static std::string GenerateRandomString(uint32_t length,
                                           const std::string& charset = 
                                           "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz");
    
    /**
     * @brief 生成UUID
     * @return UUID字符串
     */
    static std::string GenerateUUID();
    
    /**
     * @brief 生成随机整数
     * @param min 最小值
     * @param max 最大值
     * @return 随机整数
     */
    static uint32_t GenerateRandomInt(uint32_t min, uint32_t max);
};

/**
 * @brief 获取错误描述
 * @param error 错误码
 * @return 错误描述字符串
 */
std::string GetCryptoErrorDescription(CryptoError error);

} // namespace crypto
} // namespace gaokao

#endif // CRYPTO_UTILS_H