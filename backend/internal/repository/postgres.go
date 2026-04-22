package repository

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pushiq/pushiq-backend/internal/model"
)

type Repository struct {
	db *sqlx.DB
}

func NewPostgresRepository(databaseURL string) (*Repository, error) {
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	return &Repository{db: db}, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) UpsertDevice(device *model.Device) error {
	now := time.Now().UTC()
	if device.ID == uuid.Nil {
		device.ID = uuid.New()
		device.CreatedAt = now
	}
	device.UpdatedAt = now

	_, err := r.db.Exec(`
        INSERT INTO devices (id, tenant_id, user_id, platform, app_version, locale, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (user_id, platform)
        DO UPDATE SET app_version = EXCLUDED.app_version, locale = EXCLUDED.locale, updated_at = EXCLUDED.updated_at
    `, device.ID, device.TenantID, device.UserID, device.Platform, device.AppVersion, device.Locale, device.CreatedAt, device.UpdatedAt)
	return err
}

func (r *Repository) UpsertDeviceToken(token *model.DeviceToken) error {
	now := time.Now().UTC()
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
		token.CreatedAt = now
	}
	token.UpdatedAt = now
	token.LastSeenAt = now

	_, err := r.db.Exec(`
        INSERT INTO device_tokens (id, device_id, token, provider, status, last_seen_at, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (device_id, provider)
        DO UPDATE SET token = EXCLUDED.token, status = EXCLUDED.status, last_seen_at = EXCLUDED.last_seen_at, updated_at = EXCLUDED.updated_at
    `, token.ID, token.DeviceID, token.Token, token.Provider, token.Status, token.LastSeenAt, token.CreatedAt, token.UpdatedAt)
	return err
}

func (r *Repository) GetActiveDeviceToken(deviceID uuid.UUID, provider string) (*model.DeviceToken, error) {
	var token model.DeviceToken
	err := r.db.Get(&token, `
        SELECT id, device_id, token, provider, status, last_seen_at, created_at, updated_at
        FROM device_tokens
        WHERE device_id = $1 AND provider = $2 AND status = 'active'
        LIMIT 1
    `, deviceID, provider)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *Repository) GetDeviceByID(deviceID uuid.UUID) (*model.Device, error) {
	var device model.Device
	err := r.db.Get(&device, `
        SELECT id, tenant_id, user_id, platform, app_version, locale, created_at, updated_at
        FROM devices
        WHERE id = $1
    `, deviceID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &device, err
}

func (r *Repository) GetDeviceByUserAndPlatform(userID string, platform string) (*model.Device, error) {
	var device model.Device
	err := r.db.Get(&device, `
        SELECT id, tenant_id, user_id, platform, app_version, locale, created_at, updated_at
        FROM devices
        WHERE user_id = $1 AND platform = $2
        LIMIT 1
    `, userID, platform)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &device, err
}

func (r *Repository) GetDeviceTokenByToken(token string) (*model.DeviceToken, error) {
	var deviceToken model.DeviceToken
	err := r.db.Get(&deviceToken, `
        SELECT id, device_id, token, provider, status, last_seen_at, created_at, updated_at
        FROM device_tokens
        WHERE token = $1
        LIMIT 1
    `, token)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &deviceToken, err
}

func (r *Repository) CreateNotification(notification *model.Notification) error {
	now := time.Now().UTC()
	if notification.ID == uuid.Nil {
		notification.ID = uuid.New()
		notification.CreatedAt = now
	}
	notification.UpdatedAt = now
	if notification.MaxRetries == 0 {
		notification.MaxRetries = 3
	}

	dataJSON, err := json.Marshal(notification.Data)
	if err != nil {
		return err
	}
	responseJSON, err := json.Marshal(notification.ProviderResponse)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`
        INSERT INTO notifications (
            id, tenant_id, device_id, platform, provider, title, body, data, status, provider_response, 
            created_at, updated_at, sent_at, delivered_at, attempt_count, max_retries, next_retry_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
    `, notification.ID, notification.TenantID, notification.DeviceID, notification.Platform, notification.Provider,
		notification.Title, notification.Body, dataJSON, notification.Status, responseJSON, notification.CreatedAt,
		notification.UpdatedAt, notification.SentAt, notification.DeliveredAt, notification.AttemptCount,
		notification.MaxRetries, notification.NextRetryAt)
	return err
}

func (r *Repository) UpdateNotificationStatus(notificationID uuid.UUID, status model.NotificationStatus, providerResponse map[string]any) error {
	responseJSON, err := json.Marshal(providerResponse)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`
        UPDATE notifications
        SET status = $1, provider_response = $2, sent_at = NOW(), updated_at = NOW()
        WHERE id = $3
    `, status, responseJSON, notificationID)
	return err
}

