package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// StubPaymentAdapter 存根支付适配器（用于快速构建）
type StubPaymentAdapter struct {
	name string
}

// NewStubPaymentAdapter 创建存根支付适配器
func NewStubPaymentAdapter(name string) *StubPaymentAdapter {
	return &StubPaymentAdapter{
		name: name,
	}
}

// GetName 获取支付渠道名称
func (s *StubPaymentAdapter) GetName() string {
	return s.name
}

// CreatePayment 创建支付
func (s *StubPaymentAdapter) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	return &PaymentResponse{
		OrderNo:        req.OutTradeNo,
		ChannelTradeNo: "stub_" + req.OutTradeNo,
	}, nil
}

// VerifyCallback 验证回调签名
func (s *StubPaymentAdapter) VerifyCallback(ctx context.Context, data []byte, signature string) (*PaymentCallback, error) {
	return &PaymentCallback{
		OrderNo:        "stub_order",
		ChannelTradeNo: "stub_channel_trade",
		Amount:         decimal.NewFromFloat(100.00),
		ActualAmount:   decimal.NewFromFloat(100.00),
		Status:         "success",
		PaidAt:         time.Now(),
		RawData:        string(data),
		Signature:      signature,
	}, nil
}

// QueryPayment 查询支付状态
func (s *StubPaymentAdapter) QueryPayment(ctx context.Context, req *QueryRequest) (*QueryResponse, error) {
	return &QueryResponse{
		OrderNo:        req.OrderNo,
		ChannelTradeNo: "stub_" + req.OrderNo,
		Amount:         decimal.NewFromFloat(100.00),
		Status:         "success",
		PaidAt:         &[]time.Time{time.Now()}[0],
		RefundedAmount: decimal.Zero,
	}, nil
}

// CreateRefund 创建退款
func (s *StubPaymentAdapter) CreateRefund(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
	return &RefundResponse{
		RefundNo:   req.RefundNo,
		Amount:     req.RefundAmount,
		Status:     "success",
		RefundedAt: time.Now(),
	}, nil
}

// QueryRefund 查询退款状态
func (s *StubPaymentAdapter) QueryRefund(ctx context.Context, refundNo string) (*RefundResponse, error) {
	return &RefundResponse{
		RefundNo:   refundNo,
		Amount:     decimal.NewFromFloat(50.00),
		Status:     "success",
		RefundedAt: time.Now(),
	}, nil
}

// CloseOrder 关闭订单
func (s *StubPaymentAdapter) CloseOrder(ctx context.Context, orderNo string) error {
	// 存根实现，总是成功
	return nil
}

// StubWeChatAdapter 微信支付存根适配器
func NewStubWeChatAdapter(config AdapterConfig) PaymentAdapter {
	return NewStubPaymentAdapter("wechat")
}

// StubAlipayAdapter 支付宝存根适配器
func NewStubAlipayAdapter(config AdapterConfig) PaymentAdapter {
	return NewStubPaymentAdapter("alipay")
}

// StubQQAdapter QQ支付存根适配器
func NewStubQQAdapter(config AdapterConfig) PaymentAdapter {
	return NewStubPaymentAdapter("qq")
}

// StubUnionPayAdapter 银联支付存根适配器
func NewStubUnionPayAdapter(config AdapterConfig) PaymentAdapter {
	return NewStubPaymentAdapter("unionpay")
}

// StubPaymentCallback 存根支付回调结构
type StubPaymentCallback struct {
	OrderNo        string          `json:"order_no"`
	ChannelTradeNo string          `json:"channel_trade_no"`
	Amount         decimal.Decimal `json:"amount"`
	Status         string          `json:"status"`
	PaidAt         time.Time       `json:"paid_at"`
	RawData        string          `json:"raw_data"`
	Signature      string          `json:"signature"`
}

// StubPaymentAdapterFactory 存根支付适配器工厂
type StubPaymentAdapterFactory struct{}

// NewStubPaymentAdapterFactory 创建存根支付适配器工厂
func NewStubPaymentAdapterFactory() *StubPaymentAdapterFactory {
	return &StubPaymentAdapterFactory{}
}

