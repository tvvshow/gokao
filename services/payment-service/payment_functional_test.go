//go:build legacy
// +build legacy

package main

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/tvvshow/gokao/services/payment-service/internal/adapters"
	"github.com/tvvshow/gokao/services/payment-service/internal/config"
	"github.com/tvvshow/gokao/services/payment-service/internal/models"
	"github.com/tvvshow/gokao/services/payment-service/internal/services"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDatabase 模拟数据库
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) ExecContext(ctx context.Context, query string, args ...interface{}) error {
	returnArgs := m.Called(ctx, query, args)
	return returnArgs.Error(0)
}

func (m *MockDatabase) QueryRowContext(ctx context.Context, query string, args ...interface{}) interface{} {
	returnArgs := m.Called(ctx, query, args)
	return returnArgs.Get(0)
}

func (m *MockDatabase) Close() error {
	return nil
}

// MockRedis 模拟Redis
type MockRedis struct {
	mock.Mock
	data map[string]string
}

func (m *MockRedis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if m.data == nil {
		m.data = make(map[string]string)
	}
	m.data[key] = value.(string)
	return nil
}

func (m *MockRedis) Get(ctx context.Context, key string) (string, error) {
	if value, exists := m.data[key]; exists {
		return value, nil
	}
	return "", fmt.Errorf("key not found")
}

func (m *MockRedis) Del(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		delete(m.data, key)
	}
	return nil
}

func (m *MockRedis) Close() error {
	return nil
}

// MockPaymentAdapter 模拟支付适配器
type MockPaymentAdapter struct {
	mock.Mock
}

func (m *MockPaymentAdapter) CreatePayment(ctx context.Context, req *adapters.PaymentRequest) (*adapters.PaymentResponse, error) {
	returnArgs := m.Called(ctx, req)
	return returnArgs.Get(0).(*adapters.PaymentResponse), returnArgs.Error(1)
}

func (m *MockPaymentAdapter) VerifyCallback(ctx context.Context, data []byte, signature string) (*adapters.PaymentCallback, error) {
	returnArgs := m.Called(ctx, data, signature)
	return returnArgs.Get(0).(*adapters.PaymentCallback), returnArgs.Error(1)
}

func (m *MockPaymentAdapter) QueryPayment(ctx context.Context, req *adapters.QueryRequest) (*adapters.QueryResponse, error) {
	returnArgs := m.Called(ctx, req)
	return returnArgs.Get(0).(*adapters.QueryResponse), returnArgs.Error(1)
}

func (m *MockPaymentAdapter) CreateRefund(ctx context.Context, req *adapters.RefundRequest) (*adapters.RefundResponse, error) {
	returnArgs := m.Called(ctx, req)
	return returnArgs.Get(0).(*adapters.RefundResponse), returnArgs.Error(1)
}

func (m *MockPaymentAdapter) CloseOrder(ctx context.Context, orderNo string) error {
	returnArgs := m.Called(ctx, orderNo)
	return returnArgs.Error(0)
}

// MockAdapterFactory 模拟适配器工厂
type MockAdapterFactory struct {
	mock.Mock
	adapters map[string]adapters.PaymentAdapter
}

func (m *MockAdapterFactory) GetAdapter(channel string) (adapters.PaymentAdapter, error) {
	if adapter, exists := m.adapters[channel]; exists {
		return adapter, nil
	}
	return nil, fmt.Errorf("adapter not found for channel: %s", channel)
}

// TestPaymentService 测试支付服务
func TestPaymentService(t *testing.T) {
	fmt.Println("=== 支付服务单元测试 ===")

	// 创建模拟对象
	mockDB := &MockDatabase{}
	mockRedis := &MockRedis{}
	mockAdapter := &MockPaymentAdapter{}
	mockFactory := &MockAdapterFactory{
		adapters: map[string]adapters.PaymentAdapter{
			"alipay": mockAdapter,
		},
	}

	// 创建支付服务
	paymentService := services.NewPaymentService(mockDB, mockRedis, mockFactory)

	// 测试用例1: 创建支付订单
	t.Run("创建支付订单", func(t *testing.T) {
		testCreatePayment(t, paymentService, mockAdapter, mockDB)
	})

	// 测试用例2: 验证支付回调
	t.Run("验证支付回调", func(t *testing.T) {
		testVerifyCallback(t, paymentService, mockAdapter, mockDB)
	})

	// 测试用例3: 查询支付状态
	t.Run("查询支付状态", func(t *testing.T) {
		testQueryPayment(t, paymentService, mockAdapter, mockDB)
	})

	// 测试用例4: 创建退款
	t.Run("创建退款", func(t *testing.T) {
		testCreateRefund(t, paymentService, mockAdapter, mockDB)
	})

	fmt.Println("✓ 所有支付服务测试通过")
}

