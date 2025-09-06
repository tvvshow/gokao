package shared

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"
)

// HTTPClient 创建一个带超时的HTTP客户端
func HTTPClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}

// GenerateSignature 生成HMAC-SHA256签名
func GenerateSignature(clientID, nonce, timestamp, secretKey string) string {
	message := clientID + nonce + timestamp
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}