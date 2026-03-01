-- migrations/v0/005_user_preferences.up.sql

ALTER TABLE users ADD COLUMN IF NOT EXISTS notification_frequency TEXT DEFAULT 'daily'; -- immediate, daily, weekly
ALTER TABLE users ADD COLUMN IF NOT EXISTS timezone TEXT DEFAULT 'UTC';
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_notified_at TIMESTAMPTZ DEFAULT '1970-01-01 00:00:00Z';

-- Create index for faster lookup during cron
CREATE INDEX IF NOT EXISTS idx_users_notif ON users (notification_frequency, timezone);
