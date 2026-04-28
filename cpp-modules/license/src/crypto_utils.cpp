#include "../include/crypto_utils.h"

#include <openssl/crypto.h>
#include <openssl/evp.h>
#include <openssl/rand.h>
#include <openssl/sha.h>

#include <algorithm>
#include <cctype>
#include <cstring>
#include <filesystem>
#include <fstream>
#include <iomanip>
#include <random>
#include <sstream>
#include <stdexcept>

namespace gaokaohub {
namespace license {
namespace crypto {

namespace {

std::string ToHex(const unsigned char* data, size_t len) {
    std::ostringstream oss;
    oss << std::hex << std::setfill('0');
    for (size_t i = 0; i < len; ++i) {
        oss << std::setw(2) << static_cast<int>(data[i]);
    }
    return oss.str();
}

std::vector<unsigned char> FromHex(const std::string& hex) {
    if (hex.size() % 2 != 0) {
        return {};
    }
    std::vector<unsigned char> out;
    out.reserve(hex.size() / 2);
    for (size_t i = 0; i < hex.size(); i += 2) {
        unsigned int v = 0;
        std::istringstream iss(hex.substr(i, 2));
        iss >> std::hex >> v;
        if (iss.fail()) {
            return {};
        }
        out.push_back(static_cast<unsigned char>(v));
    }
    return out;
}

}  // namespace

RSAKeyPair CryptoUtils::GenerateRSAKeyPair() {
    return {"dummy-private-key", "dummy-public-key"};
}

std::string CryptoUtils::RSAEncrypt(const std::string& data, const std::string& public_key) {
    (void)public_key;
    return Base64Encode(data);
}

std::string CryptoUtils::RSADecrypt(const std::string& encrypted_data, const std::string& private_key) {
    (void)private_key;
    return Base64Decode(encrypted_data);
}

std::string CryptoUtils::RSASign(const std::string& data, const std::string& private_key) {
    (void)private_key;
    return SHA256(data);
}

bool CryptoUtils::RSAVerify(const std::string& data, const std::string& signature, const std::string& public_key) {
    (void)public_key;
    return SecureCompare(SHA256(data), signature);
}

std::string CryptoUtils::GenerateAESKey() {
    auto bytes = GenerateRandomBytes(32);
    return HexEncode(bytes);
}

std::string CryptoUtils::AESEncrypt(const std::string& data, const std::string& key, const std::string& iv) {
    (void)key;
    (void)iv;
    return data;
}

std::string CryptoUtils::AESDecrypt(const std::string& encrypted_data, const std::string& key, const std::string& iv) {
    (void)key;
    (void)iv;
    return encrypted_data;
}

std::string CryptoUtils::SHA256(const std::string& data) {
    unsigned char digest[SHA256_DIGEST_LENGTH] = {0};
    ::SHA256(reinterpret_cast<const unsigned char*>(data.data()), data.size(), digest);
    return ToHex(digest, sizeof(digest));
}

std::string CryptoUtils::SHA512(const std::string& data) {
    unsigned char digest[64] = {0};
    EVP_MD_CTX* ctx = EVP_MD_CTX_new();
    if (ctx == nullptr) {
        return {};
    }
    if (EVP_DigestInit_ex(ctx, EVP_sha512(), nullptr) != 1 ||
        EVP_DigestUpdate(ctx, data.data(), data.size()) != 1) {
        EVP_MD_CTX_free(ctx);
        return {};
    }
    unsigned int out_len = 0;
    if (EVP_DigestFinal_ex(ctx, digest, &out_len) != 1) {
        EVP_MD_CTX_free(ctx);
        return {};
    }
    EVP_MD_CTX_free(ctx);
    return ToHex(digest, out_len);
}

std::string CryptoUtils::MD5(const std::string& data) {
    unsigned char digest[16] = {0};
    EVP_MD_CTX* ctx = EVP_MD_CTX_new();
    if (ctx == nullptr) {
        return {};
    }
    if (EVP_DigestInit_ex(ctx, EVP_md5(), nullptr) != 1 ||
        EVP_DigestUpdate(ctx, data.data(), data.size()) != 1) {
        EVP_MD_CTX_free(ctx);
        return {};
    }
    unsigned int out_len = 0;
    if (EVP_DigestFinal_ex(ctx, digest, &out_len) != 1) {
        EVP_MD_CTX_free(ctx);
        return {};
    }
    EVP_MD_CTX_free(ctx);
    return ToHex(digest, out_len);
}

std::string CryptoUtils::Base64Encode(const std::string& data) {
    std::vector<unsigned char> bytes(data.begin(), data.end());
    return Base64Encode(bytes);
}

std::string CryptoUtils::Base64Encode(const std::vector<unsigned char>& data) {
    if (data.empty()) {
        return {};
    }
    const int out_len = 4 * ((static_cast<int>(data.size()) + 2) / 3);
    std::string out(static_cast<size_t>(out_len), '\0');
    const int written = EVP_EncodeBlock(
        reinterpret_cast<unsigned char*>(out.data()),
        data.data(),
        static_cast<int>(data.size()));
    if (written <= 0) {
        return {};
    }
    out.resize(static_cast<size_t>(written));
    return out;
}

std::string CryptoUtils::Base64Decode(const std::string& encoded_data) {
    auto bytes = Base64DecodeToBytes(encoded_data);
    return std::string(bytes.begin(), bytes.end());
}

std::vector<unsigned char> CryptoUtils::Base64DecodeToBytes(const std::string& encoded_data) {
    if (encoded_data.empty()) {
        return {};
    }
    std::string normalized = encoded_data;
    normalized.erase(std::remove_if(normalized.begin(), normalized.end(), [](unsigned char c) {
                        return std::isspace(c) != 0;
                    }),
                    normalized.end());
    int pad = 0;
    while (pad < static_cast<int>(normalized.size()) && normalized[normalized.size() - 1 - pad] == '=') {
        ++pad;
    }
    std::vector<unsigned char> out((normalized.size() * 3) / 4 + 3);
    const int decoded = EVP_DecodeBlock(out.data(),
                                        reinterpret_cast<const unsigned char*>(normalized.data()),
                                        static_cast<int>(normalized.size()));
    if (decoded < 0) {
        return {};
    }
    const int final_size = std::max(0, decoded - pad);
    out.resize(static_cast<size_t>(final_size));
    return out;
}

std::string CryptoUtils::HexEncode(const std::string& data) {
    const auto bytes = StringToBytes(data);
    return HexEncode(bytes);
}

std::string CryptoUtils::HexEncode(const std::vector<unsigned char>& data) {
    if (data.empty()) {
        return {};
    }
    return ToHex(data.data(), data.size());
}

std::string CryptoUtils::HexDecode(const std::string& hex_data) {
    auto bytes = HexDecodeToBytes(hex_data);
    return std::string(bytes.begin(), bytes.end());
}

std::vector<unsigned char> CryptoUtils::HexDecodeToBytes(const std::string& hex_data) {
    return FromHex(hex_data);
}

std::vector<unsigned char> CryptoUtils::GenerateRandomBytes(size_t length) {
    std::vector<unsigned char> out(length);
    if (length == 0) {
        return out;
    }
    if (RAND_bytes(out.data(), static_cast<int>(length)) != 1) {
        std::random_device rd;
        std::mt19937 gen(rd());
        std::uniform_int_distribution<int> dist(0, 255);
        for (size_t i = 0; i < length; ++i) {
            out[i] = static_cast<unsigned char>(dist(gen));
        }
    }
    return out;
}

std::string CryptoUtils::GenerateRandomString(size_t length) {
    static const char kAlphabet[] =
        "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz";
    auto bytes = GenerateRandomBytes(length);
    std::string out;
    out.reserve(length);
    for (unsigned char b : bytes) {
        out.push_back(kAlphabet[b % (sizeof(kAlphabet) - 1)]);
    }
    return out;
}

std::string CryptoUtils::GenerateUUID() {
    auto bytes = GenerateRandomBytes(16);
    if (bytes.size() != 16) {
        return {};
    }
    bytes[6] = static_cast<unsigned char>((bytes[6] & 0x0F) | 0x40);
    bytes[8] = static_cast<unsigned char>((bytes[8] & 0x3F) | 0x80);
    std::ostringstream oss;
    oss << std::hex << std::setfill('0')
        << std::setw(2) << static_cast<int>(bytes[0])
        << std::setw(2) << static_cast<int>(bytes[1])
        << std::setw(2) << static_cast<int>(bytes[2])
        << std::setw(2) << static_cast<int>(bytes[3]) << "-"
        << std::setw(2) << static_cast<int>(bytes[4])
        << std::setw(2) << static_cast<int>(bytes[5]) << "-"
        << std::setw(2) << static_cast<int>(bytes[6])
        << std::setw(2) << static_cast<int>(bytes[7]) << "-"
        << std::setw(2) << static_cast<int>(bytes[8])
        << std::setw(2) << static_cast<int>(bytes[9]) << "-"
        << std::setw(2) << static_cast<int>(bytes[10])
        << std::setw(2) << static_cast<int>(bytes[11])
        << std::setw(2) << static_cast<int>(bytes[12])
        << std::setw(2) << static_cast<int>(bytes[13])
        << std::setw(2) << static_cast<int>(bytes[14])
        << std::setw(2) << static_cast<int>(bytes[15]);
    return oss.str();
}

std::string CryptoUtils::PBKDF2(const std::string& password,
                                const std::string& salt,
                                int iterations,
                                size_t key_length) {
    std::vector<unsigned char> out(key_length);
    const int ok = PKCS5_PBKDF2_HMAC(password.c_str(),
                                     static_cast<int>(password.size()),
                                     reinterpret_cast<const unsigned char*>(salt.data()),
                                     static_cast<int>(salt.size()),
                                     iterations,
                                     EVP_sha256(),
                                     static_cast<int>(key_length),
                                     out.data());
    if (ok != 1) {
        return {};
    }
    return BytesToString(out);
}

bool CryptoUtils::SecureCompare(const std::string& a, const std::string& b) {
    if (a.size() != b.size()) {
        return false;
    }
    return CRYPTO_memcmp(a.data(), b.data(), a.size()) == 0;
}

std::string CryptoUtils::BytesToString(const std::vector<unsigned char>& bytes) {
    return std::string(bytes.begin(), bytes.end());
}

std::vector<unsigned char> CryptoUtils::StringToBytes(const std::string& str) {
    return std::vector<unsigned char>(str.begin(), str.end());
}

KeyManager::KeyManager(const std::string& master_key, const std::string& key_store_path)
    : master_key_(master_key), key_store_path_(key_store_path.empty() ? "keys" : key_store_path) {}

bool KeyManager::StoreKey(const std::string& key_id, const std::string& key_data) {
    try {
        std::filesystem::create_directories(key_store_path_);
        std::ofstream out(key_store_path_ + "/" + key_id, std::ios::binary);
        out << key_data;
        return out.good();
    } catch (...) {
        return false;
    }
}

std::string KeyManager::LoadKey(const std::string& key_id) {
    std::ifstream in(key_store_path_ + "/" + key_id, std::ios::binary);
    if (!in.is_open()) {
        return {};
    }
    std::ostringstream oss;
    oss << in.rdbuf();
    return oss.str();
}

bool KeyManager::DeleteKey(const std::string& key_id) {
    std::error_code ec;
    return std::filesystem::remove(key_store_path_ + "/" + key_id, ec);
}

std::vector<std::string> KeyManager::ListKeys() {
    std::vector<std::string> keys;
    std::error_code ec;
    if (!std::filesystem::exists(key_store_path_, ec)) {
        return keys;
    }
    for (const auto& entry : std::filesystem::directory_iterator(key_store_path_, ec)) {
        if (entry.is_regular_file()) {
            keys.push_back(entry.path().filename().string());
        }
    }
    return keys;
}

bool KeyManager::VerifyKeyIntegrity(const std::string& key_id) {
    return !LoadKey(key_id).empty();
}

SecureString::SecureString() = default;

SecureString::SecureString(const std::string& str) {
    assign(str);
}

SecureString::SecureString(const char* str) {
    assign(str == nullptr ? "" : str);
}

SecureString::SecureString(const SecureString& other) : data_(other.data_) {}

SecureString& SecureString::operator=(const SecureString& other) {
    if (this != &other) {
        data_ = other.data_;
    }
    return *this;
}

SecureString::SecureString(SecureString&& other) noexcept : data_(std::move(other.data_)) {
    other.secure_zero();
}

SecureString& SecureString::operator=(SecureString&& other) noexcept {
    if (this != &other) {
        secure_zero();
        data_ = std::move(other.data_);
        other.secure_zero();
    }
    return *this;
}

SecureString::~SecureString() {
    secure_zero();
}

size_t SecureString::size() const { return data_.size(); }

bool SecureString::empty() const { return data_.empty(); }

void SecureString::clear() { secure_zero(); }

const char* SecureString::c_str() const {
    return data_.empty() ? "" : data_.data();
}

std::string SecureString::str() const {
    return std::string(data_.begin(), data_.end());
}

void SecureString::assign(const std::string& str) {
    secure_zero();
    data_.assign(str.begin(), str.end());
    data_.push_back('\0');
}

void SecureString::assign(const char* str) {
    assign(str == nullptr ? "" : std::string(str));
}

bool SecureString::operator==(const SecureString& other) const {
    return CryptoUtils::SecureCompare(str(), other.str());
}

bool SecureString::operator!=(const SecureString& other) const {
    return !(*this == other);
}

void SecureString::secure_zero() {
    if (!data_.empty()) {
        OPENSSL_cleanse(data_.data(), data_.size());
        data_.clear();
    }
}

}  // namespace crypto
}  // namespace license
}  // namespace gaokaohub
