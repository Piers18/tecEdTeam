# Movie Tracker Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a full-stack personal movie/series tracker with Go backend, React frontend, PostgreSQL, Redis cache, and an OpenAI chat agent that uses function calling.

**Architecture:** The backend is a Go HTTP server (chi router) exposing a REST API. TMDB results are cached in Redis (10-min TTL for search, 24h for detail). The watchlist lives in PostgreSQL with raw SQL queries. The AI agent maintains session history in memory and calls tools backed by the same DB layer.

**Tech Stack:** Go 1.22, chi v5, pgx/v5, go-redis/v9, openai-go, golang-migrate, Vite + React, PostgreSQL 15, Redis 7, Docker + docker-compose

---

## File Map

```
movie-tracker/
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/config.go
│   │   ├── db/
│   │   │   ├── db.go
│   │   │   ├── migrate.go
│   │   │   └── migrations/
│   │   │       ├── 001_watchlist_items.up.sql
│   │   │       └── 002_watch_records.up.sql
│   │   ├── cache/redis.go
│   │   ├── tmdb/
│   │   │   ├── client.go
│   │   │   └── models.go
│   │   ├── watchlist/
│   │   │   ├── models.go
│   │   │   ├── repository.go
│   │   │   └── repository_test.go
│   │   ├── chat/
│   │   │   ├── session.go
│   │   │   ├── session_test.go
│   │   │   ├── tools.go
│   │   │   └── agent.go
│   │   └── api/
│   │       ├── router.go
│   │       ├── response.go
│   │       ├── middleware/cors.go
│   │       └── handlers/
│   │           ├── search.go
│   │           ├── search_test.go
│   │           ├── media.go
│   │           ├── watchlist.go
│   │           ├── watchlist_test.go
│   │           └── chat.go
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── api/client.js
│   │   ├── components/
│   │   │   ├── SearchBar.jsx
│   │   │   ├── SearchResults.jsx
│   │   │   ├── WatchlistView.jsx
│   │   │   ├── WatchlistItem.jsx
│   │   │   ├── WatchModal.jsx
│   │   │   └── ChatWindow.jsx
│   │   ├── App.jsx
│   │   └── main.jsx
│   ├── index.html
│   ├── vite.config.js
│   ├── package.json
│   ├── nginx.conf
│   └── Dockerfile
├── docker-compose.yml
├── .env.example
└── README.md
```

---

## Task 1: Repo scaffold — docker-compose, .env.example, git init

**Files:**
- Create: `docker-compose.yml`
- Create: `.env.example`

- [ ] **Step 1: Initialize git repo and create root directory layout**

```bash
cd /path/to/movie-tracker
git init
mkdir -p backend/cmd/server
mkdir -p backend/internal/{config,db/migrations,cache,tmdb,watchlist,chat,api/{handlers,middleware}}
mkdir -p frontend/src/{api,components}
```

- [ ] **Step 2: Create `.env.example`**

```env
# Database
POSTGRES_USER=tracker
POSTGRES_PASSWORD=tracker
POSTGRES_DB=movietracker
DATABASE_URL=postgres://tracker:tracker@postgres:5432/movietracker?sslmode=disable

# Redis
REDIS_URL=redis://redis:6379

# TMDB
TMDB_API_KEY=your_tmdb_api_key_here
TMDB_BASE_URL=https://api.themoviedb.org/3

# OpenAI
OPENAI_API_KEY=your_openai_api_key_here

# Server
PORT=8080
```

- [ ] **Step 3: Create `docker-compose.yml`**

```yaml
version: "3.9"

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: ${POSTGRES_DB}
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER} -d ${POSTGRES_DB}"]
      interval: 5s
      timeout: 5s
      retries: 10

  redis:
    image: redis:7-alpine
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

  backend:
    build: ./backend
    ports:
      - "8080:8080"
    env_file: .env
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy

  frontend:
    build: ./frontend
    ports:
      - "5173:80"
    depends_on:
      - backend

volumes:
  pgdata:
```

- [ ] **Step 4: Create `.gitignore`**

```gitignore
.env
*.env.local
__debug_bin
/backend/tmp/
/frontend/node_modules/
/frontend/dist/
*.sum
```

- [ ] **Step 5: Commit**

```bash
git add docker-compose.yml .env.example .gitignore
git commit -m "chore: project scaffold with docker-compose"
```

---

## Task 2: Go module and backend Dockerfile

**Files:**
- Create: `backend/go.mod`
- Create: `backend/Dockerfile`

- [ ] **Step 1: Initialize Go module**

```bash
cd backend
go mod init movie-tracker
```

- [ ] **Step 2: Add dependencies to `backend/go.mod`**

The final `go.mod` after `go mod tidy` will include these. Install them now:

```bash
go get github.com/go-chi/chi/v5@v5.1.0
go get github.com/jackc/pgx/v5@v5.6.0
go get github.com/redis/go-redis/v9@v9.6.1
go get github.com/openai/openai-go@v0.1.0-alpha.40
go get github.com/joho/godotenv@v1.5.1
go get github.com/google/uuid@v1.6.0
```

- [ ] **Step 3: Create `backend/Dockerfile`**

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

- [ ] **Step 4: Commit**

```bash
cd backend && git add go.mod go.sum Dockerfile
git commit -m "chore: go module and backend Dockerfile"
```

---

## Task 3: Config package

**Files:**
- Create: `backend/internal/config/config.go`

- [ ] **Step 1: Write `backend/internal/config/config.go`**

```go
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	RedisURL    string
	TMDBAPIKey  string
	TMDBBaseURL string
	OpenAIKey   string
}

func Load() (*Config, error) {
	// Load .env if present (local dev). Ignored in Docker where env is injected.
	_ = godotenv.Load()

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		TMDBAPIKey:  os.Getenv("TMDB_API_KEY"),
		TMDBBaseURL: getEnv("TMDB_BASE_URL", "https://api.themoviedb.org/3"),
		OpenAIKey:   os.Getenv("OPENAI_API_KEY"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.TMDBAPIKey == "" {
		return nil, fmt.Errorf("TMDB_API_KEY is required")
	}
	if cfg.OpenAIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/config/
git commit -m "feat: config package with env validation"
```

---

## Task 4: Database setup and migrations

**Files:**
- Create: `backend/internal/db/db.go`
- Create: `backend/internal/db/migrate.go`
- Create: `backend/internal/db/migrations/001_watchlist_items.up.sql`
- Create: `backend/internal/db/migrations/002_watch_records.up.sql`

- [ ] **Step 1: Write `backend/internal/db/db.go`**

```go
package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
```

- [ ] **Step 2: Write `backend/internal/db/migrations/001_watchlist_items.up.sql`**

```sql
CREATE TABLE IF NOT EXISTS watchlist_items (
    id          SERIAL PRIMARY KEY,
    tmdb_id     INTEGER NOT NULL,
    media_type  VARCHAR(10) NOT NULL CHECK (media_type IN ('movie', 'tv')),
    title       VARCHAR(255) NOT NULL,
    poster_path VARCHAR(255),
    overview    TEXT,
    added_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tmdb_id, media_type)
);
```

- [ ] **Step 3: Write `backend/internal/db/migrations/002_watch_records.up.sql`**

```sql
CREATE TABLE IF NOT EXISTS watch_records (
    id                  SERIAL PRIMARY KEY,
    watchlist_item_id   INTEGER NOT NULL REFERENCES watchlist_items(id) ON DELETE CASCADE,
    watched_at          DATE NOT NULL,
    rating              SMALLINT NOT NULL CHECK (rating >= 1 AND rating <= 10),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (watchlist_item_id)
);
```

