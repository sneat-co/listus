// Mirrors the Go tmdbclient types (backend/movius/tmdbclient/types.go) used by
// the TMDB-backed movie search/resolve endpoints. If TMDB_API_KEY is not
// configured server-side, these endpoints transparently return realistic MOCK
// data - the shapes below are unaffected either way.

// MovieSummary is a lightweight movie search result.
export interface MovieSummary {
  tmdbID: number;
  title: string;
  year?: number;
  posterURL?: string;
}

// MovieDetails is the fully-enriched movie data used to populate a watch-list item.
export interface MovieDetails extends MovieSummary {
  overview?: string;
  trailerYouTubeKey?: string;
  cast?: string[]; // top ~5 cast member names
}
