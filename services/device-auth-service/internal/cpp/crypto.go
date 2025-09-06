package cpp

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
)

// CryptoService 加密服务
type CryptoService struct{}

// NewCryptoService 创建新的加密服务实例
func NewCryptoService() *CryptoService {
	return &CryptoService{}
}

// GenerateRSAKeyPair 生成RSA密钥对
func (c *CryptoService) GenerateRSAKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}

	return privateKey, &privateKey.PublicKey, nil
}

// EncryptWithPublicKey 使用公钥加密数据
func (c *CryptoService) EncryptWithPublicKey(data []byte, pub *rsa.PublicKey) ([]byte, error) {
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, data, nil)
}

// DecryptWithPrivateKey 使用私钥解密数据
func (c *CryptoService) DecryptWithPrivateKey(ciphertext []byte, priv *rsa.PrivateKey) ([]byte, error) {
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, ciphertext, nil)
}

// SignData 使用私钥对数据进行签名
func (c *CryptoService) SignData(data []byte, privateKeyPEM string) ([]byte, error) {
	// 解析私钥
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the private key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	// 计算数据哈希
	hashed := sha256.Sum256(data)

	// 签名
	return rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hashed[:])
}

// VerifySignature 使用公钥验证签名
func (c *CryptoService) VerifySignature(data, signature []byte, publicKeyPEM string) (bool, error) {
	// 解析公钥
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return false, errors.New("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return false, err
	}

	// 计算数据哈希
	hashed := sha256.Sum256(data)

	// 验证签名
	err = rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed[:], signature)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ExportPrivateKey 导出私钥为PEM格式
func (c *CryptoService) ExportPrivateKey(privateKey *rsa.PrivateKey) string {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		},
	)
	return string(privateKeyPEM)
}

// ExportPublicKey 导出公钥为PEM格式
func (c *CryptoService) ExportPublicKey(publicKey *rsa.PublicKey) string {
	publicKeyBytes, _ := x509.MarshalPKCS1PublicKey(publicKey)
	publicKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: publicKeyBytes,
		},
	)
	return string(publicKeyPEM)
}

// ImportPrivateKey 从PEM格式导入私钥
func (c *CryptoService) ImportPrivateKey(privateKeyPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the private key")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// ImportPublicKey 从PEM格式导入公钥
func (c *CryptoService) ImportPublicKey(publicKeyPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the public key")
	}

	return x509.ParsePKCS1PublicKey(block.Bytes)
}

// GenerateRandomBytes 生成随机字节
func (c *CryptoService) GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// bytesToHex 将字节切片转换为十六进制字符串
func (c *CryptoService) bytesToHex(data []byte) string {
	return hex.EncodeToString(data)
}

// hexToBytes 将十六进制字符串转换为字节切片
func (c *CryptoService) hexToBytes(hexStr string) ([]byte, error) {
	return hex.DecodeString(hexStr)
}

// CalculateHash 计算数据的SHA256哈希值
func (c *CryptoService) CalculateHash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// VerifyHash 验证数据的哈希值
func (c *CryptoService) VerifyHash(data []byte, hash []byte) bool {
	calculatedHash := c.CalculateHash(data)
	return fmt.Sprintf("%x", calculatedHash) == fmt.Sprintf("%x", hash)
}