- [ ] **Step 4: Write `backend/internal/db/migrate.go`**

```go
package db

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return err
	}

	var files []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, f := range files {
		version := strings.TrimSuffix(f, ".up.sql")

		var applied bool
		if err := pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`,
			version,
		).Scan(&applied); err != nil {
			return err
		}
		if applied {
			continue
		}

		content, err := migrationFS.ReadFile("migrations/" + f)
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("migration %s: %w", f, err)
		}
		if _, err := pool.Exec(ctx,
			`INSERT INTO schema_migrations (version) VALUES ($1)`, version,
		); err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/db/
git commit -m "feat: db pool setup and embedded SQL migrations"
```

---

## Task 5: Redis cache

**Files:**
- Create: `backend/internal/cache/redis.go`

- [ ] **Step 1: Write `backend/internal/cache/redis.go`**

```go
package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

func New(redisURL string) (*Cache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	c := &Cache{client: redis.NewClient(opts)}
	return c, nil
}

func (c *Cache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *Cache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *Cache) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/cache/
git commit -m "feat: Redis cache wrapper"
```

---

## Task 6: TMDB client

**Files:**
- Create: `backend/internal/tmdb/models.go`
- Create: `backend/internal/tmdb/client.go`

- [ ] **Step 1: Write `backend/internal/tmdb/models.go`**

```go
package tmdb

// MediaItem is returned from search results (both movie and TV).
type MediaItem struct {
	ID           int     `json:"id"`
	Title        string  `json:"title,omitempty"`
	Name         string  `json:"name,omitempty"`
	Overview     string  `json:"overview"`
	PosterPath   string  `json:"poster_path"`
	MediaType    string  `json:"media_type"`
	VoteAverage  float64 `json:"vote_average"`
	ReleaseDate  string  `json:"release_date,omitempty"`
	FirstAirDate string  `json:"first_air_date,omitempty"`
}

// DisplayTitle returns the correct title field regardless of media type.
func (m *MediaItem) DisplayTitle() string {
	if m.Title != "" {
		return m.Title
	}
	return m.Name
}

type SearchResult struct {
	Page       int         `json:"page"`
	Results    []MediaItem `json:"results"`
	TotalPages int         `json:"total_pages"`
	TotalItems int         `json:"total_results"`
}

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type MovieDetail struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Overview    string  `json:"overview"`
	PosterPath  string  `json:"poster_path"`
	ReleaseDate string  `json:"release_date"`
	Runtime     int     `json:"runtime"`
	VoteAverage float64 `json:"vote_average"`
	Genres      []Genre `json:"genres"`
}

type TVDetail struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	Overview         string  `json:"overview"`
	PosterPath       string  `json:"poster_path"`
	FirstAirDate     string  `json:"first_air_date"`
	NumberOfSeasons  int     `json:"number_of_seasons"`
	NumberOfEpisodes int     `json:"number_of_episodes"`
	VoteAverage      float64 `json:"vote_average"`
	Genres           []Genre `json:"genres"`
}
```

- [ ] **Step 2: Write `backend/internal/tmdb/client.go`**

```go
package tmdb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

func New(apiKey, baseURL string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) SearchMulti(ctx context.Context, query string, page int) (*SearchResult, error) {
	u := c.buildURL("/search/multi", map[string]string{
		"query": query,
		"page":  strconv.Itoa(page),
	})
	var result SearchResult
	return &result, c.get(ctx, u, &result)
}

func (c *Client) SearchMovies(ctx context.Context, query string, page int) (*SearchResult, error) {
	u := c.buildURL("/search/movie", map[string]string{
		"query": query,
		"page":  strconv.Itoa(page),
	})
	var result SearchResult
	if err := c.get(ctx, u, &result); err != nil {
		return nil, err
	}
	for i := range result.Results {
		result.Results[i].MediaType = "movie"
	}
	return &result, nil
}

func (c *Client) SearchTV(ctx context.Context, query string, page int) (*SearchResult, error) {
	u := c.buildURL("/search/tv", map[string]string{
		"query": query,
		"page":  strconv.Itoa(page),
	})
	var result SearchResult
	if err := c.get(ctx, u, &result); err != nil {
		return nil, err
	}
	for i := range result.Results {
		result.Results[i].MediaType = "tv"
	}
	return &result, nil
}

func (c *Client) GetMovie(ctx context.Context, id int) (*MovieDetail, error) {
	u := c.buildURL(fmt.Sprintf("/movie/%d", id), nil)
	var detail MovieDetail
	return &detail, c.get(ctx, u, &detail)
}

func (c *Client) GetTV(ctx context.Context, id int) (*TVDetail, error) {
	u := c.buildURL(fmt.Sprintf("/tv/%d", id), nil)
	var detail TVDetail
	return &detail, c.get(ctx, u, &detail)
}

func (c *Client) buildURL(path string, params map[string]string) string {
	u, _ := url.Parse(c.baseURL + path)
	q := u.Query()
	q.Set("api_key", c.apiKey)
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func (c *Client) get(ctx context.Context, rawURL string, dest interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("tmdb: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("tmdb: unexpected status %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(dest)
}
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/tmdb/
git commit -m "feat: TMDB client for search and detail"
```

---

## Task 7: Watchlist models and repository

**Files:**
- Create: `backend/internal/watchlist/models.go`
- Create: `backend/internal/watchlist/repository.go`

- [ ] **Step 1: Write `backend/internal/watchlist/models.go`**

```go
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
```

- [ ] **Step 2: Write `backend/internal/watchlist/repository.go`**

```go
package watchlist

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
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
		if strings.Contains(err.Error(), "unique") ||
			strings.Contains(err.Error(), "duplicate") {
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
		if strings.Contains(err.Error(), "foreign key") {
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
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/watchlist/
git commit -m "feat: watchlist models and repository with raw SQL"
```

---

## Task 8: Shared API response helpers

**Files:**
- Create: `backend/internal/api/response.go`

- [ ] **Step 1: Write `backend/internal/api/response.go`**

```go
package api

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, map[string]string{"error": msg})
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/api/response.go
git commit -m "feat: shared JSON/Error response helpers"
```

---

## Task 9: Search handler

**Files:**
- Create: `backend/internal/api/handlers/search.go`

- [ ] **Step 1: Write `backend/internal/api/handlers/search.go`**

```go
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
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/api/handlers/search.go
git commit -m "feat: search handler with Redis cache (10-min TTL)"
```

---

## Task 10: Media detail handler

**Files:**
- Create: `backend/internal/api/handlers/media.go`

- [ ] **Step 1: Write `backend/internal/api/handlers/media.go`**

```go
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
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/api/handlers/media.go
git commit -m "feat: media detail handler with 24h Redis cache"
```

---

## Task 11: Watchlist handler

**Files:**
- Create: `backend/internal/api/handlers/watchlist.go`

- [ ] **Step 1: Write `backend/internal/api/handlers/watchlist.go`**

```go
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"movie-tracker/internal/api"
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

	// Validate filter values
	if f.Status != "" && f.Status != "all" && f.Status != "watched" && f.Status != "unwatched" {
		api.Error(w, http.StatusBadRequest, "status must be all, watched, or unwatched")
		return
	}
	if f.MediaType != "" && f.MediaType != "all" && f.MediaType != "movie" && f.MediaType != "tv" {
		api.Error(w, http.StatusBadRequest, "type must be all, movie, or tv")
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
		api.Error(w, http.StatusInternalServerError, "failed to list watchlist")
		return
	}

	if items == nil {
		items = []watchlist.Item{}
	}
	api.JSON(w, http.StatusOK, items)
}

func (h *WatchlistHandler) Add(w http.ResponseWriter, r *http.Request) {
	var req watchlist.AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.TmdbID == 0 {
		api.Error(w, http.StatusBadRequest, "tmdb_id is required")
		return
	}
	if req.MediaType != "movie" && req.MediaType != "tv" {
		api.Error(w, http.StatusBadRequest, "media_type must be movie or tv")
		return
	}
	if req.Title == "" {
		api.Error(w, http.StatusBadRequest, "title is required")
		return
	}

	item, err := h.repo.Add(r.Context(), req)
	if err != nil {
		if errors.Is(err, watchlist.ErrAlreadyExists) {
			api.Error(w, http.StatusConflict, "item already in watchlist")
			return
		}
		api.Error(w, http.StatusInternalServerError, "failed to add item")
		return
	}
	api.JSON(w, http.StatusCreated, item)
}

func (h *WatchlistHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		api.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, watchlist.ErrNotFound) {
			api.Error(w, http.StatusNotFound, "item not found")
			return
		}
		api.Error(w, http.StatusInternalServerError, "failed to delete item")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WatchlistHandler) MarkWatched(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		api.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req watchlist.MarkWatchedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.WatchedAt == "" {
		api.Error(w, http.StatusBadRequest, "watched_at is required (YYYY-MM-DD)")
		return
	}
	if req.Rating < 1 || req.Rating > 10 {
		api.Error(w, http.StatusBadRequest, "rating must be between 1 and 10")
		return
	}

	wr, err := h.repo.MarkWatched(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, watchlist.ErrNotFound) {
			api.Error(w, http.StatusNotFound, "item not found")
			return
		}
		api.Error(w, http.StatusInternalServerError, "failed to mark as watched")
		return
	}
	api.JSON(w, http.StatusOK, wr)
}

