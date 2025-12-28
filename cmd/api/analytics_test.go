package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/huynguyenanh2000/url-shorterner/internal/store"
	"github.com/huynguyenanh2000/url-shorterner/internal/store/cache"
	"github.com/stretchr/testify/mock"
)

func TestURLGetStats(t *testing.T) {
	// Initialize the test application with Redis/Cache enabled
	app := newTestApplication(t, config{redisCfg: redisConfig{enable: true}})
	mux := app.mount()

	shortCode := "abcxyz"
	testURL := &store.URL{
		ID:       1,
		ShortURL: shortCode,
		LongURL:  "https://google.com",
	}

	// Define the expected statistics payload
	mockStats := &store.StatsSummary{
		TotalClicks: 100,
		Clicks24h:   15,
		TopReferrer: "Direct",
		DeviceStats: map[string]int{"Mobile": 70, "Desktop": 30},
	}

	t.Run("should return 200 and statistics when URL and stats exist", func(t *testing.T) {
		resetMocks(app)

		// Access mock stores via type assertion
		mockURLStore := app.store.URL.(*store.MockURLStore)
		mockAnalyticsStore := app.store.URLAnalytics.(*store.MockAnalyticsStore)
		mockCache := app.cacheStorage.URL.(*cache.MockURLStore)

		// --- Phase 1: Mocking urlContextMiddleware Behavior ---
		// First, the middleware attempts to fetch the URL from Cache
		mockCache.On("GetByShortURL", mock.Anything, shortCode).Return(nil, nil).Once()

		// If Cache miss, the middleware fetches from the Database
		mockURLStore.On("GetByShortURL", mock.Anything, shortCode).Return(testURL, nil).Once()

		// The middleware then updates the Cache with the fetched URL
		mockCache.On("Set", mock.Anything, testURL).Return(nil).Once()

		// --- Phase 2: Mocking urlGetStatsHandler Behavior ---
		// The handler retrieves aggregated statistics for the validated URL ID
		mockAnalyticsStore.On("GetStatsByURLID", mock.Anything, testURL.ID).Return(mockStats, nil).Once()

		// Construct and execute the GET request
		req, _ := http.NewRequest(http.MethodGet, "/v1/urls/"+shortCode+"/stats", nil)
		rr := executeRequest(req, mux)

		// Assert HTTP Status Code 200 OK
		checkResponseCode(t, http.StatusOK, rr.Code)

		var response struct {
			Data store.StatsSummary `json:"data"`
		}

		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to decode JSON response: %v", err)
		}

		// Validate specific metrics
		if response.Data.TotalClicks != mockStats.TotalClicks {
			t.Errorf("Expected %d clicks, got %d", mockStats.TotalClicks, response.Data.TotalClicks)
		}

		if response.Data.DeviceStats["Mobile"] != 70 {
			t.Errorf("Expected 70 mobile clicks, got %d", response.Data.DeviceStats["Mobile"])
		}

		// Ensure all mock expectations were satisfied
		mockURLStore.AssertExpectations(t)
		mockAnalyticsStore.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("should return 404 when the short URL is not found by middleware", func(t *testing.T) {
		resetMocks(app)
		mockURLStore := app.store.URL.(*store.MockURLStore)
		mockCache := app.cacheStorage.URL.(*cache.MockURLStore)

		// Middleware: Cache miss
		mockCache.On("GetByShortURL", mock.Anything, "missing").Return(nil, nil).Once()

		// Middleware: Database miss - returns store.ErrNotFound
		mockURLStore.On("GetByShortURL", mock.Anything, "missing").Return(nil, store.ErrNotFound).Once()

		req, _ := http.NewRequest(http.MethodGet, "/v1/urls/missing/stats", nil)
		rr := executeRequest(req, mux)

		// Assert HTTP Status Code 404 Not Found
		checkResponseCode(t, http.StatusNotFound, rr.Code)

		mockURLStore.AssertExpectations(t)
	})
}
