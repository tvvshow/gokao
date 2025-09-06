package main

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
<<<<<<< HEAD
	"github.com/gaokaohub/gaokao/services/payment-service/internal/models"
	"github.com/gaokaohub/gaokao/services/payment-service/internal/services"
=======
	"github.com/gaokaohub/payment-service/internal/models"
	"github.com/gaokaohub/payment-service/internal/services"
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
	"github.com/go-redis/redis/v8"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// TestPaymentAndMembershipIntegration 测试支付和会员系统的完整集成
func TestPaymentAndMembershipIntegration(t *testing.T) {
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
	membershipService := services.NewMembershipServiceNew(db, redisClient)

	t.Run("CompletePaymentAndMembershipFlow", func(t *testing.T) {
		userID := "user123"
		planCode := "basic"

		// 1. 获取会员套餐
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
		WHERE is_active = true
		ORDER BY price ASC`).
			WillReturnRows(planRows)

		plans, err := membershipService.GetPlans(context.Background())
		assert.NoError(t, err)
		assert.Len(t, plans, 1)
		assert.Equal(t, planCode, plans[0].PlanCode)
		assert.Equal(t, "基础版", plans[0].Name)

		// 2. 创建订单
		planRow := sqlmock.NewRows([]string{
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
			WillReturnRows(planRow)

		mock.ExpectQuery(`INSERT INTO payment_orders`).
			WithArgs(
				sqlmock.AnyArg(), userID, 29.90, "CNY", "会员套餐-基础版", "基础功能套餐",
				models.PaymentStatusPending, models.PaymentChannelAlipay, sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		createOrderReq := &models.CreateOrderRequest{
			PlanCode:       planCode,
			PaymentChannel: models.PaymentChannelAlipay,
		}

		orderResponse, err := orderService.CreateOrder(context.Background(), userID, createOrderReq)
		assert.NoError(t, err)
		assert.NotEmpty(t, orderResponse.OrderNo)
		assert.Equal(t, decimal.NewFromFloat(29.90), orderResponse.Amount)

		// 3. 模拟支付成功
		orderRows := sqlmock.NewRows([]string{
			"id", "order_no", "user_id", "amount", "currency", "subject", "description",
			"status", "payment_channel", "channel_trade_no", "client_ip", "notify_url",
			"return_url", "expire_time", "paid_at", "created_at", "updated_at", "metadata",
		}).AddRow(
			1, orderResponse.OrderNo, userID, 29.90, "CNY", "基础版会员", "基础功能套餐",
			models.PaymentStatusPending, models.PaymentChannelAlipay, "trade123", "127.0.0.1", "",
			"", time.Now(), nil, time.Now(), time.Now(), `{"plan_code": "basic", "auto_renew": false}`,
		)

		mock.ExpectQuery(`SELECT id, order_no, user_id, amount, currency, subject, description,
		       status, payment_channel, channel_trade_no, client_ip, notify_url,
		       return_url, expire_time, paid_at, created_at, updated_at, metadata
		FROM payment_orders
		WHERE order_no = \$1`).
			WithArgs(orderResponse.OrderNo).
			WillReturnRows(orderRows)

		mock.ExpectExec(`UPDATE payment_orders SET status = \$1, updated_at = \$2 WHERE order_no = \$3`).
			WithArgs(models.PaymentStatusPaid, sqlmock.AnyArg(), orderResponse.OrderNo).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err = orderService.UpdateOrderStatus(context.Background(), orderResponse.OrderNo, models.PaymentStatusPaid)
		assert.NoError(t, err)

		// 4. 订阅会员
		// 重新获取订单信息（已支付）
		paidOrderRows := sqlmock.NewRows([]string{
			"id", "order_no", "user_id", "amount", "currency", "subject", "description",
			"status", "payment_channel", "channel_trade_no", "client_ip", "notify_url",
			"return_url", "expire_time", "paid_at", "created_at", "updated_at", "metadata",
		}).AddRow(
			1, orderResponse.OrderNo, userID, 29.90, "CNY", "基础版会员", "基础功能套餐",
			models.PaymentStatusPaid, models.PaymentChannelAlipay, "trade123", "127.0.0.1", "",
			"", time.Now(), time.Now(), time.Now(), time.Now(), `{"plan_code": "basic", "auto_renew": false}`,
		)

		mock.ExpectQuery(`SELECT id, order_no, user_id, amount, currency, subject, description,
		       status, payment_channel, channel_trade_no, client_ip, notify_url,
		       return_url, expire_time, paid_at, created_at, updated_at, metadata
		FROM payment_orders
		WHERE order_no = \$1`).
			WithArgs(orderResponse.OrderNo).
			WillReturnRows(paidOrderRows)

		// 获取会员套餐信息
		mock.ExpectQuery(`SELECT id, plan_code, name, description, price, duration_days,
		       features, max_queries, max_downloads, is_active,
		       created_at, updated_at
		FROM membership_plans
		WHERE plan_code = \$1`).
			WithArgs(planCode).
			WillReturnRows(planRow)

		// 检查用户是否已有会员（应该没有）
		mock.ExpectQuery(`SELECT id, user_id, plan_code, order_no, start_time,
		       end_time, status, auto_renew, used_queries,
		       used_downloads, created_at, updated_at
		FROM user_memberships
		WHERE user_id = \$1
		ORDER BY end_time DESC
		LIMIT 1`).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		// 创建新会员
		mock.ExpectQuery(`INSERT INTO user_memberships`).
			WithArgs(
				userID, planCode, orderResponse.OrderNo, sqlmock.AnyArg(), sqlmock.AnyArg(),
				models.MembershipStatusActive, false, 0, 0, sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		err = membershipService.Subscribe(context.Background(), userID, orderResponse.OrderNo)
		assert.NoError(t, err)

		// 5. 获取会员状态
		membershipRows := sqlmock.NewRows([]string{
			"um.id", "um.user_id", "um.plan_code", "um.order_no", "um.start_time",
			"um.end_time", "um.status", "um.auto_renew", "um.used_queries",
			"um.used_downloads", "um.created_at", "um.updated_at",
			"mp.name", "mp.features", "mp.max_queries", "mp.max_downloads",
		}).AddRow(
			1, userID, planCode, orderResponse.OrderNo, time.Now(), time.Now().AddDate(0, 0, 30),
			models.MembershipStatusActive, false, 0, 0, time.Now(), time.Now(),
			"基础版", `{"basic_query": true}`, 100, 0,
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

		status, err := membershipService.GetMembershipStatus(context.Background(), userID)
		assert.NoError(t, err)
		assert.True(t, status.IsVIP)
		assert.Equal(t, planCode, status.PlanCode)
		assert.Equal(t, "基础版", status.PlanName)
		assert.Equal(t, 30, status.RemainingDays)

		// 6. 取消会员
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

		mock.ExpectExec(`UPDATE user_memberships SET status = \$1, auto_renew = false, updated_at = \$2 WHERE user_id = \$3 AND status = \$4`).
			WithArgs(models.MembershipStatusCanceled, sqlmock.AnyArg(), userID, models.MembershipStatusActive).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err = membershipService.CancelMembership(context.Background(), userID)
		assert.NoError(t, err)

		// 确保所有期望都被满足
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestAPIIntegration 测试API集成
func TestAPIIntegration(t *testing.T) {
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
	membershipService := services.NewMembershipServiceNew(db, redisClient)

	// 创建测试服务器
	// 注意：这里我们不会真正启动服务器，而是直接测试处理器方法

	t.Run("APIHandlerIntegration", func(t *testing.T) {
		// 测试创建订单API
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

		// 确保所有期望都被满足
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
