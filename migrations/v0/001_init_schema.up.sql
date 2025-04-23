CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE cve_sources(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(32) NOT NULL UNIQUE,
    type varchar(32) NOT NULL,
    url text,
    description text,
    enabled boolean DEFAULT TRUE,
    created_at timestamp DEFAULT now()
);

CREATE TABLE vendors(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(128) NOT NULL UNIQUE
);

CREATE TABLE products(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    vendor_id uuid NOT NULL REFERENCES vendors(id),
    name varchar(128) NOT NULL,
    UNIQUE (vendor_id, name)
);

CREATE TABLE cves(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
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

