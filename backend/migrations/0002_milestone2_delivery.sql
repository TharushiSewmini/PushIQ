-- PushIQ Milestone 2: Delivery attempts, retry tracking, and webhooks

ALTER TABLE notifications ADD COLUMN IF NOT EXISTS attempt_count INT NOT NULL DEFAULT 0;
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS next_retry_at timestamptz;
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS max_retries INT NOT NULL DEFAULT 3;
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS delivered_at timestamptz;

CREATE TABLE IF NOT EXISTS delivery_attempts (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    notification_id uuid NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    attempt_number INT NOT NULL,
    status text NOT NULL,
    provider_error text,
    provider_response jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS webhook_events (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    notification_id uuid NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    provider text NOT NULL,
    event_type text NOT NULL,
    provider_message_id text,
    webhook_data jsonb,
    processed BOOLEAN NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_delivery_attempts_notification_id ON delivery_attempts(notification_id);
CREATE INDEX IF NOT EXISTS idx_notifications_next_retry ON notifications(next_retry_at);
CREATE INDEX IF NOT EXISTS idx_notifications_attempt_count ON notifications(attempt_count);
CREATE INDEX IF NOT EXISTS idx_webhook_events_notification_id ON webhook_events(notification_id);
CREATE INDEX IF NOT EXISTS idx_webhook_events_processed ON webhook_events(processed);
