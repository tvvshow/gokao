package alerts

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"github.com/gaokao/monitoring-service/internal/metrics"
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
	metrics  *metrics.MetricsCollector
	rules    map[string]*AlertRule
	channels map[string]*NotificationChannel
	mu       sync.RWMutex
}

// NewAlertManager 创建告警管理器
func NewAlertManager(redis *redis.Client, logger *zap.Logger, metrics *metrics.MetricsCollector) *AlertManager {
	return &AlertManager{
		redis:    redis,
		logger:   logger,
		metrics:  metrics,
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

// EvaluateRule 评估告警规则
func (am *AlertManager) EvaluateRule(ctx context.Context, rule *AlertRule) (bool, float64, error) {
	// 这里应该实现真实的指标查询逻辑
	// 为了演示,我们使用模拟的指标值
	metricValue, err := am.getMetricValue(ctx, rule.Query)
	if err != nil {
		return false, 0, err
	}

	triggered := am.evaluateCondition(metricValue, rule.Threshold, rule.Operator)
	return triggered, metricValue, nil
}

// getMetricValue 获取指标值
func (am *AlertManager) getMetricValue(ctx context.Context, query string) (float64, error) {
	// 根据查询语句获取相应的指标值
	switch query {
	case "system_cpu_usage_percent":
		// 获取CPU使用率
		cpuPercent, err := am.metrics.GetCPUUsage(ctx)
		if err != nil {
			return 0, fmt.Errorf("failed to get CPU usage: %w", err)
		}
		return cpuPercent, nil
	case "system_memory_usage_percent":
		// 获取内存使用率
		memoryPercent, err := am.metrics.GetMemoryUsage(ctx)
		if err != nil {
			return 0, fmt.Errorf("failed to get memory usage: %w", err)
		}
		return memoryPercent, nil
	case "database_connections{state=\"active\"} / database_connections{state=\"max\"} * 100":
		// 获取数据库连接使用率
		dbUsage, err := am.metrics.GetDatabaseConnectionUsage(ctx)
		if err != nil {
			return 0, fmt.Errorf("failed to get database connection usage: %w", err)
		}
		return dbUsage, nil
	default:
		// 对于其他查询,尝试解析为Prometheus指标
		value, err := am.metrics.GetMetricByQuery(ctx, query)
		if err != nil {
			return 0, fmt.Errorf("failed to get metric by query %s: %w", query, err)
		}
		return value, nil
	}
}

// evaluateCondition 评估条件
func (am *AlertManager) evaluateCondition(value, threshold float64, operator string) bool {
	switch operator {
	case ">":
		return value > threshold
	case ">=":
		return value >= threshold
	case "<":
		return value < threshold
	case "<=":
		return value <= threshold
	case "==":
		return value == threshold
	case "!=":
		return value != threshold
	default:
		return false
	}
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
	// 获取邮件配置
	smtpHost, ok := channel.Config["smtp_host"].(string)
	if !ok {
		am.logger.Error("SMTP host not configured", zap.String("channel_id", channel.ID))
		return
	}

	smtpPort, ok := channel.Config["smtp_port"].(float64)
	if !ok {
		am.logger.Error("SMTP port not configured", zap.String("channel_id", channel.ID))
		return
	}

	username, ok := channel.Config["username"].(string)
	if !ok {
		am.logger.Error("SMTP username not configured", zap.String("channel_id", channel.ID))
		return
	}

	password, ok := channel.Config["password"].(string)
	if !ok {
		am.logger.Error("SMTP password not configured", zap.String("channel_id", channel.ID))
		return
	}

	from, ok := channel.Config["from"].(string)
	if !ok {
		am.logger.Error("Email from address not configured", zap.String("channel_id", channel.ID))
		return
	}

	to, ok := channel.Config["to"].(string)
	if !ok {
		am.logger.Error("Email to address not configured", zap.String("channel_id", channel.ID))
		return
	}

	// 构建邮件内容
	subject := fmt.Sprintf("[%s] %s", alert.Level, alert.Title)
	body := fmt.Sprintf("Alert: %s\n\nMessage: %s\n\nSource: %s\n\nTime: %s\n\nAnnotations: %v",
		alert.Title, alert.Message, alert.Source, alert.Timestamp.Format("2006-01-02 15:04:05"), alert.Annotations)

	// 发送邮件
	err := am.sendEmail(smtpHost, int(smtpPort), username, password, from, to, subject, body)
	if err != nil {
		am.logger.Error("Failed to send email notification",
			zap.String("channel_id", channel.ID),
			zap.String("alert_id", alert.ID),
			zap.Error(err))
		return
	}

	am.logger.Info("Email notification sent successfully",
		zap.String("channel_id", channel.ID),
		zap.String("alert_id", alert.ID),
	)
}

// sendEmail 发送邮件
func (am *AlertManager) sendEmail(smtpHost string, smtpPort int, username, password, from, to, subject, body string) error {
	// 创建SMTP客户端
	addr := fmt.Sprintf("%s:%d", smtpHost, smtpPort)

	// 创建认证信息
	auth := smtp.PlainAuth("", username, password, smtpHost)

	// 构建邮件消息
	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		to, subject, body)

	// 发送邮件
	err := smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// sendDingTalkNotification 发送钉钉通知
func (am *AlertManager) sendDingTalkNotification(ctx context.Context, channel *NotificationChannel, alert *Alert) {
	// 获取钉钉配置
	webhookURL, ok := channel.Config["webhook_url"].(string)
	if !ok {
		am.logger.Error("DingTalk webhook URL not configured", zap.String("channel_id", channel.ID))
		return
	}

	secret, _ := channel.Config["secret"].(string) // 可选的签名密钥

	// 构建钉钉消息
	message := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": fmt.Sprintf("[%s] %s\n\n%s\n\nTime: %s",
				alert.Level, alert.Title, alert.Message, alert.Timestamp.Format("2006-01-02 15:04:05")),
		},
	}

	// 如果有签名密钥,添加签名
	if secret != "" {
		timestamp := time.Now().UnixNano() / 1e6
		sign := am.generateDingTalkSign(secret, timestamp)
		message["timestamp"] = timestamp
		message["sign"] = sign
	}

	// 发送钉钉通知
	err := am.sendWebhookRequest(webhookURL, message)
	if err != nil {
		am.logger.Error("Failed to send DingTalk notification",
			zap.String("channel_id", channel.ID),
			zap.String("alert_id", alert.ID),
			zap.Error(err))
		return
	}

	am.logger.Info("DingTalk notification sent successfully",
		zap.String("channel_id", channel.ID),
		zap.String("alert_id", alert.ID),
	)
}

