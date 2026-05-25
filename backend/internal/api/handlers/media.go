package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"movie-tracker/internal/api"
	"movie-tracker/internal/cache"
	"movie-tracker/internal/tmdb"
)

type MediaHandler struct {
	tmdb  *tmdb.Client
	cache *cache.Cache
}

func NewMediaHandler(t *tmdb.Client, c *cache.Cache) *MediaHandler {
	return &MediaHandler{tmdb: t, cache: c}
}

func (h *MediaHandler) GetMovie(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		api.Error(w, http.StatusBadRequest, "invalid movie id")
		return
	}

	cacheKey := fmt.Sprintf("media:movie:%d", id)

	if cached, err := h.cache.Get(r.Context(), cacheKey); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write([]byte(cached))
		return
	}

	detail, err := h.tmdb.GetMovie(r.Context(), id)
	if err != nil {
		api.Error(w, http.StatusBadGateway, "TMDB movie detail failed: "+err.Error())
		return
	}

	b, _ := json.Marshal(detail)
	h.cache.Set(r.Context(), cacheKey, string(b), 24*time.Hour)

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (h *MediaHandler) GetTV(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		api.Error(w, http.StatusBadRequest, "invalid tv id")
		return
	}

	cacheKey := fmt.Sprintf("media:tv:%d", id)

	if cached, err := h.cache.Get(r.Context(), cacheKey); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write([]byte(cached))
		return
	}

	detail, err := h.tmdb.GetTV(r.Context(), id)
	if err != nil {
		api.Error(w, http.StatusBadGateway, "TMDB TV detail failed: "+err.Error())
		return
	}

	b, _ := json.Marshal(detail)
	h.cache.Set(r.Context(), cacheKey, string(b), 24*time.Hour)

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
