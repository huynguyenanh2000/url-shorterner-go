package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/huynguyenanh2000/url-shorterner/internal/idgen"
	"github.com/huynguyenanh2000/url-shorterner/internal/store"
	"github.com/huynguyenanh2000/url-shorterner/internal/store/cache"
	"go.uber.org/zap"
)

func newTestApplication(t *testing.T, cfg config) *application {
	t.Helper()

	logger := zap.NewNop().Sugar()

	mockStore := store.NewMockStore()
	mockCacheStore := cache.NewMockStore()

	idGen, err := idgen.NewSnowflakeClient(1)
	if err != nil {
		t.Fatalf("failed to initialize id generator for test: %v", err)
	}

	return &application{
		logger:       logger,
		store:        mockStore,
		cacheStorage: mockCacheStore,
		idGenerator:  idGen,
		config:       cfg,
	}
}

func executeRequest(req *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d", expected, actual)
	}
}
