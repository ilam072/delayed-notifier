CREATE TYPE notification_channel AS ENUM ('email', 'telegram');

CREATE TYPE notification_status AS ENUM ('scheduled', 'sent', 'canceled', 'failed');

CREATE TABLE IF NOT EXISTS notification (
    id UUID PRIMARY KEY,
    message TEXT NOT NULL,
    scheduled_at TIMESTAMP NOT NULL,
    channel notification_channel NOT NULL,
    status notification_status NOT NULL DEFAULT 'scheduled',
    recipient TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);