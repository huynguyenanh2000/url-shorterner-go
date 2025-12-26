package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/huynguyenanh2000/url-shorterner/internal/store"
	"github.com/redis/go-redis/v9"
)

type URLStore struct {
	rdb *redis.Client
}

const URLExpTime = time.Hour * 24 * 7

func (s *URLStore) GetByLongURLHash(ctx context.Context, longURLHash string) (*store.URL, error) {
	cacheKey := fmt.Sprintf("url:l:%s", longURLHash)
	return s.get(ctx, cacheKey)
}

func (s *URLStore) GetByShortURL(ctx context.Context, shortURL string) (*store.URL, error) {
	cacheKey := fmt.Sprintf("url:s:%s", shortURL)
	return s.get(ctx, cacheKey)
}

func (s *URLStore) get(ctx context.Context, cacheKey string) (*store.URL, error) {
	data, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var url store.URL
	if err := json.Unmarshal([]byte(data), &url); err != nil {
		return nil, err
	}

	return &url, nil
}

func (s *URLStore) Set(ctx context.Context, url *store.URL) error {
	data, err := json.Marshal(url)
	if err != nil {
		return err
	}

	longURLHash := store.ComputeHash(url.LongURL)

	pipe := s.rdb.Pipeline()

	pipe.Set(ctx, fmt.Sprintf("url:l:%s", longURLHash), data, URLExpTime)
	pipe.Set(ctx, fmt.Sprintf("url:s:%s", url.ShortURL), data, URLExpTime)

	_, err = pipe.Exec(ctx)
	return err
}
