package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"movie-tracker/internal/api/handlers"
)

// newWatchlistRouter builds a chi router wired to a real WatchlistHandler with
// a nil repository. Only tests that hit validation before the repo is touched
// can use this safely.
func newWatchlistRouter() http.Handler {
	h := handlers.NewWatchlistHandler(nil)
	r := chi.NewRouter()
	r.Post("/watchlist", h.Add)
	r.Post("/watchlist/{id}/watch", h.MarkWatched)
	r.Get("/watchlist/{id}", h.GetOne)
	r.Delete("/watchlist/{id}", h.Delete)
	r.Delete("/watchlist/{id}/watch", h.UnmarkWatched)
	return r
}

func postJSON(t *testing.T, router http.Handler, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// --- Add validation ---

func TestAdd_MissingTmdbID(t *testing.T) {
	r := newWatchlistRouter()
	w := postJSON(t, r, "/watchlist", map[string]interface{}{
		"media_type": "movie", "title": "Inception",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "tmdb_id") {
		t.Errorf("expected tmdb_id in error, got %s", w.Body.String())
	}
}

func TestAdd_InvalidMediaType(t *testing.T) {
	r := newWatchlistRouter()
	w := postJSON(t, r, "/watchlist", map[string]interface{}{
		"tmdb_id": 1, "media_type": "anime", "title": "X",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAdd_MissingTitle(t *testing.T) {
	r := newWatchlistRouter()
	w := postJSON(t, r, "/watchlist", map[string]interface{}{
		"tmdb_id": 1, "media_type": "movie",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAdd_InvalidJSON(t *testing.T) {
	r := newWatchlistRouter()
	req := httptest.NewRequest(http.MethodPost, "/watchlist", strings.NewReader("{bad json}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// --- MarkWatched validation ---

func TestMarkWatched_InvalidID(t *testing.T) {
	r := newWatchlistRouter()
	w := postJSON(t, r, "/watchlist/abc/watch", map[string]interface{}{
		"watched_at": "2026-01-01", "rating": 8,
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for non-numeric ID, got %d", w.Code)
	}
}

func TestMarkWatched_MissingWatchedAt(t *testing.T) {
	r := newWatchlistRouter()
	w := postJSON(t, r, "/watchlist/1/watch", map[string]interface{}{
		"rating": 8,
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing watched_at, got %d", w.Code)
	}
}

func TestMarkWatched_InvalidDateFormat(t *testing.T) {
	r := newWatchlistRouter()
	w := postJSON(t, r, "/watchlist/1/watch", map[string]interface{}{
		"watched_at": "not-a-date", "rating": 8,
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid date, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "YYYY-MM-DD") {
		t.Errorf("expected YYYY-MM-DD in error message, got %s", w.Body.String())
	}
}

func TestMarkWatched_RatingTooLow(t *testing.T) {
	r := newWatchlistRouter()
	w := postJSON(t, r, "/watchlist/1/watch", map[string]interface{}{
		"watched_at": "2026-01-01", "rating": 0,
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for rating=0, got %d", w.Code)
	}
}

func TestMarkWatched_RatingTooHigh(t *testing.T) {
	r := newWatchlistRouter()
	w := postJSON(t, r, "/watchlist/1/watch", map[string]interface{}{
		"watched_at": "2026-01-01", "rating": 11,
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for rating=11, got %d", w.Code)
	}
}

// --- GetOne / Delete validation ---

func TestGetOne_InvalidID(t *testing.T) {
	r := newWatchlistRouter()
	req := httptest.NewRequest(http.MethodGet, "/watchlist/notanid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for non-numeric ID, got %d", w.Code)
	}
}

func TestDelete_InvalidID(t *testing.T) {
	r := newWatchlistRouter()
	req := httptest.NewRequest(http.MethodDelete, "/watchlist/xyz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for non-numeric ID, got %d", w.Code)
	}
}
