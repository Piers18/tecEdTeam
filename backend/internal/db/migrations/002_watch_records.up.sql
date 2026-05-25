CREATE TABLE IF NOT EXISTS watch_records (
    id                  SERIAL PRIMARY KEY,
    watchlist_item_id   INTEGER NOT NULL REFERENCES watchlist_items(id) ON DELETE CASCADE,
    watched_at          DATE NOT NULL,
    rating              SMALLINT NOT NULL CHECK (rating >= 1 AND rating <= 10),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (watchlist_item_id)
);
