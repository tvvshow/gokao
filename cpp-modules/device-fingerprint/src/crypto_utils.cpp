/**
 * @file crypto_utils.cpp
 * @brief 加密工具类实现
 * @author 高考志愿填报系统开发团队
 * @version 1.0.0
 * @date 2025-01-18
 */

#include "crypto_utils.h"
#include <openssl/aes.h>
#include <openssl/rsa.h>
#include <openssl/evp.h>
#include <openssl/rand.h>
#include <openssl/sha.h>
#include <openssl/hmac.h>
#include <openssl/pem.h>
#include <openssl/err.h>
#include <openssl/kdf.h>
#include <iostream>
#include <sstream>
#include <iomanip>
#include <random>
#include <chrono>

namespace gaokao {
namespace crypto {

// Base64编码表
static const std::string base64_chars = 
    "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
    "abcdefghijklmnopqrstuvwxyz"
    "0123456789+/";

// 检查字符是否为Base64字符
static inline bool is_base64(unsigned char c) {
    return (isalnum(c) || (c == '+') || (c == '/'));
}

// AES实现类
class AESCipher::Impl {
public:
    Impl() {
        // OpenSSL初始化
        EVP_add_cipher(EVP_aes_256_cbc());
        EVP_add_cipher(EVP_aes_256_gcm());
        EVP_add_cipher(EVP_aes_256_ctr());
    }
    
    ~Impl() = default;
    
    CryptoError DeriveKey(const std::string& password, 
                         const std::vector<uint8_t>& salt,
                         std::vector<uint8_t>& key) {
        key.resize(32); // AES-256需要32字节密钥
        
        // 使用PBKDF2派生密钥
        int result = PKCS5_PBKDF2_HMAC(
            password.c_str(), password.length(),
            salt.data(), salt.size(),
            10000, // 迭代次数
            EVP_sha256(),
            key.size(),
            key.data()
        );
        
        return (result == 1) ? CryptoError::SUCCESS : CryptoError::KEY_GENERATION_FAILED;
    }
};

// RSA实现类
class RSACipher::Impl {
public:
    Impl() = default;
    ~Impl() = default;
    
