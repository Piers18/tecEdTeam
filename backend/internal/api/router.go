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
