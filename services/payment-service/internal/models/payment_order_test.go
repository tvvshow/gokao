//go:build legacy
// +build legacy

package models

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestPaymentOrderDB_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	paymentOrderDB := NewPaymentOrderDB(db)

	order := &PaymentOrder{
		OrderNo:        "ORDER202301010001",
		UserID:         "user123",
		Amount:         decimal.NewFromFloat(99.99),
		Currency:       "CNY",
		Subject:        "测试订单",
		Description:    "测试订单描述",
		Status:         PaymentStatusPending,
		PaymentChannel: PaymentChannelAlipay,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Metadata:       JSONB{"key": "value"},
	}

	// Mock the database operation
	mock.ExpectQuery(`INSERT INTO payment_orders`).
		WithArgs(
			order.OrderNo, order.UserID, order.Amount, order.Currency,
			order.Subject, order.Description, order.Status, order.PaymentChannel,
			sqlmock.AnyArg(), order.CreatedAt, order.UpdatedAt, order.Metadata,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err = paymentOrderDB.Create(order)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), order.ID)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPaymentOrderDB_GetByOrderNo(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	paymentOrderDB := NewPaymentOrderDB(db)

	orderNo := "ORDER202301010001"

	// Mock the database operation
	rows := sqlmock.NewRows([]string{
		"id", "order_no", "user_id", "amount", "currency", "subject", "description",
		"status", "payment_channel", "channel_trade_no", "client_ip", "notify_url",
		"return_url", "expire_time", "paid_at", "created_at", "updated_at", "metadata",
	}).AddRow(
		1, orderNo, "user123", 99.99, "CNY", "测试订单", "测试订单描述",
		PaymentStatusPending, PaymentChannelAlipay, "", "", "",
		"", time.Now(), nil, time.Now(), time.Now(), `{"key": "value"}`,
	)

	mock.ExpectQuery(`SELECT id, order_no, user_id, amount, currency, subject, description,
		       status, payment_channel, channel_trade_no, client_ip, notify_url,
		       return_url, expire_time, paid_at, created_at, updated_at, metadata
		FROM payment_orders
		WHERE order_no = \$1`).
		WithArgs(orderNo).
		WillReturnRows(rows)

	order, err := paymentOrderDB.GetByOrderNo(orderNo)
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, orderNo, order.OrderNo)
	assert.Equal(t, "user123", order.UserID)
	assert.Equal(t, PaymentStatusPending, order.Status)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPaymentOrderDB_UpdateStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	paymentOrderDB := NewPaymentOrderDB(db)

	orderNo := "ORDER202301010001"
	newStatus := PaymentStatusPaid

	// Mock the database operation
	mock.ExpectExec(`UPDATE payment_orders SET status = \$1, updated_at = \$2 WHERE order_no = \$3`).
		WithArgs(newStatus, sqlmock.AnyArg(), orderNo).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = paymentOrderDB.UpdateStatus(orderNo, newStatus)
	assert.NoError(t, err)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
