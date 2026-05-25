# Movie Tracker

Personal movie and TV series tracker with AI chat assistant.

## Screenshots

### Search
Search movies and TV series via TMDB, then add them to your watchlist in one click.

![Search](docs/screenshots/search.png)

### Watchlist
Filter by status, type, sort order, and mark items as watched with a rating.

![Watchlist](docs/screenshots/watchlist.png)

### AI Chat
Ask the assistant about your watchlist, get personalized recommendations, and explore movies.

![Chat](docs/screenshots/chat.png)

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

If Redis is unavailable, all cache operations are silently skipped and the request falls through to TMDB. This is handled by checking whether Get returns an error — on any error (including redis.Nil for misses), we proceed without propagating to the HTTP response.

### 4. Chat history stays in memory, not PostgreSQL

The spec says "maintain history during the session" — not across restarts. Adding a `chat_sessions` table would require schema, serialization, and TTL cleanup for zero benefit per the stated requirement. A `sync.Map` with a copy-on-read pattern is safe and correct.

### 5. SQL-side filtering and sorting

`GET /api/watchlist` builds a parameterized SQL query at runtime rather than fetching all rows and filtering in Go. This is safe (no string interpolation of user values — all values go through `$N` placeholders), correct at scale, and explicitly required by the spec.

### 6. Redis TTL strategy

Search results: 10-minute TTL (required by spec). Movie/TV detail: 24-hour TTL. Metadata like titles, posters, and overviews rarely changes for released content; 24 hours ensures freshness without hammering TMDB on every detail view. If TMDB updates a poster, the cache self-heals within a day.
