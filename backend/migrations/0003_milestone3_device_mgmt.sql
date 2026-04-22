-- PushIQ Milestone 3: Device Management & Token Refresh

-- Device presence tracking
CREATE TABLE IF NOT EXISTS device_presence (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id uuid NOT NULL UNIQUE REFERENCES devices(id) ON DELETE CASCADE,
    is_online BOOLEAN NOT NULL DEFAULT false,
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    last_online_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Token lifecycle and expiration
ALTER TABLE device_tokens ADD COLUMN IF NOT EXISTS expires_at timestamptz;
ALTER TABLE device_tokens ADD COLUMN IF NOT EXISTS is_valid BOOLEAN NOT NULL DEFAULT true;

-- Device activity audit log
CREATE TABLE IF NOT EXISTS device_activity_log (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id uuid NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    activity_type text NOT NULL,
    details jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Indexes for M3 features
CREATE INDEX IF NOT EXISTS idx_device_presence_is_online ON device_presence(is_online);
CREATE INDEX IF NOT EXISTS idx_device_presence_last_seen ON device_presence(last_seen_at);
CREATE INDEX IF NOT EXISTS idx_device_tokens_expires_at ON device_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_device_tokens_is_valid ON device_tokens(is_valid);
CREATE INDEX IF NOT EXISTS idx_activity_log_device_id ON device_activity_log(device_id);
CREATE INDEX IF NOT EXISTS idx_activity_log_created_at ON device_activity_log(created_at);