// testCreatePayment 测试创建支付订单
func testCreatePayment(t *testing.T, service *services.PaymentService, mockAdapter *MockPaymentAdapter, mockDB *MockDatabase) {
	// 准备测试数据
	req := &adapters.PaymentRequest{
		OrderNo:     "TEST20230830123456",
		Amount:      decimal.NewFromFloat(99.00),
		Subject:     "高考志愿填报助手-会员服务",
		Description: "基础会员-月度订阅",
		NotifyURL:   "/api/v1/payments/callback/alipay",
		ReturnURL:   "https://gaokaohub.com/payment/return",
		UserID:      "test_user_123",
		ClientIP:    "127.0.0.1",
		ExpireTime:  time.Hour,
		Metadata: map[string]interface{}{
			"channel": "alipay",
			"product_id": "basic_monthly",
		},
	}

	// 设置模拟期望
	expectedResp := &adapters.PaymentResponse{
		OrderNo:        req.OrderNo,
		PaymentURL:     "https://alipay.com/test/payment",
		QRCode:         "data:image/png;base64,test",
		ExpireTime:     time.Now().Add(time.Hour),
		ChannelTradeNo: "2023083022001499911234567890",
	}

	mockAdapter.On("CreatePayment", mock.Anything, req).Return(expectedResp, nil)
	mockDB.On("ExecContext", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// 执行测试
	resp, err := service.CreatePayment(context.Background(), req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, req.OrderNo, resp.OrderNo)
	assert.Equal(t, "https://alipay.com/test/payment", resp.PaymentURL)

	fmt.Println("✓ 创建支付订单测试通过")
}

// testVerifyCallback 测试验证支付回调
func testVerifyCallback(t *testing.T, service *services.PaymentService, mockAdapter *MockPaymentAdapter, mockDB *MockDatabase) {
	// 准备测试数据
	callbackData := []byte(`{"trade_no":"2023083022001499911234567890","out_trade_no":"TEST20230830123456","trade_status":"TRADE_SUCCESS"}`)
	signature := "test_signature"

	// 设置模拟期望
	expectedCallback := &adapters.PaymentCallback{
		OrderNo:        "TEST20230830123456",
		ChannelTradeNo: "2023083022001499911234567890",
		Amount:         decimal.NewFromFloat(99.00),
		Status:         "success",
		PaidAt:         time.Now(),
		Metadata:       map[string]interface{}{},
	}

	mockAdapter.On("VerifyCallback", mock.Anything, callbackData, signature).Return(expectedCallback, nil)
	mockDB.On("ExecContext", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// 执行测试
	callback, err := service.VerifyCallback(context.Background(), "alipay", callbackData, signature)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, callback)
	assert.Equal(t, "success", callback.Status)
	assert.Equal(t, "TEST20230830123456", callback.OrderNo)

	fmt.Println("✓ 验证支付回调测试通过")
}

// testQueryPayment 测试查询支付状态
func testQueryPayment(t *testing.T, service *services.PaymentService, mockAdapter *MockPaymentAdapter, mockDB *MockDatabase) {
	orderNo := "TEST20230830123456"

	// 设置模拟期望
	expectedQuery := &adapters.QueryResponse{
		OrderNo:        orderNo,
		ChannelTradeNo: "2023083022001499911234567890",
		Amount:         decimal.NewFromFloat(99.00),
		Status:         "paid",
		PaidAt:         time.Now(),
	}

	mockAdapter.On("QueryPayment", mock.Anything, &adapters.QueryRequest{
		OrderNo: orderNo,
		ChannelTradeNo: "2023083022001499911234567890",
	}).Return(expectedQuery, nil)

	mockDB.On("QueryRowContext", mock.Anything, mock.Anything, mock.Anything).Return(&models.PaymentOrder{
		OrderNo:        orderNo,
		Status:         models.PaymentStatusPending,
		ChannelTradeNo: "2023083022001499911234567890",
	})

	mockDB.On("ExecContext", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// 执行测试
	resp, err := service.QueryPayment(context.Background(), orderNo)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "paid", resp.Status)
	assert.Equal(t, orderNo, resp.OrderNo)

	fmt.Println("✓ 查询支付状态测试通过")
}

// testCreateRefund 测试创建退款
func testCreateRefund(t *testing.T, service *services.PaymentService, mockAdapter *MockPaymentAdapter, mockDB *MockDatabase) {
	// 准备测试数据
	req := &adapters.RefundRequest{
		OrderNo:   "TEST20230830123456",
		RefundNo:  "REFUND20230830123456",
		Amount:    decimal.NewFromFloat(99.00),
		Reason:    "用户申请退款",
		NotifyURL: "/api/v1/payments/refund/callback",
	}

	// 设置模拟期望
	expectedRefund := &adapters.RefundResponse{
		RefundNo:       req.RefundNo,
		OrderNo:        req.OrderNo,
		Status:         "processing",
		ChannelRefundNo: "2023083022001499911234567890",
	}

	mockAdapter.On("CreateRefund", mock.Anything, req).Return(expectedRefund, nil)

	mockDB.On("QueryRowContext", mock.Anything, mock.Anything, mock.Anything).Return(&models.PaymentOrder{
		OrderNo:        req.OrderNo,
		Status:         models.PaymentStatusPaid,
		ChannelTradeNo: "2023083022001499911234567890",
		Amount:         decimal.NewFromFloat(99.00),
	})

	mockDB.On("ExecContext", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// 执行测试
	resp, err := service.CreateRefund(context.Background(), req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "processing", resp.Status)
	assert.Equal(t, req.RefundNo, resp.RefundNo)

	fmt.Println("✓ 创建退款测试通过")
}

// 运行所有测试
// func main() {
// 	// 创建测试实例
// 	t := &testing.T{}
// 	
// 	fmt.Println("开始支付服务单元测试...")
// 	
// 	// 运行测试
// 	TestPaymentService(t)
// 	
// 	fmt.Println("\n=== 支付服务单元测试完成 ===")
// 	fmt.Println("所有核心支付功能验证通过！")
// }
