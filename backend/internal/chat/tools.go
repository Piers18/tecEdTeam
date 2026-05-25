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
