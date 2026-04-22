-- PushIQ Phase 1 schema: devices, device tokens, notifications

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS devices (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id text,
    user_id text NOT NULL,
    platform text NOT NULL,
    app_version text,
    locale text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, platform)
);

CREATE TABLE IF NOT EXISTS device_tokens (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id uuid NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    token text NOT NULL,
    provider text NOT NULL,
    status text NOT NULL,
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (device_id, provider)
);

CREATE INDEX IF NOT EXISTS idx_device_tokens_token ON device_tokens(token);
CREATE INDEX IF NOT EXISTS idx_device_tokens_device_id ON device_tokens(device_id);

CREATE TABLE IF NOT EXISTS notifications (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id text,
    device_id uuid REFERENCES devices(id) ON DELETE SET NULL,
    platform text NOT NULL,
    provider text NOT NULL,
    title text NOT NULL,
    body text NOT NULL,
    data jsonb,
    status text NOT NULL,
    provider_response jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    sent_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_notifications_device_id ON notifications(device_id);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);