// GetAdapter 获取支付适配器
func (f *StubPaymentAdapterFactory) GetAdapter(channel string) (PaymentAdapter, error) {
	config := AdapterConfig{
		AppID:     "stub_app_id",
		MchID:     "stub_mch_id",
		APIKey:    "stub_api_key",
		NotifyURL: "http://localhost:8080/notify",
		ReturnURL: "http://localhost:3000/return",
		IsProd:    false,
		Debug:     true,
	}

	switch channel {
	case "wechat":
		return NewStubWeChatAdapter(config), nil
	case "alipay":
		return NewStubAlipayAdapter(config), nil
	case "qq":
		return NewStubQQAdapter(config), nil
	case "unionpay":
		return NewStubUnionPayAdapter(config), nil
	default:
		return nil, fmt.Errorf("不支持的支付渠道: %s", channel)
	}
}

// CreateStubPaymentRequest 创建存根支付请求
func CreateStubPaymentRequest() *PaymentRequest {
	return &PaymentRequest{
		OrderNo:       "test_order_" + fmt.Sprintf("%d", time.Now().Unix()),
		OutTradeNo:    "out_trade_" + fmt.Sprintf("%d", time.Now().Unix()),
		Amount:        decimal.NewFromFloat(100.00),
		Subject:       "高考志愿填报会员",
		Description:   "高考志愿填报系统会员服务",
		NotifyURL:     "http://localhost:8080/notify",
		ReturnURL:     "http://localhost:3000/return",
		UserID:        "test_user_123",
		ClientIP:      "127.0.0.1",
		PaymentMethod: "wechat_jsapi",
		OpenID:        "test_openid_123",
		ExpireTime:    30 * time.Minute,
		Metadata: map[string]interface{}{
			"product_type": "membership",
			"plan_id":      "premium",
		},
	}
}

// CreateStubRefundRequest 创建存根退款请求
func CreateStubRefundRequest() *RefundRequest {
	return &RefundRequest{
		OrderNo:        "test_order_123",
		RefundNo:       "refund_" + fmt.Sprintf("%d", time.Now().Unix()),
		ChannelTradeNo: "channel_trade_123",
		Amount:         decimal.NewFromFloat(100.00),
		RefundAmount:   decimal.NewFromFloat(50.00),
		TotalAmount:    decimal.NewFromFloat(100.00),
		Reason:         "用户申请退款",
		NotifyURL:      "http://localhost:8080/refund_notify",
	}
}

// CreateStubQueryRequest 创建存根查询请求
func CreateStubQueryRequest() *QueryRequest {
	return &QueryRequest{
		OrderNo:        "test_order_123",
		ChannelTradeNo: "channel_trade_123",
	}
}

// TestStubAdapter 测试存根适配器
func TestStubAdapter() error {
	factory := NewStubPaymentAdapterFactory()

	// 测试微信支付
	wechatAdapter, err := factory.GetAdapter("wechat")
	if err != nil {
		return fmt.Errorf("获取微信支付适配器失败: %w", err)
	}

	// 测试创建支付
	paymentReq := CreateStubPaymentRequest()
	paymentResp, err := wechatAdapter.CreatePayment(context.Background(), paymentReq)
	if err != nil {
		return fmt.Errorf("创建支付失败: %w", err)
	}

	fmt.Printf("支付创建成功: %+v\n", paymentResp)

	// 测试查询支付
	queryReq := CreateStubQueryRequest()
	queryResp, err := wechatAdapter.QueryPayment(context.Background(), queryReq)
	if err != nil {
		return fmt.Errorf("查询支付失败: %w", err)
	}

	fmt.Printf("支付查询成功: %+v\n", queryResp)

	// 测试退款
	refundReq := CreateStubRefundRequest()
	refundResp, err := wechatAdapter.CreateRefund(context.Background(), refundReq)
	if err != nil {
		return fmt.Errorf("创建退款失败: %w", err)
	}

	fmt.Printf("退款创建成功: %+v\n", refundResp)

	return nil
}
