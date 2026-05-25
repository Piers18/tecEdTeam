package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"movie-tracker/internal/api/handlers"
	"movie-tracker/internal/cache"
	"movie-tracker/internal/tmdb"
)

func newSearchRouter(t *testing.T, tmdbHandler http.HandlerFunc) http.Handler {
	t.Helper()
	srv := httptest.NewServer(tmdbHandler)
	t.Cleanup(srv.Close)

	// Cache that always misses (nil client means Get always errors).
	c, _ := cache.New("redis://localhost:1") // unreachable → non-fatal miss
	tmdbClient := tmdb.New("testkey", srv.URL)

	h := handlers.NewSearchHandler(tmdbClient, c)
	r := chi.NewRouter()
	r.Get("/search", h.Search)
	return r
}

func TestSearch_MissingQuery(t *testing.T) {
	r := newSearchRouter(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("TMDB should not be called for missing query")
	})

	req := httptest.NewRequest(http.MethodGet, "/search", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSearch_InvalidType(t *testing.T) {
	r := newSearchRouter(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("TMDB should not be called for invalid type")
	})

	req := httptest.NewRequest(http.MethodGet, "/search?q=batman&type=anime", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid type, got %d", w.Code)
	}
}

func TestSearch_InvalidPage(t *testing.T) {
	r := newSearchRouter(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("TMDB should not be called for invalid page")
	})

	req := httptest.NewRequest(http.MethodGet, "/search?q=batman&page=0", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for page=0, got %d", w.Code)
	}
}

func TestSearch_CallsTMDBForMovieType(t *testing.T) {
	called := false
	r := newSearchRouter(t, func(w http.ResponseWriter, req *http.Request) {
		called = true
		if req.URL.Path != "/search/movie" {
			t.Errorf("expected /search/movie, got %s", req.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"page": 1, "results": []interface{}{},
		})
	})

	httpReq := httptest.NewRequest(http.MethodGet, "/search?q=batman&type=movie", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httpReq)

	if !called {
		t.Error("expected TMDB to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSearch_CallsTMDBForTVType(t *testing.T) {
	r := newSearchRouter(t, func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/search/tv" {
			t.Errorf("expected /search/tv, got %s", req.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"page": 1, "results": []interface{}{},
		})
	})

	httpReq := httptest.NewRequest(http.MethodGet, "/search?q=breaking+bad&type=tv", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestSearch_DefaultTypeCallsMulti(t *testing.T) {
	r := newSearchRouter(t, func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/search/multi" {
			t.Errorf("expected /search/multi for default type, got %s", req.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"page": 1, "results": []interface{}{},
		})
	})

	httpReq := httptest.NewRequest(http.MethodGet, "/search?q=inception", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
