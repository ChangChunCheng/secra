-- migrations/v0/006_add_notification_time.up.sql

-- Add notification time preference (e.g., "09:00", "14:30")
-- This allows users to specify when they want to receive daily/weekly notifications
ALTER TABLE users ADD COLUMN IF NOT EXISTS notification_time TEXT DEFAULT '09:00';

-- Add index for faster lookup during scheduled notification jobs
CREATE INDEX IF NOT EXISTS idx_users_notification_schedule ON users (notification_frequency, notification_time, timezone);
