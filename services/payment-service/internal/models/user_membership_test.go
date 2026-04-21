package models

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestUserMembershipDB_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	userMembershipDB := NewUserMembershipDB(db)

	membership := &UserMembership{
		UserID:        "user123",
		PlanCode:      "basic",
		OrderNo:       "ORDER202301010001",
		StartTime:     time.Now(),
		EndTime:       time.Now().AddDate(0, 0, 30),
		Status:        MembershipStatusActive,
		AutoRenew:     false,
		UsedQueries:   0,
		UsedDownloads: 0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Mock the database operation
	mock.ExpectQuery(`INSERT INTO user_memberships`).
		WithArgs(
			membership.UserID, membership.PlanCode, membership.OrderNo,
			membership.StartTime, membership.EndTime, membership.Status,
			membership.AutoRenew, membership.UsedQueries, membership.UsedDownloads,
			membership.CreatedAt, membership.UpdatedAt,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err = userMembershipDB.Create(membership)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), membership.ID)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserMembershipDB_GetByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	userMembershipDB := NewUserMembershipDB(db)

	userID := "user123"

	// Mock the database operation
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "plan_code", "order_no", "start_time", "end_time",
		"status", "auto_renew", "used_queries", "used_downloads", "created_at", "updated_at",
	}).AddRow(
		1, userID, "basic", "ORDER202301010001", time.Now(), time.Now().AddDate(0, 0, 30),
		MembershipStatusActive, false, 0, 0, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT id, user_id, plan_code, order_no, start_time,
		       end_time, status, auto_renew, used_queries,
		       used_downloads, created_at, updated_at
		FROM user_memberships
		WHERE user_id = \$1
		ORDER BY end_time DESC
		LIMIT 1`).
		WithArgs(userID).
		WillReturnRows(rows)

	membership, err := userMembershipDB.GetByUserID(userID)
	assert.NoError(t, err)
	assert.NotNil(t, membership)
	assert.Equal(t, userID, membership.UserID)
	assert.Equal(t, "basic", membership.PlanCode)
	assert.Equal(t, MembershipStatusActive, membership.Status)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserMembershipDB_UpdateStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	userMembershipDB := NewUserMembershipDB(db)

	userID := "user123"
	newStatus := MembershipStatusExpired

	// Mock the database operation
	mock.ExpectExec(`UPDATE user_memberships SET status = \$1, updated_at = \$2 WHERE user_id = \$3`).
		WithArgs(newStatus, sqlmock.AnyArg(), userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = userMembershipDB.UpdateStatus(userID, newStatus)
	assert.NoError(t, err)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserMembershipDB_ExtendMembership(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	userMembershipDB := NewUserMembershipDB(db)

	userID := "user123"
	extensionDays := 30

	// Mock the database operation
	mock.ExpectExec(`UPDATE user_memberships SET end_time = end_time \+ INTERVAL '30 days', updated_at = \$1 WHERE user_id = \$2 AND status = \$3`).
		WithArgs(sqlmock.AnyArg(), userID, MembershipStatusActive).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = userMembershipDB.ExtendMembership(userID, extensionDays)
	assert.NoError(t, err)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}