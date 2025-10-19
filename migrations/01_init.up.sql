CREATE TYPE notification_channel AS ENUM ('email', 'telegram');

CREATE TYPE notificaton_status AS ENUM ('scheduled', 'sent', 'canceled', 'failed');

CREATE TABLE IF NOT EXISTS notification (
    id UUID PRIMARY KEY,
    message TEXT NOT NULL,
    scheduled_at TIMESTAMP NOT NULL,
    retries INT NOT NULL DEFAULT 0,
    channel notification_channel NOT NULL,
    recipient TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);