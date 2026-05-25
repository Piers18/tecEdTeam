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
