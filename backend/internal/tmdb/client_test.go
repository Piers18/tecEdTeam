package tmdb_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"movie-tracker/internal/tmdb"
)

func newMockServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *tmdb.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client := tmdb.New("testkey", srv.URL)
	return srv, client
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func TestSearchMulti_ReturnsResults(t *testing.T) {
	_, client := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search/multi" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("query") != "batman" {
			t.Errorf("expected query=batman")
		}
		if r.URL.Query().Get("api_key") != "testkey" {
			t.Errorf("expected api_key=testkey")
		}
		writeJSON(w, map[string]interface{}{
			"page":          1,
			"total_pages":   1,
			"total_results": 1,
			"results": []map[string]interface{}{
				{"id": 272, "title": "Batman Begins", "media_type": "movie"},
			},
		})
	})

	result, err := client.SearchMulti(context.Background(), "batman", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Page)
	}
	if len(result.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Results))
	}
	if result.Results[0].Title != "Batman Begins" {
		t.Errorf("expected Batman Begins, got %s", result.Results[0].Title)
	}
}

func TestSearchMovies_InjectsMovieMediaType(t *testing.T) {
	_, client := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]interface{}{
			"page":    1,
			"results": []map[string]interface{}{{"id": 1, "title": "Film"}},
		})
	})

	result, err := client.SearchMovies(context.Background(), "film", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, item := range result.Results {
		if item.MediaType != "movie" {
			t.Errorf("expected media_type=movie, got %s", item.MediaType)
		}
	}
}

func TestSearchTV_InjectsTVMediaType(t *testing.T) {
	_, client := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]interface{}{
			"page":    1,
			"results": []map[string]interface{}{{"id": 1, "name": "Show"}},
		})
	})

	result, err := client.SearchTV(context.Background(), "show", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, item := range result.Results {
		if item.MediaType != "tv" {
			t.Errorf("expected media_type=tv, got %s", item.MediaType)
		}
	}
}

func TestGetMovie_DecodesDetail(t *testing.T) {
	_, client := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/movie/27205" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		writeJSON(w, map[string]interface{}{
			"id":           27205,
			"title":        "Inception",
			"vote_average": 8.4,
			"runtime":      148,
		})
	})

	detail, err := client.GetMovie(context.Background(), 27205)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.ID != 27205 {
		t.Errorf("expected id=27205, got %d", detail.ID)
	}
	if detail.Title != "Inception" {
		t.Errorf("expected Inception, got %s", detail.Title)
	}
	if detail.Runtime != 148 {
		t.Errorf("expected runtime=148, got %d", detail.Runtime)
	}
}

func TestGetTV_DecodesDetail(t *testing.T) {
	_, client := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]interface{}{
			"id":                 1396,
			"name":               "Breaking Bad",
			"number_of_seasons":  5,
			"number_of_episodes": 62,
		})
	})

	detail, err := client.GetTV(context.Background(), 1396)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.Name != "Breaking Bad" {
		t.Errorf("expected Breaking Bad, got %s", detail.Name)
	}
	if detail.NumberOfSeasons != 5 {
		t.Errorf("expected 5 seasons, got %d", detail.NumberOfSeasons)
	}
}

func TestClient_HTTPError(t *testing.T) {
	_, client := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})

	_, err := client.SearchMulti(context.Background(), "x", 1)
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
}

func TestMediaItem_DisplayTitle_UsesTitle(t *testing.T) {
	item := &tmdb.MediaItem{Title: "Inception", Name: ""}
	if item.DisplayTitle() != "Inception" {
		t.Errorf("expected Inception, got %s", item.DisplayTitle())
	}
}

func TestMediaItem_DisplayTitle_FallsBackToName(t *testing.T) {
	item := &tmdb.MediaItem{Title: "", Name: "Breaking Bad"}
	if item.DisplayTitle() != "Breaking Bad" {
		t.Errorf("expected Breaking Bad, got %s", item.DisplayTitle())
	}
}
