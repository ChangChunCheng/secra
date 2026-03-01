-- 啟用 UUID 擴展與加密擴展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 定義 Secra Namespace UUID (隨機生成一個作為種子)
-- 這裡我們固定使用一個 Namespace，確保跨環境一致性
-- Namespace: 6ba7b810-9dad-11d1-80b4-00c04fd430c8 (這是標準 DNS namespace，或者我們可以自定)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'secra_ns') THEN
        -- 使用自定義的 Namespace UUID
        PERFORM uuid_generate_v5('6ba7b810-9dad-11d1-80b4-00c04fd430c8', 'secra.io');
    END IF;
END $$;

-- 輔助函數：簡化 UUID v5 生成
CREATE OR REPLACE FUNCTION secra_uuid_v5(name text) RETURNS uuid AS $$
BEGIN
    -- 基於 secra.io namespace 生成具名 UUID
    RETURN uuid_generate_v5('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', name);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE TABLE cve_sources(
    id uuid PRIMARY KEY, -- 由應用程式或 Trigger 決定
    name varchar(32) NOT NULL UNIQUE,
    type varchar(32) NOT NULL,
    url text,
    description text,
    enabled boolean DEFAULT TRUE,
    created_at timestamp DEFAULT now()
);

CREATE TABLE vendors(
    id uuid PRIMARY KEY,
    name varchar(128) NOT NULL UNIQUE
);

CREATE TABLE products(
    id uuid PRIMARY KEY,
    vendor_id uuid NOT NULL REFERENCES vendors(id),
    name varchar(128) NOT NULL,
    UNIQUE (vendor_id, name)
);

CREATE TABLE cves(
    id uuid PRIMARY KEY,
    source_id uuid NOT NULL REFERENCES cve_sources(id),
    source_uid varchar(16) NOT NULL UNIQUE,
    title varchar(256) NOT NULL,
    description text NOT NULL,
    severity varchar(8),
    cvss_score float,
    status varchar(16) DEFAULT 'active',
    published_at timestamp,
    updated_at timestamp
);

CREATE TABLE cve_products(
    cve_id uuid NOT NULL REFERENCES cves(id),
    product_id uuid NOT NULL REFERENCES products(id),
    PRIMARY KEY (cve_id, product_id)
);

CREATE TABLE IF NOT EXISTS cve_references (
    id UUID PRIMARY KEY,
    cve_id UUID NOT NULL REFERENCES cves(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    source TEXT,
    tags TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    UNIQUE (cve_id, url) 
);

CREATE TABLE IF NOT EXISTS cve_weaknesses (
    id UUID PRIMARY KEY,
    cve_id UUID NOT NULL REFERENCES cves(id) ON DELETE CASCADE,
    weakness TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    UNIQUE (cve_id, weakness) 
);
