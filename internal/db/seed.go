package db

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/huynguyenanh2000/url-shorterner/internal/idgen"
	"github.com/huynguyenanh2000/url-shorterner/internal/pkg/base62"
	"github.com/huynguyenanh2000/url-shorterner/internal/store"
)

var domains = []string{"google.com", "amazon.com", "facebook.com", "shopee.vn", "tiki.vn", "github.com", "medium.com"}

var longPaths = []string{
	"shop/v2/products/apple-iphone-15-pro-max-256gb-blue-titanium?utm_source=facebook&utm_medium=cpc&utm_campaign=black_friday_2025&affiliate_id=998877",
	"news/tech/2025/12/26/google-gemini-vs-openai-latest-model-comparison.html?display=full&share=true&ref=homepage",
	"search/filter?category=electronics&price_range=500-1000&brand=sony&warranty=true&shipping=fast&sort=rating_high_to_low",
	"docs/v1/api/reference/endpoints/webhooks/configurations#setup-instructions-for-production-environment-v2",
	"blog/article/deep-learning-concurrency-patterns-in-golang-distributed-systems?author=huynguyen&topic=backend&level=advanced",
	"watch?v=dQw4w9WgXcQ&list=RDdQw4w9WgXcQ&start_radio=1&ab_channel=OfficialArtist",
}

func SeedURLs(st store.Storage, db *sql.DB, idGen idgen.Client) {
	ctx := context.Background()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := range 100 {
		domain := domains[r.Intn(len(domains))]
		path := longPaths[r.Intn(len(longPaths))]

		longURL := fmt.Sprintf("https://www.%s/%s&seed_id=%d&ts=%d",
			domain, path, i, time.Now().UnixNano())

		id := idGen.Generate()
		shortCode := base62.Encode(id)

		url := &store.URL{
			ID:        id,
			LongURL:   longURL,
			ShortURL:  shortCode,
			CreatedAt: time.Now().Add(-time.Duration(r.Intn(1000)) * time.Hour),
			UpdatedAt: time.Now(),
		}

		_ = st.URL.Create(ctx, url)
	}
}
