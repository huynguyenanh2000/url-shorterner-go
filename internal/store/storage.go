package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("resource not found")
	QueryTimeoutDuration = time.Second * 5
)

type Storage struct {
	URL interface {
		Create(context.Context, *URL) error
		GetByLongURL(context.Context, string) (*URL, error)
		GetByShortURL(context.Context, string) (*URL, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		URL: &URLStore{db},
	}
}