func (h *WatchlistHandler) UnmarkWatched(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		api.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.repo.UnmarkWatched(r.Context(), id); err != nil {
		if errors.Is(err, watchlist.ErrNotFound) {
			api.Error(w, http.StatusNotFound, "no watch record for this item")
			return
		}
		api.Error(w, http.StatusInternalServerError, "failed to unmark")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WatchlistHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		api.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	item, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, watchlist.ErrNotFound) {
			api.Error(w, http.StatusNotFound, "item not found")
			return
		}
		api.Error(w, http.StatusInternalServerError, "failed to get item")
		return
	}
	api.JSON(w, http.StatusOK, item)
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/api/handlers/watchlist.go
git commit -m "feat: watchlist CRUD handlers with input validation"
```

---

## Task 12: Chat session store

**Files:**
- Create: `backend/internal/chat/session.go`
- Create: `backend/internal/chat/session_test.go`

- [ ] **Step 1: Write `backend/internal/chat/session.go`**

```go
package chat

import (
	"sync"

	"github.com/openai/openai-go"
)

// SessionStore holds in-memory conversation histories keyed by session ID.
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string][]openai.ChatCompletionMessageParamUnion
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string][]openai.ChatCompletionMessageParamUnion),
	}
}

func (s *SessionStore) Get(id string) []openai.ChatCompletionMessageParamUnion {
	s.mu.RLock()
	defer s.mu.RUnlock()
	hist := s.sessions[id]
	cp := make([]openai.ChatCompletionMessageParamUnion, len(hist))
	copy(cp, hist)
	return cp
}

func (s *SessionStore) Append(id string, msgs ...openai.ChatCompletionMessageParamUnion) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = append(s.sessions[id], msgs...)
}

func (s *SessionStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
}

func (s *SessionStore) Exists(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.sessions[id]
	return ok
}
```

- [ ] **Step 2: Write `backend/internal/chat/session_test.go`**

```go
package chat

import (
	"testing"

	"github.com/openai/openai-go"
)

func TestSessionStore_AppendAndGet(t *testing.T) {
	s := NewSessionStore()
	s.Append("sess1", openai.UserMessage("hello"))
	s.Append("sess1", openai.AssistantMessage("hi there"))

	hist := s.Get("sess1")
	if len(hist) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(hist))
	}
}

func TestSessionStore_Delete(t *testing.T) {
	s := NewSessionStore()
	s.Append("sess2", openai.UserMessage("msg"))
	s.Delete("sess2")

	if s.Exists("sess2") {
		t.Fatal("session should not exist after delete")
	}
	if len(s.Get("sess2")) != 0 {
		t.Fatal("get on deleted session should return empty slice")
	}
}

func TestSessionStore_IsolatesSessionsFromEachOther(t *testing.T) {
	s := NewSessionStore()
	s.Append("a", openai.UserMessage("msg-a"))
	s.Append("b", openai.UserMessage("msg-b"))

	if len(s.Get("a")) != 1 {
		t.Fatal("session a should have 1 message")
	}
	if len(s.Get("b")) != 1 {
		t.Fatal("session b should have 1 message")
	}
}
```

- [ ] **Step 3: Run the test**

```bash
cd backend && go test ./internal/chat/ -v -run TestSessionStore
```

Expected output:
```
--- PASS: TestSessionStore_AppendAndGet (0.00s)
--- PASS: TestSessionStore_Delete (0.00s)
--- PASS: TestSessionStore_IsolatesSessionsFromEachOther (0.00s)
PASS
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/chat/session.go backend/internal/chat/session_test.go
git commit -m "feat: in-memory chat session store with tests"
```

---

## Task 13: Chat tools and agent

**Files:**
- Create: `backend/internal/chat/tools.go`
- Create: `backend/internal/chat/agent.go`

- [ ] **Step 1: Write `backend/internal/chat/tools.go`**

```go
package chat

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"

	"movie-tracker/internal/tmdb"
	"movie-tracker/internal/watchlist"
)

// toolDefinitions declares all tools the agent can call.
var toolDefinitions = []openai.ChatCompletionToolParam{
	{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),
		Function: openai.F(openai.FunctionDefinitionParam{
			Name:        openai.String("get_watchlist"),
			Description: openai.String("Get the user's watchlist. Optionally filter by status (all/watched/unwatched) and media_type (all/movie/tv)."),
			Parameters: openai.F(openai.FunctionParameters{
				"type": "object",
				"properties": map[string]interface{}{
					"status": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"all", "watched", "unwatched"},
						"description": "Filter by watch status",
					},
					"media_type": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"all", "movie", "tv"},
						"description": "Filter by media type",
					},
				},
			}),
		}),
	},
	{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),
		Function: openai.F(openai.FunctionDefinitionParam{
			Name:        openai.String("get_watched_with_ratings"),
			Description: openai.String("Get items the user has already watched, ordered by rating descending. Use to understand the user's taste before making recommendations."),
			Parameters: openai.F(openai.FunctionParameters{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Max number of results to return. Use 0 for all.",
					},
				},
			}),
		}),
	},
	{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),
		Function: openai.F(openai.FunctionDefinitionParam{
			Name:        openai.String("get_watchlist_item"),
			Description: openai.String("Get full details of a specific watchlist item including its watch record if it exists."),
			Parameters: openai.F(openai.FunctionParameters{
				"type":     "object",
				"required": []string{"id"},
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "integer",
						"description": "Watchlist item ID",
					},
				},
			}),
		}),
	},
	{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),
		Function: openai.F(openai.FunctionDefinitionParam{
			Name:        openai.String("search_tmdb"),
			Description: openai.String("Search TMDB for movies or TV series. Use to find content to recommend to the user."),
			Parameters: openai.F(openai.FunctionParameters{
				"type":     "object",
				"required": []string{"query"},
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query string",
					},
					"type": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"movie", "tv", "all"},
						"description": "Type of content to search",
					},
				},
			}),
		}),
	},
}

