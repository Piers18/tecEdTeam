package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"movie-tracker/internal/api"
	"movie-tracker/internal/cache"
	"movie-tracker/internal/tmdb"
)

type SearchHandler struct {
	tmdb  *tmdb.Client
	cache *cache.Cache
}

func NewSearchHandler(t *tmdb.Client, c *cache.Cache) *SearchHandler {
	return &SearchHandler{tmdb: t, cache: c}
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		api.Error(w, http.StatusBadRequest, "q parameter is required")
		return
	}

	mediaType := r.URL.Query().Get("type")
	if mediaType == "" {
		mediaType = "all"
	}
	if mediaType != "all" && mediaType != "movie" && mediaType != "tv" {
		api.Error(w, http.StatusBadRequest, "type must be all, movie, or tv")
		return
	}

	page := 1
	if ps := r.URL.Query().Get("page"); ps != "" {
		p, err := strconv.Atoi(ps)
		if err != nil || p < 1 {
			api.Error(w, http.StatusBadRequest, "page must be a positive integer")
			return
		}
		page = p
	}

	cacheKey := fmt.Sprintf("search:%s:%s:%d", q, mediaType, page)

	if cached, err := h.cache.Get(r.Context(), cacheKey); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write([]byte(cached))
		return
	}

	var (
		result *tmdb.SearchResult
		err    error
	)
	switch mediaType {
	case "movie":
		result, err = h.tmdb.SearchMovies(r.Context(), q, page)
	case "tv":
		result, err = h.tmdb.SearchTV(r.Context(), q, page)
	default:
		result, err = h.tmdb.SearchMulti(r.Context(), q, page)
	}
	if err != nil {
		api.Error(w, http.StatusBadGateway, "TMDB search failed: "+err.Error())
		return
	}

	b, _ := json.Marshal(result)
	h.cache.Set(r.Context(), cacheKey, string(b), 10*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
