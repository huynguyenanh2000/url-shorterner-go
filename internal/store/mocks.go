package store

import (
	"context"

	"github.com/stretchr/testify/mock"
)

func NewMockStore() Storage {
	return Storage{
		URL:          &MockURLStore{},
		URLAnalytics: &MockAnalyticsStore{},
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

func (s *MockURLStore) GetByShortURL(ctx context.Context, shortURL string) (*URL, error) {
	args := s.Called(ctx, shortURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*URL), args.Error(1)
}

type MockAnalyticsStore struct {
	mock.Mock
}

func (s *MockAnalyticsStore) Create(ctx context.Context, analytics *URLAnalytics) error {
	args := s.Called(ctx, analytics)
	return args.Error(0)
}

func (s *MockAnalyticsStore) GetStatsByURLID(ctx context.Context, urlID uint64) (*StatsSummary, error) {
	args := s.Called(ctx, urlID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*StatsSummary), args.Error(1)
}
