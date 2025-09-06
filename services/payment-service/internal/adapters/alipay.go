package adapters

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"github.com/smartwalle/alipay/v3"

	"github.com/gaokaohub/gaokao/services/payment-service/internal/models"
)

// AlipayAdapter 支付宝支付适配器
type AlipayAdapter struct {
	client    *alipay.Client
	appID     string
	notifyURL string
	returnURL string
}

// AlipayConfig 支付宝配置
type AlipayConfig struct {
	AppID      string `json:"app_id"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	NotifyURL  string `json:"notify_url"`
	ReturnURL  string `json:"return_url"`
	IsProd     bool   `json:"is_prod"`
}

// NewAlipayAdapter 创建支付宝适配器
func NewAlipayAdapter(config AlipayConfig) (PaymentAdapter, error) {
	// 创建支付宝客户端
	client, err := alipay.New(config.AppID, config.PrivateKey, !config.IsProd)
	if err != nil {
		return nil, fmt.Errorf("failed to create alipay client: %w", err)
	}

	// 加载支付宝公钥
	err = client.LoadAliPayPublicKey(config.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load alipay public key: %w", err)
	}

	return &AlipayAdapter{
		client:    client,
		appID:     config.AppID,
		notifyURL: config.NotifyURL,
		returnURL: config.ReturnURL,
	}, nil
}

// CreateOrder 创建支付订单
func (a *AlipayAdapter) CreateOrder(ctx context.Context, order *models.PaymentOrder) (*models.PaymentOrderResponse, error) {
	// 构建支付请求
	var request alipay.TradePagePay
	request.NotifyURL = a.notifyURL
	request.ReturnURL = a.returnURL
	request.Subject = order.Description
	request.OutTradeNo = order.OrderNo
	request.TotalAmount = fmt.Sprintf("%.2f", order.Amount)
	request.ProductCode = "FAST_INSTANT_TRADE_PAY"

	// 设置过期时间
	request.TimeExpire = time.Now().Add(30 * time.Minute).Format("2006-01-02 15:04:05")

	// 生成支付URL
	payURL, err := a.client.TradePagePay(request)
	if err != nil {
		return nil, fmt.Errorf("alipay create order failed: %w", err)
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
		PaymentURL: payURL.String(),
		ExpiredAt:  &expiredAt,
		CreatedAt:  order.CreatedAt,
	}

	return paymentResp, nil
}

// CreateQROrder 创建扫码支付订单
func (a *AlipayAdapter) CreateQROrder(ctx context.Context, order *models.PaymentOrder) (*models.PaymentOrderResponse, error) {
	// 构建扫码支付请求
	var request alipay.TradePreCreate
	request.NotifyURL = a.notifyURL
	request.Subject = order.Description
	request.OutTradeNo = order.OrderNo
	request.TotalAmount = order.Amount.String()

	// 设置过期时间
	request.TimeExpire = time.Now().Add(30 * time.Minute).Format("2006-01-02 15:04:05")

	// 调用支付宝API
	resp, err := a.client.TradePreCreate(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("alipay create qr order failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("alipay create qr order failed: %s - %s", resp.Code, resp.Msg)
	}

	// 构建支付响应
	expiredAt := time.Now().Add(30 * time.Minute)
	paymentResp := &models.PaymentOrderResponse{
		ID:        order.ID,
		OrderNo:   order.OrderNo,
		Amount:    order.Amount,
		Currency:  order.Currency,
		Subject:   order.Subject,
		Channel:   order.Channel,
		Status:    order.Status,
		QRCode:    resp.QRCode, // 二维码内容
		ExpiredAt: &expiredAt,
		CreatedAt: order.CreatedAt,
	}

	return paymentResp, nil
}

// QueryOrder 查询订单状态
func (a *AlipayAdapter) QueryOrder(ctx context.Context, orderID string) (*models.PaymentStatus, error) {
	var request alipay.TradeQuery
	request.OutTradeNo = orderID

	resp, err := a.client.TradeQuery(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("alipay query order failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("alipay query order failed: %s - %s", resp.Code, resp.Msg)
	}

	// 转换支付状态
	status := convertAlipayStatus(string(resp.TradeStatus))

	// 解析金额
	amount, _ := strconv.ParseFloat(resp.TotalAmount, 64)

	// 解析支付时间
	var paidAt *time.Time
	if resp.SendPayDate != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", resp.SendPayDate); err == nil {
			paidAt = &t
		}
	}

	paymentStatus := &models.PaymentStatus{
		OrderNo:       orderID,
		Status:        status,
		PaymentMethod: "alipay",
		TransactionID: resp.TradeNo,
		PaidAt:        paidAt,
		Amount:        decimal.NewFromFloat(amount),
		Currency:      "CNY",
		Extra: models.PaymentJSONB{
			"trade_status":   string(resp.TradeStatus),
			"trade_no":       resp.TradeNo,
			"buyer_logon_id": resp.BuyerLogonId,
			"buyer_user_id":  resp.BuyerUserId,
		},
	}

	return paymentStatus, nil
}

// RefundOrder 退款
func (a *AlipayAdapter) RefundOrder(ctx context.Context, refund *models.RefundRequest) (*models.RefundResponse, error) {
	var request alipay.TradeRefund
	request.OutTradeNo = refund.OrderNo
	request.RefundAmount = refund.Amount.String()
	request.RefundReason = refund.Reason
	request.OutRequestNo = refund.RefundID

	resp, err := a.client.TradeRefund(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("alipay refund failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("alipay refund failed: %s - %s", resp.Code, resp.Msg)
	}

	// 解析退款金额
	refundAmount, _ := strconv.ParseFloat(resp.RefundFee, 64)

	refundResp := &models.RefundResponse{
		RefundID:      refund.RefundID,
		OrderNo:       refund.OrderNo,
		Status:        "completed",
		Amount:        decimal.NewFromFloat(refundAmount),
		Currency:      "CNY",
		RefundedAt:    time.Now(),
		PaymentMethod: "alipay",
		Extra: models.PaymentJSONB{
			"trade_no":   resp.TradeNo,
			"refund_fee": resp.RefundFee,
			// 注意: GmtRefundPay字段在新版SDK中可能已移除或重命名
			// "gmt_refund_pay": resp.GmtRefundPay,
		},
	}

	return refundResp, nil
}

// HandleCallback 处理支付回调
func (a *AlipayAdapter) HandleCallback(ctx context.Context, request *http.Request) (*models.CallbackResult, error) {
	// 解析表单数据
	err := request.ParseForm()
	if err != nil {
		return nil, fmt.Errorf("failed to parse form data: %w", err)
	}

	// 验证签名 - 新版SDK返回error，nil表示验证成功
	if err := a.client.VerifySign(request.Form); err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	// 获取回调参数
	orderID := request.Form.Get("out_trade_no")
	tradeNo := request.Form.Get("trade_no")
	tradeStatus := request.Form.Get("trade_status")
	totalAmount := request.Form.Get("total_amount")
	gmtPayment := request.Form.Get("gmt_payment")

	// 解析金额
	amount, _ := strconv.ParseFloat(totalAmount, 64)

	// 解析支付时间
	var paidAt *time.Time
	if gmtPayment != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", gmtPayment); err == nil {
			paidAt = &t
		}
	}

	// 构建回调结果
	result := &models.CallbackResult{
		OrderNo:       orderID,
		Status:        convertAlipayStatus(tradeStatus),
		PaymentMethod: "alipay",
		TransactionID: tradeNo,
		Amount:        decimal.NewFromFloat(amount),
		Currency:      "CNY",
		PaidAt:        paidAt,
		Extra: models.PaymentJSONB{
			"trade_status":   tradeStatus,
			"trade_no":       tradeNo,
			"buyer_id":       request.Form.Get("buyer_id"),
			"buyer_logon_id": request.Form.Get("buyer_logon_id"),
			"seller_id":      request.Form.Get("seller_id"),
			"app_id":         request.Form.Get("app_id"),
		},
	}

	return result, nil
}

// CloseOrder 关闭订单
func (a *AlipayAdapter) CloseOrder(ctx context.Context, orderNo string) error {
	var request alipay.TradeClose
	request.OutTradeNo = orderNo

	resp, err := a.client.TradeClose(ctx, request)
	if err != nil {
		return fmt.Errorf("alipay close order failed: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("alipay close order failed: %s - %s", resp.Code, resp.Msg)
	}

	return nil
}

// GetName 获取支付渠道名称
func (a *AlipayAdapter) GetName() string {
	return "alipay"
}

// CreatePayment 创建支付
func (a *AlipayAdapter) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// 构建支付请求
	var request alipay.TradePagePay
	request.NotifyURL = a.notifyURL
	request.ReturnURL = a.returnURL
	request.Subject = req.Subject
	request.OutTradeNo = req.OrderNo
	request.TotalAmount = req.Amount.String()
	request.ProductCode = "FAST_INSTANT_TRADE_PAY"

	// 设置过期时间
	if req.ExpireTime > 0 {
		request.TimeExpire = time.Now().Add(req.ExpireTime).Format("2006-01-02 15:04:05")
	} else {
		request.TimeExpire = time.Now().Add(30 * time.Minute).Format("2006-01-02 15:04:05")
	}

	// 生成支付URL
	payURL, err := a.client.TradePagePay(request)
	if err != nil {
		return nil, fmt.Errorf("alipay create payment failed: %w", err)
	}

	// 构建支付响应
	expiredAt := time.Now().Add(30 * time.Minute)
	if req.ExpireTime > 0 {
		expiredAt = time.Now().Add(req.ExpireTime)
	}

	paymentResp := &PaymentResponse{
		OrderNo:   req.OrderNo,
		PayURL:    payURL.String(),
		ExpiredAt: expiredAt,
	}

	return paymentResp, nil
}

// VerifyCallback 验证回调签名
func (a *AlipayAdapter) VerifyCallback(ctx context.Context, data []byte, signature string) (*PaymentCallback, error) {
	// 解析表单数据
	form, err := url.ParseQuery(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse callback data: %w", err)
	}

	// 验证签名 - 新版SDK返回error，nil表示验证成功
	if err := a.client.VerifySign(form); err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	// 获取回调参数
	orderNo := form.Get("out_trade_no")
	channelTradeNo := form.Get("trade_no")
	tradeStatus := form.Get("trade_status")
	totalAmount := form.Get("total_amount")
	gmtPayment := form.Get("gmt_payment")

	// 解析金额
	amount, _ := strconv.ParseFloat(totalAmount, 64)
	actualAmount := decimal.NewFromFloat(amount)

	// 解析支付时间
	var paidAt time.Time
	if gmtPayment != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", gmtPayment); err == nil {
			paidAt = t
		}
	}

	// 转换状态
	status := convertAlipayStatus(tradeStatus)

	// 构建回调结果
	callback := &PaymentCallback{
		OrderNo:        orderNo,
		ChannelTradeNo: channelTradeNo,
		Amount:         actualAmount,
		ActualAmount:   actualAmount,
		Status:         status,
		PaidAt:         paidAt,
		RawData:        string(data),
		Signature:      signature,
	}

	return callback, nil
}

// QueryPayment 查询支付状态
func (a *AlipayAdapter) QueryPayment(ctx context.Context, req *QueryRequest) (*QueryResponse, error) {
	var request alipay.TradeQuery
	request.OutTradeNo = req.OrderNo

	resp, err := a.client.TradeQuery(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("alipay query payment failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("alipay query payment failed: %s - %s", resp.Code, resp.Msg)
	}

	// 转换支付状态
	status := convertAlipayStatus(string(resp.TradeStatus))

	// 解析金额
	amount, _ := strconv.ParseFloat(resp.TotalAmount, 64)

	// 解析支付时间
	var paidAt *time.Time
	if resp.SendPayDate != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", resp.SendPayDate); err == nil {
			paidAt = &t
		}
	}

	queryResp := &QueryResponse{
		OrderNo:        req.OrderNo,
		ChannelTradeNo: resp.TradeNo,
		Amount:         decimal.NewFromFloat(amount),
		Status:         status,
		PaidAt:         paidAt,
		RefundedAmount: decimal.Zero,
	}

	return queryResp, nil
}

// CreateRefund 创建退款
func (a *AlipayAdapter) CreateRefund(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
	var request alipay.TradeRefund
	request.OutTradeNo = req.OrderNo
	request.RefundAmount = req.RefundAmount.String()
	request.RefundReason = req.Reason
	request.OutRequestNo = req.RefundNo

	resp, err := a.client.TradeRefund(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("alipay create refund failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("alipay create refund failed: %s - %s", resp.Code, resp.Msg)
	}

	// 解析退款金额
	refundAmount, _ := strconv.ParseFloat(resp.RefundFee, 64)

	refundResp := &RefundResponse{
		RefundNo:        req.RefundNo,
		ChannelRefundNo: resp.TradeNo,
		Amount:          decimal.NewFromFloat(refundAmount),
		Status:          "success",
		RefundedAt:      time.Now(),
	}

	return refundResp, nil
}

// QueryRefund 查询退款状态
func (a *AlipayAdapter) QueryRefund(ctx context.Context, refundNo string) (*RefundResponse, error) {
	// 支付宝没有专门的退款查询接口，我们返回一个默认的响应
	refundResp := &RefundResponse{
		RefundNo: refundNo,
		Status:   "unknown",
	}

	return refundResp, nil
}

// 辅助函数



// convertAlipayStatus 转换支付宝支付状态
func convertAlipayStatus(alipayStatus string) string {
	switch alipayStatus {
	case "TRADE_SUCCESS":
		return "completed"
	case "TRADE_FINISHED":
		return "completed"
	case "WAIT_BUYER_PAY":
		return "pending"
	case "TRADE_CLOSED":
		return "cancelled"
	default:
		return "unknown"
	}
}

// getAlipayEnv 获取环境变量
func getAlipayEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
