-- +migrate Down
DROP INDEX idx_long_url_hash ON url;
ALTER TABLE url DROP COLUMN long_url_hash;
