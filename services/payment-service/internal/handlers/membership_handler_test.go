package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oktetopython/gaokao/services/payment-service/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// MockMembershipService 模拟会员服务
type MockMembershipService struct {
	GetPlansFunc            func(ctx context.Context) ([]*models.MembershipPlan, error)
	SubscribeFunc           func(ctx context.Context, userID, orderNo string) error
	GetMembershipStatusFunc func(ctx context.Context, userID string) (*models.MembershipStatusResponse, error)
	RenewMembershipFunc     func(ctx context.Context, userID, planCode string) (string, error)
	CancelMembershipFunc    func(ctx context.Context, userID string) error
	GetMemberBenefitsFunc   func(ctx context.Context, userID string) (map[string]interface{}, error)
}

func (m *MockMembershipService) GetPlans(ctx context.Context) ([]*models.MembershipPlan, error) {
	if m.GetPlansFunc != nil {
		return m.GetPlansFunc(ctx)
	}
	return nil, nil
}

func (m *MockMembershipService) Subscribe(ctx context.Context, userID, orderNo string) error {
	if m.SubscribeFunc != nil {
		return m.SubscribeFunc(ctx, userID, orderNo)
	}
	return nil
}

func (m *MockMembershipService) GetMembershipStatus(ctx context.Context, userID string) (*models.MembershipStatusResponse, error) {
	if m.GetMembershipStatusFunc != nil {
		return m.GetMembershipStatusFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockMembershipService) RenewMembership(ctx context.Context, userID, planCode string) (string, error) {
	if m.RenewMembershipFunc != nil {
		return m.RenewMembershipFunc(ctx, userID, planCode)
	}
	return "", nil
}

func (m *MockMembershipService) CancelMembership(ctx context.Context, userID string) error {
	if m.CancelMembershipFunc != nil {
		return m.CancelMembershipFunc(ctx, userID)
	}
	return nil
}

func (m *MockMembershipService) GetMemberBenefits(ctx context.Context, userID string) (map[string]interface{}, error) {
	if m.GetMemberBenefitsFunc != nil {
		return m.GetMemberBenefitsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockMembershipService) ConsumeQuery(ctx context.Context, userID string) error {
	return nil
}

func (m *MockMembershipService) ConsumeDownload(ctx context.Context, userID string) error {
	return nil
}

func TestMembershipHandler_GetPlans(t *testing.T) {
	// 创建模拟服务
	mockService := &MockMembershipService{
		GetPlansFunc: func(ctx context.Context) ([]*models.MembershipPlan, error) {
			return []*models.MembershipPlan{
				{
					PlanCode:     "basic",
					Name:         "基础版",
					Price:        decimal.NewFromFloat(29.90),
					DurationDays: 30,
					Features:     models.JSONB{"basic_query": true},
					MaxQueries:   100,
					IsActive:     true,
				},
			}, nil
		},
	}

	// 创建处理器
	handler := NewMembershipHandler(mockService)

	// 创建请求
	req, err := http.NewRequest("GET", "/api/v1/membership/plans", nil)
	assert.NoError(t, err)

	// 创建响应记录器
	rr := httptest.NewRecorder()

	// 调用处理器方法
	handler.GetPlans(rr, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, rr.Code)

	var plans []*models.MembershipPlan
	err = json.Unmarshal(rr.Body.Bytes(), &plans)
	assert.NoError(t, err)
	assert.Len(t, plans, 1)
	assert.Equal(t, "basic", plans[0].PlanCode)
	assert.Equal(t, "基础版", plans[0].Name)
}

func TestMembershipHandler_Subscribe(t *testing.T) {
	// 创建模拟服务
	mockService := &MockMembershipService{
		SubscribeFunc: func(ctx context.Context, userID, orderNo string) error {
			return nil
		},
	}

	// 创建处理器
	handler := NewMembershipHandler(mockService)

	// 创建请求
	req, err := http.NewRequest("POST", "/api/v1/membership/subscribe?order_no=ORDER202301010001", nil)
	assert.NoError(t, err)
	req.Header.Set("X-User-ID", "user123")

	// 创建响应记录器
	rr := httptest.NewRecorder()

	// 调用处理器方法
	handler.Subscribe(rr, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "会员订阅成功", response["message"])
}

func TestMembershipHandler_GetMembershipStatus(t *testing.T) {
	// 创建模拟服务
	mockService := &MockMembershipService{
		GetMembershipStatusFunc: func(ctx context.Context, userID string) (*models.MembershipStatusResponse, error) {
			return &models.MembershipStatusResponse{
				IsVIP:    true,
				PlanCode: "basic",
				PlanName: "基础版",
			}, nil
		},
	}

	// 创建处理器
	handler := NewMembershipHandler(mockService)

	// 创建请求
	req, err := http.NewRequest("GET", "/api/v1/membership/status", nil)
	assert.NoError(t, err)
	req.Header.Set("X-User-ID", "user123")

	// 创建响应记录器
	rr := httptest.NewRecorder()

	// 调用处理器方法
	handler.GetMembershipStatus(rr, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, rr.Code)

	var status models.MembershipStatusResponse
	err = json.Unmarshal(rr.Body.Bytes(), &status)
	assert.NoError(t, err)
	assert.True(t, status.IsVIP)
	assert.Equal(t, "basic", status.PlanCode)
	assert.Equal(t, "基础版", status.PlanName)
}