func (r *Repository) GetNotificationByID(notificationID uuid.UUID) (*model.Notification, error) {
	var notification model.Notification
	err := r.db.Get(&notification, `
        SELECT id, tenant_id, device_id, platform, provider, title, body, data, status, provider_response,
               created_at, updated_at, sent_at, delivered_at, attempt_count, max_retries, next_retry_at
        FROM notifications
        WHERE id = $1
    `, notificationID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &notification, err
}

func (r *Repository) RecordDeliveryAttempt(notificationID uuid.UUID, attemptNumber int, status string, providerError *string, providerResponse map[string]any) error {
	now := time.Now().UTC()
	responseJSON, err := json.Marshal(providerResponse)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`
        INSERT INTO delivery_attempts (id, notification_id, attempt_number, status, provider_error, provider_response, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `, uuid.New(), notificationID, attemptNumber, status, providerError, responseJSON, now, now)
	return err
}

func (r *Repository) UpdateNotificationRetry(notificationID uuid.UUID, attemptCount int, nextRetryAt time.Time) error {
	_, err := r.db.Exec(`
        UPDATE notifications
        SET attempt_count = $1, next_retry_at = $2, updated_at = NOW()
        WHERE id = $3
    `, attemptCount, nextRetryAt, notificationID)
	return err
}

func (r *Repository) GetPendingRetries(limit int) ([]model.Notification, error) {
	var notifications []model.Notification
	err := r.db.Select(&notifications, `
        SELECT id, tenant_id, device_id, platform, provider, title, body, data, status, provider_response,
               created_at, updated_at, sent_at, delivered_at, attempt_count, max_retries, next_retry_at
        FROM notifications
        WHERE status = 'failed' AND attempt_count < max_retries AND next_retry_at IS NOT NULL AND next_retry_at <= NOW()
        ORDER BY next_retry_at ASC
        LIMIT $1
    `, limit)
	return notifications, err
}

func (r *Repository) RecordWebhookEvent(notificationID uuid.UUID, provider string, eventType string, msgID *string, data map[string]any) error {
	webhookJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`
        INSERT INTO webhook_events (id, notification_id, provider, event_type, provider_message_id, webhook_data, processed, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `, uuid.New(), notificationID, provider, eventType, msgID, webhookJSON, false, time.Now().UTC())
	return err
}

func (r *Repository) MarkNotificationDelivered(notificationID uuid.UUID) error {
	_, err := r.db.Exec(`
        UPDATE notifications
        SET status = $1, delivered_at = NOW(), updated_at = NOW()
        WHERE id = $2
    `, model.NotificationStatusDelivered, notificationID)
	return err
}

// Milestone 3: Device Management & Token Refresh

func (r *Repository) UpsertDevicePresence(deviceID *uuid.UUID, isOnline bool, lastOnlineAt *time.Time) error {
	_, err := r.db.Exec(`
        INSERT INTO device_presence (id, device_id, is_online, last_seen_at, last_online_at, created_at, updated_at)
        VALUES ($1, $2, $3, NOW(), $4, NOW(), NOW())
        ON CONFLICT (device_id)
        DO UPDATE SET is_online = EXCLUDED.is_online, last_seen_at = NOW(), last_online_at = COALESCE(EXCLUDED.last_online_at, device_presence.last_online_at), updated_at = NOW()
    `, uuid.New(), deviceID, isOnline, lastOnlineAt)
	return err
}

func (r *Repository) GetDevicePresence(deviceID uuid.UUID) (*model.DevicePresence, error) {
	var presence model.DevicePresence
	err := r.db.Get(&presence, `
        SELECT id, device_id, is_online, last_seen_at, last_online_at, created_at, updated_at
        FROM device_presence
        WHERE device_id = $1
    `, deviceID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &presence, err
}

func (r *Repository) InvalidateExpiredTokens(limit int) (int64, error) {
	result, err := r.db.Exec(`
        UPDATE device_tokens
        SET is_valid = false, updated_at = NOW()
        WHERE expires_at IS NOT NULL AND expires_at <= NOW() AND is_valid = true
    `)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *Repository) CleanupStalePresence(staleThreshold time.Time) (int64, error) {
	result, err := r.db.Exec(`
        DELETE FROM device_presence
        WHERE last_seen_at < $1
    `, staleThreshold)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *Repository) MarkInactiveDevicesOffline(inactiveThreshold time.Time) (int64, error) {
	result, err := r.db.Exec(`
        UPDATE device_presence
        SET is_online = false, updated_at = NOW()
        WHERE last_seen_at < $1 AND is_online = true
    `, inactiveThreshold)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *Repository) UpdateTokenExpiration(tokenID uuid.UUID, expiresAt time.Time) error {
	_, err := r.db.Exec(`
        UPDATE device_tokens
        SET expires_at = $1, updated_at = NOW()
        WHERE id = $2
    `, expiresAt, tokenID)
	return err
}

func (r *Repository) LogDeviceActivity(activity model.DeviceActivity) error {
	detailsJSON, err := json.Marshal(activity.Details)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`
        INSERT INTO device_activity_log (id, device_id, activity_type, details, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `, activity.ID, activity.DeviceID, activity.ActivityType, detailsJSON, activity.CreatedAt)
	return err
}

