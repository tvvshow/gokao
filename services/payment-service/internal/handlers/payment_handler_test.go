package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gaokaohub/payment-service/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// MockOrderService 模拟订单服务
type MockOrderService struct {
	CreateOrderFunc       func(ctx context.Context, userID string, req *models.CreateOrderRequest) (*models.CreateOrderResponse, error)
	GetOrderFunc          func(ctx context.Context, userID, orderNo string) (*models.PaymentOrder, error)
	CancelOrderFunc       func(ctx context.Context, userID, orderNo string) error
	GetInvoiceFunc        func(ctx context.Context, userID, orderNo string) (map[string]interface{}, error)
	UpdateOrderStatusFunc func(ctx context.Context, orderNo, status string) error
}

func (m *MockOrderService) CreateOrder(ctx context.Context, userID string, req *models.CreateOrderRequest) (*models.CreateOrderResponse, error) {
	if m.CreateOrderFunc != nil {
		return m.CreateOrderFunc(ctx, userID, req)
	}
	return nil, nil
}

func (m *MockOrderService) GetOrder(ctx context.Context, userID, orderNo string) (*models.PaymentOrder, error) {
	if m.GetOrderFunc != nil {
		return m.GetOrderFunc(ctx, userID, orderNo)
	}
	return nil, nil
}

func (m *MockOrderService) CancelOrder(ctx context.Context, userID, orderNo string) error {
	if m.CancelOrderFunc != nil {
		return m.CancelOrderFunc(ctx, userID, orderNo)
	}
	return nil
}

func (m *MockOrderService) GetInvoice(ctx context.Context, userID, orderNo string) (map[string]interface{}, error) {
	if m.GetInvoiceFunc != nil {
		return m.GetInvoiceFunc(ctx, userID, orderNo)
	}
	return nil, nil
}

func (m *MockOrderService) UpdateOrderStatus(ctx context.Context, orderNo, status string) error {
	if m.UpdateOrderStatusFunc != nil {
		return m.UpdateOrderStatusFunc(ctx, orderNo, status)
	}
	return nil
}

func TestPaymentHandler_CreateOrder(t *testing.T) {
	// 创建模拟服务
	mockService := &MockOrderService{
		CreateOrderFunc: func(ctx context.Context, userID string, req *models.CreateOrderRequest) (*models.CreateOrderResponse, error) {
			return &models.CreateOrderResponse{
				OrderNo:   "ORDER202301010001",
				Amount:    decimal.NewFromFloat(29.90),
				ExpiredAt: time.Now().Add(30 * time.Minute),
			}, nil
		},
	}

	// 创建处理器
	handler := NewPaymentHandler(mockService)

	// 创建请求
	reqBody := `{"plan_code": "basic", "payment_channel": "alipay"}`
	req, err := http.NewRequest("POST", "/api/v1/payment/orders", bytes.NewBufferString(reqBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "user123")

	// 创建响应记录器
	rr := httptest.NewRecorder()

	// 调用处理器方法
	handler.CreateOrder(rr, req)

	// 验证响应
	assert.Equal(t, http.StatusCreated, rr.Code)

	var response models.CreateOrderResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ORDER202301010001", response.OrderNo)
	assert.Equal(t, decimal.NewFromFloat(29.90), response.Amount)
}

func TestPaymentHandler_GetOrder(t *testing.T) {
	// 创建模拟服务
	mockService := &MockOrderService{
		GetOrderFunc: func(ctx context.Context, userID, orderNo string) (*models.PaymentOrder, error) {
			return &models.PaymentOrder{
				OrderNo:  "ORDER202301010001",
				UserID:   "user123",
				Amount:   decimal.NewFromFloat(29.90),
				Currency: "CNY",
				Subject:  "基础版会员",
				Status:   models.PaymentStatusPending,
			}, nil
		},
	}

	// 创建处理器
	handler := NewPaymentHandler(mockService)

	// 创建请求
	req, err := http.NewRequest("GET", "/api/v1/payment/orders?order_no=ORDER202301010001", nil)
	assert.NoError(t, err)
	req.Header.Set("X-User-ID", "user123")

	// 创建响应记录器
	rr := httptest.NewRecorder()

	// 调用处理器方法
	handler.GetOrder(rr, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, rr.Code)

	var order models.PaymentOrder
	err = json.Unmarshal(rr.Body.Bytes(), &order)
	assert.NoError(t, err)
	assert.Equal(t, "ORDER202301010001", order.OrderNo)
	assert.Equal(t, "user123", order.UserID)
	assert.Equal(t, models.PaymentStatusPending, order.Status)
}

func TestPaymentHandler_CancelOrder(t *testing.T) {
	// 创建模拟服务
	mockService := &MockOrderService{
		CancelOrderFunc: func(ctx context.Context, userID, orderNo string) error {
			return nil
		},
	}

	// 创建处理器
	handler := NewPaymentHandler(mockService)

	// 创建请求
	req, err := http.NewRequest("POST", "/api/v1/payment/orders/cancel?order_no=ORDER202301010001", nil)
	assert.NoError(t, err)
	req.Header.Set("X-User-ID", "user123")

	// 创建响应记录器
	rr := httptest.NewRecorder()

	// 调用处理器方法
	handler.CancelOrder(rr, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "订单已取消", response["message"])
}
