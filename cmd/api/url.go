package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/huynguyenanh2000/url-shorterner/internal/pkg/base62"
	"github.com/huynguyenanh2000/url-shorterner/internal/store"
)

type urlKey string

const urlCtx urlKey = "url"

type ShorternURLPayload struct {
	LongURL string `json:"long_url" validate:"required,http_url"`
}

// Shortern URL godoc
//
//	@Summary		Shortern an URL
//	@Description	Shortern an URL
//	@Tags			urls
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		ShorternURLPayload	true	"URL payload"
//	@Success		201		{object}	store.URL
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/urls/shorten [post]
func (app *application) urlShortenHandler(w http.ResponseWriter, r *http.Request) {
	var payload ShorternURLPayload
	if err := readJson(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	longURLHash := store.ComputeHash(payload.LongURL)

	// Check cache
	existingURL, err := app.cacheStorage.URL.GetByLongURLHash(ctx, longURLHash)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// If longURLHash exists => return 200 OK
	if existingURL != nil {
		if err := jsonResponse(w, http.StatusOK, existingURL); err != nil {
			app.internalServerError(w, r, err)
			return
		}
		return
	}

	// Cache miss -> Check database
	existingURL, err = app.store.URL.GetByLongURL(ctx, payload.LongURL)
	if err != nil && err != store.ErrNotFound {
		app.internalServerError(w, r, err)
		return
	}

	if existingURL != nil {
		_ = app.cacheStorage.URL.Set(ctx, existingURL)

		if err := jsonResponse(w, http.StatusOK, existingURL); err != nil {
			app.internalServerError(w, r, err)
			return
		}
		return
	}

	// Generate ID
	id := app.idGenerator.Generate()

	// Encode to Base62
	shortURL := base62.Encode(id)

	url := &store.URL{
		ID:       id,
		LongURL:  payload.LongURL,
		ShortURL: shortURL,
	}

	// Save to DB
	if err := app.store.URL.Create(ctx, url); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// Save to cache
	if err := app.cacheStorage.URL.Set(ctx, url); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusCreated, url); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// Redirect URL godoc
//
//	@Summary		Redirect to long URL
//	@Description	Redirect to the original long URL based on the short url provided in the path
//	@Tags			urls
//
//	@Accept			json
//	@Produce		json
//
//	@Param			shortURL	path		string	true	"Short URL"
//	@Success		308			{string}	string	"Permanent Redirect"
//	@Failure		404			{object}	error	"URL not found"
//	@Failure		500			{object}	error	"Internal server error"
//
// Security ApiKeyAuth
//
//	@Router			/urls/{shortURL} [get]
func (app *application) urlRedirectHandler(w http.ResponseWriter, r *http.Request) {
	url := getURLFromCtx(r)

	http.Redirect(w, r, url.LongURL, http.StatusPermanentRedirect)
}

func (app *application) urlContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shortURL := chi.URLParam(r, "shortURL")
		ctx := r.Context()

		url, err := app.cacheStorage.URL.GetByShortURL(ctx, shortURL)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		if url == nil {
			url, err = app.store.URL.GetByShortURL(ctx, shortURL)
			if err != nil {
				switch {
				case errors.Is(err, store.ErrNotFound):
					app.notFoundResponse(w, r, err)
				default:
					app.internalServerError(w, r, err)
				}
				return
			}

			if err := app.cacheStorage.URL.Set(ctx, url); err != nil {
				app.internalServerError(w, r, err)
				return
			}
		}

		ctx = context.WithValue(ctx, urlCtx, url)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getURLFromCtx(r *http.Request) *store.URL {
	url, _ := r.Context().Value(urlCtx).(*store.URL)
	return url
}
