package store

import (
	"context"

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

func (s *MockURLStore) Create(ctx context.Context, url *URL) error {
	args := s.Called(ctx, url)
	return args.Error(0)
}

func (s *MockURLStore) GetByLongURL(ctx context.Context, longURL string) (*URL, error) {
	args := s.Called(ctx, longURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*URL), args.Error(1)
}
