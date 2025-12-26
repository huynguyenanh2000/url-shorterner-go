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

func resetMocks(app *application) {
	if mockStore, ok := app.store.URL.(*store.MockURLStore); ok {
		mockStore.ExpectedCalls = nil
		mockStore.Calls = nil
	}
	if mockCache, ok := app.cacheStorage.URL.(*cache.MockURLStore); ok {
		mockCache.ExpectedCalls = nil
		mockCache.Calls = nil
	}
}

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

	t.Run("should return 200 if URL exists in cache (Cache Hit)", func(t *testing.T) {
		resetMocks(app)
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
		resetMocks(app)
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
		resetMocks(app)
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
		resetMocks(app)
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

func TestURLRedirect(t *testing.T) {
	app := newTestApplication(t, config{redisCfg: redisConfig{enable: true}})
	mux := app.mount()

	shortCode := "abcxyz"
	longURL := "https://google.com"
	testURL := &store.URL{
		ShortURL: shortCode,
		LongURL:  longURL,
	}

	t.Run("should redirect (308) when URL exists in cache", func(t *testing.T) {
		resetMocks(app)
		mockCacheStore := app.cacheStorage.URL.(*cache.MockURLStore)

		// Setup: Cache Hit
		mockCacheStore.On("GetByShortURL", mock.Anything, shortCode).Return(testURL, nil).Once()

		req, _ := http.NewRequest(http.MethodGet, "/v1/urls/"+shortCode, nil)
		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusPermanentRedirect, rr.Code)

		if rr.Header().Get("Location") != longURL {
			t.Errorf("expected location %s, got %s", longURL, rr.Header().Get("Location"))
		}

		mockCacheStore.AssertExpectations(t)
	})

	t.Run("should redirect (308) when cache miss but exists in DB", func(t *testing.T) {
		resetMocks(app)
		mockStore := app.store.URL.(*store.MockURLStore)
		mockCacheStore := app.cacheStorage.URL.(*cache.MockURLStore)

		// Setup: Cache Miss -> DB Hit -> Set Cache
		mockCacheStore.On("GetByShortURL", mock.Anything, shortCode).Return(nil, nil).Once()
		mockStore.On("GetByShortURL", mock.Anything, shortCode).Return(testURL, nil).Once()
		mockCacheStore.On("Set", mock.Anything, testURL).Return(nil).Once()

		req, _ := http.NewRequest(http.MethodGet, "/v1/urls/"+shortCode, nil)
		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusPermanentRedirect, rr.Code)
		if rr.Header().Get("Location") != longURL {
			t.Errorf("expected location %s, got %s", longURL, rr.Header().Get("Location"))
		}

		mockCacheStore.AssertExpectations(t)
		mockStore.AssertExpectations(t)
	})

	t.Run("should return 404 if URL does not exist anywhere", func(t *testing.T) {
		resetMocks(app)
		mockStore := app.store.URL.(*store.MockURLStore)
		mockCacheStore := app.cacheStorage.URL.(*cache.MockURLStore)

		// Setup: Cache Miss -> DB Miss (ErrNotFound)
		mockCacheStore.On("GetByShortURL", mock.Anything, "nonexistent").Return(nil, nil).Once()
		mockStore.On("GetByShortURL", mock.Anything, "nonexistent").Return(nil, store.ErrNotFound).Once()

		req, _ := http.NewRequest(http.MethodGet, "/v1/urls/nonexistent", nil)
		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusNotFound, rr.Code)

		mockStore.AssertExpectations(t)
		mockCacheStore.AssertExpectations(t)
	})
}
