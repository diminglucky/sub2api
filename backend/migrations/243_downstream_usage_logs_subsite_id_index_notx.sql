-- usage_logs 是最大的表，必须使用 CONCURRENTLY 建立子站归属索引。
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_usage_logs_subsite_id
    ON usage_logs(subsite_id)
    WHERE subsite_id IS NOT NULL;
