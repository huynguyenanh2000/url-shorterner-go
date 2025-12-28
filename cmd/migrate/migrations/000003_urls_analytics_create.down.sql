-- +migrate Down
-- Reverse the changes by dropping the analytics table first
-- This must be done before dropping the 'url' table due to the Foreign Key constraint
DROP TABLE IF EXISTS url_analytics;
