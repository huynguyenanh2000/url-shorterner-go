-- +migrate Up
-- Create the analytics table with a Foreign Key relationship to the url table
CREATE TABLE IF NOT EXISTS url_analytics (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    url_id BIGINT UNSIGNED NOT NULL,
    click_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT NOT NULL,
    referrer TEXT,

    PRIMARY KEY (id),

    -- Indexing for performance
    INDEX idx_url_id (url_id),
    INDEX idx_click_time (click_time),

    -- Constraint to ensure data integrity
    CONSTRAINT fk_url_analytics_url
        FOREIGN KEY (url_id)
        REFERENCES url (id)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
