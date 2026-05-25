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

---

## Prerequisites

- **Docker** and **Docker Compose** (v2)
- **TMDB API key** — free registration at https://www.themoviedb.org/settings/api
- **OpenAI API key** — https://platform.openai.com/api-keys

No local Go, Node.js, or PostgreSQL installation required; everything runs inside containers.

---

## Environment Variables

Copy the example file and fill in the two secret keys:

```bash
cp .env.example .env
# Open .env and replace the placeholder values for TMDB_API_KEY and OPENAI_API_KEY
```

### `.env.example` reference

| Variable | Description | Example / Default |
|---|---|---|
| `POSTGRES_USER` | PostgreSQL username | `tracker` |
| `POSTGRES_PASSWORD` | PostgreSQL password | `tracker` |
| `POSTGRES_DB` | PostgreSQL database name | `movietracker` |
| `DATABASE_URL` | Full connection string used by the backend | `postgres://tracker:tracker@postgres:5432/movietracker?sslmode=disable` |
| `REDIS_URL` | Redis connection URL | `redis://redis:6379` |
| `TMDB_API_KEY` | **Required.** Your TMDB v3 API key | `your_tmdb_api_key_here` |
| `TMDB_BASE_URL` | TMDB API base URL | `https://api.themoviedb.org/3` |
| `OPENAI_API_KEY` | **Required.** Your OpenAI API key | `your_openai_api_key_here` |
| `PORT` | HTTP port exposed by the backend | `8080` |

> **Never commit `.env` with real values.** Only `.env.example` (with placeholder values) is tracked in git.

---

## Run

```bash
docker compose up --build
```

| Service | URL |
|---|---|
| Frontend | http://localhost:5173 |
| Backend API | http://localhost:8080 |
| API Docs (Swagger) | Import `docs/swagger.yaml` into https://editor.swagger.io |

On first start the backend runs database migrations automatically. No manual steps needed.

---

## Database Schema

The database has two tables with a one-to-one relationship:

```
watchlist_items (1) ──── (0..1) watch_records
```

### `watchlist_items`

Stores every movie or TV series the user has added to their list.

| Column | Type | Constraints | Notes |
|---|---|---|---|
| `id` | `SERIAL` | PK | Auto-incremented internal ID |
| `tmdb_id` | `INTEGER` | NOT NULL | TMDB content ID |
| `media_type` | `VARCHAR(10)` | NOT NULL, CHECK `movie\|tv` | Discriminates movies from series |
| `title` | `VARCHAR(255)` | NOT NULL | Display title |
| `poster_path` | `VARCHAR(255)` | — | TMDB relative poster path |
| `overview` | `TEXT` | — | Plot summary from TMDB |
| `added_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT NOW() | When the user added the item |

**Unique constraint:** `(tmdb_id, media_type)` — the same TMDB title cannot be added twice.

### `watch_records`

Stores a single watch event per item (date and personal rating).

| Column | Type | Constraints | Notes |
|---|---|---|---|
| `id` | `SERIAL` | PK | |
| `watchlist_item_id` | `INTEGER` | NOT NULL, FK → `watchlist_items.id` CASCADE DELETE | |
| `watched_at` | `DATE` | NOT NULL | User-supplied watch date |
| `rating` | `SMALLINT` | NOT NULL, CHECK 1–10 | Personal score |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT NOW() | Record creation timestamp |

**Unique constraint:** `(watchlist_item_id)` — enforces at most one watch record per item.

**Relation:** one-to-one from `watch_records` → `watchlist_items`. Deleting a watchlist item cascades and removes its watch record automatically.

---

## API Endpoints

Full OpenAPI 3.0 specification: [`docs/swagger.yaml`](docs/swagger.yaml)

### Search & Media

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/search` | Search TMDB. Params: `q` (required), `type` (`all`/`movie`/`tv`), `page`. Results cached 10 min in Redis. |
| `GET` | `/api/media/movie/:id` | TMDB movie detail by TMDB ID. Cached 24 h. |
| `GET` | `/api/media/tv/:id` | TMDB TV series detail by TMDB ID. Cached 24 h. |

### Watchlist

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/watchlist` | List items. Filters: `status` (`all`/`watched`/`unwatched`), `type` (`all`/`movie`/`tv`), `sort` (`added_at`/`title`/`rating`), `order` (`asc`/`desc`). |
| `POST` | `/api/watchlist` | Add item. Body: `{ tmdb_id, media_type, title, poster_path?, overview? }`. Returns `409` if already present. |
| `GET` | `/api/watchlist/:id` | Get single item (includes watch record if watched). |
| `DELETE` | `/api/watchlist/:id` | Remove item (cascades watch record). Returns `204`. |
| `POST` | `/api/watchlist/:id/watch` | Mark as watched (upsert). Body: `{ watched_at: "YYYY-MM-DD", rating: 1-10 }`. |
| `DELETE` | `/api/watchlist/:id/watch` | Unmark as watched. Returns `204`. |

### Chat

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/chat` | Send message. Body: `{ session_id?: uuid, message: string }`. Returns `{ session_id, reply }`. |
| `DELETE` | `/api/chat/:sessionId` | Clear session history. Returns `204`. |

