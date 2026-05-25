package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"movie-tracker/internal/respond"
	"movie-tracker/internal/watchlist"
)

type WatchlistHandler struct {
	repo *watchlist.Repository
}

func NewWatchlistHandler(repo *watchlist.Repository) *WatchlistHandler {
	return &WatchlistHandler{repo: repo}
}

func (h *WatchlistHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := watchlist.ListFilter{
		Status:    q.Get("status"),
		MediaType: q.Get("type"),
		Sort:      q.Get("sort"),
		Order:     q.Get("order"),
	}

	if f.Status != "" && f.Status != "all" && f.Status != "watched" && f.Status != "unwatched" {
		respond.Error(w, http.StatusBadRequest, "status must be all, watched, or unwatched")
		return
	}
	if f.MediaType != "" && f.MediaType != "all" && f.MediaType != "movie" && f.MediaType != "tv" {
		respond.Error(w, http.StatusBadRequest, "type must be all, movie, or tv")
		return
	}
	if f.Sort == "" {
		f.Sort = "added_at"
	}
	if f.Order == "" {
		f.Order = "desc"
	}

	items, err := h.repo.List(r.Context(), f)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to list watchlist")
		return
	}

	if items == nil {
		items = []watchlist.Item{}
	}
	respond.JSON(w, http.StatusOK, items)
}

func (h *WatchlistHandler) Add(w http.ResponseWriter, r *http.Request) {
	var req watchlist.AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.TmdbID == 0 {
		respond.Error(w, http.StatusBadRequest, "tmdb_id is required")
		return
	}
	if req.MediaType != "movie" && req.MediaType != "tv" {
		respond.Error(w, http.StatusBadRequest, "media_type must be movie or tv")
		return
	}
	if req.Title == "" {
		respond.Error(w, http.StatusBadRequest, "title is required")
		return
	}

	item, err := h.repo.Add(r.Context(), req)
	if err != nil {
		if errors.Is(err, watchlist.ErrAlreadyExists) {
			respond.Error(w, http.StatusConflict, "item already in watchlist")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "failed to add item")
		return
	}
	respond.JSON(w, http.StatusCreated, item)
}

func (h *WatchlistHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, watchlist.ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "item not found")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "failed to delete item")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WatchlistHandler) MarkWatched(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req watchlist.MarkWatchedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.WatchedAt == "" {
		respond.Error(w, http.StatusBadRequest, "watched_at is required (YYYY-MM-DD)")
		return
	}
	if req.Rating < 1 || req.Rating > 10 {
		respond.Error(w, http.StatusBadRequest, "rating must be between 1 and 10")
		return
	}

	wr, err := h.repo.MarkWatched(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, watchlist.ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "item not found")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "failed to mark as watched")
		return
	}
	respond.JSON(w, http.StatusOK, wr)
}

func (h *WatchlistHandler) UnmarkWatched(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.repo.UnmarkWatched(r.Context(), id); err != nil {
		if errors.Is(err, watchlist.ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "no watch record for this item")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "failed to unmark")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WatchlistHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	item, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, watchlist.ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "item not found")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "failed to get item")
		return
	}
	respond.JSON(w, http.StatusOK, item)
}
