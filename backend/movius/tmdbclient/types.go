// Package tmdbclient is a small TMDB (The Movie Database) v3 API client used
// to enrich listus "watch" list items with movie metadata: poster, overview,
// release year, cast & trailer.
//
// If TMDB_API_KEY is not configured the client transparently falls back to
// realistic MOCK data (see mock.go) so the whole watch-list feature works
// end-to-end without provisioning a key.
package tmdbclient

// MovieSummary is a lightweight movie search result.
type MovieSummary struct {
	TmdbID    int    `json:"tmdbID"`
	Title     string `json:"title"`
	Year      int    `json:"year,omitempty"`
	PosterURL string `json:"posterURL,omitempty"`
}

// MovieDetails is the fully-enriched movie data used to populate a watch-list item.
type MovieDetails struct {
	MovieSummary
	Overview          string   `json:"overview,omitempty"`
	TrailerYouTubeKey string   `json:"trailerYouTubeKey,omitempty"`
	Cast              []string `json:"cast,omitempty"` // top ~5 cast member names
}