**Error responses** always return JSON `{ "error": "<message>" }` with an appropriate HTTP status code.

---

## AI Chat Agent

### Overview

The agent is built with **GPT-4o Mini** and the OpenAI function-calling API. It acts as a personal watchlist assistant: it can answer questions about what the user has saved, understand their taste from past ratings, and recommend new content via TMDB.

### Tools

| Tool | Parameters | When the agent uses it |
|---|---|---|
| `get_watchlist` | `status?`, `media_type?` | User asks "what's on my list", or agent needs list context before answering |
| `get_watched_with_ratings` | `limit?` | Agent wants to understand user preferences before making recommendations (called first in recommendation flows) |
| `get_watchlist_item` | `id` (required) | User refers to a specific item and the agent needs its full details or watch record |
| `search_tmdb` | `query` (required), `type?` | Agent wants to suggest content not already in the watchlist |

The agent **never** receives the full watchlist in the system prompt. It calls tools on demand, so token cost scales with what the request actually needs — not with watchlist size.

### Conversation History

History is stored **in memory** (`sync.Map`) keyed by `session_id` (UUID). Each entry is an ordered slice of OpenAI message objects (user, assistant, tool calls, tool results).

- History persists for the lifetime of the server process.
- Omitting `session_id` in the request starts a fresh conversation; the new UUID is returned in the response.
- Call `DELETE /api/chat/:sessionId` to reset a session explicitly.
- A copy-on-read pattern (`sync.RWMutex` + slice copy in `Get`) ensures concurrent requests to different sessions do not race.

### Agentic loop

Each `POST /api/chat` request runs an internal loop (capped at **10 rounds**):

1. Append user message to history.
2. Send full history + system prompt to the model.
3. If the model returns `finish_reason: tool_calls` → execute all requested tools, append results, repeat.
4. If the model returns a natural-language response → append to history, return to client.

---

## Architecture Decisions

### 1. Raw SQL with `pgx/v5` — no ORM

All database access uses handwritten SQL via `pgx`. The watchlist query requires a **dynamic `WHERE` + `ORDER BY` clause** built at runtime from user filter parameters. An ORM's query builder would be equally complex but add indirection. `pgx` parameterised queries (`$1`, `$2`, …) are safe by construction — no string interpolation of user-supplied values occurs anywhere.

### 2. Frontend served by a dedicated nginx container

The React app is compiled in a `node:alpine` build stage and the output is served by `nginx:alpine` (~15 MB image). The Go binary serves **only** the JSON API. nginx proxies `/api/*` to the `backend` container by its Compose service name, keeping inter-service networking clean. This mirrors a real production pattern where static assets go to a CDN while the API runs separately.

### 3. Redis failure is non-fatal (graceful degradation)

If Redis is unavailable, all cache operations are silently skipped. `cache.Get` returning any error (including `redis.Nil`) causes the handler to fall through to a live TMDB request. This means a Redis restart never takes down the API — users simply get slightly slower responses during the outage.

### 4. Chat history in memory, not PostgreSQL

The requirement is to maintain history *during the session*, not across server restarts. Persisting chat to a `chat_sessions` table would require schema, serialization, TTL cleanup, and a join on every request — all cost for zero specified benefit. A `sync.Map` with a copy-on-read slice pattern is safe, correct, and operationally simpler.

### 5. SQL-side filtering and sorting

`GET /api/watchlist` builds a parameterized SQL query at runtime instead of fetching all rows and filtering in Go. This is correct at any scale, explicitly required by the spec, and avoids shipping unnecessary data over the network. The sort column is validated against an allowlist before interpolation so there is no SQL injection risk.

### 6. Redis TTL strategy

Search results use a **10-minute TTL** (specified requirement). Movie/TV detail responses use a **24-hour TTL**. Metadata like titles, posters, and overviews rarely changes for released content; 24 hours avoids hammering the TMDB API on every detail view. If TMDB updates a poster the cache self-heals within a day.

---

## Project Structure

```
.
├── backend/
│   ├── cmd/server/          # main.go — wires dependencies, starts HTTP server
│   └── internal/
│       ├── api/
│       │   ├── handlers/    # HTTP handlers: search, media, watchlist, chat
│       │   ├── middleware/  # CORS middleware
│       │   └── router.go    # chi router setup
│       ├── cache/           # Redis wrapper (Get/Set with graceful degradation)
│       ├── chat/            # AI agent, tool definitions, session store
│       ├── config/          # Environment variable loading
│       ├── db/              # pgx pool, migrations runner, SQL migration files
│       ├── respond/         # JSON / error response helpers
│       ├── tmdb/            # TMDB API client
│       └── watchlist/       # Repository (all DB queries), domain types
├── docs/
│   └── swagger.yaml         # OpenAPI 3.0 specification
├── frontend/
│   └── src/                 # React + Vite application
├── docker-compose.yml
├── .env.example
└── README.md
```
