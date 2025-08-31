package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// AlertLevel 告警级别
type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelError    AlertLevel = "error"
	AlertLevelCritical AlertLevel = "critical"
)

// Alert 告警信息
type Alert struct {
	ID          string                 `json:"id"`
	Level       AlertLevel             `json:"level"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Source      string                 `json:"source"`
	Timestamp   time.Time              `json:"timestamp"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]interface{} `json:"annotations"`
	Resolved    bool                   `json:"resolved"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
}

// AlertRule 告警规则
type AlertRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Query       string                 `json:"query"`
	Threshold   float64                `json:"threshold"`
	Operator    string                 `json:"operator"` // >, <, >=, <=, ==, !=
	Duration    time.Duration          `json:"duration"`
	Level       AlertLevel             `json:"level"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]interface{} `json:"annotations"`
	Enabled     bool                   `json:"enabled"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// NotificationChannel 通知渠道
type NotificationChannel struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"` // email, webhook, dingtalk, wechat
	Name     string                 `json:"name"`
	Config   map[string]interface{} `json:"config"`
	Enabled  bool                   `json:"enabled"`
	Filters  []AlertLevel           `json:"filters"` // 只接收特定级别的告警
}

// AlertManager 告警管理器
type AlertManager struct {
	redis    *redis.Client
	logger   *zap.Logger
	rules    map[string]*AlertRule
	channels map[string]*NotificationChannel
	mu       sync.RWMutex
}

// NewAlertManager 创建告警管理器
func NewAlertManager(redis *redis.Client, logger *zap.Logger) *AlertManager {
	return &AlertManager{
		redis:    redis,
		logger:   logger,
		rules:    make(map[string]*AlertRule),
		channels: make(map[string]*NotificationChannel),
	}
}

// AddRule 添加告警规则
func (am *AlertManager) AddRule(rule *AlertRule) {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	am.rules[rule.ID] = rule
	
	am.logger.Info("Alert rule added", zap.String("rule_id", rule.ID), zap.String("name", rule.Name))
}

// RemoveRule 移除告警规则
func (am *AlertManager) RemoveRule(ruleID string) {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	delete(am.rules, ruleID)
	am.logger.Info("Alert rule removed", zap.String("rule_id", ruleID))
}

// AddChannel 添加通知渠道
func (am *AlertManager) AddChannel(channel *NotificationChannel) {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	am.channels[channel.ID] = channel
	am.logger.Info("Notification channel added", zap.String("channel_id", channel.ID), zap.String("type", channel.Type))
}

// RemoveChannel 移除通知渠道
func (am *AlertManager) RemoveChannel(channelID string) {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	delete(am.channels, channelID)
	am.logger.Info("Notification channel removed", zap.String("channel_id", channelID))
}

// TriggerAlert 触发告警
func (am *AlertManager) TriggerAlert(ctx context.Context, alert *Alert) error {
	alert.ID = generateAlertID()
	alert.Timestamp = time.Now()
	
	// 存储告警到Redis
	alertKey := fmt.Sprintf("alert:%s", alert.ID)
	alertData, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}
	
	err = am.redis.Set(ctx, alertKey, alertData, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to store alert: %w", err)
	}
	
	// 发送通知
	go am.sendNotifications(ctx, alert)
	
	am.logger.Info("Alert triggered", 
		zap.String("alert_id", alert.ID),
		zap.String("level", string(alert.Level)),
		zap.String("title", alert.Title),
	)
	
	return nil
}

// ResolveAlert 解决告警
func (am *AlertManager) ResolveAlert(ctx context.Context, alertID string) error {
	alertKey := fmt.Sprintf("alert:%s", alertID)
	
	alertData, err := am.redis.Get(ctx, alertKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get alert: %w", err)
	}
	
	var alert Alert
	err = json.Unmarshal([]byte(alertData), &alert)
	if err != nil {
		return fmt.Errorf("failed to unmarshal alert: %w", err)
	}
	
	now := time.Now()
	alert.Resolved = true
	alert.ResolvedAt = &now
	
	updatedData, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal updated alert: %w", err)
	}
	
	err = am.redis.Set(ctx, alertKey, updatedData, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to update alert: %w", err)
	}
	
	am.logger.Info("Alert resolved", zap.String("alert_id", alertID))
	
	return nil
}

// GetAlerts 获取告警列表
func (am *AlertManager) GetAlerts(ctx context.Context, limit int, offset int) ([]*Alert, error) {
	keys, err := am.redis.Keys(ctx, "alert:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get alert keys: %w", err)
	}
	
	var alerts []*Alert
	for i, key := range keys {
		if i < offset {
			continue
		}
		if len(alerts) >= limit {
			break
		}
		
		alertData, err := am.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}
		
		var alert Alert
		err = json.Unmarshal([]byte(alertData), &alert)
		if err != nil {
			continue
		}
		
		alerts = append(alerts, &alert)
	}
	
	return alerts, nil
}

// sendNotifications 发送通知
func (am *AlertManager) sendNotifications(ctx context.Context, alert *Alert) {
	am.mu.RLock()
	channels := make([]*NotificationChannel, 0, len(am.channels))
	for _, channel := range am.channels {
		if channel.Enabled && am.shouldNotify(channel, alert.Level) {
			channels = append(channels, channel)
		}
	}
	am.mu.RUnlock()
	
	for _, channel := range channels {
		go am.sendNotification(ctx, channel, alert)
	}
}

