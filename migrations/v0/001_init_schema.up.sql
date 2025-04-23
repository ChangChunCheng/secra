CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE cve_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL,
    url TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE vendors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vendor_id UUID NOT NULL REFERENCES vendors(id),
    name TEXT NOT NULL,
    UNIQUE (vendor_id, name)
);

CREATE TABLE cves (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id UUID NOT NULL REFERENCES cve_sources(id),
    source_uid TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    severity TEXT,
    cvss_score FLOAT,
    status TEXT DEFAULT 'active',
    published_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE cve_products (
    cve_id UUID NOT NULL REFERENCES cves(id),
    product_id UUID NOT NULL REFERENCES products(id),
    PRIMARY KEY (cve_id, product_id)
);