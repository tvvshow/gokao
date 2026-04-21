//go:build cgo && !windows
// +build cgo,!windows

package cpp

/*
#cgo CPPFLAGS: -I../../../cpp-modules/device-fingerprint/include
#cgo LDFLAGS: -L../../../cpp-modules/device-fingerprint/lib -ldevice_fingerprint -lstdc++

#include "c_interface.h"
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"unsafe"
)

// CryptoService 加密服务接口
type CryptoService struct {
	initialized bool
}

// NewCryptoService 创建新的加密服务
func NewCryptoService() *CryptoService {
	return &CryptoService{
		initialized: true,
	}
}

// EncryptData 加密数据
func (c *CryptoService) EncryptData(data []byte, key string) ([]byte, error) {
	if !c.initialized {
		return nil, errors.New("crypto service not initialized")
	}

	if len(data) == 0 {
		return nil, errors.New("data is empty")
	}

	if key == "" {
		return nil, errors.New("encryption key is empty")
	}

	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))

	// 准备输出缓冲区（通常加密后数据会比原数据大）
	encryptedBuffer := make([]byte, len(data)+256) // 预留足够空间
	var actualSize C.size_t

	result := C.DeviceFingerprint_Encrypt(
		(*C.char)(unsafe.Pointer(&data[0])),
		C.size_t(len(data)),
		cKey,
		(*C.char)(unsafe.Pointer(&encryptedBuffer[0])),
		C.size_t(len(encryptedBuffer)),
		&actualSize)

	if result != C.C_SUCCESS {
		return nil, c.convertError(result)
	}

	// 返回实际大小的数据
	return encryptedBuffer[:actualSize], nil
}

// DecryptData 解密数据
func (c *CryptoService) DecryptData(encryptedData []byte, key string) ([]byte, error) {
	if !c.initialized {
		return nil, errors.New("crypto service not initialized")
	}

	if len(encryptedData) == 0 {
		return nil, errors.New("encrypted data is empty")
	}

	if key == "" {
		return nil, errors.New("decryption key is empty")
	}

	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))

	// 准备输出缓冲区
	decryptedBuffer := make([]byte, len(encryptedData)+256)
	var actualSize C.size_t

	result := C.DeviceFingerprint_Decrypt(
		(*C.char)(unsafe.Pointer(&encryptedData[0])),
		C.size_t(len(encryptedData)),
		cKey,
		(*C.char)(unsafe.Pointer(&decryptedBuffer[0])),
		C.size_t(len(decryptedBuffer)),
		&actualSize)

	if result != C.C_SUCCESS {
		return nil, c.convertError(result)
	}

	// 返回实际大小的数据
	return decryptedBuffer[:actualSize], nil
}

// SignData 生成数字签名
func (c *CryptoService) SignData(data []byte, privateKey string) ([]byte, error) {
	if !c.initialized {
		return nil, errors.New("crypto service not initialized")
	}

	if len(data) == 0 {
		return nil, errors.New("data is empty")
	}

	if privateKey == "" {
		return nil, errors.New("private key is empty")
	}

	cPrivateKey := C.CString(privateKey)
	defer C.free(unsafe.Pointer(cPrivateKey))

	// 准备签名缓冲区
	signatureBuffer := make([]byte, 512) // 通常签名大小固定
	var actualSize C.size_t

	result := C.DeviceFingerprint_Sign(
		(*C.char)(unsafe.Pointer(&data[0])),
		C.size_t(len(data)),
		cPrivateKey,
		(*C.char)(unsafe.Pointer(&signatureBuffer[0])),
		C.size_t(len(signatureBuffer)),
		&actualSize)

	if result != C.C_SUCCESS {
		return nil, c.convertError(result)
	}

	return signatureBuffer[:actualSize], nil
}

// VerifySignature 验证数字签名
func (c *CryptoService) VerifySignature(data []byte, signature []byte, publicKey string) (bool, error) {
	if !c.initialized {
		return false, errors.New("crypto service not initialized")
	}

	if len(data) == 0 {
		return false, errors.New("data is empty")
	}

	if len(signature) == 0 {
		return false, errors.New("signature is empty")
	}

	if publicKey == "" {
		return false, errors.New("public key is empty")
	}

	cPublicKey := C.CString(publicKey)
	defer C.free(unsafe.Pointer(cPublicKey))

	var isValid C.int

	result := C.DeviceFingerprint_VerifySignature(
		(*C.char)(unsafe.Pointer(&data[0])),
		C.size_t(len(data)),
		(*C.char)(unsafe.Pointer(&signature[0])),
		C.size_t(len(signature)),
		cPublicKey,
		&isValid)

	if result != C.C_SUCCESS {
		return false, c.convertError(result)
	}

	return isValid != 0, nil
}

// EncryptFingerprint 加密设备指纹
func (c *CryptoService) EncryptFingerprint(fingerprint *DeviceFingerprint, key string) (*DeviceFingerprint, error) {
	if fingerprint == nil {
		return nil, errors.New("fingerprint is nil")
	}

	// 序列化指纹为JSON
	collector := NewDeviceFingerprintCollector()
	if err := collector.Initialize(""); err != nil {
		return nil, err
	}
	defer collector.Uninitialize()

	jsonData, err := collector.SerializeToJSON(fingerprint)
	if err != nil {
		return nil, err
	}

	// 加密JSON数据
	encryptedData, err := c.EncryptData([]byte(jsonData), key)
	if err != nil {
		return nil, err
	}

	// 创建新的指纹副本，包含加密数据
	encryptedFingerprint := *fingerprint
	encryptedFingerprint.ErrorMessage = "" // 清除错误信息

	// 将加密数据存储在特定字段中（这里需要扩展结构体）
	// 暂时将加密数据转换为十六进制字符串存储在设备ID中
	encryptedFingerprint.DeviceID = c.bytesToHex(encryptedData)

	return &encryptedFingerprint, nil
}

// DecryptFingerprint 解密设备指纹
func (c *CryptoService) DecryptFingerprint(encryptedFingerprint *DeviceFingerprint, key string) (*DeviceFingerprint, error) {
	if encryptedFingerprint == nil {
		return nil, errors.New("encrypted fingerprint is nil")
	}

	// 从设备ID中提取加密数据（十六进制字符串）
	encryptedData, err := c.hexToBytes(encryptedFingerprint.DeviceID)
	if err != nil {
		return nil, err
	}

	// 解密数据
	decryptedData, err := c.DecryptData(encryptedData, key)
	if err != nil {
		return nil, err
	}

	// 反序列化JSON数据
	collector := NewDeviceFingerprintCollector()
	if err := collector.Initialize(""); err != nil {
		return nil, err
	}
	defer collector.Uninitialize()

	fingerprint, err := collector.DeserializeFromJSON(string(decryptedData))
	if err != nil {
		return nil, err
	}

	return fingerprint, nil
}

// HashData 计算数据哈希
func (c *CryptoService) HashData(data []byte) (string, error) {
	if !c.initialized {
		return "", errors.New("crypto service not initialized")
	}

	if len(data) == 0 {
		return "", errors.New("data is empty")
	}

	// 创建临时指纹结构体
	tempFingerprint := &DeviceFingerprint{
		DeviceID: string(data),
	}

	// 使用设备指纹采集器生成哈希
	collector := NewDeviceFingerprintCollector()
	if err := collector.Initialize(""); err != nil {
		return "", err
	}
	defer collector.Uninitialize()

	hash, err := collector.GenerateHash(tempFingerprint)
	if err != nil {
		return "", err
	}

	return hash, nil
}

// ValidateDataIntegrity 验证数据完整性
func (c *CryptoService) ValidateDataIntegrity(data []byte, expectedHash string) (bool, error) {
	if len(data) == 0 {
		return false, errors.New("data is empty")
	}

	if expectedHash == "" {
		return false, errors.New("expected hash is empty")
	}

	// 计算数据哈希
	actualHash, err := c.HashData(data)
	if err != nil {
		return false, err
	}

	// 比较哈希值
	return actualHash == expectedHash, nil
}

// GenerateSecureKey 生成安全密钥
func (c *CryptoService) GenerateSecureKey(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("key length must be positive")
	}

	if length > 256 {
		return "", errors.New("key length too large")
	}

	// 使用系统时间和设备信息生成种子
	fingerprint, err := QuickCollectFingerprint()
	if err != nil {
		return "", err
	}

	// 基于设备指纹生成密钥
	seedData := fingerprint.DeviceID + fingerprint.CPUID + fingerprint.MotherboardSerial
	hash, err := c.HashData([]byte(seedData))
	if err != nil {
		return "", err
	}

	// 截取指定长度
	if len(hash) > length {
		return hash[:length], nil
	}

	return hash, nil
}

// 内部辅助函数

// convertError 转换C错误码为Go错误
func (c *CryptoService) convertError(cError C.CErrorCode) error {
	return errors.New(getErrorDescription(cError))
}

// bytesToHex 将字节数组转换为十六进制字符串
func (c *CryptoService) bytesToHex(data []byte) string {
	const hexChars = "0123456789abcdef"
	result := make([]byte, len(data)*2)

	for i, b := range data {
		result[i*2] = hexChars[b>>4]
		result[i*2+1] = hexChars[b&0x0f]
	}

	return string(result)
}

// hexToBytes 将十六进制字符串转换为字节数组
func (c *CryptoService) hexToBytes(hexStr string) ([]byte, error) {
	if len(hexStr)%2 != 0 {
		return nil, errors.New("invalid hex string length")
	}

	result := make([]byte, len(hexStr)/2)

	for i := 0; i < len(hexStr); i += 2 {
		b1, err := c.hexCharToByte(hexStr[i])
		if err != nil {
			return nil, err
		}

		b2, err := c.hexCharToByte(hexStr[i+1])
		if err != nil {
			return nil, err
		}

		result[i/2] = (b1 << 4) | b2
	}

	return result, nil
}

// hexCharToByte 将十六进制字符转换为字节
func (c *CryptoService) hexCharToByte(ch byte) (byte, error) {
	switch {
	case ch >= '0' && ch <= '9':
		return ch - '0', nil
	case ch >= 'a' && ch <= 'f':
		return ch - 'a' + 10, nil
	case ch >= 'A' && ch <= 'F':
		return ch - 'A' + 10, nil
	default:
		return 0, errors.New("invalid hex character")
	}
}

// SecureClearMemory 安全清除内存中的敏感数据
func (c *CryptoService) SecureClearMemory(data []byte) {
	for i := range data {
		data[i] = 0
	}
}

// ValidateKey 验证密钥格式
func (c *CryptoService) ValidateKey(key string) error {
	if key == "" {
		return errors.New("key is empty")
	}

	if len(key) < 8 {
		return errors.New("key too short (minimum 8 characters)")
	}

	if len(key) > 256 {
		return errors.New("key too long (maximum 256 characters)")
	}

	return nil
}
