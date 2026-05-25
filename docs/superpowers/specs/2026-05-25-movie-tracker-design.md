# Movie Tracker — Design Document
**Date:** 2026-05-25  
**Level:** Mid-level Technical Test  
**Delivery:** GitHub (24h)

---

## 1. Overview

A personal movie and series tracker. The user can search for content via TMDB, build a personal watchlist, mark items as watched with a rating, and chat with an AI agent that knows their viewing history.

**Single-user, no authentication.** The system is designed for one user; auth is explicitly out of scope.

---

## 2. Stack

| Technology      | Role                                     |
|-----------------|------------------------------------------|
| Go              | Backend language                         |
| PostgreSQL      | Primary database (raw SQL, no ORM)       |
| Redis           | Cache only (search results + details)    |
| OpenAI Go SDK   | AI chat agent with function calling      |
| React           | Frontend (functional, not design-focused)|
| Docker / docker-compose | Execution environment            |
| TMDB API        | External source for movies and series    |

---

## 3. Project Structure

```
movie-tracker/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              # Entry point: wires config, DB, Redis, router
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go            # Env-based config (TMDB key, OpenAI key, DSNs)
│   │   ├── db/
│   │   │   ├── db.go                # pgx pool setup
│   │   │   └── migrations/          # Ordered SQL migration files
│   │   │       ├── 001_create_watchlist_items.sql
│   │   │       └── 002_create_watch_records.sql
│   │   ├── cache/
│   │   │   └── redis.go             # Redis client + helpers (Get/Set/Del with TTL)
│   │   ├── tmdb/
│   │   │   └── client.go            # TMDB HTTP client (search, detail)
│   │   ├── watchlist/
│   │   │   ├── repository.go        # All SQL for watchlist_items and watch_records
│   │   │   ├── service.go           # Business logic (cache-aside for TMDB calls)
│   │   │   └── models.go            # Go structs for DB rows
│   │   ├── chat/
│   │   │   ├── agent.go             # OpenAI client + session history in memory
│   │   │   ├── tools.go             # Tool definitions and dispatch
│   │   │   └── session.go           # In-memory session map (id → []messages)
│   │   └── api/
│   │       ├── router.go            # chi router, mounts all handlers
│   │       ├── handlers/
│   │       │   ├── search.go
│   │       │   ├── media.go
│   │       │   ├── watchlist.go
│   │       │   └── chat.go
│   │       └── middleware/
│   │           └── cors.go
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── SearchBar.jsx
│   │   │   ├── SearchResults.jsx
│   │   │   ├── WatchlistItem.jsx
│   │   │   ├── WatchlistView.jsx
│   │   │   ├── WatchModal.jsx       # Mark as watched + rating
│   │   │   └── ChatWindow.jsx
│   │   ├── api/
│   │   │   └── client.js            # fetch wrappers for each backend endpoint
│   │   ├── App.jsx
│   │   └── main.jsx
│   ├── Dockerfile                   # Multi-stage: node build → nginx serve
│   ├── nginx.conf
│   └── package.json
├── docker-compose.yml
├── .env.example
└── README.md
```

---

## 4. Database Schema

### Table: `watchlist_items`

Represents a piece of content the user has added to their list.

```sql
CREATE TABLE watchlist_items (
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

### Table: `watch_records`

Represents a "watched" event for an item. An item can only be marked watched once (one record per item).

```sql
CREATE TABLE watch_records (
    id                  SERIAL PRIMARY KEY,
    watchlist_item_id   INTEGER NOT NULL REFERENCES watchlist_items(id) ON DELETE CASCADE,
    watched_at          DATE NOT NULL,
    rating              SMALLINT NOT NULL CHECK (rating >= 1 AND rating <= 10),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (watchlist_item_id)
);
```

**Relation:** `watch_records.watchlist_item_id → watchlist_items.id` (1-to-1, cascade delete).

**Design note:** Separating "in watchlist" from "watched" allows clean filtering by status and keeps both concepts independently queryable without nullable columns.

---

## 5. API Endpoints

### Search

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/search` | Search TMDB for movies/series. Params: `q`, `type` (movie\|tv\|all, default all), `page` (default 1). Cached 10 min. |

