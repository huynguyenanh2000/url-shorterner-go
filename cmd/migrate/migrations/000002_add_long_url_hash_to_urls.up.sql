-- +migrate Up
ALTER TABLE url
ADD COLUMN long_url_hash CHAR(40) NOT NULL AFTER id;

CREATE INDEX idx_long_url_hash ON url(long_url_hash);
