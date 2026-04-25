package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/oktetopython/gaokao/services/payment-service/internal/models"
	"github.com/oktetopython/gaokao/services/payment-service/internal/services"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestPaymentAPIIntegration(t *testing.T) {
	// 创建模拟数据库
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// 创建模拟Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 创建服务
	orderService := services.NewOrderServiceNew(db, redisClient)

	// 创建处理器
	paymentHandler := NewPaymentHandler(orderService)

	// 创建路由器
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/payment/orders", paymentHandler.CreateOrder).Methods("POST")
	router.HandleFunc("/api/v1/payment/orders", paymentHandler.GetOrder).Methods("GET")
	router.HandleFunc("/api/v1/payment/orders/cancel", paymentHandler.CancelOrder).Methods("POST")

	t.Run("CreateOrder", func(t *testing.T) {
		// 模拟数据库查询
		planRows := sqlmock.NewRows([]string{
			"id", "plan_code", "name", "description", "price", "duration_days",
			"features", "max_queries", "max_downloads", "is_active", "created_at", "updated_at",
		}).AddRow(
			1, "basic", "基础版", "基础功能套餐", 29.90, 30,
			`{"basic_query": true}`, 100, 0, true, time.Now(), time.Now(),
		)

		mock.ExpectQuery(`SELECT id, plan_code, name, description, price, duration_days,
		       features, max_queries, max_downloads, is_active,
		       created_at, updated_at
		FROM membership_plans
		WHERE plan_code = \$1`).
			WithArgs("basic").
			WillReturnRows(planRows)

		mock.ExpectQuery(`INSERT INTO payment_orders`).
			WithArgs(
				sqlmock.AnyArg(), "user123", 29.90, "CNY", "会员套餐-基础版", "基础功能套餐",
				models.PaymentStatusPending, models.PaymentChannelAlipay, sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// 创建请求
		reqBody := `{"plan_code": "basic", "payment_channel": "alipay"}`
		req, err := http.NewRequest("POST", "/api/v1/payment/orders", bytes.NewBufferString(reqBody))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user123")

		// 创建响应记录器
		rr := httptest.NewRecorder()

		// 调用API
		router.ServeHTTP(rr, req)

		// 验证响应
		assert.Equal(t, http.StatusCreated, rr.Code)

		var response models.CreateOrderResponse
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.OrderNo)
		assert.Equal(t, decimal.NewFromFloat(29.90), response.Amount)
	})

	t.Run("GetOrder", func(t *testing.T) {
		// 模拟数据库查询
		orderRows := sqlmock.NewRows([]string{
			"id", "order_no", "user_id", "amount", "currency", "subject", "description",
			"status", "payment_channel", "channel_trade_no", "client_ip", "notify_url",
			"return_url", "expire_time", "paid_at", "created_at", "updated_at", "metadata",
		}).AddRow(
			1, "ORDER202301010001", "user123", 29.90, "CNY", "基础版会员", "基础功能套餐",
			models.PaymentStatusPending, models.PaymentChannelAlipay, "", "", "",
			"", time.Now(), nil, time.Now(), time.Now(), `{"plan_code": "basic"}`,
		)

		mock.ExpectQuery(`SELECT id, order_no, user_id, amount, currency, subject, description,
		       status, payment_channel, channel_trade_no, client_ip, notify_url,
		       return_url, expire_time, paid_at, created_at, updated_at, metadata
		FROM payment_orders
		WHERE order_no = \$1`).
			WithArgs("ORDER202301010001").
			WillReturnRows(orderRows)

		// 创建请求
		req, err := http.NewRequest("GET", "/api/v1/payment/orders?order_no=ORDER202301010001", nil)
		assert.NoError(t, err)
		req.Header.Set("X-User-ID", "user123")

		// 创建响应记录器
		rr := httptest.NewRecorder()

		// 调用API
		router.ServeHTTP(rr, req)

		// 验证响应
		assert.Equal(t, http.StatusOK, rr.Code)

		var order models.PaymentOrder
		err = json.Unmarshal(rr.Body.Bytes(), &order)
		assert.NoError(t, err)
		assert.Equal(t, "ORDER202301010001", order.OrderNo)
		assert.Equal(t, "user123", order.UserID)
		assert.Equal(t, models.PaymentStatusPending, order.Status)
	})

	// 确保所有期望都被满足
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMembershipAPIIntegration(t *testing.T) {
	// 创建模拟数据库
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// 创建模拟Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 创建服务
	membershipService := services.NewMembershipServiceNew(db, redisClient)

	// 创建处理器
	membershipHandler := NewMembershipHandler(membershipService)

	// 创建路由器
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/membership/plans", membershipHandler.GetPlans).Methods("GET")
	router.HandleFunc("/api/v1/membership/subscribe", membershipHandler.Subscribe).Methods("POST")
	router.HandleFunc("/api/v1/membership/status", membershipHandler.GetMembershipStatus).Methods("GET")

	t.Run("GetPlans", func(t *testing.T) {
		// 模拟数据库查询
		rows := sqlmock.NewRows([]string{
			"id", "plan_code", "name", "description", "price", "duration_days",
			"features", "max_queries", "max_downloads", "is_active", "created_at", "updated_at",
		}).
			AddRow(1, "basic", "基础版", "基础功能套餐", 29.90, 30, `{"basic_query": true}`, 100, 0, true, time.Now(), time.Now()).
			AddRow(2, "premium", "高级版", "高级功能套餐", 99.90, 90, `{"basic_query": true, "data_export": true}`, 1000, 50, true, time.Now(), time.Now())

		mock.ExpectQuery(`SELECT id, plan_code, name, description, price, duration_days,
		       features, max_queries, max_downloads, is_active,
		       created_at, updated_at
		FROM membership_plans
		WHERE is_active = true
		ORDER BY price ASC`).
			WillReturnRows(rows)

		// 创建请求
		req, err := http.NewRequest("GET", "/api/v1/membership/plans", nil)
		assert.NoError(t, err)

		// 创建响应记录器
		rr := httptest.NewRecorder()

		// 调用API
		router.ServeHTTP(rr, req)

		// 验证响应
		assert.Equal(t, http.StatusOK, rr.Code)

		var plans []*models.MembershipPlan
		err = json.Unmarshal(rr.Body.Bytes(), &plans)
		assert.NoError(t, err)
		assert.Len(t, plans, 2)
		assert.Equal(t, "basic", plans[0].PlanCode)
		assert.Equal(t, "premium", plans[1].PlanCode)
	})

	// 确保所有期望都被满足
	assert.NoError(t, mock.ExpectationsWereMet())
}
