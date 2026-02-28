-- 確保表存在
CREATE TABLE IF NOT EXISTS daily_cve_counts (
    day DATE PRIMARY KEY,
    count INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 刪除舊統計並重新根據 cves 表計算（全量校準）
TRUNCATE TABLE daily_cve_counts;

INSERT INTO daily_cve_counts (day, count)
SELECT published_at::date as day, count(*) as count
FROM cves
GROUP BY day
ON CONFLICT (day) DO UPDATE SET count = EXCLUDED.count;
