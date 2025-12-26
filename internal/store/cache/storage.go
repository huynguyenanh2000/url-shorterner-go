package cache

import (
	"context"

	"github.com/huynguyenanh2000/url-shorterner/internal/store"
	"github.com/redis/go-redis/v9"
)

type Storage struct {
	URL interface {
		GetByLongURLHash(context.Context, string) (*store.URL, error)
		GetByShortURL(context.Context, string) (*store.URL, error)
		Set(context.Context, *store.URL) error
	}
}

func NewRedisStorage(rdb *redis.Client) Storage {
	return Storage{
		URL: &URLStore{rdb: rdb},
	}
}
