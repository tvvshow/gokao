package adapters

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// AlipayAdapter 支付宝适配器
type AlipayAdapter struct {
	config     AdapterConfig
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	gatewayURL string
}

// NewAlipayAdapter 创建支付宝适配器
func NewAlipayAdapter(config AdapterConfig) PaymentAdapter {
	adapter := &AlipayAdapter{
		config: config,
	}

	// 解析私钥
	if config.PrivateKey != "" {
		if key, err := parsePrivateKey(config.PrivateKey); err == nil {
			adapter.privateKey = key
		}
	}

	// 解析公钥
	if config.PublicKey != "" {
		if key, err := parsePublicKey(config.PublicKey); err == nil {
			adapter.publicKey = key
		}
	}

	// 设置网关URL
	if config.Sandbox {
		adapter.gatewayURL = "https://openapi.alipaydev.com/gateway.do"
	} else {
		adapter.gatewayURL = "https://openapi.alipay.com/gateway.do"
	}

	return adapter
}

// GetName 获取支付渠道名称
func (a *AlipayAdapter) GetName() string {
	return "alipay"
}

// CreatePayment 创建支付
func (a *AlipayAdapter) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// 构建业务参数
	bizContent := map[string]interface{}{
		"out_trade_no":    req.OrderNo,
		"total_amount":    req.Amount.String(),
		"subject":         req.Subject,
		"body":            req.Description,
		"timeout_express": fmt.Sprintf("%dm", int(req.ExpireTime.Minutes())),
		"product_code":    "FAST_INSTANT_TRADE_PAY",
	}

	bizContentJSON, _ := json.Marshal(bizContent)

	// 构建请求参数
	params := map[string]string{
		"app_id":      a.config.AppID,
		"method":      "alipay.trade.page.pay",
		"charset":     "utf-8",
		"sign_type":   a.config.SignType,
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  req.NotifyURL,
		"return_url":  req.ReturnURL,
		"biz_content": string(bizContentJSON),
	}

	// 生成签名
	sign, err := a.generateSign(params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate sign: %w", err)
	}
	params["sign"] = sign

	// 构建支付URL
	payURL := a.buildPayURL(params)

	return &PaymentResponse{
		OrderNo:   req.OrderNo,
		PayURL:    payURL,
		ExpiredAt: time.Now().Add(req.ExpireTime),
	}, nil
}

// VerifyCallback 验证回调签名
func (a *AlipayAdapter) VerifyCallback(ctx context.Context, data []byte, signature string) (*PaymentCallback, error) {
	// 解析回调数据
	values, err := url.ParseQuery(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse callback data: %w", err)
	}

	// 提取签名
	sign := values.Get("sign")
	if sign == "" {
		return nil, fmt.Errorf("missing signature in callback")
	}

	// 移除签名字段
	values.Del("sign")
	values.Del("sign_type")

	// 验证签名
	if !a.verifySign(values, sign) {
		return nil, fmt.Errorf("invalid signature")
	}

	// 构建回调响应
	callback := &PaymentCallback{
		OrderNo:        values.Get("out_trade_no"),
		ChannelTradeNo: values.Get("trade_no"),
		Status:         "success",
		RawData:        string(data),
		Signature:      sign,
	}

	// 解析金额
	if totalAmount := values.Get("total_amount"); totalAmount != "" {
		if amount, err := decimal.NewFromString(totalAmount); err == nil {
			callback.Amount = amount
			callback.ActualAmount = amount
		}
	}

	// 解析支付时间
	if gmtPayment := values.Get("gmt_payment"); gmtPayment != "" {
		if paidAt, err := time.Parse("2006-01-02 15:04:05", gmtPayment); err == nil {
			callback.PaidAt = paidAt
		}
	}

	// 检查交易状态
	tradeStatus := values.Get("trade_status")
	if tradeStatus != "TRADE_SUCCESS" && tradeStatus != "TRADE_FINISHED" {
		callback.Status = "failed"
	}

	return callback, nil
}

