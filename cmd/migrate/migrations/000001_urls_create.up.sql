-- +migrate Up
CREATE TABLE IF NOT EXISTS url (
    id BIGINT UNSIGNED NOT NULL,

    short_url VARCHAR(11) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,

    long_url TEXT NOT NULL,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    PRIMARY KEY (id),

    UNIQUE INDEX idx_short_url (short_url),

    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