func (r *Repository) GetDeviceActivityHistory(deviceID uuid.UUID, limit int) ([]model.DeviceActivity, error) {
	var activities []model.DeviceActivity
	err := r.db.Select(&activities, `
        SELECT id, device_id, activity_type, details, created_at
        FROM device_activity_log
        WHERE device_id = $1
        ORDER BY created_at DESC
        LIMIT $2
    `, deviceID, limit)
	return activities, err
}

func (r *Repository) ListDevicesWithPresence(appID uuid.UUID, onlineOnly bool) ([]model.DeviceWithPresence, error) {
	var devices []model.DeviceWithPresence
	query := `
        SELECT 
            d.id, d.tenant_id, d.user_id, d.platform, d.app_version, d.locale, d.created_at, d.updated_at,
            COALESCE(dp.is_online, false) as is_online, COALESCE(dp.last_seen_at, d.created_at) as last_seen_at,
            dp.last_online_at, dt.token as active_token, dt.expires_at as token_expires_at
        FROM devices d
        LEFT JOIN device_presence dp ON d.id = dp.device_id
        LEFT JOIN device_tokens dt ON d.id = dt.device_id AND dt.is_valid = true
        WHERE d.tenant_id = $1
    `
	args := []interface{}{appID}

	if onlineOnly {
		query += ` AND dp.is_online = true`
	}

	query += ` ORDER BY dp.last_seen_at DESC`
	err := r.db.Select(&devices, query, args...)
	return devices, err
}

// Milestone 4: Analytics & React Dashboard

func (r *Repository) GetDeliveryMetrics(tenantID string) (*model.DeliveryMetrics, error) {
	var metrics model.DeliveryMetrics

	err := r.db.Get(&metrics, `
        SELECT 
            COUNT(*) FILTER (WHERE status != 'pending') as total_sent,
            COUNT(*) FILTER (WHERE status = 'delivered') as total_delivered,
            COUNT(*) FILTER (WHERE status = 'failed') as total_failed
        FROM notifications
        WHERE tenant_id = $1
    `, tenantID)

	if err != nil {
		return nil, err
	}

	if metrics.TotalSent > 0 {
		metrics.DeliveryRate = float64(metrics.TotalDelivered) / float64(metrics.TotalSent)
		metrics.FailureRate = float64(metrics.TotalFailed) / float64(metrics.TotalSent)
	}

	return &metrics, nil
}

func (r *Repository) GetDeliveryFunnel(tenantID string) (*model.DeliveryFunnel, error) {
	var funnel model.DeliveryFunnel

	err := r.db.Get(&funnel, `
        SELECT 
            COUNT(DISTINCT d.id) as registered_devices,
            COUNT(DISTINCT CASE WHEN dp.is_online THEN d.id END) as online_devices,
            COUNT(n.id) FILTER (WHERE n.status != 'pending') as notifications_sent,
            COUNT(n.id) FILTER (WHERE n.status = 'delivered') as delivered,
            COUNT(n.id) FILTER (WHERE n.status = 'failed') as failed,
            COALESCE(AVG(EXTRACT(EPOCH FROM (n.delivered_at - n.created_at)))::text || 's', '0s') as avg_delivery_time
        FROM devices d
        LEFT JOIN device_presence dp ON d.id = dp.device_id
        LEFT JOIN notifications n ON d.id = n.device_id
        WHERE d.tenant_id = $1
    `, tenantID)

	if err != nil {
		return nil, err
	}

	if funnel.RegisteredDevices > 0 {
		funnel.OnlineRate = float64(funnel.OnlineDevices) / float64(funnel.RegisteredDevices)
	}

	return &funnel, nil
}

func (r *Repository) GetPlatformAnalytics(tenantID string) ([]model.PlatformAnalytics, error) {
	var analytics []model.PlatformAnalytics

	err := r.db.Select(&analytics, `
        SELECT 
            d.platform,
            COUNT(DISTINCT d.id) as total_devices,
            COUNT(DISTINCT CASE WHEN dp.is_online THEN d.id END) as online_devices,
            COUNT(DISTINCT CASE WHEN dt.is_valid THEN dt.id END) as tokens_valid,
            COUNT(DISTINCT CASE WHEN NOT dt.is_valid THEN dt.id END) as tokens_expired,
            COUNT(n.id) FILTER (WHERE n.status != 'pending') as notifications_sent,
            CASE WHEN COUNT(n.id) FILTER (WHERE n.status != 'pending') > 0
                THEN CAST(COUNT(n.id) FILTER (WHERE n.status = 'delivered') AS FLOAT) / COUNT(n.id) FILTER (WHERE n.status != 'pending')
                ELSE 0
            END as delivery_rate
        FROM devices d
        LEFT JOIN device_presence dp ON d.id = dp.device_id
        LEFT JOIN device_tokens dt ON d.id = dt.device_id
        LEFT JOIN notifications n ON d.id = n.device_id
        WHERE d.tenant_id = $1
        GROUP BY d.platform
    `, tenantID)

	return analytics, err
}

