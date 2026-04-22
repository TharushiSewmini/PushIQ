package model

import (
    "database/sql"
    "time"

    "github.com/google/uuid"
)

type Platform string

const (
    PlatformAndroid Platform = "android"
    PlatformiOS     Platform = "ios"
)

type TokenStatus string

const (
    TokenStatusActive TokenStatus = "active"
    TokenStatusStale  TokenStatus = "stale"
)

type DeliveryProvider string

const (
    ProviderFCM  DeliveryProvider = "fcm"
    ProviderAPNS DeliveryProvider = "apns"
)

type Device struct {
    ID         uuid.UUID      `db:"id" json:"id"`
    TenantID   sql.NullString `db:"tenant_id" json:"tenant_id,omitempty"`
    UserID     string         `db:"user_id" json:"user_id"`
    Platform   Platform       `db:"platform" json:"platform"`
    AppVersion string         `db:"app_version" json:"app_version,omitempty"`
    Locale     string         `db:"locale" json:"locale,omitempty"`
    CreatedAt  time.Time      `db:"created_at" json:"created_at"`
    UpdatedAt  time.Time      `db:"updated_at" json:"updated_at"`
}

type DeviceToken struct {
    ID         uuid.UUID   `db:"id" json:"id"`
    DeviceID   uuid.UUID   `db:"device_id" json:"device_id"`
    Token      string      `db:"token" json:"token"`
    Provider   string      `db:"provider" json:"provider"`
    Status     TokenStatus `db:"status" json:"status"`
    LastSeenAt time.Time   `db:"last_seen_at" json:"last_seen_at"`
    CreatedAt  time.Time   `db:"created_at" json:"created_at"`
    UpdatedAt  time.Time   `db:"updated_at" json:"updated_at"`
}

type NotificationStatus string

const (
    NotificationStatusPending   NotificationStatus = "pending"
    NotificationStatusSent      NotificationStatus = "sent"
    NotificationStatusFailed    NotificationStatus = "failed"
    NotificationStatusDelivered NotificationStatus = "delivered"
)

type Notification struct {
    ID               uuid.UUID          `db:"id" json:"id"`
    TenantID         sql.NullString     `db:"tenant_id" json:"tenant_id,omitempty"`
    DeviceID         uuid.UUID          `db:"device_id" json:"device_id"`
    Platform         Platform           `db:"platform" json:"platform"`
    Provider         DeliveryProvider   `db:"provider" json:"provider"`
    Title            string             `db:"title" json:"title"`
    Body             string             `db:"body" json:"body"`
    Data             map[string]any     `db:"data" json:"data,omitempty"`
    Status           NotificationStatus `db:"status" json:"status"`
    ProviderResponse map[string]any     `db:"provider_response" json:"provider_response,omitempty"`
    CreatedAt        time.Time          `db:"created_at" json:"created_at"`
    UpdatedAt        time.Time          `db:"updated_at" json:"updated_at"`
    SentAt           sql.NullTime       `db:"sent_at" json:"sent_at,omitempty"`
    DeliveredAt      sql.NullTime       `db:"delivered_at" json:"delivered_at,omitempty"`
    AttemptCount     int                `db:"attempt_count" json:"attempt_count"`
    MaxRetries       int                `db:"max_retries" json:"max_retries"`
    NextRetryAt      sql.NullTime       `db:"next_retry_at" json:"next_retry_at,omitempty"`
}

type DeliveryAttempt struct {
    ID               uuid.UUID      `db:"id" json:"id"`
    NotificationID   uuid.UUID      `db:"notification_id" json:"notification_id"`
    AttemptNumber    int            `db:"attempt_number" json:"attempt_number"`
    Status           string         `db:"status" json:"status"`
    ProviderError    sql.NullString `db:"provider_error" json:"provider_error,omitempty"`
    ProviderResponse map[string]any `db:"provider_response" json:"provider_response,omitempty"`
    CreatedAt        time.Time      `db:"created_at" json:"created_at"`
    UpdatedAt        time.Time      `db:"updated_at" json:"updated_at"`
}

type WebhookEvent struct {
    ID              uuid.UUID      `db:"id" json:"id"`
    NotificationID  uuid.UUID      `db:"notification_id" json:"notification_id"`
    Provider        string         `db:"provider" json:"provider"`
    EventType       string         `db:"event_type" json:"event_type"`
    ProviderMsgID   sql.NullString `db:"provider_message_id" json:"provider_message_id,omitempty"`
    WebhookData     map[string]any `db:"webhook_data" json:"webhook_data,omitempty"`
    Processed       bool           `db:"processed" json:"processed"`
    CreatedAt       time.Time      `db:"created_at" json:"created_at"`
}

// Milestone 3: Device Management & Token Refresh

