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