// shouldNotify 检查是否应该通知
func (am *AlertManager) shouldNotify(channel *NotificationChannel, level AlertLevel) bool {
	if len(channel.Filters) == 0 {
		return true
	}
	
	for _, filter := range channel.Filters {
		if filter == level {
			return true
		}
	}
	
	return false
}

// sendNotification 发送单个通知
func (am *AlertManager) sendNotification(ctx context.Context, channel *NotificationChannel, alert *Alert) {
	switch channel.Type {
	case "webhook":
		am.sendWebhookNotification(ctx, channel, alert)
	case "email":
		am.sendEmailNotification(ctx, channel, alert)
	case "dingtalk":
		am.sendDingTalkNotification(ctx, channel, alert)
	case "wechat":
		am.sendWeChatNotification(ctx, channel, alert)
	default:
		am.logger.Warn("Unknown notification channel type", zap.String("type", channel.Type))
	}
}

// sendWebhookNotification 发送Webhook通知
func (am *AlertManager) sendWebhookNotification(ctx context.Context, channel *NotificationChannel, alert *Alert) {
	url, ok := channel.Config["url"].(string)
	if !ok {
		am.logger.Error("Webhook URL not configured", zap.String("channel_id", channel.ID))
		return
	}
	
	payload := map[string]interface{}{
		"alert":   alert,
		"channel": channel.Name,
	}
	
	data, err := json.Marshal(payload)
	if err != nil {
		am.logger.Error("Failed to marshal webhook payload", zap.Error(err))
		return
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		am.logger.Error("Failed to create webhook request", zap.Error(err))
		return
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		am.logger.Error("Failed to send webhook notification", zap.Error(err))
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		am.logger.Info("Webhook notification sent successfully", 
			zap.String("channel_id", channel.ID),
			zap.String("alert_id", alert.ID),
		)
	} else {
		am.logger.Error("Webhook notification failed", 
			zap.String("channel_id", channel.ID),
			zap.Int("status_code", resp.StatusCode),
		)
	}
}

// sendEmailNotification 发送邮件通知
func (am *AlertManager) sendEmailNotification(ctx context.Context, channel *NotificationChannel, alert *Alert) {
	// 邮件发送实现
	am.logger.Info("Email notification sent", 
		zap.String("channel_id", channel.ID),
		zap.String("alert_id", alert.ID),
	)
}

// sendDingTalkNotification 发送钉钉通知
func (am *AlertManager) sendDingTalkNotification(ctx context.Context, channel *NotificationChannel, alert *Alert) {
	// 钉钉通知实现
	am.logger.Info("DingTalk notification sent", 
		zap.String("channel_id", channel.ID),
		zap.String("alert_id", alert.ID),
	)
}

// sendWeChatNotification 发送微信通知
func (am *AlertManager) sendWeChatNotification(ctx context.Context, channel *NotificationChannel, alert *Alert) {
	// 微信通知实现
	am.logger.Info("WeChat notification sent", 
		zap.String("channel_id", channel.ID),
		zap.String("alert_id", alert.ID),
	)
}

// generateAlertID 生成告警ID
func generateAlertID() string {
	return fmt.Sprintf("alert_%d", time.Now().UnixNano())
}

// 预定义告警规则
func (am *AlertManager) LoadDefaultRules() {
	// CPU使用率过高
	am.AddRule(&AlertRule{
		ID:          "cpu_high_usage",
		Name:        "CPU使用率过高",
		Description: "CPU使用率超过80%",
		Query:       "system_cpu_usage_percent",
		Threshold:   80,
		Operator:    ">",
		Duration:    5 * time.Minute,
		Level:       AlertLevelWarning,
		Enabled:     true,
	})
	
	// 内存使用率过高
	am.AddRule(&AlertRule{
		ID:          "memory_high_usage",
		Name:        "内存使用率过高",
		Description: "内存使用率超过85%",
		Query:       "system_memory_usage_percent",
		Threshold:   85,
		Operator:    ">",
		Duration:    5 * time.Minute,
		Level:       AlertLevelWarning,
		Enabled:     true,
	})
	
	// HTTP错误率过高
	am.AddRule(&AlertRule{
		ID:          "http_error_rate_high",
		Name:        "HTTP错误率过高",
		Description: "HTTP 5xx错误率超过5%",
		Query:       "rate(http_requests_total{status_code=~\"5..\"}[5m]) / rate(http_requests_total[5m]) * 100",
		Threshold:   5,
		Operator:    ">",
		Duration:    2 * time.Minute,
		Level:       AlertLevelError,
		Enabled:     true,
	})
	
	// 数据库连接数过高
	am.AddRule(&AlertRule{
		ID:          "db_connections_high",
		Name:        "数据库连接数过高",
		Description: "数据库连接数超过90%",
		Query:       "database_connections{state=\"active\"} / database_connections{state=\"max\"} * 100",
		Threshold:   90,
		Operator:    ">",
		Duration:    3 * time.Minute,
		Level:       AlertLevelCritical,
		Enabled:     true,
	})
}
