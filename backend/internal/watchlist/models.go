package watchlist

import "time"

type Item struct {
	ID         int          `json:"id"`
	TmdbID     int          `json:"tmdb_id"`
	MediaType  string       `json:"media_type"`
	Title      string       `json:"title"`
	PosterPath string       `json:"poster_path"`
	Overview   string       `json:"overview"`
	AddedAt    time.Time    `json:"added_at"`
	Watched    *WatchRecord `json:"watched,omitempty"`
}

type WatchRecord struct {
	ID        int       `json:"id"`
	WatchedAt string    `json:"watched_at"` // "YYYY-MM-DD"
	Rating    int       `json:"rating"`
	CreatedAt time.Time `json:"created_at"`
}

type AddItemRequest struct {
	TmdbID     int    `json:"tmdb_id"`
	MediaType  string `json:"media_type"`
	Title      string `json:"title"`
	PosterPath string `json:"poster_path"`
	Overview   string `json:"overview"`
}

type MarkWatchedRequest struct {
	WatchedAt string `json:"watched_at"` // "YYYY-MM-DD"
	Rating    int    `json:"rating"`
}

type ListFilter struct {
	Status    string // all | watched | unwatched
	MediaType string // all | movie | tv
	Sort      string // added_at | watched_at | rating | title
	Order     string // asc | desc
}
