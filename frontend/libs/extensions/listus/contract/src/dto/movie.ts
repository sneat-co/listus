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
  // Overview is the movie synopsis. TMDB returns it on search results too, so
  // candidate lists (e.g. the Discover "identify by description" flow) can
  // show a short overview without resolving each candidate.
  overview?: string;
}

// MovieDetails is the fully-enriched movie data used to populate a watch-list item.
export interface MovieDetails extends MovieSummary {
  // Fable refactoring: `overview` moved up into MovieSummary (mirroring the Go
  // change in backend/movius/tmdbclient/types.go) so search/identify
  // candidates carry it; MovieDetails still exposes it via the extends.
  trailerYouTubeKey?: string;
  cast?: string[]; // top ~5 cast member names
}
