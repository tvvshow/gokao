package services

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redis/v8"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/gaokaohub/payment-service/internal/models"
)

func TestOrderServiceNew_CreateOrder(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Create a mock Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	orderService := NewOrderServiceNew(db, redisClient)

	userID := "user123"
	planCode := "basic"

	// Mock the membership plan query
	planRows := sqlmock.NewRows([]string{
		"id", "plan_code", "name", "description", "price", "duration_days",
		"features", "max_queries", "max_downloads", "is_active", "created_at", "updated_at",
	}).AddRow(
		1, planCode, "基础版", "基础功能套餐", 29.90, 30,
		`{"basic_query": true}`, 100, 0, true, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT id, plan_code, name, description, price, duration_days,
		       features, max_queries, max_downloads, is_active,
		       created_at, updated_at
		FROM membership_plans
		WHERE plan_code = \$1`).
		WithArgs(planCode).
		WillReturnRows(planRows)

	// Mock the payment order creation
	mock.ExpectQuery(`INSERT INTO payment_orders`).
		WithArgs(
			sqlmock.AnyArg(), userID, 29.90, "CNY", "会员套餐-基础版", "基础功能套餐",
			models.PaymentStatusPending, models.PaymentChannelAlipay, sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Test the CreateOrder method
	req := &models.CreateOrderRequest{
		PlanCode:       planCode,
		PaymentChannel: models.PaymentChannelAlipay,
	}

	response, err := orderService.CreateOrder(context.Background(), userID, req)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.OrderNo)
	assert.Equal(t, decimal.NewFromFloat(29.90), response.Amount)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderServiceNew_GetOrder(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Create a mock Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	orderService := NewOrderServiceNew(db, redisClient)

	userID := "user123"
	orderNo := "ORDER202301010001"

	// Mock the payment order query
	rows := sqlmock.NewRows([]string{
		"id", "order_no", "user_id", "amount", "currency", "subject", "description",
		"status", "payment_channel", "channel_trade_no", "client_ip", "notify_url",
		"return_url", "expire_time", "paid_at", "created_at", "updated_at", "metadata",
	}).AddRow(
		1, orderNo, userID, 29.90, "CNY", "基础版会员", "基础功能套餐",
		models.PaymentStatusPending, models.PaymentChannelAlipay, "", "", "",
		"", time.Now(), nil, time.Now(), time.Now(), `{"plan_code": "basic"}`,
	)

	mock.ExpectQuery(`SELECT id, order_no, user_id, amount, currency, subject, description,
		       status, payment_channel, channel_trade_no, client_ip, notify_url,
		       return_url, expire_time, paid_at, created_at, updated_at, metadata
		FROM payment_orders
		WHERE order_no = \$1`).
		WithArgs(orderNo).
		WillReturnRows(rows)

	// Test the GetOrder method
	order, err := orderService.GetOrder(context.Background(), userID, orderNo)
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, orderNo, order.OrderNo)
	assert.Equal(t, userID, order.UserID)
	assert.Equal(t, models.PaymentStatusPending, order.Status)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderServiceNew_CancelOrder(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Create a mock Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	orderService := NewOrderServiceNew(db, redisClient)

	userID := "user123"
	orderNo := "ORDER202301010001"

	// Mock the payment order query
	rows := sqlmock.NewRows([]string{
		"id", "order_no", "user_id", "amount", "currency", "subject", "description",
		"status", "payment_channel", "channel_trade_no", "client_ip", "notify_url",
		"return_url", "expire_time", "paid_at", "created_at", "updated_at", "metadata",
	}).AddRow(
		1, orderNo, userID, 29.90, "CNY", "基础版会员", "基础功能套餐",
		models.PaymentStatusPending, models.PaymentChannelAlipay, "", "", "",
		"", time.Now(), nil, time.Now(), time.Now(), `{"plan_code": "basic"}`,
	)

	mock.ExpectQuery(`SELECT id, order_no, user_id, amount, currency, subject, description,
		       status, payment_channel, channel_trade_no, client_ip, notify_url,
		       return_url, expire_time, paid_at, created_at, updated_at, metadata
		FROM payment_orders
		WHERE order_no = \$1`).
		WithArgs(orderNo).
		WillReturnRows(rows)

	// Mock the payment order update
	mock.ExpectExec(`UPDATE payment_orders SET status = \$1, updated_at = \$2 WHERE order_no = \$3 AND user_id = \$4`).
		WithArgs(models.PaymentStatusCanceled, sqlmock.AnyArg(), orderNo, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Test the CancelOrder method
	err = orderService.CancelOrder(context.Background(), userID, orderNo)
	assert.NoError(t, err)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderServiceNew_UpdateOrderStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Create a mock Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	orderService := NewOrderServiceNew(db, redisClient)

	orderNo := "ORDER202301010001"
	newStatus := models.PaymentStatusPaid

	// Mock the payment order update
	mock.ExpectExec(`UPDATE payment_orders SET status = \$1, updated_at = \$2 WHERE order_no = \$3`).
		WithArgs(newStatus, sqlmock.AnyArg(), orderNo).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Test the UpdateOrderStatus method
	err = orderService.UpdateOrderStatus(context.Background(), orderNo, newStatus)
	assert.NoError(t, err)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}