    EVP_PKEY* CreateRSAKeyPair(int key_size) {
        EVP_PKEY_CTX* ctx = EVP_PKEY_CTX_new_id(EVP_PKEY_RSA, NULL);
        if (!ctx) return nullptr;
        
        if (EVP_PKEY_keygen_init(ctx) <= 0) {
            EVP_PKEY_CTX_free(ctx);
            return nullptr;
        }
        
        if (EVP_PKEY_CTX_set_rsa_keygen_bits(ctx, key_size) <= 0) {
            EVP_PKEY_CTX_free(ctx);
            return nullptr;
        }
        
        EVP_PKEY* key = nullptr;
        if (EVP_PKEY_keygen(ctx, &key) <= 0) {
            EVP_PKEY_CTX_free(ctx);
            return nullptr;
        }
        
        EVP_PKEY_CTX_free(ctx);
        return key;
    }
};

// AESCipher实现
AESCipher::AESCipher() : pimpl_(std::make_unique<Impl>()) {}
AESCipher::~AESCipher() = default;

CryptoError AESCipher::GenerateKey(uint32_t key_size, std::vector<uint8_t>& key) {
    if (key_size != 128 && key_size != 192 && key_size != 256) {
        return CryptoError::INVALID_PARAMETER;
    }
    
    key.resize(key_size / 8);
    
    if (RAND_bytes(key.data(), key.size()) != 1) {
        return CryptoError::RANDOM_GENERATION_FAILED;
    }
    
    return CryptoError::SUCCESS;
}

CryptoError AESCipher::Encrypt(const std::vector<uint8_t>& plaintext,
                              const std::vector<uint8_t>& key,
                              CryptoAlgorithm algorithm,
                              EncryptionResult& result) {
    
    const EVP_CIPHER* cipher = nullptr;
    switch (algorithm) {
        case CryptoAlgorithm::AES_256_CBC:
            cipher = EVP_aes_256_cbc();
            break;
        case CryptoAlgorithm::AES_256_GCM:
            cipher = EVP_aes_256_gcm();
            break;
        case CryptoAlgorithm::AES_256_CTR:
            cipher = EVP_aes_256_ctr();
            break;
        default:
            return CryptoError::ALGORITHM_NOT_SUPPORTED;
    }
    
    EVP_CIPHER_CTX* ctx = EVP_CIPHER_CTX_new();
    if (!ctx) {
        return CryptoError::MEMORY_ERROR;
    }
    
    // 生成随机IV
    result.iv.resize(EVP_CIPHER_iv_length(cipher));
    if (RAND_bytes(result.iv.data(), result.iv.size()) != 1) {
        EVP_CIPHER_CTX_free(ctx);
        return CryptoError::RANDOM_GENERATION_FAILED;
    }
    
    // 初始化加密
    if (EVP_EncryptInit_ex(ctx, cipher, NULL, key.data(), result.iv.data()) != 1) {
        EVP_CIPHER_CTX_free(ctx);
        return CryptoError::ENCRYPTION_FAILED;
    }
    
    // 加密数据
    result.encrypted_data.resize(plaintext.size() + EVP_CIPHER_block_size(cipher));
    int len = 0;
    int ciphertext_len = 0;
    
    if (EVP_EncryptUpdate(ctx, result.encrypted_data.data(), &len, 
                         plaintext.data(), plaintext.size()) != 1) {
        EVP_CIPHER_CTX_free(ctx);
        return CryptoError::ENCRYPTION_FAILED;
    }
    ciphertext_len = len;
    
    if (EVP_EncryptFinal_ex(ctx, result.encrypted_data.data() + len, &len) != 1) {
        EVP_CIPHER_CTX_free(ctx);
        return CryptoError::ENCRYPTION_FAILED;
    }
    ciphertext_len += len;
    
    result.encrypted_data.resize(ciphertext_len);
    
    // 如果是GCM模式，获取认证标签
    if (algorithm == CryptoAlgorithm::AES_256_GCM) {
        result.tag.resize(16);
        if (EVP_CIPHER_CTX_ctrl(ctx, EVP_CTRL_GCM_GET_TAG, 16, result.tag.data()) != 1) {
            EVP_CIPHER_CTX_free(ctx);
            return CryptoError::ENCRYPTION_FAILED;
        }
    }
    
    result.algorithm = algorithm;
    EVP_CIPHER_CTX_free(ctx);
    return CryptoError::SUCCESS;
}

CryptoError AESCipher::Decrypt(const EncryptionResult& result,
                              const std::vector<uint8_t>& key,
                              std::vector<uint8_t>& plaintext) {
    
    const EVP_CIPHER* cipher = nullptr;
    switch (result.algorithm) {
        case CryptoAlgorithm::AES_256_CBC:
            cipher = EVP_aes_256_cbc();
            break;
        case CryptoAlgorithm::AES_256_GCM:
            cipher = EVP_aes_256_gcm();
            break;
        case CryptoAlgorithm::AES_256_CTR:
            cipher = EVP_aes_256_ctr();
            break;
        default:
            return CryptoError::ALGORITHM_NOT_SUPPORTED;
    }
    
    EVP_CIPHER_CTX* ctx = EVP_CIPHER_CTX_new();
    if (!ctx) {
        return CryptoError::MEMORY_ERROR;
    }
    
    // 初始化解密
    if (EVP_DecryptInit_ex(ctx, cipher, NULL, key.data(), result.iv.data()) != 1) {
        EVP_CIPHER_CTX_free(ctx);
        return CryptoError::DECRYPTION_FAILED;
    }
    
    // 如果是GCM模式，设置认证标签
    if (result.algorithm == CryptoAlgorithm::AES_256_GCM) {
        if (EVP_CIPHER_CTX_ctrl(ctx, EVP_CTRL_GCM_SET_TAG, 
                               result.tag.size(), (void*)result.tag.data()) != 1) {
            EVP_CIPHER_CTX_free(ctx);
            return CryptoError::DECRYPTION_FAILED;
        }
    }
    
    // 解密数据
    plaintext.resize(result.encrypted_data.size() + EVP_CIPHER_block_size(cipher));
    int len = 0;
    int plaintext_len = 0;
    
    if (EVP_DecryptUpdate(ctx, plaintext.data(), &len, 
                         result.encrypted_data.data(), result.encrypted_data.size()) != 1) {
        EVP_CIPHER_CTX_free(ctx);
        return CryptoError::DECRYPTION_FAILED;
    }
    plaintext_len = len;
    
    if (EVP_DecryptFinal_ex(ctx, plaintext.data() + len, &len) != 1) {
        EVP_CIPHER_CTX_free(ctx);
        return CryptoError::VERIFICATION_FAILED;
    }
    plaintext_len += len;
    
    plaintext.resize(plaintext_len);
    EVP_CIPHER_CTX_free(ctx);
    return CryptoError::SUCCESS;
}

CryptoError AESCipher::EncryptString(const std::string& plaintext,
                                    const std::string& password,
                                    std::string& encrypted_base64) {
    
    // 生成盐值
    std::vector<uint8_t> salt(16);
    if (RAND_bytes(salt.data(), salt.size()) != 1) {
        return CryptoError::RANDOM_GENERATION_FAILED;
    }
    
    // 派生密钥
    std::vector<uint8_t> key;
    auto result = pimpl_->DeriveKey(password, salt, key);
    if (result != CryptoError::SUCCESS) {
        return result;
    }
    
    // 加密
    std::vector<uint8_t> plaintext_bytes(plaintext.begin(), plaintext.end());
    EncryptionResult encrypt_result;
    encrypt_result.salt = salt;
    
    result = Encrypt(plaintext_bytes, key, CryptoAlgorithm::AES_256_CBC, encrypt_result);
    if (result != CryptoError::SUCCESS) {
        return result;
    }
    
    // 组合盐值、IV和密文
    std::vector<uint8_t> combined;
    combined.insert(combined.end(), salt.begin(), salt.end());
    combined.insert(combined.end(), encrypt_result.iv.begin(), encrypt_result.iv.end());
    combined.insert(combined.end(), encrypt_result.encrypted_data.begin(), encrypt_result.encrypted_data.end());
    
    // Base64编码
    encrypted_base64 = EncodingUtils::Base64Encode(combined);
    
    return CryptoError::SUCCESS;
}

CryptoError AESCipher::DecryptString(const std::string& encrypted_base64,
                                    const std::string& password,
                                    std::string& plaintext) {
    
    // Base64解码
    std::vector<uint8_t> combined;
    if (!EncodingUtils::Base64Decode(encrypted_base64, combined)) {
        return CryptoError::INVALID_PARAMETER;
    }
    
    if (combined.size() < 32) { // 至少需要16字节盐值 + 16字节IV
        return CryptoError::INVALID_PARAMETER;
    }
    
    // 提取盐值、IV和密文
    std::vector<uint8_t> salt(combined.begin(), combined.begin() + 16);
    std::vector<uint8_t> iv(combined.begin() + 16, combined.begin() + 32);
    std::vector<uint8_t> encrypted_data(combined.begin() + 32, combined.end());
    
    // 派生密钥
    std::vector<uint8_t> key;
    auto result = pimpl_->DeriveKey(password, salt, key);
    if (result != CryptoError::SUCCESS) {
        return result;
    }
    
    // 构建解密结果
    EncryptionResult encrypt_result;
    encrypt_result.encrypted_data = encrypted_data;
    encrypt_result.iv = iv;
    encrypt_result.algorithm = CryptoAlgorithm::AES_256_CBC;
    
    // 解密
    std::vector<uint8_t> plaintext_bytes;
    result = Decrypt(encrypt_result, key, plaintext_bytes);
    if (result != CryptoError::SUCCESS) {
        return result;
    }
    
    plaintext = std::string(plaintext_bytes.begin(), plaintext_bytes.end());
    return CryptoError::SUCCESS;
}

// RSACipher实现
RSACipher::RSACipher() : pimpl_(std::make_unique<Impl>()) {}
RSACipher::~RSACipher() = default;

CryptoError RSACipher::GenerateKeyPair(uint32_t key_size, KeyPair& key_pair) {
    if (key_size != 2048 && key_size != 4096) {
        return CryptoError::INVALID_PARAMETER;
    }
    
    EVP_PKEY* pkey = pimpl_->CreateRSAKeyPair(key_size);
    if (!pkey) {
        return CryptoError::KEY_GENERATION_FAILED;
    }
    
    // 提取公钥
    BIO* pub_bio = BIO_new(BIO_s_mem());
    if (PEM_write_bio_PUBKEY(pub_bio, pkey) != 1) {
        EVP_PKEY_free(pkey);
        BIO_free(pub_bio);
        return CryptoError::KEY_GENERATION_FAILED;
    }
    
    char* pub_data = nullptr;
    long pub_len = BIO_get_mem_data(pub_bio, &pub_data);
    key_pair.public_key.assign(pub_data, pub_data + pub_len);
    BIO_free(pub_bio);
    
    // 提取私钥
    BIO* priv_bio = BIO_new(BIO_s_mem());
    if (PEM_write_bio_PrivateKey(priv_bio, pkey, NULL, NULL, 0, NULL, NULL) != 1) {
        EVP_PKEY_free(pkey);
        BIO_free(priv_bio);
        return CryptoError::KEY_GENERATION_FAILED;
    }
    
    char* priv_data = nullptr;
    long priv_len = BIO_get_mem_data(priv_bio, &priv_data);
    key_pair.private_key.assign(priv_data, priv_data + priv_len);
    BIO_free(priv_bio);
    
    key_pair.algorithm = (key_size == 2048) ? CryptoAlgorithm::RSA_2048 : CryptoAlgorithm::RSA_4096;
    key_pair.key_size = key_size;
    
    EVP_PKEY_free(pkey);
    return CryptoError::SUCCESS;
}

// 哈希工具实现
CryptoError HashUtils::CalculateHash(const std::vector<uint8_t>& data,
                                    HashAlgorithm algorithm,
                                    std::vector<uint8_t>& hash) {
    
    const EVP_MD* md = nullptr;
    switch (algorithm) {
        case HashAlgorithm::SHA256:
            md = EVP_sha256();
            hash.resize(SHA256_DIGEST_LENGTH);
            break;
        case HashAlgorithm::SHA384:
            md = EVP_sha384();
            hash.resize(SHA384_DIGEST_LENGTH);
            break;
        case HashAlgorithm::SHA512:
            md = EVP_sha512();
            hash.resize(SHA512_DIGEST_LENGTH);
            break;
        default:
            return CryptoError::ALGORITHM_NOT_SUPPORTED;
    }
    
    EVP_MD_CTX* ctx = EVP_MD_CTX_new();
    if (!ctx) {
        return CryptoError::MEMORY_ERROR;
    }
    
    if (EVP_DigestInit_ex(ctx, md, NULL) != 1 ||
        EVP_DigestUpdate(ctx, data.data(), data.size()) != 1 ||
        EVP_DigestFinal_ex(ctx, hash.data(), NULL) != 1) {
        EVP_MD_CTX_free(ctx);
        return CryptoError::HASH_FAILED;
    }
    
    EVP_MD_CTX_free(ctx);
    return CryptoError::SUCCESS;
}

CryptoError HashUtils::CalculateStringHash(const std::string& data,
                                          HashAlgorithm algorithm,
                                          std::string& hash_hex) {
    
    std::vector<uint8_t> data_bytes(data.begin(), data.end());
    std::vector<uint8_t> hash;
    
    auto result = CalculateHash(data_bytes, algorithm, hash);
    if (result != CryptoError::SUCCESS) {
        return result;
    }
    
    hash_hex = EncodingUtils::HexEncode(hash);
    return CryptoError::SUCCESS;
}

CryptoError HashUtils::GenerateSalt(uint32_t size, std::vector<uint8_t>& salt) {
    salt.resize(size);
    
    if (RAND_bytes(salt.data(), size) != 1) {
        return CryptoError::RANDOM_GENERATION_FAILED;
    }
    
    return CryptoError::SUCCESS;
}

// 编码工具实现
std::string EncodingUtils::Base64Encode(const std::vector<uint8_t>& data) {
    std::string encoded;
    int val = 0, valb = -6;
    
    for (uint8_t c : data) {
        val = (val << 8) + c;
        valb += 8;
        while (valb >= 0) {
            encoded.push_back(base64_chars[(val >> valb) & 0x3F]);
            valb -= 6;
        }
    }
    
    if (valb > -6) {
        encoded.push_back(base64_chars[((val << 8) >> (valb + 8)) & 0x3F]);
    }
    
    while (encoded.size() % 4) {
        encoded.push_back('=');
    }
    
    return encoded;
}

bool EncodingUtils::Base64Decode(const std::string& encoded, std::vector<uint8_t>& data) {
    std::vector<int> table(256, -1);
    for (int i = 0; i < 64; i++) {
        table[base64_chars[i]] = i;
    }
    
    int val = 0, valb = -8;
    for (unsigned char c : encoded) {
        if (table[c] == -1) break;
        val = (val << 6) + table[c];
        valb += 6;
        if (valb >= 0) {
            data.push_back(char((val >> valb) & 0xFF));
            valb -= 8;
        }
    }
    
    return true;
}

std::string EncodingUtils::HexEncode(const std::vector<uint8_t>& data, bool uppercase) {
    std::ostringstream oss;
    oss << std::hex << std::setfill('0');
    if (uppercase) {
        oss << std::uppercase;
    }
    
    for (uint8_t byte : data) {
        oss << std::setw(2) << static_cast<unsigned>(byte);
    }
    
    return oss.str();
}

<<<<<<< HEAD
bool EncodingUtils::HexDecode(const std::string& hex_string, std::vector<uint8_t>& data) {
    data.clear();
    
    // 检查字符串长度是否为偶数
    if (hex_string.length() % 2 != 0) {
        return false;
    }
    
    data.reserve(hex_string.length() / 2);
    
    for (size_t i = 0; i < hex_string.length(); i += 2) {
        std::string byte_string = hex_string.substr(i, 2);
        
        // 检查字符是否为有效的十六进制字符
        for (char c : byte_string) {
            if (!std::isxdigit(c)) {
                data.clear();
                return false;
            }
        }
        
        try {
            uint8_t byte = static_cast<uint8_t>(std::stoul(byte_string, nullptr, 16));
            data.push_back(byte);
        } catch (const std::exception&) {
            data.clear();
            return false;
        }
    }
    
    return true;
}

=======
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
std::vector<uint8_t> EncodingUtils::StringToBytes(const std::string& str) {
    return std::vector<uint8_t>(str.begin(), str.end());
}

std::string EncodingUtils::BytesToString(const std::vector<uint8_t>& data) {
    return std::string(data.begin(), data.end());
}

// 随机数生成器实现
CryptoError RandomGenerator::GenerateRandomBytes(uint32_t size, 
                                                std::vector<uint8_t>& random_bytes) {
    random_bytes.resize(size);
    
    if (RAND_bytes(random_bytes.data(), size) != 1) {
        return CryptoError::RANDOM_GENERATION_FAILED;
    }
    
    return CryptoError::SUCCESS;
}

std::string RandomGenerator::GenerateRandomString(uint32_t length, 
                                                 const std::string& charset) {
    std::random_device rd;
    std::mt19937 gen(rd());
    std::uniform_int_distribution<> dis(0, charset.size() - 1);
    
    std::string result;
    result.reserve(length);
    
    for (uint32_t i = 0; i < length; ++i) {
        result += charset[dis(gen)];
    }
    
    return result;
}

std::string RandomGenerator::GenerateUUID() {
    std::vector<uint8_t> bytes;
    auto result = GenerateRandomBytes(16, bytes);
    if (result != CryptoError::SUCCESS) {
        return "";
    }
    
    // 设置UUID版本(4)和变体位
    bytes[6] = (bytes[6] & 0x0F) | 0x40; // 版本4
    bytes[8] = (bytes[8] & 0x3F) | 0x80; // 变体10
    
    std::ostringstream oss;
<<<<<<< HEAD
    oss << std::hex << std::setfill('0') << std::nouppercase;
=======
    oss << std::hex << std::setfill('0') << std::lowercase;
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
    
    for (size_t i = 0; i < 16; ++i) {
        if (i == 4 || i == 6 || i == 8 || i == 10) {
            oss << '-';
        }
        oss << std::setw(2) << static_cast<unsigned>(bytes[i]);
    }
    
    return oss.str();
}

uint32_t RandomGenerator::GenerateRandomInt(uint32_t min, uint32_t max) {
    std::random_device rd;
    std::mt19937 gen(rd());
    std::uniform_int_distribution<uint32_t> dis(min, max);
    
    return dis(gen);
}

// 错误描述
std::string GetCryptoErrorDescription(CryptoError error) {
    switch (error) {
        case CryptoError::SUCCESS:
            return "操作成功";
        case CryptoError::INVALID_PARAMETER:
            return "无效参数";
        case CryptoError::KEY_GENERATION_FAILED:
            return "密钥生成失败";
        case CryptoError::ENCRYPTION_FAILED:
            return "加密失败";
        case CryptoError::DECRYPTION_FAILED:
            return "解密失败";
        case CryptoError::SIGNATURE_FAILED:
            return "签名失败";
        case CryptoError::VERIFICATION_FAILED:
            return "验证失败";
        case CryptoError::HASH_FAILED:
            return "哈希计算失败";
        case CryptoError::KEY_INVALID:
            return "无效密钥";
        case CryptoError::ALGORITHM_NOT_SUPPORTED:
            return "算法不支持";
        case CryptoError::MEMORY_ERROR:
            return "内存错误";
        case CryptoError::RANDOM_GENERATION_FAILED:
            return "随机数生成失败";
        default:
            return "未知错误";
    }
}

} // namespace crypto
} // namespace gaokao