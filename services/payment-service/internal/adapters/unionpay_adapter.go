package adapters

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// UnionPayAdapter 银联支付适配器
type UnionPayAdapter struct {
	config     AdapterConfig
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	gatewayURL string
}

// NewUnionPayAdapter 创建银联支付适配器
func NewUnionPayAdapter(config AdapterConfig) PaymentAdapter {
	adapter := &UnionPayAdapter{
		config: config,
	}

	// 解析私钥
	if config.PrivateKey != "" {
		if key, err := parseRSAPrivateKey(config.PrivateKey); err == nil {
			adapter.privateKey = key
		}
	}

	// 解析公钥
	if config.PublicKey != "" {
		if key, err := parseRSAPublicKey(config.PublicKey); err == nil {
			adapter.publicKey = key
		}
	}

	// 设置网关URL
	if !config.IsProd {
		adapter.gatewayURL = "https://gateway.test.95516.com/gateway/api"
	} else {
		adapter.gatewayURL = "https://gateway.95516.com/gateway/api"
	}

	return adapter
}

// GetName 获取支付渠道名称
func (u *UnionPayAdapter) GetName() string {
	return "unionpay"
}

// CreatePayment 创建支付
func (u *UnionPayAdapter) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// 构建支付参数
	params := map[string]string{
		"version":      "5.1.0",
		"encoding":     "UTF-8",
		"signMethod":   "01", // RSA签名
		"txnType":      "01", // 消费
		"txnSubType":   "01", // 前台消费
		"bizType":      "000201", // B2C网关支付
		"channelType":  "07", // PC网银
		"accessType":   "0",  // 直连商户
		"merId":        u.config.AppID,
		"orderId":      req.OrderNo,
		"txnTime":      time.Now().Format("20060102150405"),
		"txnAmt":       strconv.FormatInt(req.Amount.Mul(decimal.NewFromInt(100)).IntPart(), 10),
		"currencyCode": "156", // 人民币
		"backUrl":      req.NotifyURL,
		"frontUrl":     req.ReturnURL,
		"orderDesc":    req.Subject,
	}

	// 生成签名
	signature, err := u.generateSignature(params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signature: %w", err)
	}
	params["signature"] = signature

	// 构建表单HTML
	formHTML := u.buildFormHTML(params)

	return &PaymentResponse{
		OrderNo:   req.OrderNo,
		FormData:  formHTML,
		ExpiredAt: time.Now().Add(req.ExpireTime),
	}, nil
}

// VerifyCallback 验证回调签名
func (u *UnionPayAdapter) VerifyCallback(ctx context.Context, data []byte, signature string) (*PaymentCallback, error) {
	// 解析回调数据
	values, err := url.ParseQuery(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse callback data: %w", err)
	}

	// 提取签名
	callbackSignature := values.Get("signature")
	if callbackSignature == "" {
		return nil, fmt.Errorf("missing signature in callback")
	}

	// 验证签名
	if !u.verifySignature(values, callbackSignature) {
		return nil, fmt.Errorf("invalid signature")
	}

	// 检查响应码
	respCode := values.Get("respCode")
	if respCode != "00" {
		return nil, fmt.Errorf("payment failed: %s - %s", respCode, values.Get("respMsg"))
	}

	// 构建回调响应
	callback := &PaymentCallback{
		OrderNo:        values.Get("orderId"),
		ChannelTradeNo: values.Get("queryId"),
		Status:         "success",
		RawData:        string(data),
		Signature:      callbackSignature,
	}

	// 解析金额（银联返回的是分）
	if txnAmt := values.Get("txnAmt"); txnAmt != "" {
		if amount, err := strconv.ParseInt(txnAmt, 10, 64); err == nil {
			amountDecimal := decimal.NewFromInt(amount).Div(decimal.NewFromInt(100))
			callback.Amount = amountDecimal
			callback.ActualAmount = amountDecimal
		}
	}

	// 解析支付时间
	if txnTime := values.Get("txnTime"); txnTime != "" {
		if paidAt, err := time.Parse("20060102150405", txnTime); err == nil {
			callback.PaidAt = paidAt
		}
	}

	return callback, nil
}

