package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/oktetopython/gaokao/services/payment-service/internal/models"
)

type membershipServiceStub struct {
	getPlans            func(ctx context.Context) ([]*models.MembershipPlan, error)
	subscribe           func(ctx context.Context, userID, orderNo string) error
	getMembershipStatus func(ctx context.Context, userID string) (*models.MembershipStatusResponse, error)
	renewMembership     func(ctx context.Context, userID, planCode string) (string, error)
	cancelMembership    func(ctx context.Context, userID string) error
	getMemberBenefits   func(ctx context.Context, userID string) (map[string]interface{}, error)
	updateAutoRenew     func(ctx context.Context, userID string, autoRenew bool) error
}

func (s membershipServiceStub) GetPlans(ctx context.Context) ([]*models.MembershipPlan, error) {
	return s.getPlans(ctx)
}

func (s membershipServiceStub) Subscribe(ctx context.Context, userID, orderNo string) error {
	return s.subscribe(ctx, userID, orderNo)
}

func (s membershipServiceStub) GetMembershipStatus(ctx context.Context, userID string) (*models.MembershipStatusResponse, error) {
	return s.getMembershipStatus(ctx, userID)
}

func (s membershipServiceStub) RenewMembership(ctx context.Context, userID, planCode string) (string, error) {
	return s.renewMembership(ctx, userID, planCode)
}

func (s membershipServiceStub) CancelMembership(ctx context.Context, userID string) error {
	return s.cancelMembership(ctx, userID)
}

func (s membershipServiceStub) GetMemberBenefits(ctx context.Context, userID string) (map[string]interface{}, error) {
	return s.getMemberBenefits(ctx, userID)
}

func (s membershipServiceStub) UpdateAutoRenew(ctx context.Context, userID string, autoRenew bool) error {
	return s.updateAutoRenew(ctx, userID, autoRenew)
}

func TestMembershipHandlerRegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewMembershipHandler(membershipServiceStub{
		getPlans: func(ctx context.Context) ([]*models.MembershipPlan, error) {
			return []*models.MembershipPlan{
				{
					ID:           1,
					PlanCode:     "basic",
					Name:         "基础版",
					DurationDays: 30,
					Price:        29.9,
					IsActive:     true,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				},
			}, nil
		},
		subscribe: func(ctx context.Context, userID, orderNo string) error {
			if userID != "user-1" || orderNo != "ORDER-1" {
				t.Fatalf("unexpected subscribe args: userID=%s orderNo=%s", userID, orderNo)
			}
			return nil
		},
		getMembershipStatus: func(ctx context.Context, userID string) (*models.MembershipStatusResponse, error) {
			return &models.MembershipStatusResponse{
				IsVIP:         true,
				PlanCode:      "premium",
				PlanName:      "高级版",
				RemainingDays: 15,
				Features: map[string]interface{}{
					"ai_recommendation": true,
				},
			}, nil
		},
		renewMembership: func(ctx context.Context, userID, planCode string) (string, error) {
			if planCode != "premium" {
				t.Fatalf("unexpected renew plan code: %s", planCode)
			}
			return "RN-1", nil
		},
		cancelMembership: func(ctx context.Context, userID string) error {
			return nil
		},
		getMemberBenefits: func(ctx context.Context, userID string) (map[string]interface{}, error) {
			return map[string]interface{}{
				"features": map[string]interface{}{
					"ai_recommendation": true,
				},
			}, nil
		},
		updateAutoRenew: func(ctx context.Context, userID string, autoRenew bool) error {
			if !autoRenew {
				t.Fatalf("expected auto renew true")
			}
			return nil
		},
	})

	router := gin.New()
	api := router.Group("/api/v1")
	handler.RegisterRoutes(api.Group("/membership"))

	t.Run("plans", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/membership/plans", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("subscribe", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/membership/subscribe", strings.NewReader(`{"order_no":"ORDER-1"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-1")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/membership/status", nil)
		req.Header.Set("X-User-ID", "user-1")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
		}

		var resp struct {
			Success bool                            `json:"success"`
			Data    models.MembershipStatusResponse `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if !resp.Success || resp.Data.PlanCode != "premium" {
			t.Fatalf("unexpected status response: %+v", resp)
		}
	})

	t.Run("renew", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/membership/renew", strings.NewReader(`{"plan_code":"premium"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-1")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("cancel", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/membership/cancel", nil)
		req.Header.Set("X-User-ID", "user-1")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("benefits", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/membership/benefits", nil)
		req.Header.Set("X-User-ID", "user-1")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("auto renew", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/membership/auto-renew", strings.NewReader(`{"auto_renew":true}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-1")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
		}
	})
}

func TestMembershipHandlerUserIDRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewMembershipHandler(membershipServiceStub{
		getPlans:  func(ctx context.Context) ([]*models.MembershipPlan, error) { return nil, nil },
		subscribe: func(ctx context.Context, userID, orderNo string) error { return nil },
		getMembershipStatus: func(ctx context.Context, userID string) (*models.MembershipStatusResponse, error) {
			return &models.MembershipStatusResponse{}, nil
		},
		renewMembership:   func(ctx context.Context, userID, planCode string) (string, error) { return "", nil },
		cancelMembership:  func(ctx context.Context, userID string) error { return nil },
		getMemberBenefits: func(ctx context.Context, userID string) (map[string]interface{}, error) { return nil, nil },
		updateAutoRenew:   func(ctx context.Context, userID string, autoRenew bool) error { return nil },
	})

	router := gin.New()
	api := router.Group("/api/v1")
	handler.RegisterRoutes(api.Group("/membership"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/membership/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", w.Code, w.Body.String())
	}
}
