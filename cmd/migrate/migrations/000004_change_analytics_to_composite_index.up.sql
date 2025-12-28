-- +migrate Up
-- 1. Create the new high-performance composite index
-- This index supports both (url_id) and (url_id, click_time) queries
CREATE INDEX idx_url_metrics ON url_analytics (url_id, click_time);

-- 2. Drop the old redundant separate indexes to save disk space and speed up INSERTs
DROP INDEX idx_url_id ON url_analytics;
DROP INDEX idx_click_time ON url_analytics;