// QueryPayment 查询支付状态
func (u *UnionPayAdapter) QueryPayment(ctx context.Context, req *QueryRequest) (*QueryResponse, error) {
	// 构建查询参数
	params := map[string]string{
		"version":      "5.1.0",
		"encoding":     "UTF-8",
		"signMethod":   "01",
		"txnType":      "00", // 查询
		"txnSubType":   "00",
		"bizType":      "000000",
		"accessType":   "0",
		"merId":        u.config.AppID,
		"orderId":      req.OrderNo,
		"txnTime":      time.Now().Format("20060102150405"),
	}

	// 生成签名
	signature, err := u.generateSignature(params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signature: %w", err)
	}
	params["signature"] = signature

	// 发送请求
	respData, err := u.sendRequest(params)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// 解析响应
	respValues, err := url.ParseQuery(string(respData))
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 验证响应签名
	respSignature := respValues.Get("signature")
	if !u.verifySignature(respValues, respSignature) {
		return nil, fmt.Errorf("invalid response signature")
	}

	// 构建响应
	response := &QueryResponse{
		OrderNo:        respValues.Get("orderId"),
		ChannelTradeNo: respValues.Get("queryId"),
	}

	// 解析金额
	if txnAmt := respValues.Get("txnAmt"); txnAmt != "" {
		if amount, err := strconv.ParseInt(txnAmt, 10, 64); err == nil {
			response.Amount = decimal.NewFromInt(amount).Div(decimal.NewFromInt(100))
		}
	}

	// 解析状态
	respCode := respValues.Get("respCode")
	switch respCode {
	case "00":
		response.Status = "paid"
		if txnTime := respValues.Get("txnTime"); txnTime != "" {
			if paidAt, err := time.Parse("20060102150405", txnTime); err == nil {
				response.PaidAt = &paidAt
			}
		}
	case "03", "04", "05":
		response.Status = "pending"
	case "12":
		response.Status = "canceled"
	default:
		response.Status = "failed"
	}

	return response, nil
}

// CreateRefund 创建退款
func (u *UnionPayAdapter) CreateRefund(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
	// 构建退款参数
	params := map[string]string{
		"version":      "5.1.0",
		"encoding":     "UTF-8",
		"signMethod":   "01",
		"txnType":      "04", // 退货
		"txnSubType":   "00",
		"bizType":      "000201",
		"accessType":   "0",
		"merId":        u.config.AppID,
		"orderId":      req.RefundNo,
		"origQryId":    req.ChannelTradeNo,
		"txnTime":      time.Now().Format("20060102150405"),
		"txnAmt":       strconv.FormatInt(req.Amount.Mul(decimal.NewFromInt(100)).IntPart(), 10),
		"backUrl":      req.NotifyURL,
	}

	// 生成签名
	signature, err := u.generateSignature(params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signature: %w", err)
	}
	params["signature"] = signature

	// 发送请求
	respData, err := u.sendRequest(params)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// 解析响应
	respValues, err := url.ParseQuery(string(respData))
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 构建响应
	response := &RefundResponse{
		RefundNo:   req.RefundNo,
		RefundedAt: time.Now(),
	}

	// 检查响应码
	respCode := respValues.Get("respCode")
	if respCode == "00" {
		response.Status = "success"
	} else {
		response.Status = "failed"
	}

	// 解析退款金额
	if txnAmt := respValues.Get("txnAmt"); txnAmt != "" {
		if amount, err := strconv.ParseInt(txnAmt, 10, 64); err == nil {
			response.Amount = decimal.NewFromInt(amount).Div(decimal.NewFromInt(100))
		}
	}

	return response, nil
}

