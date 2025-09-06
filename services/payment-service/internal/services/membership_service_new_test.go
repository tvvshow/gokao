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

func TestMembershipServiceNew_Subscribe(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Create a mock Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	membershipService := NewMembershipServiceNew(db, redisClient)

	userID := "user123"
	orderNo := "ORDER202301010001"
	planCode := "basic"

	// Mock the payment order query
	orderRows := sqlmock.NewRows([]string{
		"id", "order_no", "user_id", "amount", "currency", "subject", "description",
		"status", "payment_channel", "channel_trade_no", "client_ip", "notify_url",
		"return_url", "expire_time", "paid_at", "created_at", "updated_at", "metadata",
	}).AddRow(
		1, orderNo, userID, 29.90, "CNY", "基础版会员", "基础功能套餐",
		models.PaymentStatusPaid, models.PaymentChannelAlipay, "", "", "",
		"", time.Now(), time.Now(), time.Now(), time.Now(), `{"plan_code": "basic", "auto_renew": false}`,
	)

	mock.ExpectQuery(`SELECT id, order_no, user_id, amount, currency, subject, description,
		       status, payment_channel, channel_trade_no, client_ip, notify_url,
		       return_url, expire_time, paid_at, created_at, updated_at, metadata
		FROM payment_orders
		WHERE order_no = \$1`).
		WithArgs(orderNo).
		WillReturnRows(orderRows)

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

	// Mock the user membership query (no existing membership)
	mock.ExpectQuery(`SELECT id, user_id, plan_code, order_no, start_time,
		       end_time, status, auto_renew, used_queries,
		       used_downloads, created_at, updated_at
		FROM user_memberships
		WHERE user_id = \$1
		ORDER BY end_time DESC
		LIMIT 1`).
		WithArgs(userID).
		WillReturnError(sqlmock.ErrCancelled)

	// Mock the user membership creation
	mock.ExpectQuery(`INSERT INTO user_memberships`).
		WithArgs(
			userID, planCode, orderNo, sqlmock.AnyArg(), sqlmock.AnyArg(),
			models.MembershipStatusActive, false, 0, 0, sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Test the Subscribe method
	err = membershipService.Subscribe(context.Background(), userID, orderNo)
	assert.NoError(t, err)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMembershipServiceNew_GetPlans(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Create a mock Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	membershipService := NewMembershipServiceNew(db, redisClient)

	// Mock the membership plans query
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

	// Test the GetPlans method
	plans, err := membershipService.GetPlans(context.Background())
	assert.NoError(t, err)
	assert.Len(t, plans, 2)
	assert.Equal(t, "basic", plans[0].PlanCode)
	assert.Equal(t, "premium", plans[1].PlanCode)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMembershipServiceNew_CancelMembership(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Create a mock Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	membershipService := NewMembershipServiceNew(db, redisClient)

	userID := "user123"

	// Mock the user membership query
	membershipRows := sqlmock.NewRows([]string{
		"id", "user_id", "plan_code", "order_no", "start_time", "end_time",
		"status", "auto_renew", "used_queries", "used_downloads", "created_at", "updated_at",
	}).AddRow(
		1, userID, "basic", "ORDER202301010001", time.Now().AddDate(0, 0, -10), time.Now().AddDate(0, 0, 20),
		models.MembershipStatusActive, false, 0, 0, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT um.id, um.user_id, um.plan_code, um.order_no, um.start_time,
		       um.end_time, um.status, um.auto_renew, um.used_queries,
		       um.used_downloads, um.created_at, um.updated_at,
		       mp.name, mp.features, mp.max_queries, mp.max_downloads
		FROM user_memberships um
		JOIN membership_plans mp ON um.plan_code = mp.plan_code
		WHERE um.user_id = \$1
		ORDER BY um.end_time DESC
		LIMIT 1`).
		WithArgs(userID).
		WillReturnRows(membershipRows)

	// Mock the user membership update
	mock.ExpectExec(`UPDATE user_memberships SET status = \$1, auto_renew = false, updated_at = \$2 WHERE user_id = \$3 AND status = \$4`).
		WithArgs(models.MembershipStatusCanceled, sqlmock.AnyArg(), userID, models.MembershipStatusActive).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Test the CancelMembership method
	err = membershipService.CancelMembership(context.Background(), userID)
	assert.NoError(t, err)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}