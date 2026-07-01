package facade4listus

import (
	"context"
	"fmt"

	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/listus/backend/movius/tmdbclient"
)

// movieResolver is the seam movius/tmdbclient is called through - lets tests
// (and, in future, the listusbot / AI-identify callers) substitute their own
// client. Defaults to reading TMDB_API_KEY from env; falls back to MOCK data
// automatically when the key is absent (see tmdbclient.Client.IsMock()).
type movieResolver interface {
	SearchMovie(ctx context.Context, query string) ([]tmdbclient.MovieSummary, error)
	SearchPerson(ctx context.Context, actor string) ([]tmdbclient.MovieSummary, error)
	Resolve(ctx context.Context, tmdbID int) (tmdbclient.MovieDetails, error)
}

var movieClient movieResolver = tmdbclient.NewClientFromEnv()

// SearchMovies searches TMDB by title, merged with actor/person search
// results (a query can match either a movie title or an actor name).
func SearchMovies(ctx context.Context, request dto4listus.MovieSearchRequest) (response dto4listus.MovieSearchResponse, err error) {
	if err = request.Validate(); err != nil {
		return
	}
	byTitle, err := movieClient.SearchMovie(ctx, request.Query)
	if err != nil {
		err = fmt.Errorf("failed to search movies by title: %w", err)
		return
	}
	byPerson, err := movieClient.SearchPerson(ctx, request.Query)
	if err != nil {
		err = fmt.Errorf("failed to search movies by actor: %w", err)
		return
	}
	response.Movies = mergeMovieSummaries(byTitle, byPerson)
	return
}

// ResolveMovie fully resolves a single movie (overview, poster, cast, trailer) by its TMDB id.
func ResolveMovie(ctx context.Context, request dto4listus.MovieResolveRequest) (response dto4listus.MovieResolveResponse, err error) {
	if err = request.Validate(); err != nil {
		return
	}
	response.Movie, err = movieClient.Resolve(ctx, request.TmdbID)
	if err != nil {
		err = fmt.Errorf("failed to resolve movie tmdbID=%d: %w", request.TmdbID, err)
	}
	return
}

func mergeMovieSummaries(lists ...[]tmdbclient.MovieSummary) []tmdbclient.MovieSummary {
	seen := make(map[int]bool)
	var out []tmdbclient.MovieSummary
	for _, list := range lists {
		for _, m := range list {
			if seen[m.TmdbID] {
				continue
			}
			seen[m.TmdbID] = true
			out = append(out, m)
		}
	}
	return out
}