// QueryRefund 查询退款状态
func (u *UnionPayAdapter) QueryRefund(ctx context.Context, refundNo string) (*RefundResponse, error) {
	// 构建查询参数
	params := map[string]string{
		"version":      "5.1.0",
		"encoding":     "UTF-8",
		"signMethod":   "01",
		"txnType":      "00",
		"txnSubType":   "00",
		"bizType":      "000000",
		"accessType":   "0",
		"merId":        u.config.AppID,
		"orderId":      refundNo,
		"txnTime":      time.Now().Format("20060102150405"),
	}

	// 生成签名
	signature, err := u.generateSignature(params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signature: %w", err)
	}
	params["signature"] = signature

	// 发送请求
	respData, err := u.sendRequest(params)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// 解析响应
	respValues, err := url.ParseQuery(string(respData))
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 构建响应
	response := &RefundResponse{
		RefundNo: refundNo,
	}

	// 检查响应码
	respCode := respValues.Get("respCode")
	if respCode == "00" {
		response.Status = "success"
	} else if respCode == "03" || respCode == "04" || respCode == "05" {
		response.Status = "pending"
	} else {
		response.Status = "failed"
	}

	return response, nil
}

// CloseOrder 关闭订单
func (u *UnionPayAdapter) CloseOrder(ctx context.Context, orderNo string) error {
	// 银联没有专门的关闭订单接口，订单会自动过期
	return nil
}

// 私有方法

// generateSignature 生成RSA签名
func (u *UnionPayAdapter) generateSignature(params map[string]string) (string, error) {
	if u.privateKey == nil {
		return "", fmt.Errorf("private key not configured")
	}

	// 构建签名字符串
	signString := u.buildSignString(params)

	// SHA256哈希
	hash := sha256.Sum256([]byte(signString))

	// RSA签名
	signature, err := rsa.SignPKCS1v15(rand.Reader, u.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// verifySignature 验证RSA签名
func (u *UnionPayAdapter) verifySignature(values url.Values, signature string) bool {
	if u.publicKey == nil {
		return false
	}

	// 移除签名字段
	params := make(map[string]string)
	for k, v := range values {
		if k != "signature" && len(v) > 0 {
			params[k] = v[0]
		}
	}

	// 构建签名字符串
	signString := u.buildSignString(params)

	// 解码签名
	signBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false
	}

	// SHA256哈希
	hash := sha256.Sum256([]byte(signString))

	// 验证签名
	err = rsa.VerifyPKCS1v15(u.publicKey, crypto.SHA256, hash[:], signBytes)
	return err == nil
}

// buildSignString 构建签名字符串
func (u *UnionPayAdapter) buildSignString(params map[string]string) string {
	// 排序参数
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "signature" && params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 构建签名字符串
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}

	return strings.Join(parts, "&")
}

// buildFormHTML 构建支付表单HTML
func (u *UnionPayAdapter) buildFormHTML(params map[string]string) string {
	var formFields []string
	for k, v := range params {
		formFields = append(formFields, fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`, k, v))
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>银联支付</title>
</head>
<body>
    <form id="unionpay_form" action="%s" method="post">
        %s
        <input type="submit" value="立即支付" style="display:none;">
    </form>
    <script>
        document.getElementById('unionpay_form').submit();
    </script>
</body>
</html>`, u.gatewayURL, strings.Join(formFields, "\n        "))

	return html
}

// sendRequest 发送HTTP请求
func (u *UnionPayAdapter) sendRequest(params map[string]string) ([]byte, error) {
	// 构建POST数据
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}

	// 发送请求
	resp, err := http.PostForm(u.gatewayURL, values)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// parseRSAPrivateKey 解析RSA私钥
func parseRSAPrivateKey(keyData string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(keyData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode private key")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// 尝试PKCS8格式
		if parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
			if rsaKey, ok := parsedKey.(*rsa.PrivateKey); ok {
				return rsaKey, nil
			}
		}
		return nil, err
	}

	return key, nil
}

// parseRSAPublicKey 解析RSA公钥
func parseRSAPublicKey(keyData string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(keyData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode public key")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaKey, nil
}