func (r *Repository) GetRetryAnalytics(tenantID string) (*model.RetryAnalytics, error) {
	var analytics model.RetryAnalytics

	err := r.db.Get(&analytics, `
        SELECT 
            COUNT(*) FILTER (WHERE attempt_count > 0) as total_retried,
            COUNT(*) FILTER (WHERE attempt_count > 0 AND status = 'delivered') as retry_success,
            COUNT(*) FILTER (WHERE attempt_count > 0 AND status = 'failed') as retry_failure,
            CASE WHEN COUNT(*) FILTER (WHERE attempt_count > 0) > 0
                THEN CAST(COUNT(*) FILTER (WHERE attempt_count > 0 AND status = 'delivered') AS FLOAT) / COUNT(*) FILTER (WHERE attempt_count > 0)
                ELSE 0
            END as retry_success_rate,
            COALESCE(AVG(attempt_count) FILTER (WHERE attempt_count > 0), 0) as avg_retry_attempts
        FROM notifications
        WHERE tenant_id = $1
    `, tenantID)

	return &analytics, err
}

func (r *Repository) GetTopNotifications(tenantID string, limit int) ([]model.NotificationStats, error) {
	var notifications []model.NotificationStats

	err := r.db.Select(&notifications, `
        SELECT 
            n.id,
            n.title,
            1 as devices_targeted,
            COUNT(n.id) FILTER (WHERE n.status != 'pending') as sent,
            COUNT(n.id) FILTER (WHERE n.status = 'delivered') as delivered,
            COUNT(n.id) FILTER (WHERE n.status = 'failed') as failed,
            CASE WHEN COUNT(n.id) FILTER (WHERE n.status != 'pending') > 0
                THEN CAST(COUNT(n.id) FILTER (WHERE n.status = 'delivered') AS FLOAT) / COUNT(n.id) FILTER (WHERE n.status != 'pending')
                ELSE 0
            END as delivery_rate,
            n.created_at,
            n.sent_at
        FROM notifications n
        WHERE n.tenant_id = $1
        GROUP BY n.id, n.title, n.created_at, n.sent_at
        ORDER BY sent DESC
        LIMIT $2
    `, tenantID, limit)

	return notifications, err
}

func (r *Repository) GetHourlyTrends(tenantID string, hours int) ([]model.TimeSeriesData, error) {
	var trends []model.TimeSeriesData

	err := r.db.Select(&trends, `
        SELECT 
            DATE_TRUNC('hour', n.created_at) as timestamp,
            COUNT(n.id) as value
        FROM notifications n
        WHERE n.tenant_id = $1 AND n.created_at >= NOW() - INTERVAL '1 hour' * $2
        GROUP BY DATE_TRUNC('hour', n.created_at)
        ORDER BY timestamp ASC
    `, tenantID, hours)

	return trends, err
}

func (r *Repository) GetAnalyticsDashboard(tenantID string) (*model.AnalyticsDashboard, error) {
	dashboard := &model.AnalyticsDashboard{
		DateRange: "Last 30 days",
	}

	// Get delivery metrics
	metrics, err := r.GetDeliveryMetrics(tenantID)
	if err != nil {
		return nil, err
	}
	dashboard.Metrics = *metrics

	// Get platform analytics
	platformAnalytics, err := r.GetPlatformAnalytics(tenantID)
	if err != nil {
		return nil, err
	}
	dashboard.ByPlatform = platformAnalytics

	// Get funnel data
	funnel, err := r.GetDeliveryFunnel(tenantID)
	if err != nil {
		return nil, err
	}
	dashboard.FunnelData = *funnel

	// Get retry metrics
	retry, err := r.GetRetryAnalytics(tenantID)
	if err != nil {
		return nil, err
	}
	dashboard.RetryMetrics = *retry

	// Get top notifications
	topNotifications, err := r.GetTopNotifications(tenantID, 10)
	if err != nil {
		return nil, err
	}
	dashboard.TopNotifications = topNotifications

	// Get hourly trends (last 24 hours)
	trends, err := r.GetHourlyTrends(tenantID, 24)
	if err != nil {
		return nil, err
	}
	dashboard.HourlyTrends = trends

	return dashboard, nil
}