type toolExecutor struct {
	repo *watchlist.Repository
	tmdb *tmdb.Client
}

func newToolExecutor(repo *watchlist.Repository, tmdbClient *tmdb.Client) *toolExecutor {
	return &toolExecutor{repo: repo, tmdb: tmdbClient}
}

func (e *toolExecutor) execute(ctx context.Context, name, argsJSON string) string {
	result, err := e.dispatch(ctx, name, argsJSON)
	if err != nil {
		errJSON, _ := json.Marshal(map[string]string{"error": err.Error()})
		return string(errJSON)
	}
	return result
}

func (e *toolExecutor) dispatch(ctx context.Context, name, argsJSON string) (string, error) {
	switch name {
	case "get_watchlist":
		return e.getWatchlist(ctx, argsJSON)
	case "get_watched_with_ratings":
		return e.getWatchedWithRatings(ctx, argsJSON)
	case "get_watchlist_item":
		return e.getWatchlistItem(ctx, argsJSON)
	case "search_tmdb":
		return e.searchTMDB(ctx, argsJSON)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func (e *toolExecutor) getWatchlist(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Status    string `json:"status"`
		MediaType string `json:"media_type"`
	}
	json.Unmarshal([]byte(argsJSON), &args) //nolint
	items, err := e.repo.List(ctx, watchlist.ListFilter{
		Status:    args.Status,
		MediaType: args.MediaType,
		Sort:      "added_at",
		Order:     "desc",
	})
	if err != nil {
		return "", err
	}
	b, _ := json.Marshal(items)
	return string(b), nil
}

func (e *toolExecutor) getWatchedWithRatings(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Limit int `json:"limit"`
	}
	json.Unmarshal([]byte(argsJSON), &args) //nolint
	items, err := e.repo.GetWatchedWithRatings(ctx, args.Limit)
	if err != nil {
		return "", err
	}
	b, _ := json.Marshal(items)
	return string(b), nil
}

func (e *toolExecutor) getWatchlistItem(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}
	item, err := e.repo.GetByID(ctx, args.ID)
	if err != nil {
		return "", err
	}
	b, _ := json.Marshal(item)
	return string(b), nil
}

func (e *toolExecutor) searchTMDB(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Query string `json:"query"`
		Type  string `json:"type"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}
	var result *tmdb.SearchResult
	var err error
	switch args.Type {
	case "movie":
		result, err = e.tmdb.SearchMovies(ctx, args.Query, 1)
	case "tv":
		result, err = e.tmdb.SearchTV(ctx, args.Query, 1)
	default:
		result, err = e.tmdb.SearchMulti(ctx, args.Query, 1)
	}
	if err != nil {
		return "", err
	}
	b, _ := json.Marshal(result.Results)
	return string(b), nil
}
```

- [ ] **Step 2: Write `backend/internal/chat/agent.go`**

```go
package chat

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"movie-tracker/internal/tmdb"
	"movie-tracker/internal/watchlist"
)

const systemPrompt = `You are a personal movie and TV series assistant. 
You have access to the user's watchlist and viewing history through tools.
Use them to answer questions, provide recommendations, and help the user 
decide what to watch next. When making recommendations, consider ratings 
the user has given to previously watched content.`

// Agent manages chat completions with function-calling and session history.
type Agent struct {
	client   *openai.Client
	executor *toolExecutor
	sessions *SessionStore
}

func NewAgent(apiKey string, repo *watchlist.Repository, tmdbClient *tmdb.Client) *Agent {
	return &Agent{
		client:   openai.NewClient(option.WithAPIKey(apiKey)),
		executor: newToolExecutor(repo, tmdbClient),
		sessions: NewSessionStore(),
	}
}

// Chat sends a user message, runs tool calls as needed, and returns the final reply.
func (a *Agent) Chat(ctx context.Context, sessionID, userMessage string) (string, error) {
	a.sessions.Append(sessionID, openai.UserMessage(userMessage))

	for {
		history := a.sessions.Get(sessionID)
		messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(history)+1)
		messages = append(messages, openai.SystemMessage(systemPrompt))
		messages = append(messages, history...)

		resp, err := a.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model:    openai.F(openai.ChatModelGPT4oMini),
			Messages: openai.F(messages),
			Tools:    openai.F(toolDefinitions),
		})
		if err != nil {
			return "", fmt.Errorf("openai: %w", err)
		}

		choice := resp.Choices[0]

		if choice.FinishReason == openai.ChatCompletionChoicesFinishReasonToolCalls {
			// Append the assistant message containing tool_calls
			a.sessions.Append(sessionID, choice.Message.ToParam())

			// Execute every requested tool and append results
			for _, tc := range choice.Message.ToolCalls {
				result := a.executor.execute(ctx, tc.Function.Name, tc.Function.Arguments)
				a.sessions.Append(sessionID, openai.ToolMessage(tc.ID, result))
			}
			continue // loop: send results back to model
		}

		// Natural language response
		content := choice.Message.Content
		a.sessions.Append(sessionID, openai.AssistantMessage(content))
		return content, nil
	}
}

// ClearSession deletes a session's history.
func (a *Agent) ClearSession(sessionID string) {
	a.sessions.Delete(sessionID)
}
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/chat/
git commit -m "feat: OpenAI chat agent with function-calling tools"
```

---

## Task 14: Chat handler

**Files:**
- Create: `backend/internal/api/handlers/chat.go`

- [ ] **Step 1: Write `backend/internal/api/handlers/chat.go`**

```go
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"movie-tracker/internal/api"
	"movie-tracker/internal/chat"
)

type ChatHandler struct {
	agent *chat.Agent
}

func NewChatHandler(agent *chat.Agent) *ChatHandler {
	return &ChatHandler{agent: agent}
}

func (h *ChatHandler) Send(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SessionID string `json:"session_id"`
		Message   string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Message == "" {
		api.Error(w, http.StatusBadRequest, "message is required")
		return
	}

	sessionID := body.SessionID
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	reply, err := h.agent.Chat(r.Context(), sessionID, body.Message)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "agent error: "+err.Error())
		return
	}

	api.JSON(w, http.StatusOK, map[string]string{
		"session_id": sessionID,
		"reply":      reply,
	})
}

func (h *ChatHandler) ClearSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		api.Error(w, http.StatusBadRequest, "sessionId is required")
		return
	}
	h.agent.ClearSession(sessionID)
	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/api/handlers/chat.go
git commit -m "feat: chat handler — send message and clear session"
```

---

## Task 15: CORS middleware and router

**Files:**
- Create: `backend/internal/api/middleware/cors.go`
- Create: `backend/internal/api/router.go`

- [ ] **Step 1: Write `backend/internal/api/middleware/cors.go`**

```go
package middleware

import "net/http"

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
```

- [ ] **Step 2: Write `backend/internal/api/router.go`**

```go
package api

import (
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"movie-tracker/internal/api/handlers"
	"movie-tracker/internal/api/middleware"
	"movie-tracker/internal/cache"
	"movie-tracker/internal/chat"
	"movie-tracker/internal/tmdb"
	"movie-tracker/internal/watchlist"
)

func NewRouter(
	tmdbClient *tmdb.Client,
	cacheClient *cache.Cache,
	repo *watchlist.Repository,
	agent *chat.Agent,
) *chi.Mux {
	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(middleware.CORS)

	searchH := handlers.NewSearchHandler(tmdbClient, cacheClient)
	mediaH := handlers.NewMediaHandler(tmdbClient, cacheClient)
	wlH := handlers.NewWatchlistHandler(repo)
	chatH := handlers.NewChatHandler(agent)

	r.Route("/api", func(r chi.Router) {
		r.Get("/search", searchH.Search)

		r.Get("/media/movie/{id}", mediaH.GetMovie)
		r.Get("/media/tv/{id}", mediaH.GetTV)

		r.Route("/watchlist", func(r chi.Router) {
			r.Get("/", wlH.List)
			r.Post("/", wlH.Add)
			r.Get("/{id}", wlH.GetOne)
			r.Delete("/{id}", wlH.Delete)
			r.Post("/{id}/watch", wlH.MarkWatched)
			r.Delete("/{id}/watch", wlH.UnmarkWatched)
		})

		r.Post("/chat", chatH.Send)
		r.Delete("/chat/{sessionId}", chatH.ClearSession)
	})

	return r
}
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/api/
git commit -m "feat: chi router with all endpoints and CORS middleware"
```

---

## Task 16: Main entry point

**Files:**
- Create: `backend/cmd/server/main.go`

- [ ] **Step 1: Write `backend/cmd/server/main.go`**

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"movie-tracker/internal/api"
	"movie-tracker/internal/cache"
	"movie-tracker/internal/chat"
	"movie-tracker/internal/config"
	"movie-tracker/internal/db"
	"movie-tracker/internal/tmdb"
	"movie-tracker/internal/watchlist"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	pool, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatalf("db migrate: %v", err)
	}
	log.Println("migrations applied")

	cacheClient, err := cache.New(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	if err := cacheClient.Ping(ctx); err != nil {
		log.Printf("redis ping failed (cache will be skipped): %v", err)
	}

	tmdbClient := tmdb.New(cfg.TMDBAPIKey, cfg.TMDBBaseURL)
	repo := watchlist.NewRepository(pool)
	agent := chat.NewAgent(cfg.OpenAIKey, repo, tmdbClient)

	router := api.NewRouter(tmdbClient, cacheClient, repo, agent)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server: %v", err)
	}
}
```

- [ ] **Step 2: Verify the backend compiles**

```bash
cd backend && go build ./...
```

Expected: no output (success). Fix any import errors before continuing.

- [ ] **Step 3: Commit**

```bash
git add backend/cmd/server/main.go
git commit -m "feat: main entry point — wires all dependencies and starts server"
```

---

## Task 17: Frontend scaffold

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/vite.config.js`
- Create: `frontend/index.html`
- Create: `frontend/src/main.jsx`
- Create: `frontend/src/App.jsx`
- Create: `frontend/src/App.css`

