-- 001_add_user_features_and_system_registration.up.sql

-- 3. 訂閱 CVE 通報主表
CREATE TABLE subscriptions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  severity_threshold VARCHAR(20) DEFAULT 'LOW',
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 4. 訂閱目標細節表（多重目標支援）
-- 4. 定義訂閱目標類型表
CREATE TABLE target_types (
  id SERIAL PRIMARY KEY,
  name VARCHAR(50) NOT NULL UNIQUE
);

-- 預填預設目標類型
INSERT INTO target_types (name) VALUES ('cve_source'), ('vendor'), ('product');

-- 修改 subscription_targets 使用 FK 參照 target_types
CREATE TABLE subscription_targets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  subscription_id UUID NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
  target_type_id INT NOT NULL REFERENCES target_types(id),
  target_id       UUID NOT NULL
);

-- 5. 通知偏好設定
CREATE TABLE notification_preferences (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  channel VARCHAR(50) NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  config JSONB,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
