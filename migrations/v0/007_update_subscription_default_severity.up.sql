-- Update subscription severity_threshold default from LOW (2) to MEDIUM (3)
ALTER TABLE subscriptions ALTER COLUMN severity_threshold SET DEFAULT 3;

-- Update existing subscriptions with severity_threshold = 2 (LOW) to 3 (MEDIUM)
-- Comment this line if you want to keep existing subscriptions unchanged
-- UPDATE subscriptions SET severity_threshold = 3 WHERE severity_threshold = 2;
