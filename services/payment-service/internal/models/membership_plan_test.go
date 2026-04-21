package models

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestMembershipPlanDB_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	membershipPlanDB := NewMembershipPlanDB(db)

	plan := &MembershipPlan{
		PlanCode:     "basic",
		Name:         "基础版",
		Description:  "基础功能套餐",
		Price:        decimal.NewFromFloat(29.90),
		DurationDays: 30,
		Features:     JSONB{"basic_query": true},
		MaxQueries:   100,
		MaxDownloads: 0,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Mock the database operation
	mock.ExpectQuery(`INSERT INTO membership_plans`).
		WithArgs(
			plan.PlanCode, plan.Name, plan.Description, plan.Price, plan.DurationDays,
			plan.Features, plan.MaxQueries, plan.MaxDownloads, plan.IsActive,
			plan.CreatedAt, plan.UpdatedAt,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err = membershipPlanDB.Create(plan)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), plan.ID)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMembershipPlanDB_GetByPlanCode(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	membershipPlanDB := NewMembershipPlanDB(db)

	planCode := "basic"

	// Mock the database operation
	rows := sqlmock.NewRows([]string{
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
		WillReturnRows(rows)

	plan, err := membershipPlanDB.GetByPlanCode(planCode)
	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, planCode, plan.PlanCode)
	assert.Equal(t, "基础版", plan.Name)
	assert.True(t, plan.IsActive)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMembershipPlanDB_GetAllActive(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	membershipPlanDB := NewMembershipPlanDB(db)

	// Mock the database operation
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

	plans, err := membershipPlanDB.GetAllActive()
	assert.NoError(t, err)
	assert.Len(t, plans, 2)
	assert.Equal(t, "basic", plans[0].PlanCode)
	assert.Equal(t, "premium", plans[1].PlanCode)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}