// QueryPayment 查询支付状态
func (a *AlipayAdapter) QueryPayment(ctx context.Context, req *QueryRequest) (*QueryResponse, error) {
	// 构建业务参数
	bizContent := map[string]interface{}{
		"out_trade_no": req.OrderNo,
	}

	if req.ChannelTradeNo != "" {
		bizContent["trade_no"] = req.ChannelTradeNo
	}

	bizContentJSON, _ := json.Marshal(bizContent)

	// 构建请求参数
	params := map[string]string{
		"app_id":      a.config.AppID,
		"method":      "alipay.trade.query",
		"charset":     "utf-8",
		"sign_type":   a.config.SignType,
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": string(bizContentJSON),
	}

	// 生成签名
	sign, err := a.generateSign(params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate sign: %w", err)
	}
	params["sign"] = sign

	// 发送请求
	respData, err := a.sendRequest(params)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// 解析响应
	var response struct {
		AlipayTradeQueryResponse struct {
			Code         string `json:"code"`
			Msg          string `json:"msg"`
			OutTradeNo   string `json:"out_trade_no"`
			TradeNo      string `json:"trade_no"`
			TotalAmount  string `json:"total_amount"`
			TradeStatus  string `json:"trade_status"`
			SendPayDate  string `json:"send_pay_date"`
		} `json:"alipay_trade_query_response"`
	}

	if err := json.Unmarshal(respData, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	resp := &QueryResponse{
		OrderNo:        response.AlipayTradeQueryResponse.OutTradeNo,
		ChannelTradeNo: response.AlipayTradeQueryResponse.TradeNo,
	}

	// 解析金额
	if response.AlipayTradeQueryResponse.TotalAmount != "" {
		if amount, err := decimal.NewFromString(response.AlipayTradeQueryResponse.TotalAmount); err == nil {
			resp.Amount = amount
		}
	}

	// 解析状态
	switch response.AlipayTradeQueryResponse.TradeStatus {
	case "TRADE_SUCCESS", "TRADE_FINISHED":
		resp.Status = "paid"
		if response.AlipayTradeQueryResponse.SendPayDate != "" {
			if paidAt, err := time.Parse("2006-01-02 15:04:05", response.AlipayTradeQueryResponse.SendPayDate); err == nil {
				resp.PaidAt = &paidAt
			}
		}
	case "WAIT_BUYER_PAY":
		resp.Status = "pending"
	case "TRADE_CLOSED":
		resp.Status = "canceled"
	default:
		resp.Status = "failed"
	}

	return resp, nil
}

// CreateRefund 创建退款
func (a *AlipayAdapter) CreateRefund(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
	// 构建业务参数
	bizContent := map[string]interface{}{
		"out_trade_no":   req.OrderNo,
		"refund_amount":  req.Amount.String(),
		"refund_reason":  req.Reason,
		"out_request_no": req.RefundNo,
	}

	if req.ChannelTradeNo != "" {
		bizContent["trade_no"] = req.ChannelTradeNo
	}

	bizContentJSON, _ := json.Marshal(bizContent)

	// 构建请求参数
	params := map[string]string{
		"app_id":      a.config.AppID,
		"method":      "alipay.trade.refund",
		"charset":     "utf-8",
		"sign_type":   a.config.SignType,
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": string(bizContentJSON),
	}

	// 生成签名
	sign, err := a.generateSign(params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate sign: %w", err)
	}
	params["sign"] = sign

	// 发送请求
	respData, err := a.sendRequest(params)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// 解析响应
	var response struct {
		AlipayTradeRefundResponse struct {
			Code           string `json:"code"`
			Msg            string `json:"msg"`
			OutTradeNo     string `json:"out_trade_no"`
			TradeNo        string `json:"trade_no"`
			RefundFee      string `json:"refund_fee"`
			OutRequestNo   string `json:"out_request_no"`
			GmtRefundPay   string `json:"gmt_refund_pay"`
		} `json:"alipay_trade_refund_response"`
	}

	if err := json.Unmarshal(respData, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	resp := &RefundResponse{
		RefundNo: response.AlipayTradeRefundResponse.OutRequestNo,
		Status:   "success",
	}

	// 解析退款金额
	if response.AlipayTradeRefundResponse.RefundFee != "" {
		if amount, err := decimal.NewFromString(response.AlipayTradeRefundResponse.RefundFee); err == nil {
			resp.Amount = amount
		}
	}

	// 解析退款时间
	if response.AlipayTradeRefundResponse.GmtRefundPay != "" {
		if refundedAt, err := time.Parse("2006-01-02 15:04:05", response.AlipayTradeRefundResponse.GmtRefundPay); err == nil {
			resp.RefundedAt = refundedAt
		}
	}

	// 检查结果码
	if response.AlipayTradeRefundResponse.Code != "10000" {
		resp.Status = "failed"
	}

	return resp, nil
}

// QueryRefund 查询退款状态
func (a *AlipayAdapter) QueryRefund(ctx context.Context, refundNo string) (*RefundResponse, error) {
	// 支付宝没有单独的退款查询接口，可以通过交易查询来获取退款信息
	return nil, fmt.Errorf("refund query not supported by alipay")
}

// CloseOrder 关闭订单
func (a *AlipayAdapter) CloseOrder(ctx context.Context, orderNo string) error {
	// 构建业务参数
	bizContent := map[string]interface{}{
		"out_trade_no": orderNo,
	}

	bizContentJSON, _ := json.Marshal(bizContent)

	// 构建请求参数
	params := map[string]string{
		"app_id":      a.config.AppID,
		"method":      "alipay.trade.close",
		"charset":     "utf-8",
		"sign_type":   a.config.SignType,
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": string(bizContentJSON),
	}

	// 生成签名
	sign, err := a.generateSign(params)
	if err != nil {
		return fmt.Errorf("failed to generate sign: %w", err)
	}
	params["sign"] = sign

	// 发送请求
	_, err = a.sendRequest(params)
	return err
}

// 私有方法

// generateSign 生成签名
func (a *AlipayAdapter) generateSign(params map[string]string) (string, error) {
	if a.privateKey == nil {
		return "", fmt.Errorf("private key not configured")
	}

	// 排序参数
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建签名字符串
	var parts []string
	for _, k := range keys {
		if params[k] != "" {
			parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
		}
	}
	signString := strings.Join(parts, "&")

	// 生成签名
	hash := sha256.Sum256([]byte(signString))
	signature, err := rsa.SignPKCS1v15(rand.Reader, a.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// verifySign 验证签名
func (a *AlipayAdapter) verifySign(values url.Values, sign string) bool {
	if a.publicKey == nil {
		return false
	}

	// 构建签名字符串
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		if values.Get(k) != "" {
			parts = append(parts, fmt.Sprintf("%s=%s", k, values.Get(k)))
		}
	}
	signString := strings.Join(parts, "&")

	// 验证签名
	signature, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false
	}

	hash := sha256.Sum256([]byte(signString))
	err = rsa.VerifyPKCS1v15(a.publicKey, crypto.SHA256, hash[:], signature)
	return err == nil
}

// buildPayURL 构建支付URL
func (a *AlipayAdapter) buildPayURL(params map[string]string) string {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	return a.gatewayURL + "?" + values.Encode()
}

// sendRequest 发送请求
func (a *AlipayAdapter) sendRequest(params map[string]string) ([]byte, error) {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}

	resp, err := http.PostForm(a.gatewayURL, values)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// parsePrivateKey 解析私钥
func parsePrivateKey(keyData string) (*rsa.PrivateKey, error) {
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

// parsePublicKey 解析公钥
func parsePublicKey(keyData string) (*rsa.PublicKey, error) {
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