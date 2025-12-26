package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/huynguyenanh2000/url-shorterner/internal/store"
	"github.com/huynguyenanh2000/url-shorterner/internal/store/cache"
	"github.com/stretchr/testify/mock"
)

func TestShorternURL(t *testing.T) {
	withRedis := config{
		redisCfg: redisConfig{
			enable: true,
		},
	}
	app := newTestApplication(t, withRedis)
	mux := app.mount()

	longURL := "https://google.com"
	longURLHash := store.ComputeHash(longURL)
	payload := ShorternURLPayload{LongURL: longURL}
	body, _ := json.Marshal(payload)

	resetMocks := func() {
		mockStore := app.store.URL.(*store.MockURLStore)
		mockCacheStore := app.cacheStorage.URL.(*cache.MockURLStore)

		mockStore.ExpectedCalls = nil
		mockStore.Calls = nil

		mockCacheStore.ExpectedCalls = nil
		mockCacheStore.Calls = nil
	}

	t.Run("should return 200 if URL exists in cache (Cache Hit)", func(t *testing.T) {
		resetMocks()
		mockCacheStore := app.cacheStorage.URL.(*cache.MockURLStore)

		existingURL := &store.URL{
			ID:       12345,
			LongURL:  longURL,
			ShortURL: "abcxyz",
		}

		mockCacheStore.On("GetByLongURLHash", mock.Anything, longURLHash).Return(existingURL, nil)

		req, _ := http.NewRequest(http.MethodPost, "/v1/urls/shorten", bytes.NewBuffer(body))
		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusOK, rr.Code)
		mockCacheStore.AssertExpectations(t)
	})

	t.Run("should return 200 and set cache if cache miss but DB exists", func(t *testing.T) {
		resetMocks()
		mockStore := app.store.URL.(*store.MockURLStore)
		mockCacheStore := app.cacheStorage.URL.(*cache.MockURLStore)

		existingURL := &store.URL{
			ID:       12345,
			LongURL:  longURL,
			ShortURL: "abcxyz",
		}

		// Logic: Cache Miss -> DB Hit -> Set Cache
		mockCacheStore.On("GetByLongURLHash", mock.Anything, longURLHash).Return(nil, nil).Once()
		mockStore.On("GetByLongURL", mock.Anything, longURL).Return(existingURL, nil).Once()
		mockCacheStore.On("Set", mock.Anything, existingURL).Return(nil).Once()

		req, _ := http.NewRequest(http.MethodPost, "/v1/urls/shorten", bytes.NewBuffer(body))
		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusOK, rr.Code)

		mockCacheStore.AssertExpectations(t)
		mockStore.AssertExpectations(t)
	})

	t.Run("should return 201 and set cache if URL is completely new", func(t *testing.T) {
		resetMocks()
		mockStore := app.store.URL.(*store.MockURLStore)
		mockCacheStore := app.cacheStorage.URL.(*cache.MockURLStore)

		// Logic: Cache Miss -> DB Miss -> Create -> Set Cache
		mockCacheStore.On("GetByLongURLHash", mock.Anything, longURLHash).Return(nil, nil)
		mockStore.On("GetByLongURL", mock.Anything, longURL).Return(nil, store.ErrNotFound)

		mockStore.On("Create", mock.Anything, mock.MatchedBy(func(u *store.URL) bool {
			return u.LongURL == longURL
		})).Return(nil).Once()
		mockCacheStore.On("Set", mock.Anything, mock.MatchedBy(func(u *store.URL) bool {
			return u.LongURL == longURL
		})).Return(nil).Once()

		req, _ := http.NewRequest(http.MethodPost, "/v1/urls/shorten", bytes.NewBuffer(body))
		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusCreated, rr.Code)

		mockCacheStore.AssertExpectations(t)
		mockStore.AssertExpectations(t)
	})

	t.Run("should return 400 if payload is not a valid URL (e.g., 'aaa')", func(t *testing.T) {
		resetMocks()
		mockCacheStore := app.cacheStorage.URL.(*cache.MockURLStore)
		mockStore := app.store.URL.(*store.MockURLStore)
		mockCacheStore.ExpectedCalls = nil
		mockStore.ExpectedCalls = nil

		invalidPayload := ShorternURLPayload{LongURL: "aaa"}
		body, _ := json.Marshal(invalidPayload)

		req, _ := http.NewRequest(http.MethodPost, "/v1/urls/shorten", bytes.NewBuffer(body))
		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusBadRequest, rr.Code)

		mockCacheStore.AssertNotCalled(t, "GetByLongURLHash", mock.Anything, mock.Anything)
		mockStore.AssertNotCalled(t, "GetByLongURL", mock.Anything, mock.Anything)
		mockStore.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)

		mockCacheStore.AssertExpectations(t)
		mockStore.AssertExpectations(t)
	})
}
