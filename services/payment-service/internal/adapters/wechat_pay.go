package adapters

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"

	"github.com/gaokaohub/payment-service/internal/models"
)

// WechatPayAdapter 微信支付适配器
type WechatPayAdapter struct {
	client    *core.Client
	appID     string
	mchID     string
	notifyURL string
}

// WechatPayConfig 微信支付配置
type WechatPayConfig struct {
	AppID        string `json:"app_id"`
	MchID        string `json:"mch_id"`
	APIKey       string `json:"api_key"`
	CertPath     string `json:"cert_path"`
	KeyPath      string `json:"key_path"`
	NotifyURL    string `json:"notify_url"`
	SerialNumber string `json:"serial_number"`
}

// NewWechatPayAdapter 创建微信支付适配器
func NewWechatPayAdapter() (*WechatPayAdapter, error) {
	config, err := loadWechatPayConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load wechat pay config: %w", err)
	}

	// 加载商户私钥
	privateKey, err := loadPrivateKey(config.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	// 创建微信支付客户端
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(config.MchID, config.SerialNumber, privateKey, config.APIKey),
	}

	client, err := core.NewClient(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create wechat pay client: %w", err)
	}

	return &WechatPayAdapter{
		client:    client,
		appID:     config.AppID,
		mchID:     config.MchID,
		notifyURL: config.NotifyURL,
	}, nil
}

// CreateOrder 创建支付订单
func (w *WechatPayAdapter) CreateOrder(ctx context.Context, order *models.PaymentOrder) (*models.PaymentOrderResponse, error) {
	// 构建支付请求
	request := &native.PrepayRequest{
		Appid:       core.String(w.appID),
		Mchid:       core.String(w.mchID),
		Description: core.String(order.Description),
		OutTradeNo:  core.String(order.OrderNo),
		NotifyUrl:   core.String(w.notifyURL),
		Amount: &native.Amount{
			Total:    core.Int64(order.Amount.Mul(decimal.NewFromInt(100)).IntPart()), // 转换为分
			Currency: core.String("CNY"),
		},
		TimeExpire: core.Time(time.Now().Add(30 * time.Minute)), // 30分钟过期
	}

	// 调用微信支付API
	svc := native.NativeApiService{Client: w.client}
	resp, result, err := svc.Prepay(ctx, *request)
	if err != nil {
		return nil, fmt.Errorf("wechat pay prepay failed: %w", err)
	}

	// 检查响应状态
	if result.Response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wechat pay prepay failed with status: %d", result.Response.StatusCode)
	}

	// 构建支付响应
	expiredAt := time.Now().Add(30 * time.Minute)
	paymentResp := &models.PaymentOrderResponse{
		ID:         order.ID,
		OrderNo:    order.OrderNo,
		Amount:     order.Amount,
		Currency:   order.Currency,
		Subject:    order.Subject,
		Channel:    order.Channel,
		Status:     order.Status,
		PaymentURL: *resp.CodeUrl, // 二维码链接
		ExpiredAt:  &expiredAt,
		CreatedAt:  order.CreatedAt,
	}

	return paymentResp, nil
}

// QueryOrder 查询订单状态
func (w *WechatPayAdapter) QueryOrder(ctx context.Context, orderID string) (*models.PaymentStatus, error) {
	svc := native.NativeApiService{Client: w.client}

	request := native.QueryOrderByOutTradeNoRequest{
		OutTradeNo: core.String(orderID),
		Mchid:      core.String(w.mchID),
	}

	resp, result, err := svc.QueryOrderByOutTradeNo(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("wechat pay query order failed: %w", err)
	}

	if result.Response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wechat pay query order failed with status: %d", result.Response.StatusCode)
	}

	// 转换支付状态
	status := convertWechatPayStatus(*resp.TradeState)

	paymentStatus := &models.PaymentStatus{
		OrderNo:       orderID,
		Status:        status,
		PaymentMethod: "wechat_pay",
		TransactionID: getStringValue(resp.TransactionId),
		PaidAt:        parseWechatPayTime(resp.SuccessTime),
		Amount:        decimal.NewFromFloat(float64(*resp.Amount.Total) / 100), // 转换为元并使用decimal
		Currency:      *resp.Amount.Currency,
		Extra: map[string]interface{}{
			"trade_state":      resp.TradeState,
			"trade_state_desc": resp.TradeStateDesc,
			"transaction_id":   resp.TransactionId,
		},
	}

	return paymentStatus, nil
}

// RefundOrder 退款
func (w *WechatPayAdapter) RefundOrder(ctx context.Context, refund *models.RefundRequest) (*models.RefundResponse, error) {
	// 微信支付退款API实现
	// 这里需要根据微信支付API文档实现退款逻辑
	return nil, fmt.Errorf("wechat pay refund not implemented yet")
}

