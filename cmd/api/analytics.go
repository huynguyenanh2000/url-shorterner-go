package main

import (
	"errors"
	"net/http"

	"github.com/huynguyenanh2000/url-shorterner/internal/store"
)

// Get URL Stats godoc
//
//	@Summary		Get URL analytics
//	@Description	Get analytics and click statistics for a specific short URL
//	@Tags			urls
//	@Accept			json
//	@Produce		json
//	@Param			shortURL	path		string	true	"Short URL"
//	@Success		200			{object}	store.StatsSummary
//	@Failure		404			{object}	error
//	@Failure		500			{object}	error
//	@Router			/urls/{shortURL}/stats [get]
func (app *application) urlGetStatsHandler(w http.ResponseWriter, r *http.Request) {
	url := getURLFromCtx(r)
	ctx := r.Context()

	// Call the store method which returns a *store.StatsSummary instance
	stats, err := app.store.URLAnalytics.GetStatsByURLID(ctx, url.ID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	// We return the stats instance directly.
	// The jsonResponse function will handle the serialization using your struct tags.
	if err := jsonResponse(w, http.StatusOK, stats); err != nil {
		app.internalServerError(w, r, err)
	}
}
