-- +migrate Down
-- 1. Re-create the separate indexes
CREATE INDEX idx_url_id ON url_analytics (url_id);
CREATE INDEX idx_click_time ON url_analytics (click_time);

-- 2. Drop the composite index
DROP INDEX idx_url_metrics ON url_analytics;