// generateDingTalkSign 生成钉钉签名
func (am *AlertManager) generateDingTalkSign(secret string, timestamp int64) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)

	// 创建HMAC-SHA256哈希
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))

	// Base64编码
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// sendWeChatNotification 发送微信通知
func (am *AlertManager) sendWeChatNotification(ctx context.Context, channel *NotificationChannel, alert *Alert) {
	// 获取微信配置
	webhookURL, ok := channel.Config["webhook_url"].(string)
	if !ok {
		am.logger.Error("WeChat webhook URL not configured", zap.String("channel_id", channel.ID))
		return
	}

	// 构建微信消息
	message := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]interface{}{
			"content": fmt.Sprintf("[%s] %s\n\n%s\n\nTime: %s",
				alert.Level, alert.Title, alert.Message, alert.Timestamp.Format("2006-01-02 15:04:05")),
		},
	}

	// 发送微信通知
	err := am.sendWebhookRequest(webhookURL, message)
	if err != nil {
		am.logger.Error("Failed to send WeChat notification",
			zap.String("channel_id", channel.ID),
			zap.String("alert_id", alert.ID),
			zap.Error(err))
		return
	}

	am.logger.Info("WeChat notification sent successfully",
		zap.String("channel_id", channel.ID),
		zap.String("alert_id", alert.ID),
	)
}

// sendWebhookRequest 发送Webhook请求
func (am *AlertManager) sendWebhookRequest(url string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook request failed with status code: %d", resp.StatusCode)
	}

	return nil
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
