package models

import (
	"database/sql"
	"fmt"
	"time"
)

// UserMembershipRepository 用户会员仓库接口
type UserMembershipRepository interface {
	Create(membership *UserMembership) error
	GetByUserID(userID string) (*UserMembership, error)
	GetByOrderNo(orderNo string) (*UserMembership, error)
	Update(membership *UserMembership) error
	UpdateStatus(userID, status string) error
	ExtendMembership(userID string, extensionDays int) error
	Delete(userID string) error
}

// UserMembershipDB 实现UserMembershipRepository接口
type UserMembershipDB struct {
	db *sql.DB
}

// NewUserMembershipDB 创建UserMembershipDB实例
func NewUserMembershipDB(db *sql.DB) *UserMembershipDB {
	return &UserMembershipDB{db: db}
}

// Create 创建用户会员
func (u *UserMembershipDB) Create(membership *UserMembership) error {
	query := `
		INSERT INTO user_memberships (
			user_id, plan_code, order_no, start_time, end_time,
			status, auto_renew, used_queries, used_downloads,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	return u.db.QueryRow(
		query,
		membership.UserID, membership.PlanCode, membership.OrderNo,
		membership.StartTime, membership.EndTime, membership.Status,
		membership.AutoRenew, membership.UsedQueries, membership.UsedDownloads,
		membership.CreatedAt, membership.UpdatedAt,
	).Scan(&membership.ID)
}

// GetByUserID 根据用户ID获取用户会员信息
func (u *UserMembershipDB) GetByUserID(userID string) (*UserMembership, error) {
	query := `
		SELECT id, user_id, plan_code, order_no, start_time,
		       end_time, status, auto_renew, used_queries,
		       used_downloads, created_at, updated_at
		FROM user_memberships
		WHERE user_id = $1
		ORDER BY end_time DESC
		LIMIT 1
	`

	membership := &UserMembership{}
	err := u.db.QueryRow(query, userID).Scan(
		&membership.ID, &membership.UserID, &membership.PlanCode,
		&membership.OrderNo, &membership.StartTime, &membership.EndTime,
		&membership.Status, &membership.AutoRenew, &membership.UsedQueries,
		&membership.UsedDownloads, &membership.CreatedAt, &membership.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user membership not found")
		}
		return nil, fmt.Errorf("failed to get user membership: %w", err)
	}

	return membership, nil
}

// GetByOrderNo 根据订单号获取用户会员信息
func (u *UserMembershipDB) GetByOrderNo(orderNo string) (*UserMembership, error) {
	query := `
		SELECT id, user_id, plan_code, order_no, start_time,
		       end_time, status, auto_renew, used_queries,
		       used_downloads, created_at, updated_at
		FROM user_memberships
		WHERE order_no = $1
	`

	membership := &UserMembership{}
	err := u.db.QueryRow(query, orderNo).Scan(
		&membership.ID, &membership.UserID, &membership.PlanCode,
		&membership.OrderNo, &membership.StartTime, &membership.EndTime,
		&membership.Status, &membership.AutoRenew, &membership.UsedQueries,
		&membership.UsedDownloads, &membership.CreatedAt, &membership.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user membership not found")
		}
		return nil, fmt.Errorf("failed to get user membership: %w", err)
	}

	return membership, nil
}

// Update 更新用户会员信息
func (u *UserMembershipDB) Update(membership *UserMembership) error {
	query := `
		UPDATE user_memberships 
		SET plan_code = $1, start_time = $2, end_time = $3,
		    status = $4, auto_renew = $5, used_queries = $6,
		    used_downloads = $7, updated_at = $8
		WHERE user_id = $9
	`

	_, err := u.db.Exec(
		query,
		membership.PlanCode, membership.StartTime, membership.EndTime,
		membership.Status, membership.AutoRenew, membership.UsedQueries,
		membership.UsedDownloads, membership.UpdatedAt, membership.UserID,
	)

	return err
}

// UpdateStatus 更新用户会员状态
func (u *UserMembershipDB) UpdateStatus(userID, status string) error {
	query := `
		UPDATE user_memberships 
		SET status = $1, updated_at = $2
		WHERE user_id = $3
	`

	_, err := u.db.Exec(query, status, time.Now(), userID)
	return err
}

// ExtendMembership 延长用户会员期限
func (u *UserMembershipDB) ExtendMembership(userID string, extensionDays int) error {
	query := `
		UPDATE user_memberships 
		SET end_time = end_time + INTERVAL '%d days', updated_at = $1
		WHERE user_id = $2 AND status = $3
	`

	_, err := u.db.Exec(query, time.Now(), userID, MembershipStatusActive)
	return err
}

// Delete 删除用户会员信息
func (u *UserMembershipDB) Delete(userID string) error {
	query := `DELETE FROM user_memberships WHERE user_id = $1`
	_, err := u.db.Exec(query, userID)
	return err
}