// HandleCallback 处理支付回调
func (w *WechatPayAdapter) HandleCallback(ctx context.Context, request *http.Request) (*models.CallbackResult, error) {
	// 读取请求体
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read callback body: %w", err)
	}

	// 验证签名
	if err := w.verifySignature(request, body); err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	// 解析回调数据
	var callbackData WechatPayCallback
	if err := json.Unmarshal(body, &callbackData); err != nil {
		return nil, fmt.Errorf("failed to parse callback data: %w", err)
	}

	// 解密资源数据
	resource, err := w.decryptResource(callbackData.Resource)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt resource: %w", err)
	}

	// 构建回调结果
	result := &models.CallbackResult{
		OrderNo:       resource.OutTradeNo,
		Status:        convertWechatPayStatus(resource.TradeState),
		PaymentMethod: "wechat_pay",
		TransactionID: resource.TransactionId,
		Amount:        decimal.NewFromFloat(float64(resource.Amount.Total) / 100),
		Currency:      resource.Amount.Currency,
		PaidAt:        parseWechatPayTime(&resource.SuccessTime),
		Extra: map[string]interface{}{
			"event_type":     callbackData.EventType,
			"trade_state":    resource.TradeState,
			"transaction_id": resource.TransactionId,
		},
	}

	return result, nil
}

// 辅助函数

// loadWechatPayConfig 加载微信支付配置
func loadWechatPayConfig() (*WechatPayConfig, error) {
	return &WechatPayConfig{
		AppID:        getEnv("WECHAT_APP_ID", ""),
		MchID:        getEnv("WECHAT_MCH_ID", ""),
		APIKey:       getEnv("WECHAT_API_KEY", ""),
		CertPath:     getEnv("WECHAT_CERT_PATH", ""),
		KeyPath:      getEnv("WECHAT_KEY_PATH", ""),
		NotifyURL:    getEnv("WECHAT_NOTIFY_URL", ""),
		SerialNumber: getEnv("WECHAT_SERIAL_NUMBER", ""),
	}, nil
}

// loadPrivateKey 加载私钥
func loadPrivateKey(keyPath string) (*rsa.PrivateKey, error) {
	keyData, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not RSA key")
	}

	return rsaKey, nil
}

// convertWechatPayStatus 转换微信支付状态
func convertWechatPayStatus(wechatStatus string) string {
	switch wechatStatus {
	case "SUCCESS":
		return "completed"
	case "REFUND":
		return "refunded"
	case "NOTPAY":
		return "pending"
	case "CLOSED":
		return "cancelled"
	case "REVOKED":
		return "cancelled"
	case "USERPAYING":
		return "processing"
	case "PAYERROR":
		return "failed"
	default:
		return "unknown"
	}
}

// parseWechatPayTime 解析微信支付时间
func parseWechatPayTime(timeStr *string) *time.Time {
	if timeStr == nil || *timeStr == "" {
		return nil
	}

	t, err := time.Parse(time.RFC3339, *timeStr)
	if err != nil {
		return nil
	}

	return &t
}

// getStringValue 安全获取字符串值
func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// verifySignature 验证签名
func (w *WechatPayAdapter) verifySignature(request *http.Request, body []byte) error {
	// 微信支付签名验证逻辑
	// 这里需要根据微信支付API文档实现签名验证
	return nil
}

// decryptResource 解密资源数据
func (w *WechatPayAdapter) decryptResource(resource WechatPayResource) (*WechatPayResourceData, error) {
	// 微信支付资源解密逻辑
	// 这里需要根据微信支付API文档实现资源解密
	return nil, fmt.Errorf("resource decryption not implemented yet")
}

// getEnv 获取环境变量
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 微信支付回调数据结构
type WechatPayCallback struct {
	ID           string            `json:"id"`
	CreateTime   string            `json:"create_time"`
	ResourceType string            `json:"resource_type"`
	EventType    string            `json:"event_type"`
	Summary      string            `json:"summary"`
	Resource     WechatPayResource `json:"resource"`
}

type WechatPayResource struct {
	OriginalType   string `json:"original_type"`
	Algorithm      string `json:"algorithm"`
	Ciphertext     string `json:"ciphertext"`
	AssociatedData string `json:"associated_data"`
	Nonce          string `json:"nonce"`
}

type WechatPayResourceData struct {
	Appid          string          `json:"appid"`
	Mchid          string          `json:"mchid"`
	OutTradeNo     string          `json:"out_trade_no"`
	TransactionId  string          `json:"transaction_id"`
	TradeType      string          `json:"trade_type"`
	TradeState     string          `json:"trade_state"`
	TradeStateDesc string          `json:"trade_state_desc"`
	BankType       string          `json:"bank_type"`
	Attach         string          `json:"attach"`
	SuccessTime    string          `json:"success_time"`
	Payer          WechatPayPayer  `json:"payer"`
	Amount         WechatPayAmount `json:"amount"`
}

type WechatPayPayer struct {
	Openid string `json:"openid"`
}

type WechatPayAmount struct {
	Total         int    `json:"total"`
	PayerTotal    int    `json:"payer_total"`
	Currency      string `json:"currency"`
	PayerCurrency string `json:"payer_currency"`
}
