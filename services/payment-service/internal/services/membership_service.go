package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/tvvshow/gokao/services/payment-service/internal/models"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// MembershipService 会员服务
type MembershipService struct {
	db    *sql.DB
	redis *redis.Client
}

// NewMembershipService 创建会员服务
func NewMembershipService(db *sql.DB, redis *redis.Client) *MembershipService {
	return &MembershipService{
		db:    db,
		redis: redis,
	}
}

// GetPlans 获取会员套餐列表
func (s *MembershipService) GetPlans(ctx context.Context) ([]*models.MembershipPlan, error) {
	query := `
		SELECT id, plan_code, name, description, price, duration_days,
		       features, max_queries, max_downloads, is_active,
		       created_at, updated_at
		FROM membership_plans
		WHERE is_active = true
		ORDER BY price ASC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query membership plans: %w", err)
	}
	defer rows.Close()

	var plans []*models.MembershipPlan
	for rows.Next() {
		plan := &models.MembershipPlan{}
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

// Subscribe 订阅会员
func (s *MembershipService) Subscribe(ctx context.Context, userID, orderNo string) error {
	// 获取订单信息
	order, err := s.getPaymentOrder(ctx, orderNo)
	if err != nil {
		return fmt.Errorf("failed to get payment order: %w", err)
	}

	// 检查订单是否已支付
	if order.Status != models.PaymentStatusPaid {
		return fmt.Errorf("order is not paid")
	}

	// 检查订单是否属于该用户
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}
	if order.UserID != userUUID {
		return fmt.Errorf("order does not belong to user")
	}

	// 获取套餐信息
	planCode, ok := order.Metadata["plan_code"].(string)
	if !ok {
		return fmt.Errorf("invalid plan code in order metadata")
	}

	plan, err := s.getMembershipPlan(ctx, planCode)
	if err != nil {
		return fmt.Errorf("failed to get membership plan: %w", err)
	}

	// 检查是否已经有有效会员
	currentMembership, err := s.GetMembershipStatus(ctx, userID)
	if err == nil && currentMembership.IsVIP {
		// 如果已有会员，则延长会员期限
		return s.extendMembership(ctx, userID, plan, orderNo)
	}

	// 创建新会员
	startTime := time.Now()
	endTime := startTime.AddDate(0, 0, plan.DurationDays)

	autoRenew := false
	if renew, ok := order.Metadata["auto_renew"].(bool); ok {
		autoRenew = renew
	}

	membership := &models.UserMembership{
		UserID:        userID,
		PlanCode:      planCode,
		OrderNo:       orderNo,
		StartTime:     startTime,
		EndTime:       endTime,
		Status:        models.MembershipStatusActive,
		AutoRenew:     autoRenew,
		UsedQueries:   0,
		UsedDownloads: 0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	return s.saveMembership(ctx, membership)
}

// GetMembershipStatus 获取会员状态
func (s *MembershipService) GetMembershipStatus(ctx context.Context, userID string) (*models.MembershipStatusResponse, error) {
	// 先从缓存获取
	cacheKey := fmt.Sprintf("membership:%s", userID)
	if cached, err := s.redis.Get(ctx, cacheKey).Result(); err == nil && cached != "" {
		// 解析缓存数据
		// 这里简化处理，实际应该序列化/反序列化
	}

	// 从数据库获取最新的会员信息
	query := `
		SELECT um.id, um.user_id, um.plan_code, um.order_no, um.start_time,
		       um.end_time, um.status, um.auto_renew, um.used_queries,
		       um.used_downloads, um.created_at, um.updated_at,
		       mp.name, mp.features, mp.max_queries, mp.max_downloads
		FROM user_memberships um
		JOIN membership_plans mp ON um.plan_code = mp.plan_code
		WHERE um.user_id = $1
		ORDER BY um.end_time DESC
		LIMIT 1
	`

	membership := &models.UserMembership{Plan: &models.MembershipPlan{}}
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&membership.ID, &membership.UserID, &membership.PlanCode,
		&membership.OrderNo, &membership.StartTime, &membership.EndTime,
		&membership.Status, &membership.AutoRenew, &membership.UsedQueries,
		&membership.UsedDownloads, &membership.CreatedAt, &membership.UpdatedAt,
		&membership.Plan.Name, &membership.Plan.Features,
		&membership.Plan.MaxQueries, &membership.Plan.MaxDownloads,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// 没有会员记录，返回非VIP状态
			return &models.MembershipStatusResponse{
				IsVIP:     false,
				Features:  make(map[string]interface{}),
				AutoRenew: false,
			}, nil
		}
		return nil, fmt.Errorf("failed to get membership: %w", err)
	}

	// 构造响应
	response := &models.MembershipStatusResponse{
		IsVIP:         membership.IsActive(),
		PlanCode:      membership.PlanCode,
		PlanName:      membership.Plan.Name,
		StartTime:     &membership.StartTime,
		EndTime:       &membership.EndTime,
		RemainingDays: membership.RemainingDays(),
		UsedQueries:   membership.UsedQueries,
		MaxQueries:    membership.Plan.MaxQueries,
		UsedDownloads: membership.UsedDownloads,
		MaxDownloads:  membership.Plan.MaxDownloads,
		Features:      membership.Plan.Features,
		AutoRenew:     membership.AutoRenew,
	}

	// 缓存结果（5分钟）
	s.cacheMembershipStatus(ctx, userID, response)

	return response, nil
}

// RenewMembership 续费会员
func (s *MembershipService) RenewMembership(ctx context.Context, userID, planCode string) (string, error) {
	// 检查当前会员状态
	currentStatus, err := s.GetMembershipStatus(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get current membership: %w", err)
	}

	// 获取套餐信息
	plan, err := s.getMembershipPlan(ctx, planCode)
	if err != nil {
		return "", fmt.Errorf("failed to get membership plan: %w", err)
	}

	// 生成续费订单号
	orderNo := fmt.Sprintf("RN%d", time.Now().Unix())

	// 计算续费开始时间
	var startTime time.Time
	if currentStatus.IsVIP && currentStatus.EndTime != nil {
		// 如果当前还是VIP，从当前到期时间开始
		startTime = *currentStatus.EndTime
	} else {
		// 如果已过期，从现在开始
		startTime = time.Now()
	}

	endTime := startTime.AddDate(0, 0, plan.DurationDays)

	// 创建续费会员记录
	membership := &models.UserMembership{
		UserID:        userID,
		PlanCode:      planCode,
		OrderNo:       orderNo,
		StartTime:     startTime,
		EndTime:       endTime,
		Status:        models.MembershipStatusActive,
		AutoRenew:     currentStatus.AutoRenew,
		UsedQueries:   0,
		UsedDownloads: 0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.saveMembership(ctx, membership); err != nil {
		return "", fmt.Errorf("failed to save renewal membership: %w", err)
	}

	// 清除缓存
	s.clearMembershipCache(ctx, userID)

	return orderNo, nil
}

// CancelMembership 取消会员
func (s *MembershipService) CancelMembership(ctx context.Context, userID string) error {
	// 获取当前会员信息
	currentStatus, err := s.GetMembershipStatus(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get current membership: %w", err)
	}

	if !currentStatus.IsVIP {
		return fmt.Errorf("user is not a VIP member")
	}

	// 更新会员状态为取消
	query := `
		UPDATE user_memberships 
		SET status = $1, auto_renew = false, updated_at = $2
		WHERE user_id = $3 AND status = $4
	`

	_, err = s.db.ExecContext(ctx, query,
		models.MembershipStatusCanceled, time.Now(), userID, models.MembershipStatusActive,
	)

	if err != nil {
		return fmt.Errorf("failed to cancel membership: %w", err)
	}

	// 清除缓存
	s.clearMembershipCache(ctx, userID)

	return nil
}

// GetMemberBenefits 获取会员权益
func (s *MembershipService) GetMemberBenefits(ctx context.Context, userID string) (map[string]interface{}, error) {
	status, err := s.GetMembershipStatus(ctx, userID)
	if err != nil {
		return nil, err
	}

	benefits := make(map[string]interface{})

	if status.IsVIP {
		benefits["features"] = status.Features
		benefits["query_limit"] = map[string]interface{}{
			"used":      status.UsedQueries,
			"max":       status.MaxQueries,
			"unlimited": status.MaxQueries == -1,
		}
		benefits["download_limit"] = map[string]interface{}{
			"used":      status.UsedDownloads,
			"max":       status.MaxDownloads,
			"unlimited": status.MaxDownloads == -1,
		}
		benefits["expire_info"] = map[string]interface{}{
			"end_time":       status.EndTime,
			"remaining_days": status.RemainingDays,
			"auto_renew":     status.AutoRenew,
		}
	} else {
		benefits["features"] = map[string]interface{}{
			"basic_query": true,
		}
		benefits["query_limit"] = map[string]interface{}{
			"used":      0,
			"max":       10, // 非VIP用户限制
			"unlimited": false,
		}
		benefits["download_limit"] = map[string]interface{}{
			"used":      0,
			"max":       0,
			"unlimited": false,
		}
	}

	return benefits, nil
}

// ConsumeQuery 消费查询次数
func (s *MembershipService) ConsumeQuery(ctx context.Context, userID string) error {
	status, err := s.GetMembershipStatus(ctx, userID)
	if err != nil {
		return err
	}

	if !status.IsVIP {
		return fmt.Errorf("VIP membership required")
	}

	if status.MaxQueries != -1 && status.UsedQueries >= status.MaxQueries {
		return fmt.Errorf("query limit exceeded")
	}

	// 更新使用次数
	query := `
		UPDATE user_memberships
		SET used_queries = used_queries + 1, updated_at = $1
		WHERE user_id = $2 AND status = $3
	`

	_, err = s.db.ExecContext(ctx, query, time.Now(), userID, models.MembershipStatusActive)
	if err != nil {
		return fmt.Errorf("failed to consume query: %w", err)
	}

	// 清除缓存
	s.clearMembershipCache(ctx, userID)

	return nil
}

// CheckMembershipPermission 检查会员权限
func (s *MembershipService) CheckMembershipPermission(ctx context.Context, userID string, feature string) (bool, error) {
	status, err := s.GetMembershipStatus(ctx, userID)
	if err != nil {
		return false, err
	}

	// 非VIP用户只能使用基础功能
	if !status.IsVIP {
		basicFeatures := map[string]bool{
			"basic_query":      true,
			"basic_university": true,
		}
		return basicFeatures[feature], nil
	}

	// VIP用户检查具体功能权限
	features := status.Features
	if enabled, exists := features[feature]; exists {
		if enabledBool, ok := enabled.(bool); ok {
			return enabledBool, nil
		}
	}

	return false, nil
}

// UpdateAutoRenew 更新自动续费设置
func (s *MembershipService) UpdateAutoRenew(ctx context.Context, userID string, autoRenew bool) error {
	query := `
		UPDATE user_memberships
		SET auto_renew = $1, updated_at = $2
		WHERE user_id = $3 AND status = $4
	`

	result, err := s.db.ExecContext(ctx, query, autoRenew, time.Now(), userID, models.MembershipStatusActive)
	if err != nil {
		return fmt.Errorf("failed to update auto renew: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no active membership found for user")
	}

	// 清除缓存
	s.clearMembershipCache(ctx, userID)

	return nil
}

// ConsumeDownload 消费下载次数
func (s *MembershipService) ConsumeDownload(ctx context.Context, userID string) error {
	status, err := s.GetMembershipStatus(ctx, userID)
	if err != nil {
		return err
	}

	if !status.IsVIP {
		return fmt.Errorf("VIP membership required")
	}

	if status.MaxDownloads != -1 && status.UsedDownloads >= status.MaxDownloads {
		return fmt.Errorf("download limit exceeded")
	}

	// 更新使用次数
	query := `
		UPDATE user_memberships 
		SET used_downloads = used_downloads + 1, updated_at = $1
		WHERE user_id = $2 AND status = $3
	`

	_, err = s.db.ExecContext(ctx, query, time.Now(), userID, models.MembershipStatusActive)
	if err != nil {
		return fmt.Errorf("failed to consume download: %w", err)
	}

	// 清除缓存
	s.clearMembershipCache(ctx, userID)

	return nil
}

// 私有方法

func (s *MembershipService) getPaymentOrder(ctx context.Context, orderNo string) (*models.PaymentOrder, error) {
	query := `
		SELECT id, order_no, user_id, amount, currency, subject, description,
		       status, payment_channel, channel_trade_no, client_ip, notify_url,
		       return_url, expire_time, paid_at, created_at, updated_at, metadata
		FROM payment_orders
		WHERE order_no = $1
	`

	order := &models.PaymentOrder{}
	err := s.db.QueryRowContext(ctx, query, orderNo).Scan(
		&order.ID, &order.OrderNo, &order.UserID, &order.Amount,
		&order.Currency, &order.Subject, &order.Description, &order.Status,
		&order.Channel, &order.ChannelTradeNo, &order.ClientIP,
		&order.NotifyURL, &order.ReturnURL, &order.ExpiredAt,
		&order.PaidAt, &order.CreatedAt, &order.UpdatedAt, &order.Metadata,
	)

	return order, err
}

func (s *MembershipService) getMembershipPlan(ctx context.Context, planCode string) (*models.MembershipPlan, error) {
	query := `
		SELECT id, plan_code, name, description, price, duration_days,
		       features, max_queries, max_downloads, is_active,
		       created_at, updated_at
		FROM membership_plans
		WHERE plan_code = $1
	`

	plan := &models.MembershipPlan{}
	err := s.db.QueryRowContext(ctx, query, planCode).Scan(
		&plan.ID, &plan.PlanCode, &plan.Name, &plan.Description,
		&plan.Price, &plan.DurationDays, &plan.Features,
		&plan.MaxQueries, &plan.MaxDownloads, &plan.IsActive,
		&plan.CreatedAt, &plan.UpdatedAt,
	)

	return plan, err
}

func (s *MembershipService) saveMembership(ctx context.Context, membership *models.UserMembership) error {
	query := `
		INSERT INTO user_memberships (
			user_id, plan_code, order_no, start_time, end_time,
			status, auto_renew, used_queries, used_downloads,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := s.db.ExecContext(ctx, query,
		membership.UserID, membership.PlanCode, membership.OrderNo,
		membership.StartTime, membership.EndTime, membership.Status,
		membership.AutoRenew, membership.UsedQueries, membership.UsedDownloads,
		membership.CreatedAt, membership.UpdatedAt,
	)

	return err
}

func (s *MembershipService) extendMembership(ctx context.Context, userID string, plan *models.MembershipPlan, orderNo string) error {
	// 获取当前会员信息
	query := `
		SELECT end_time 
		FROM user_memberships 
		WHERE user_id = $1 AND status = $2 
		ORDER BY end_time DESC 
		LIMIT 1
	`

	var currentEndTime time.Time
	err := s.db.QueryRowContext(ctx, query, userID, models.MembershipStatusActive).Scan(&currentEndTime)
	if err != nil {
		return err
	}

	// 延长会员期限
	newEndTime := currentEndTime.AddDate(0, 0, plan.DurationDays)

	updateQuery := `
		UPDATE user_memberships 
		SET end_time = $1, updated_at = $2
		WHERE user_id = $3 AND status = $4
	`

	_, err = s.db.ExecContext(ctx, updateQuery, newEndTime, time.Now(), userID, models.MembershipStatusActive)
	return err
}

func (s *MembershipService) cacheMembershipStatus(ctx context.Context, userID string, status *models.MembershipStatusResponse) {
	// 实际实现中应该序列化status并缓存
	cacheKey := fmt.Sprintf("membership:%s", userID)
	s.redis.Set(ctx, cacheKey, "cached", 5*time.Minute)
}

func (s *MembershipService) clearMembershipCache(ctx context.Context, userID string) {
	cacheKey := fmt.Sprintf("membership:%s", userID)
	s.redis.Del(ctx, cacheKey)
}