### Media Detail

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/media/movie/:id` | TMDB movie detail. Cached 24h. |
| `GET` | `/api/media/tv/:id` | TMDB TV detail. Cached 24h. |

### Watchlist

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/watchlist` | List watchlist items. Filters: `status` (all\|watched\|unwatched), `type` (movie\|tv\|all), `sort` (added_at\|watched_at\|rating\|title), `order` (asc\|desc). All filtering and sorting done in SQL. |
| `POST` | `/api/watchlist` | Add item. Body: `{ tmdb_id, media_type, title, poster_path, overview }`. |
| `DELETE` | `/api/watchlist/:id` | Remove item (cascades watch record). |
| `POST` | `/api/watchlist/:id/watch` | Mark as watched. Body: `{ watched_at, rating }`. Upserts watch_records. |
| `DELETE` | `/api/watchlist/:id/watch` | Unmark as watched (deletes watch record). |

### Chat

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/chat` | Send a message. Body: `{ session_id, message }`. Returns AI response. Creates session if new. |
| `DELETE` | `/api/chat/:session_id` | Clear session history from memory. |

---

## 6. Redis Cache Strategy

| Cache Key Pattern | Content | TTL |
|---|---|---|
| `search:{q}:{type}:{page}` | TMDB search results (serialized JSON) | **10 min** (required by spec) |
| `media:movie:{id}` | TMDB movie detail | **24h** — movie metadata rarely changes; short-lived enough to pick up poster updates |
| `media:tv:{id}` | TMDB TV series detail | **24h** — same reasoning |

**Strategy:** Cache-aside. On every TMDB call: check Redis first → on miss, call TMDB → store result → return. On Redis error, fall through to TMDB and log the error (never fail a user request because of cache).

**Redis is not used as primary storage.** The watchlist lives exclusively in PostgreSQL.

---

## 7. AI Agent Design

### Overview

The agent uses OpenAI's chat completions API with function calling. The session message history is stored in memory (Go map keyed by `session_id`). On each `/api/chat` request:

1. Append the user message to the session history.
2. Call OpenAI with the history + tool definitions.
3. If OpenAI returns a `tool_calls` response: execute the tool(s), append results, call OpenAI again.
4. Repeat until OpenAI returns a plain assistant message.
5. Append the final assistant message to history, return it to the frontend.

### Session History

Stored as `[]openai.ChatCompletionMessage` in a Go `sync.Map`. Each session is identified by a `session_id` (UUID) generated client-side or by the backend on first message. History is not persisted to disk — it lives only for the duration of the server process.

### System Prompt

A short, static system prompt describing the agent's role. It does **not** include the user's watchlist. The agent fetches data through tools when it needs it.

```
You are a personal movie and TV series assistant. 
You have access to the user's watchlist and viewing history through tools.
Use them to answer questions, provide recommendations, and help the user 
decide what to watch next. When making recommendations, consider ratings 
the user has given to previously watched content.
```

### Tools

| Tool Name | Description | When used |
|---|---|---|
| `get_watchlist` | Returns all items in the user's watchlist. Optional filter: `status` (all\|watched\|unwatched), `media_type` (movie\|tv\|all). | When the user asks "what's on my list" or the agent needs the full list. |
| `get_watched_with_ratings` | Returns only watched items with their ratings, ordered by rating descending. Optional `limit`. | When the agent needs to understand user taste for recommendations. |
| `get_watchlist_item` | Returns full detail of one watchlist item including watch record. Param: `id` (int). | When the agent needs specifics about one item. |
| `search_tmdb` | Searches TMDB for content to make recommendations. Params: `query` (string), `type` (movie\|tv\|all). | When recommending content not yet on the user's list. |

**Function calling is the only way the agent accesses data.** The full watchlist is never injected into the context on every turn. This keeps prompt cost low and scales with large lists.

---

## 8. Frontend Architecture

### Decision: Separate nginx service

The React app is served by nginx as a separate container in docker-compose, not embedded in the Go binary.

**Reasoning:**
- Cleaner separation of concerns: Go serves the API, nginx serves static assets.
- The React build runs in a `node:alpine` stage; the final image is `nginx:alpine` (~20MB vs bloating the Go image).
- In a real deployment, static assets would go to a CDN — this structure mirrors that.
- No need for the Go server to handle static file routing edge cases.

### Key Pages / Views

1. **Search tab** — search bar + results grid. Each result has "Add to watchlist" button.
2. **Watchlist tab** — filterable list. Each item has "Mark watched" (opens rating modal) and "Remove" actions.
3. **Chat tab** — simple chat window with session management.

The frontend communicates with the backend via `VITE_API_BASE_URL` (set to `http://localhost:8080` in dev, or the backend container name in docker-compose networking).

