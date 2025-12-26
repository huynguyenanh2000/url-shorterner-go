package cache

import (
	"context"

	"github.com/huynguyenanh2000/url-shorterner/internal/store"
	"github.com/stretchr/testify/mock"
)

func NewMockStore() Storage {
	return Storage{
		URL: &MockURLStore{},
	}
}

type MockURLStore struct {
	mock.Mock
}

func (m *MockURLStore) GetByLongURLHash(ctx context.Context, longURLHash string) (*store.URL, error) {
	args := m.Called(ctx, longURLHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.URL), args.Error(1)
}

func (m *MockURLStore) GetByShortURL(ctx context.Context, shortURL string) (*store.URL, error) {
	args := m.Called(ctx, shortURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.URL), args.Error(1)
}

func (m *MockURLStore) Set(ctx context.Context, url *store.URL) error {
	args := m.Called(ctx, url)
	return args.Error(0)
}