type DevicePresence struct {
    ID           uuid.UUID  `db:"id" json:"id"`
    DeviceID     uuid.UUID  `db:"device_id" json:"device_id"`
    IsOnline     bool       `db:"is_online" json:"is_online"`
    LastSeenAt   time.Time  `db:"last_seen_at" json:"last_seen_at"`
    LastOnlineAt *time.Time `db:"last_online_at" json:"last_online_at,omitempty"`
    CreatedAt    time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

type DeviceActivity struct {
    ID           uuid.UUID              `db:"id" json:"id"`
    DeviceID     uuid.UUID              `db:"device_id" json:"device_id"`
    ActivityType string                 `db:"activity_type" json:"activity_type"`
    Details      map[string]interface{} `db:"details" json:"details,omitempty"`
    CreatedAt    time.Time              `db:"created_at" json:"created_at"`
}

type DeviceWithPresence struct {
    ID               uuid.UUID      `db:"id" json:"id"`
    TenantID         sql.NullString `db:"tenant_id" json:"tenant_id,omitempty"`
    UserID           string         `db:"user_id" json:"user_id"`
    Platform         Platform       `db:"platform" json:"platform"`
    AppVersion       string         `db:"app_version" json:"app_version,omitempty"`
    Locale           string         `db:"locale" json:"locale,omitempty"`
    CreatedAt        time.Time      `db:"created_at" json:"created_at"`
    UpdatedAt        time.Time      `db:"updated_at" json:"updated_at"`
    IsOnline         bool           `db:"is_online" json:"is_online"`
    LastSeenAt       time.Time      `db:"last_seen_at" json:"last_seen_at"`
    LastOnlineAt     *time.Time     `db:"last_online_at" json:"last_online_at,omitempty"`
    ActiveToken      string         `db:"active_token" json:"active_token,omitempty"`
    TokenExpiresAt   *time.Time     `db:"token_expires_at" json:"token_expires_at,omitempty"`
}

type TokenLifecycle struct {
    TokenID   uuid.UUID  `db:"token_id" json:"token_id"`
    DeviceID  uuid.UUID  `db:"device_id" json:"device_id"`
    ExpiresAt *time.Time `db:"expires_at" json:"expires_at,omitempty"`
    IsValid   bool       `db:"is_valid" json:"is_valid"`
    CreatedAt time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
}

// Milestone 4: Analytics & React Dashboard

type DeliveryMetrics struct {
    TotalSent      int64   `json:"total_sent"`
    TotalDelivered int64   `json:"total_delivered"`
    TotalFailed    int64   `json:"total_failed"`
    DeliveryRate   float64 `json:"delivery_rate"`
    FailureRate    float64 `json:"failure_rate"`
}

type NotificationStats struct {
    NotificationID uuid.UUID       `json:"notification_id"`
    Title          string          `json:"title"`
    DevicesTargeted int64          `json:"devices_targeted"`
    Sent           int64           `json:"sent"`
    Delivered      int64           `json:"delivered"`
    Failed         int64           `json:"failed"`
    DeliveryRate   float64         `json:"delivery_rate"`
    CreatedAt      time.Time       `json:"created_at"`
    SentAt         *time.Time      `json:"sent_at,omitempty"`
}

type DeliveryFunnel struct {
    RegisteredDevices int64   `json:"registered_devices"`
    OnlineDevices     int64   `json:"online_devices"`
    OnlineRate        float64 `json:"online_rate"`
    NotificationsSent int64   `json:"notifications_sent"`
    Delivered         int64   `json:"delivered"`
    Failed            int64   `json:"failed"`
    AvgDeliveryTime   string  `json:"avg_delivery_time"`
}

type PlatformAnalytics struct {
    Platform        Platform `json:"platform"`
    TotalDevices    int64    `json:"total_devices"`
    OnlineDevices   int64    `json:"online_devices"`
    TokensValid     int64    `json:"tokens_valid"`
    TokensExpired   int64    `json:"tokens_expired"`
    NotificationsSent int64  `json:"notifications_sent"`
    DeliveryRate    float64  `json:"delivery_rate"`
}

type RetryAnalytics struct {
    TotalRetried      int64   `json:"total_retried"`
    RetrySuccess      int64   `json:"retry_success"`
    RetryFailure      int64   `json:"retry_failure"`
    RetrySuccessRate  float64 `json:"retry_success_rate"`
    AvgRetryAttempts  float64 `json:"avg_retry_attempts"`
}

type TimeSeriesData struct {
    Timestamp time.Time `json:"timestamp"`
    Value     int64     `json:"value"`
}

type AnalyticsDashboard struct {
    DateRange     string             `json:"date_range"`
    Metrics       DeliveryMetrics    `json:"metrics"`
    ByPlatform    []PlatformAnalytics `json:"by_platform"`
    FunnelData    DeliveryFunnel     `json:"funnel_data"`
    RetryMetrics  RetryAnalytics     `json:"retry_metrics"`
    TopNotifications []NotificationStats `json:"top_notifications"`
    HourlyTrends  []TimeSeriesData   `json:"hourly_trends"`
}