- [ ] **Step 1: Create `frontend/package.json`**

```json
{
  "name": "movie-tracker-frontend",
  "private": true,
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "react": "^18.3.1",
    "react-dom": "^18.3.1"
  },
  "devDependencies": {
    "@vitejs/plugin-react": "^4.3.1",
    "vite": "^5.4.0"
  }
}
```

- [ ] **Step 2: Create `frontend/vite.config.js`**

```js
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080'
    }
  }
})
```

- [ ] **Step 3: Create `frontend/index.html`**

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Movie Tracker</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.jsx"></script>
  </body>
</html>
```

- [ ] **Step 4: Create `frontend/src/main.jsx`**

```jsx
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from './App.jsx'
import './App.css'

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
```

- [ ] **Step 5: Create `frontend/src/App.jsx`**

```jsx
import { useState } from 'react'
import SearchBar from './components/SearchBar.jsx'
import SearchResults from './components/SearchResults.jsx'
import WatchlistView from './components/WatchlistView.jsx'
import ChatWindow from './components/ChatWindow.jsx'

export default function App() {
  const [activeTab, setActiveTab] = useState('search')
  const [searchResults, setSearchResults] = useState(null)
  const [watchlistRefresh, setWatchlistRefresh] = useState(0)

  function triggerWatchlistRefresh() {
    setWatchlistRefresh(n => n + 1)
  }

  return (
    <div className="app">
      <header className="app-header">
        <h1>Movie Tracker</h1>
        <nav>
          <button
            className={activeTab === 'search' ? 'active' : ''}
            onClick={() => setActiveTab('search')}
          >Search</button>
          <button
            className={activeTab === 'watchlist' ? 'active' : ''}
            onClick={() => setActiveTab('watchlist')}
          >Watchlist</button>
          <button
            className={activeTab === 'chat' ? 'active' : ''}
            onClick={() => setActiveTab('chat')}
          >Chat</button>
        </nav>
      </header>

      <main className="app-main">
        {activeTab === 'search' && (
          <div>
            <SearchBar onResults={setSearchResults} />
            {searchResults && (
              <SearchResults
                results={searchResults}
                onAdded={triggerWatchlistRefresh}
              />
            )}
          </div>
        )}
        {activeTab === 'watchlist' && (
          <WatchlistView key={watchlistRefresh} />
        )}
        {activeTab === 'chat' && (
          <ChatWindow />
        )}
      </main>
    </div>
  )
}
```

- [ ] **Step 6: Create `frontend/src/App.css`**

```css
* { box-sizing: border-box; margin: 0; padding: 0; }
body { font-family: system-ui, sans-serif; background: #0f0f0f; color: #e0e0e0; }
.app { max-width: 960px; margin: 0 auto; padding: 1rem; }
.app-header { display: flex; align-items: center; justify-content: space-between; padding: 1rem 0; border-bottom: 1px solid #333; margin-bottom: 1.5rem; }
.app-header h1 { font-size: 1.5rem; }
nav { display: flex; gap: 0.5rem; }
nav button { background: #1e1e1e; color: #ccc; border: 1px solid #333; padding: 0.4rem 1rem; border-radius: 4px; cursor: pointer; }
nav button.active { background: #e50914; color: white; border-color: #e50914; }
button { cursor: pointer; }
.app-main { min-height: 80vh; }
input, select, textarea { background: #1e1e1e; color: #e0e0e0; border: 1px solid #444; border-radius: 4px; padding: 0.5rem; }
```

- [ ] **Step 7: Install dependencies**

```bash
cd frontend && npm install
```

- [ ] **Step 8: Commit**

```bash
git add frontend/
git commit -m "feat: React frontend scaffold with tab navigation"
```

---

## Task 18: Frontend API client

**Files:**
- Create: `frontend/src/api/client.js`

- [ ] **Step 1: Write `frontend/src/api/client.js`**

```js
const BASE = import.meta.env.VITE_API_BASE_URL || ''

async function request(method, path, body) {
  const opts = {
    method,
    headers: { 'Content-Type': 'application/json' },
  }
  if (body !== undefined) opts.body = JSON.stringify(body)

  const res = await fetch(BASE + path, opts)
  if (res.status === 204) return null

  const data = await res.json()
  if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`)
  return data
}

export const api = {
  search: (q, type = 'all', page = 1) =>
    request('GET', `/api/search?q=${encodeURIComponent(q)}&type=${type}&page=${page}`),

  getMovie: (id) => request('GET', `/api/media/movie/${id}`),
  getTV: (id) => request('GET', `/api/media/tv/${id}`),

  getWatchlist: (params = {}) => {
    const qs = new URLSearchParams(params).toString()
    return request('GET', `/api/watchlist${qs ? '?' + qs : ''}`)
  },
  addToWatchlist: (item) => request('POST', '/api/watchlist', item),
  removeFromWatchlist: (id) => request('DELETE', `/api/watchlist/${id}`),
  markWatched: (id, data) => request('POST', `/api/watchlist/${id}/watch`, data),
  unmarkWatched: (id) => request('DELETE', `/api/watchlist/${id}/watch`),

  chat: (sessionId, message) =>
    request('POST', '/api/chat', { session_id: sessionId, message }),
  clearChat: (sessionId) => request('DELETE', `/api/chat/${sessionId}`),
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/api/client.js
git commit -m "feat: frontend API client for all backend endpoints"
```

---

## Task 19: Search components

**Files:**
- Create: `frontend/src/components/SearchBar.jsx`
- Create: `frontend/src/components/SearchResults.jsx`

- [ ] **Step 1: Write `frontend/src/components/SearchBar.jsx`**

```jsx
import { useState } from 'react'
import { api } from '../api/client.js'

export default function SearchBar({ onResults }) {
  const [query, setQuery] = useState('')
  const [type, setType] = useState('all')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function handleSearch(e) {
    e.preventDefault()
    if (!query.trim()) return
    setLoading(true)
    setError('')
    try {
      const results = await api.search(query.trim(), type)
      onResults(results)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSearch} style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem' }}>
      <input
        type="text"
        value={query}
        onChange={e => setQuery(e.target.value)}
        placeholder="Search movies or series..."
        style={{ flex: 1 }}
      />
      <select value={type} onChange={e => setType(e.target.value)}>
        <option value="all">All</option>
        <option value="movie">Movies</option>
        <option value="tv">TV Series</option>
      </select>
      <button type="submit" disabled={loading}
        style={{ background: '#e50914', color: 'white', border: 'none', padding: '0.5rem 1.2rem', borderRadius: '4px' }}>
        {loading ? 'Searching...' : 'Search'}
      </button>
      {error && <span style={{ color: '#ff6b6b', alignSelf: 'center' }}>{error}</span>}
    </form>
  )
}
```

- [ ] **Step 2: Write `frontend/src/components/SearchResults.jsx`**

```jsx
import { useState } from 'react'
import { api } from '../api/client.js'

const IMG_BASE = 'https://image.tmdb.org/t/p/w200'

export default function SearchResults({ results, onAdded }) {
  const [adding, setAdding] = useState(null)
  const [messages, setMessages] = useState({})

  async function handleAdd(item) {
    const title = item.title || item.name
    setAdding(item.id)
    try {
      await api.addToWatchlist({
        tmdb_id: item.id,
        media_type: item.media_type || 'movie',
        title,
        poster_path: item.poster_path || '',
        overview: item.overview || '',
      })
      setMessages(m => ({ ...m, [item.id]: 'Added!' }))
      onAdded()
    } catch (err) {
      setMessages(m => ({ ...m, [item.id]: err.message }))
    } finally {
      setAdding(null)
    }
  }

  if (!results || results.results?.length === 0) {
    return <p style={{ color: '#888' }}>No results found.</p>
  }

  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(160px, 1fr))', gap: '1rem' }}>
      {results.results.map(item => {
        const title = item.title || item.name
        return (
          <div key={item.id} style={{ background: '#1e1e1e', borderRadius: '8px', overflow: 'hidden' }}>
            {item.poster_path
              ? <img src={IMG_BASE + item.poster_path} alt={title} style={{ width: '100%' }} />
              : <div style={{ height: '240px', background: '#333', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#666' }}>No image</div>
            }
            <div style={{ padding: '0.5rem' }}>
              <p style={{ fontSize: '0.85rem', marginBottom: '0.4rem', fontWeight: 600 }}>{title}</p>
              <p style={{ fontSize: '0.75rem', color: '#888', marginBottom: '0.5rem' }}>
                {item.media_type === 'tv' ? 'TV' : 'Movie'} • ★ {item.vote_average?.toFixed(1)}
              </p>
              <button
                onClick={() => handleAdd(item)}
                disabled={adding === item.id}
                style={{ width: '100%', background: '#e50914', color: 'white', border: 'none', padding: '0.3rem', borderRadius: '4px', fontSize: '0.8rem' }}
              >
                {messages[item.id] || '+ Watchlist'}
              </button>
            </div>
          </div>
        )
      })}
    </div>
  )
}
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/SearchBar.jsx frontend/src/components/SearchResults.jsx
git commit -m "feat: search UI components"
```

---

## Task 20: Watchlist components

**Files:**
- Create: `frontend/src/components/WatchlistView.jsx`
- Create: `frontend/src/components/WatchlistItem.jsx`
- Create: `frontend/src/components/WatchModal.jsx`

- [ ] **Step 1: Write `frontend/src/components/WatchModal.jsx`**

```jsx
import { useState } from 'react'

export default function WatchModal({ item, onConfirm, onClose }) {
  const today = new Date().toISOString().split('T')[0]
  const [watchedAt, setWatchedAt] = useState(
    item.watched?.watched_at || today
  )
  const [rating, setRating] = useState(item.watched?.rating || 7)

  return (
    <div style={{
      position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.7)',
      display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 100
    }}>
      <div style={{ background: '#1e1e1e', padding: '1.5rem', borderRadius: '8px', minWidth: '320px' }}>
        <h3 style={{ marginBottom: '1rem' }}>Mark as Watched</h3>
        <p style={{ marginBottom: '1rem', color: '#aaa' }}>{item.title}</p>

        <label style={{ display: 'block', marginBottom: '0.8rem' }}>
          <span style={{ fontSize: '0.85rem', color: '#aaa' }}>Date watched</span>
          <input type="date" value={watchedAt} onChange={e => setWatchedAt(e.target.value)}
            style={{ display: 'block', width: '100%', marginTop: '0.3rem' }} />
        </label>

        <label style={{ display: 'block', marginBottom: '1rem' }}>
          <span style={{ fontSize: '0.85rem', color: '#aaa' }}>Rating: {rating}/10</span>
          <input type="range" min="1" max="10" value={rating} onChange={e => setRating(Number(e.target.value))}
            style={{ display: 'block', width: '100%', marginTop: '0.3rem' }} />
        </label>

        <div style={{ display: 'flex', gap: '0.5rem', justifyContent: 'flex-end' }}>
          <button onClick={onClose} style={{ background: '#333', color: '#ccc', border: 'none', padding: '0.4rem 1rem', borderRadius: '4px' }}>
            Cancel
          </button>
          <button onClick={() => onConfirm(watchedAt, rating)}
            style={{ background: '#e50914', color: 'white', border: 'none', padding: '0.4rem 1rem', borderRadius: '4px' }}>
            Save
          </button>
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Write `frontend/src/components/WatchlistItem.jsx`**

```jsx
import { useState } from 'react'
import { api } from '../api/client.js'
import WatchModal from './WatchModal.jsx'

const IMG_BASE = 'https://image.tmdb.org/t/p/w92'

export default function WatchlistItem({ item, onChanged }) {
  const [showModal, setShowModal] = useState(false)
  const [loading, setLoading] = useState(false)

  async function handleDelete() {
    if (!confirm(`Remove "${item.title}" from watchlist?`)) return
    setLoading(true)
    try {
      await api.removeFromWatchlist(item.id)
      onChanged()
    } finally {
      setLoading(false)
    }
  }

  async function handleWatch(watchedAt, rating) {
    setShowModal(false)
    setLoading(true)
    try {
      await api.markWatched(item.id, { watched_at: watchedAt, rating })
      onChanged()
    } finally {
      setLoading(false)
    }
  }

  async function handleUnwatch() {
    setLoading(true)
    try {
      await api.unmarkWatched(item.id)
      onChanged()
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{
      display: 'flex', gap: '0.75rem', background: '#1e1e1e',
      borderRadius: '8px', padding: '0.75rem', alignItems: 'center'
    }}>
      {item.poster_path
        ? <img src={IMG_BASE + item.poster_path} alt={item.title} style={{ width: 46, borderRadius: 4, flexShrink: 0 }} />
        : <div style={{ width: 46, height: 69, background: '#333', borderRadius: 4, flexShrink: 0 }} />
      }
      <div style={{ flex: 1, minWidth: 0 }}>
        <p style={{ fontWeight: 600, fontSize: '0.9rem', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
          {item.title}
        </p>
        <p style={{ fontSize: '0.75rem', color: '#888' }}>
          {item.media_type === 'tv' ? 'TV' : 'Movie'}
          {item.watched && <> · ★ {item.watched.rating}/10 · {item.watched.watched_at}</>}
        </p>
      </div>
      <div style={{ display: 'flex', gap: '0.4rem', flexShrink: 0 }}>
        {item.watched
          ? <button onClick={handleUnwatch} disabled={loading}
              style={{ background: '#333', color: '#ccc', border: 'none', padding: '0.3rem 0.7rem', borderRadius: 4, fontSize: '0.8rem' }}>
              Unwatch
            </button>
          : <button onClick={() => setShowModal(true)} disabled={loading}
              style={{ background: '#1a6e3a', color: 'white', border: 'none', padding: '0.3rem 0.7rem', borderRadius: 4, fontSize: '0.8rem' }}>
              Mark watched
            </button>
        }
        <button onClick={handleDelete} disabled={loading}
          style={{ background: '#4a1a1a', color: '#ff8080', border: 'none', padding: '0.3rem 0.7rem', borderRadius: 4, fontSize: '0.8rem' }}>
          Remove
        </button>
      </div>
      {showModal && <WatchModal item={item} onConfirm={handleWatch} onClose={() => setShowModal(false)} />}
    </div>
  )
}
```

- [ ] **Step 3: Write `frontend/src/components/WatchlistView.jsx`**

```jsx
import { useState, useEffect } from 'react'
import { api } from '../api/client.js'
import WatchlistItem from './WatchlistItem.jsx'

export default function WatchlistView() {
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [filters, setFilters] = useState({ status: 'all', type: 'all', sort: 'added_at', order: 'desc' })

  async function load() {
    setLoading(true)
    setError('')
    try {
      const data = await api.getWatchlist(filters)
      setItems(data || [])
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [filters])

  return (
    <div>
      <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem', flexWrap: 'wrap' }}>
        <select value={filters.status} onChange={e => setFilters(f => ({ ...f, status: e.target.value }))}>
          <option value="all">All</option>
          <option value="watched">Watched</option>
          <option value="unwatched">Unwatched</option>
        </select>
        <select value={filters.type} onChange={e => setFilters(f => ({ ...f, type: e.target.value }))}>
          <option value="all">All types</option>
          <option value="movie">Movies</option>
          <option value="tv">TV Series</option>
        </select>
        <select value={filters.sort} onChange={e => setFilters(f => ({ ...f, sort: e.target.value }))}>
          <option value="added_at">Date added</option>
          <option value="watched_at">Date watched</option>
          <option value="rating">Rating</option>
          <option value="title">Title</option>
        </select>
        <select value={filters.order} onChange={e => setFilters(f => ({ ...f, order: e.target.value }))}>
          <option value="desc">Desc</option>
          <option value="asc">Asc</option>
        </select>
      </div>

      {loading && <p style={{ color: '#888' }}>Loading...</p>}
      {error && <p style={{ color: '#ff6b6b' }}>{error}</p>}
      {!loading && items.length === 0 && <p style={{ color: '#888' }}>Your watchlist is empty.</p>}

      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
        {items.map(item => (
          <WatchlistItem key={item.id} item={item} onChanged={load} />
        ))}
      </div>
    </div>
  )
}
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/
git commit -m "feat: watchlist components with filter, mark-watched modal"
```

---

## Task 21: Chat component

**Files:**
- Create: `frontend/src/components/ChatWindow.jsx`

- [ ] **Step 1: Write `frontend/src/components/ChatWindow.jsx`**

```jsx
import { useState, useRef, useEffect } from 'react'
import { api } from '../api/client.js'

export default function ChatWindow() {
  const [sessionId, setSessionId] = useState(null)
  const [messages, setMessages] = useState([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const bottomRef = useRef(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  async function handleSend(e) {
    e.preventDefault()
    const text = input.trim()
    if (!text || loading) return

    setInput('')
    setMessages(m => [...m, { role: 'user', content: text }])
    setLoading(true)

    try {
      const res = await api.chat(sessionId, text)
      setSessionId(res.session_id)
      setMessages(m => [...m, { role: 'assistant', content: res.reply }])
    } catch (err) {
      setMessages(m => [...m, { role: 'error', content: err.message }])
    } finally {
      setLoading(false)
    }
  }

  async function handleClear() {
    if (sessionId) await api.clearChat(sessionId)
    setSessionId(null)
    setMessages([])
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '70vh' }}>
      <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: '0.5rem' }}>
        <button onClick={handleClear}
          style={{ background: '#333', color: '#ccc', border: 'none', padding: '0.3rem 0.8rem', borderRadius: '4px', fontSize: '0.8rem' }}>
          Clear chat
        </button>
      </div>

      <div style={{
        flex: 1, overflowY: 'auto', background: '#1e1e1e',
        borderRadius: '8px', padding: '1rem', display: 'flex', flexDirection: 'column', gap: '0.75rem'
      }}>
        {messages.length === 0 && (
          <p style={{ color: '#555', textAlign: 'center', marginTop: '2rem' }}>
            Ask me about your watchlist, get recommendations, or anything about movies and TV.
          </p>
        )}
        {messages.map((msg, i) => (
          <div key={i} style={{
            alignSelf: msg.role === 'user' ? 'flex-end' : 'flex-start',
            maxWidth: '75%',
            background: msg.role === 'user' ? '#e50914' : msg.role === 'error' ? '#4a1a1a' : '#2d2d2d',
            color: msg.role === 'error' ? '#ff8080' : 'white',
            padding: '0.6rem 0.9rem',
            borderRadius: '12px',
            fontSize: '0.9rem',
            lineHeight: '1.4',
            whiteSpace: 'pre-wrap',
          }}>
            {msg.content}
          </div>
        ))}
        {loading && (
          <div style={{ alignSelf: 'flex-start', color: '#888', fontSize: '0.85rem' }}>
            Thinking...
          </div>
        )}
        <div ref={bottomRef} />
      </div>

      <form onSubmit={handleSend} style={{ display: 'flex', gap: '0.5rem', marginTop: '0.5rem' }}>
        <input
          type="text"
          value={input}
          onChange={e => setInput(e.target.value)}
          placeholder="Type a message..."
          disabled={loading}
          style={{ flex: 1 }}
        />
        <button type="submit" disabled={loading || !input.trim()}
          style={{ background: '#e50914', color: 'white', border: 'none', padding: '0.5rem 1.2rem', borderRadius: '4px' }}>
          Send
        </button>
      </form>
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/ChatWindow.jsx
git commit -m "feat: chat window component with session management"
```

---

## Task 22: Frontend Dockerfile and nginx config

**Files:**
- Create: `frontend/nginx.conf`
- Create: `frontend/Dockerfile`

- [ ] **Step 1: Write `frontend/nginx.conf`**

```nginx
server {
    listen 80;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

- [ ] **Step 2: Write `frontend/Dockerfile`**

```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm install
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
```

- [ ] **Step 3: Commit**

```bash
git add frontend/Dockerfile frontend/nginx.conf
git commit -m "feat: frontend multi-stage Dockerfile with nginx proxy"
```

---

## Task 23: End-to-end smoke test with docker-compose

- [ ] **Step 1: Copy `.env.example` to `.env` and fill in real keys**

```bash
cp .env.example .env
# Edit .env: set TMDB_API_KEY and OPENAI_API_KEY to real values
```

- [ ] **Step 2: Build and start all services**

```bash
docker compose up --build
```

Expected: all 4 services start, backend logs `migrations applied` and `server listening on :8080`.

- [ ] **Step 3: Verify backend health**

```bash
curl -s "http://localhost:8080/api/search?q=inception&type=movie" | head -c 200
```

Expected: JSON with `results` array containing movie data.

- [ ] **Step 4: Verify watchlist endpoints**

```bash
# Add an item
curl -s -X POST http://localhost:8080/api/watchlist \
  -H "Content-Type: application/json" \
  -d '{"tmdb_id":27205,"media_type":"movie","title":"Inception","poster_path":"/path.jpg","overview":"A thief..."}' | python3 -m json.tool

# List watchlist
curl -s "http://localhost:8080/api/watchlist" | python3 -m json.tool
```

Expected: item returned on add, item in list.

- [ ] **Step 5: Open frontend at `http://localhost:5173` and test**

- Search for "Inception", add it to watchlist
- Switch to Watchlist tab, verify it appears
- Mark it as watched with a rating
- Switch to Chat, ask "What have I watched?" — the agent should call a tool and respond with your watched items

- [ ] **Step 6: Commit final state**

```bash
git add -A
git commit -m "chore: verified end-to-end with docker-compose"
```

---

## Task 24: README

**Files:**
- Create: `README.md`

- [ ] **Step 1: Write `README.md`**

```markdown
# Movie Tracker

Personal movie and TV series tracker with AI chat assistant.

## Prerequisites

- Docker and Docker Compose
- TMDB API key — register at https://www.themoviedb.org
- OpenAI API key

## Setup

```bash
cp .env.example .env
# Edit .env and fill in TMDB_API_KEY and OPENAI_API_KEY
```

## Run

```bash
docker compose up --build
```

- Backend API: http://localhost:8080
- Frontend: http://localhost:5173

On first start the backend runs migrations automatically. No manual steps needed.

## Database Schema

### `watchlist_items`
| Column | Type | Notes |
|---|---|---|
| id | SERIAL PK | |
| tmdb_id | INTEGER | TMDB content ID |
| media_type | VARCHAR(10) | `movie` or `tv` |
| title | VARCHAR(255) | |
| poster_path | VARCHAR(255) | TMDB poster path |
| overview | TEXT | |
| added_at | TIMESTAMPTZ | default NOW() |

Unique constraint on `(tmdb_id, media_type)`.

### `watch_records`
| Column | Type | Notes |
|---|---|---|
| id | SERIAL PK | |
| watchlist_item_id | INTEGER FK | → watchlist_items.id, CASCADE DELETE |
| watched_at | DATE | user-supplied date |
| rating | SMALLINT | 1–10 |
| created_at | TIMESTAMPTZ | default NOW() |

Unique constraint on `watchlist_item_id` (one watch record per item).

**Relation:** one-to-one from `watch_records` to `watchlist_items`.

## API Endpoints

| Method | Path | Description |
|---|---|---|
| GET | `/api/search?q=&type=&page=` | Search TMDB. `type`: all/movie/tv. Cached 10 min. |
| GET | `/api/media/movie/:id` | TMDB movie detail. Cached 24h. |
| GET | `/api/media/tv/:id` | TMDB TV detail. Cached 24h. |
| GET | `/api/watchlist` | List watchlist. Filters: `status`, `type`, `sort`, `order`. |
| POST | `/api/watchlist` | Add item. Body: `{tmdb_id, media_type, title, poster_path, overview}`. |
| GET | `/api/watchlist/:id` | Get single watchlist item. |
| DELETE | `/api/watchlist/:id` | Remove item. |
| POST | `/api/watchlist/:id/watch` | Mark watched. Body: `{watched_at, rating}`. Upserts. |
| DELETE | `/api/watchlist/:id/watch` | Unmark watched. |
| POST | `/api/chat` | Send message. Body: `{session_id?, message}`. Returns `{session_id, reply}`. |
| DELETE | `/api/chat/:sessionId` | Clear session history. |

## AI Agent

### Tools

| Tool | When used |
|---|---|
| `get_watchlist` | User asks "what's on my list" or agent needs full list context |
| `get_watched_with_ratings` | Agent needs to understand user taste before recommending |
| `get_watchlist_item` | User asks about a specific item by reference |
| `search_tmdb` | Agent wants to recommend content not already on the list |

The agent **never** receives the full watchlist in the system prompt. It calls tools on demand, keeping token cost proportional to what it actually needs.

### Session history

Stored in-memory (`sync.Map`) keyed by `session_id` (UUID). History is retained for the server process lifetime. A new `session_id` starts a fresh conversation. Call `DELETE /api/chat/:sessionId` to reset manually.

## Architecture Decisions

### 1. Raw SQL with pgx/v5 (no ORM)

All database access uses handwritten SQL via `pgx`. The spec forbids ORMs, but this is also the right call: the watchlist query needs a dynamic `WHERE` + `ORDER BY` clause built at runtime based on user filter params. An ORM's query builder would be equally complex but with more indirection.

### 2. Frontend as a separate nginx service

The React app is built in a `node:alpine` stage and served by `nginx:alpine` (~15MB image). The Go binary serves only the API. This mirrors a real production deployment where static assets go to a CDN. nginx proxies `/api/*` to the backend container by name, keeping networking clean.

### 3. Redis failure is non-fatal

If Redis is unavailable, all cache operations are silently skipped and the request falls through to TMDB. This is handled by checking `err == redis.Nil` vs a real error and proceeding in both paths without propagating to the HTTP response.

### 4. Chat history stays in memory, not PostgreSQL

The spec says "maintain history during the session" — not across restarts. Adding a `chat_sessions` table would require schema, serialization, and TTL cleanup for zero benefit per the stated requirement. A `sync.Map` with a copy-on-read pattern is safe and correct.

### 5. SQL-side filtering and sorting

`GET /api/watchlist` builds a parameterized SQL query at runtime rather than fetching all rows and filtering in Go. This is safe (no string interpolation of user values), correct at scale, and explicitly required by the spec.
```

- [ ] **Step 2: Commit**

```bash
git add README.md
git commit -m "docs: comprehensive README with schema, endpoints, agent, and decisions"
```
