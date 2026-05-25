package watchlist

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")
var ErrAlreadyExists = errors.New("already in watchlist")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Add(ctx context.Context, req AddItemRequest) (*Item, error) {
	var item Item
	err := r.db.QueryRow(ctx, `
		INSERT INTO watchlist_items (tmdb_id, media_type, title, poster_path, overview)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, tmdb_id, media_type, title,
		          COALESCE(poster_path, ''), COALESCE(overview, ''), added_at
	`, req.TmdbID, req.MediaType, req.Title, req.PosterPath, req.Overview).
		Scan(&item.ID, &item.TmdbID, &item.MediaType, &item.Title,
			&item.PosterPath, &item.Overview, &item.AddedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrAlreadyExists
		}
		return nil, err
	}
	return &item, nil
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	res, err := r.db.Exec(ctx, `DELETE FROM watchlist_items WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) MarkWatched(ctx context.Context, itemID int, req MarkWatchedRequest) (*WatchRecord, error) {
	var wr WatchRecord
	err := r.db.QueryRow(ctx, `
		INSERT INTO watch_records (watchlist_item_id, watched_at, rating)
		VALUES ($1, $2, $3)
		ON CONFLICT (watchlist_item_id)
		DO UPDATE SET watched_at = EXCLUDED.watched_at, rating = EXCLUDED.rating
		RETURNING id, watched_at::text, rating, created_at
	`, itemID, req.WatchedAt, req.Rating).
		Scan(&wr.ID, &wr.WatchedAt, &wr.Rating, &wr.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &wr, nil
}

func (r *Repository) UnmarkWatched(ctx context.Context, itemID int) error {
	res, err := r.db.Exec(ctx,
		`DELETE FROM watch_records WHERE watchlist_item_id = $1`, itemID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id int) (*Item, error) {
	var item Item
	var wrID *int
	var wrWatchedAt *string
	var wrRating *int
	var wrCreatedAt *time.Time

	err := r.db.QueryRow(ctx, `
		SELECT wi.id, wi.tmdb_id, wi.media_type, wi.title,
		       COALESCE(wi.poster_path, ''), COALESCE(wi.overview, ''), wi.added_at,
		       wr.id, wr.watched_at::text, wr.rating, wr.created_at
		FROM watchlist_items wi
		LEFT JOIN watch_records wr ON wr.watchlist_item_id = wi.id
		WHERE wi.id = $1
	`, id).Scan(
		&item.ID, &item.TmdbID, &item.MediaType, &item.Title,
		&item.PosterPath, &item.Overview, &item.AddedAt,
		&wrID, &wrWatchedAt, &wrRating, &wrCreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if wrID != nil {
		item.Watched = &WatchRecord{
			ID:        *wrID,
			WatchedAt: *wrWatchedAt,
			Rating:    *wrRating,
			CreatedAt: *wrCreatedAt,
		}
	}
	return &item, nil
}

func (r *Repository) List(ctx context.Context, f ListFilter) ([]Item, error) {
	var conditions []string
	var args []interface{}
	argN := 1

	if f.Status == "watched" {
		conditions = append(conditions, "wr.id IS NOT NULL")
	} else if f.Status == "unwatched" {
		conditions = append(conditions, "wr.id IS NULL")
	}

	if f.MediaType != "" && f.MediaType != "all" {
		conditions = append(conditions, fmt.Sprintf("wi.media_type = $%d", argN))
		args = append(args, f.MediaType)
		argN++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	sortCol := "wi.added_at"
	switch f.Sort {
	case "watched_at":
		sortCol = "wr.watched_at"
	case "rating":
		sortCol = "wr.rating"
	case "title":
		sortCol = "wi.title"
	}

	ord := "DESC"
	if strings.EqualFold(f.Order, "asc") {
		ord = "ASC"
	}

	query := fmt.Sprintf(`
		SELECT wi.id, wi.tmdb_id, wi.media_type, wi.title,
		       COALESCE(wi.poster_path, ''), COALESCE(wi.overview, ''), wi.added_at,
		       wr.id, wr.watched_at::text, wr.rating, wr.created_at
		FROM watchlist_items wi
		LEFT JOIN watch_records wr ON wr.watchlist_item_id = wi.id
		%s
		ORDER BY %s %s NULLS LAST
	`, where, sortCol, ord)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		var wrID *int
		var wrWatchedAt *string
		var wrRating *int
		var wrCreatedAt *time.Time

		if err := rows.Scan(
			&item.ID, &item.TmdbID, &item.MediaType, &item.Title,
			&item.PosterPath, &item.Overview, &item.AddedAt,
			&wrID, &wrWatchedAt, &wrRating, &wrCreatedAt,
		); err != nil {
			return nil, err
		}
		if wrID != nil {
			item.Watched = &WatchRecord{
				ID:        *wrID,
				WatchedAt: *wrWatchedAt,
				Rating:    *wrRating,
				CreatedAt: *wrCreatedAt,
			}
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) GetWatchedWithRatings(ctx context.Context, limit int) ([]Item, error) {
	q := `
		SELECT wi.id, wi.tmdb_id, wi.media_type, wi.title,
		       COALESCE(wi.poster_path, ''), COALESCE(wi.overview, ''), wi.added_at,
		       wr.id, wr.watched_at::text, wr.rating, wr.created_at
		FROM watchlist_items wi
		INNER JOIN watch_records wr ON wr.watchlist_item_id = wi.id
		ORDER BY wr.rating DESC, wr.watched_at DESC
	`
	if limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		var wr WatchRecord
		if err := rows.Scan(
			&item.ID, &item.TmdbID, &item.MediaType, &item.Title,
			&item.PosterPath, &item.Overview, &item.AddedAt,
			&wr.ID, &wr.WatchedAt, &wr.Rating, &wr.CreatedAt,
		); err != nil {
			return nil, err
		}
		item.Watched = &wr
		items = append(items, item)
	}
	return items, rows.Err()
}