---

## 9. Architecture Decisions

### Decision 1: `chi` as HTTP router
Standard `net/http` requires manual route parsing for path params. `chi` gives clean route groups, middleware support, and URL params without reflection-based magic. It's thin — just routing, nothing else — which keeps the codebase in control of everything else.

### Decision 2: `pgx/v5` for PostgreSQL, not `database/sql + lib/pq`
`pgx` is faster, has better support for PostgreSQL-specific types, and its connection pool is built-in. `database/sql` is an abstraction for portability across DBs — we don't need that, we're PostgreSQL-only.

### Decision 3: SQL migrations embedded in binary via `embed.FS`, run on startup
Using `golang-migrate` with embedded files means migrations run automatically when the container starts — no manual steps. Migration files are versioned (`001_`, `002_`) and idempotent via `CREATE TABLE IF NOT EXISTS`. This satisfies the "docker-compose up and it works" requirement.

### Decision 4: In-memory chat sessions, not DB-persisted
The spec says "maintain history during the session." Persisting chat to PostgreSQL adds schema, serialization overhead, and TTL cleanup complexity for zero spec-required benefit. An in-memory `sync.Map` is correct and simpler. If the server restarts, the session is lost — that's acceptable per the spec.

### Decision 5: Filter and sort watchlist in SQL, not Go
Fetching all rows and filtering in Go is an anti-pattern at scale. The spec explicitly calls this out. The `GET /api/watchlist` handler builds a parameterized SQL query dynamically (safe, no string interpolation of user input) based on the filter/sort params.

### Decision 6: Redis failure is non-fatal
If Redis is down, the system falls back to calling TMDB directly. Losing cache never fails a user request — it only makes it slower. Redis errors are logged but not propagated to the HTTP response. This is the right trade-off for a cache layer.

---

## 10. Docker Compose Services

```yaml
services:
  postgres:     # PostgreSQL 15 with health check
  redis:        # Redis 7 alpine
  backend:      # Go app, depends_on postgres + redis, runs migrations on start
  frontend:     # nginx serving built React app
```

The backend does not start accepting connections until postgres health check passes (`pg_isready`).

---

## 11. Environment Variables (.env.example)

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

---

## 12. Error Handling

- TMDB errors: if Redis has stale data, return it with a warning header. If no cache, return `502 Bad Gateway` with descriptive JSON.
- Validation errors: `400 Bad Request` with a JSON body `{ "error": "..." }`.
- Not found: `404` with JSON.
- Internal errors: `500` with JSON, error logged server-side (never expose stack traces to client).
- All errors follow the shape: `{ "error": "human-readable message" }`.

---

## 13. Out of Scope

- Authentication / multi-user
- Pagination on watchlist (can be added; not required)
- Persistent chat history across server restarts
- Rate limiting on TMDB or OpenAI calls
- HTTPS / TLS (local docker-compose setup)
