package store

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"time"
)

type URL struct {
	ID        uint64    `json:"id"`
	ShortURL  string    `json:"short_url"`
	LongURL   string    `json:"long_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type URLStore struct {
	db *sql.DB
}

func ComputeHash(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *URLStore) Create(ctx context.Context, url *URL) error {
	now := time.Now().UTC()
	url.CreatedAt = now
	url.UpdatedAt = now

	longURLHash := ComputeHash(url.LongURL)

	query := `
		INSERT INTO url (id, long_url_hash, short_url, long_url, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		url.ID,
		longURLHash,
		url.ShortURL,
		url.LongURL,
		url.CreatedAt,
		url.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *URLStore) GetByLongURL(ctx context.Context, longURL string) (*URL, error) {
	longURLHash := ComputeHash(longURL)

	query := `
		SELECT id, short_url, long_url, created_at, updated_at
		FROM url
		WHERE long_url_hash = ? AND long_url = ?
		LIMIT 1
	`

	url := &URL{}

	err := s.db.QueryRowContext(
		ctx,
		query,
		longURLHash,
		longURL,
	).Scan(
		&url.ID,
		&url.ShortURL,
		&url.LongURL,
		&url.CreatedAt,
		&url.UpdatedAt,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return url, nil
}

func (s *URLStore) GetByShortURL(ctx context.Context, shortURL string) (*URL, error) {
	query := `
		SELECT id, short_url, long_url, created_at, updated_at
		FROM url
		WHERE short_url = ?
		LIMIT 1
	`

	url := &URL{}

	err := s.db.QueryRowContext(
		ctx,
		query,
		shortURL,
	).Scan(
		&url.ID,
		&url.ShortURL,
		&url.LongURL,
		&url.CreatedAt,
		&url.UpdatedAt,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return url, nil
}
