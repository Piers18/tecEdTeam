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
