package watchlist_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"movie-tracker/internal/watchlist"
)

// dbPool returns a connection pool to the test database.
// Tests are skipped if DATABASE_URL is not set.
func dbPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set — skipping integration test")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("failed to connect to test DB: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// cleanTable removes all rows from watchlist_items (cascades to watch_records).
func cleanTable(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), "DELETE FROM watchlist_items")
	if err != nil {
		t.Fatalf("failed to clean table: %v", err)
	}
}

func TestRepository_AddAndGet(t *testing.T) {
	pool := dbPool(t)
	cleanTable(t, pool)
	repo := watchlist.NewRepository(pool)
	ctx := context.Background()

	item, err := repo.Add(ctx, watchlist.AddItemRequest{
		TmdbID:    999,
		MediaType: "movie",
		Title:     "Test Movie",
	})
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if item.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if item.Title != "Test Movie" {
		t.Errorf("expected Test Movie, got %s", item.Title)
	}

	got, err := repo.GetByID(ctx, item.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.TmdbID != 999 {
		t.Errorf("expected tmdb_id=999, got %d", got.TmdbID)
	}
}

func TestRepository_Add_DuplicateReturnsErrAlreadyExists(t *testing.T) {
	pool := dbPool(t)
	cleanTable(t, pool)
	repo := watchlist.NewRepository(pool)
	ctx := context.Background()

	req := watchlist.AddItemRequest{TmdbID: 100, MediaType: "movie", Title: "Dup"}
	if _, err := repo.Add(ctx, req); err != nil {
		t.Fatalf("first Add failed: %v", err)
	}
	_, err := repo.Add(ctx, req)
	if !errors.Is(err, watchlist.ErrAlreadyExists) {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	pool := dbPool(t)
	cleanTable(t, pool)
	repo := watchlist.NewRepository(pool)

	_, err := repo.GetByID(context.Background(), 999999)
	if !errors.Is(err, watchlist.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestRepository_Delete(t *testing.T) {
	pool := dbPool(t)
	cleanTable(t, pool)
	repo := watchlist.NewRepository(pool)
	ctx := context.Background()

	item, _ := repo.Add(ctx, watchlist.AddItemRequest{TmdbID: 200, MediaType: "tv", Title: "Show"})
	if err := repo.Delete(ctx, item.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, err := repo.GetByID(ctx, item.ID)
	if !errors.Is(err, watchlist.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestRepository_Delete_NotFound(t *testing.T) {
	pool := dbPool(t)
	cleanTable(t, pool)
	repo := watchlist.NewRepository(pool)

	err := repo.Delete(context.Background(), 999999)
	if !errors.Is(err, watchlist.ErrNotFound) {
		t.Errorf("expected ErrNotFound for missing item, got %v", err)
	}
}

func TestRepository_MarkWatched_AndUnmark(t *testing.T) {
	pool := dbPool(t)
	cleanTable(t, pool)
	repo := watchlist.NewRepository(pool)
	ctx := context.Background()

	item, _ := repo.Add(ctx, watchlist.AddItemRequest{TmdbID: 300, MediaType: "movie", Title: "Film"})

	wr, err := repo.MarkWatched(ctx, item.ID, watchlist.MarkWatchedRequest{
		WatchedAt: "2026-01-15",
		Rating:    8,
	})
	if err != nil {
		t.Fatalf("MarkWatched failed: %v", err)
	}
	if wr.Rating != 8 {
		t.Errorf("expected rating=8, got %d", wr.Rating)
	}
	if wr.WatchedAt != "2026-01-15" {
		t.Errorf("expected watched_at=2026-01-15, got %s", wr.WatchedAt)
	}

	got, _ := repo.GetByID(ctx, item.ID)
	if got.Watched == nil {
		t.Fatal("expected Watched to be non-nil")
	}

	if err := repo.UnmarkWatched(ctx, item.ID); err != nil {
		t.Fatalf("UnmarkWatched failed: %v", err)
	}

	got, _ = repo.GetByID(ctx, item.ID)
	if got.Watched != nil {
		t.Error("expected Watched to be nil after unmark")
	}
}

func TestRepository_MarkWatched_NotFound(t *testing.T) {
	pool := dbPool(t)
	cleanTable(t, pool)
	repo := watchlist.NewRepository(pool)

	_, err := repo.MarkWatched(context.Background(), 999999, watchlist.MarkWatchedRequest{
		WatchedAt: "2026-01-01",
		Rating:    5,
	})
	if !errors.Is(err, watchlist.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestRepository_List_EmptyWithNoItems(t *testing.T) {
	pool := dbPool(t)
	cleanTable(t, pool)
	repo := watchlist.NewRepository(pool)

	items, err := repo.List(context.Background(), watchlist.ListFilter{
		Sort: "added_at", Order: "desc",
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestRepository_List_FilterByMediaType(t *testing.T) {
	pool := dbPool(t)
	cleanTable(t, pool)
	repo := watchlist.NewRepository(pool)
	ctx := context.Background()

	repo.Add(ctx, watchlist.AddItemRequest{TmdbID: 1, MediaType: "movie", Title: "Movie A"})
	repo.Add(ctx, watchlist.AddItemRequest{TmdbID: 2, MediaType: "tv", Title: "TV Show B"})

	movies, err := repo.List(ctx, watchlist.ListFilter{MediaType: "movie", Sort: "added_at", Order: "desc"})
	if err != nil {
		t.Fatalf("List(movie) failed: %v", err)
	}
	if len(movies) != 1 || movies[0].MediaType != "movie" {
		t.Errorf("expected 1 movie, got %d", len(movies))
	}

	shows, err := repo.List(ctx, watchlist.ListFilter{MediaType: "tv", Sort: "added_at", Order: "desc"})
	if err != nil {
		t.Fatalf("List(tv) failed: %v", err)
	}
	if len(shows) != 1 || shows[0].MediaType != "tv" {
		t.Errorf("expected 1 tv item, got %d", len(shows))
	}
}

func TestRepository_List_FilterByWatchStatus(t *testing.T) {
	pool := dbPool(t)
	cleanTable(t, pool)
	repo := watchlist.NewRepository(pool)
	ctx := context.Background()

	item, _ := repo.Add(ctx, watchlist.AddItemRequest{TmdbID: 10, MediaType: "movie", Title: "Watched"})
	repo.Add(ctx, watchlist.AddItemRequest{TmdbID: 11, MediaType: "movie", Title: "Unwatched"})
	repo.MarkWatched(ctx, item.ID, watchlist.MarkWatchedRequest{WatchedAt: "2026-01-01", Rating: 7})

	watched, _ := repo.List(ctx, watchlist.ListFilter{Status: "watched", Sort: "added_at", Order: "desc"})
	if len(watched) != 1 {
		t.Errorf("expected 1 watched item, got %d", len(watched))
	}

	unwatched, _ := repo.List(ctx, watchlist.ListFilter{Status: "unwatched", Sort: "added_at", Order: "desc"})
	if len(unwatched) != 1 {
		t.Errorf("expected 1 unwatched item, got %d", len(unwatched))
	}
}

func TestRepository_GetWatchedWithRatings(t *testing.T) {
	pool := dbPool(t)
	cleanTable(t, pool)
	repo := watchlist.NewRepository(pool)
	ctx := context.Background()

	a, _ := repo.Add(ctx, watchlist.AddItemRequest{TmdbID: 20, MediaType: "movie", Title: "A"})
	b, _ := repo.Add(ctx, watchlist.AddItemRequest{TmdbID: 21, MediaType: "movie", Title: "B"})
	repo.MarkWatched(ctx, a.ID, watchlist.MarkWatchedRequest{WatchedAt: "2026-01-01", Rating: 6})
	repo.MarkWatched(ctx, b.ID, watchlist.MarkWatchedRequest{WatchedAt: "2026-01-02", Rating: 9})

	items, err := repo.GetWatchedWithRatings(ctx, 0)
	if err != nil {
		t.Fatalf("GetWatchedWithRatings failed: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	// Should be ordered by rating DESC.
	if items[0].Watched.Rating < items[1].Watched.Rating {
		t.Error("results should be ordered by rating descending")
	}
}
