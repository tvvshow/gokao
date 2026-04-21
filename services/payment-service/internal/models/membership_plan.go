package models

import (
	"database/sql"
	"fmt"
	"time"
)

// MembershipPlanRepository 会员套餐仓库接口
type MembershipPlanRepository interface {
	Create(plan *MembershipPlan) error
	GetByPlanCode(planCode string) (*MembershipPlan, error)
	GetAllActive() ([]*MembershipPlan, error)
	Update(plan *MembershipPlan) error
	Delete(planCode string) error
}

// MembershipPlanDB 实现MembershipPlanRepository接口
type MembershipPlanDB struct {
	db *sql.DB
}

// NewMembershipPlanDB 创建MembershipPlanDB实例
func NewMembershipPlanDB(db *sql.DB) *MembershipPlanDB {
	return &MembershipPlanDB{db: db}
}

// Create 创建会员套餐
func (m *MembershipPlanDB) Create(plan *MembershipPlan) error {
	query := `
		INSERT INTO membership_plans (
			plan_code, name, description, price, duration_days, features,
			max_queries, max_downloads, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	return m.db.QueryRow(
		query,
		plan.PlanCode, plan.Name, plan.Description, plan.Price, plan.DurationDays,
		plan.Features, plan.MaxQueries, plan.MaxDownloads, plan.IsActive,
		plan.CreatedAt, plan.UpdatedAt,
	).Scan(&plan.ID)
}

// GetByPlanCode 根据套餐代码获取会员套餐
func (m *MembershipPlanDB) GetByPlanCode(planCode string) (*MembershipPlan, error) {
	query := `
		SELECT id, plan_code, name, description, price, duration_days,
		       features, max_queries, max_downloads, is_active,
		       created_at, updated_at
		FROM membership_plans
		WHERE plan_code = $1
	`

	plan := &MembershipPlan{}
	err := m.db.QueryRow(query, planCode).Scan(
		&plan.ID, &plan.PlanCode, &plan.Name, &plan.Description,
		&plan.Price, &plan.DurationDays, &plan.Features,
		&plan.MaxQueries, &plan.MaxDownloads, &plan.IsActive,
		&plan.CreatedAt, &plan.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("membership plan not found")
		}
		return nil, fmt.Errorf("failed to get membership plan: %w", err)
	}

	return plan, nil
}

// GetAllActive 获取所有激活的会员套餐
func (m *MembershipPlanDB) GetAllActive() ([]*MembershipPlan, error) {
	query := `
		SELECT id, plan_code, name, description, price, duration_days,
		       features, max_queries, max_downloads, is_active,
		       created_at, updated_at
		FROM membership_plans
		WHERE is_active = true
		ORDER BY price ASC
	`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query membership plans: %w", err)
	}
	defer rows.Close()

	var plans []*MembershipPlan
	for rows.Next() {
		plan := &MembershipPlan{}
		err := rows.Scan(
			&plan.ID, &plan.PlanCode, &plan.Name, &plan.Description,
			&plan.Price, &plan.DurationDays, &plan.Features,
			&plan.MaxQueries, &plan.MaxDownloads, &plan.IsActive,
			&plan.CreatedAt, &plan.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan membership plan: %w", err)
		}
		plans = append(plans, plan)
	}

	return plans, nil
}

// Update 更新会员套餐
func (m *MembershipPlanDB) Update(plan *MembershipPlan) error {
	query := `
		UPDATE membership_plans 
		SET name = $1, description = $2, price = $3, duration_days = $4,
		    features = $5, max_queries = $6, max_downloads = $7, is_active = $8,
		    updated_at = $9
		WHERE plan_code = $10
	`

	_, err := m.db.Exec(
		query,
		plan.Name, plan.Description, plan.Price, plan.DurationDays,
		plan.Features, plan.MaxQueries, plan.MaxDownloads, plan.IsActive,
		plan.UpdatedAt, plan.PlanCode,
	)

	return err
}

// Delete 删除会员套餐
func (m *MembershipPlanDB) Delete(planCode string) error {
	// 注意：在实际应用中，我们通常不会真正删除套餐，而是将其设置为非激活状态
	query := `UPDATE membership_plans SET is_active = false, updated_at = $1 WHERE plan_code = $2`
	_, err := m.db.Exec(query, time.Now(), planCode)
	return err
}
