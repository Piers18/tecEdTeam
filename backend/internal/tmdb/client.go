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
