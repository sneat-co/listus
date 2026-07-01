package facade4listus

import (
	"fmt"

	"github.com/sneat-co/listus/backend/dbo4listus"
	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/listus/backend/movius/tmdbclient"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/strongo/random"
)

// AddMovieToWatchlist resolves a movie (by tmdbID, or by taking the first hit
// of a free-text search query) via movius/tmdbclient and appends it as a
// fully-enriched item to the space's canonical watch!movies list
// (dbo4listus.WatchMoviesListID), creating the list on first use.
func AddMovieToWatchlist(ctx facade.ContextWithUser, request dto4listus.AddMovieToWatchlistRequest) (response dto4listus.AddMovieToWatchlistResponse, err error) {
	if err = request.Validate(); err != nil {
		return
	}

	tmdbID := request.TmdbID
	if tmdbID <= 0 {
		var candidates []tmdbclient.MovieSummary
		if candidates, err = movieClient.SearchMovie(ctx, request.Query); err != nil {
			err = fmt.Errorf("failed to search for movie by query=%q: %w", request.Query, err)
			return
		}
		if len(candidates) == 0 {
			err = fmt.Errorf("no movie found for query=%q", request.Query)
			return
		}
		tmdbID = candidates[0].TmdbID
	}

	var movie tmdbclient.MovieDetails
	if movie, err = movieClient.Resolve(ctx, tmdbID); err != nil {
		err = fmt.Errorf("failed to resolve movie tmdbID=%d: %w", tmdbID, err)
		return
	}

	watchWith := request.WatchWith
	if watchWith == nil {
		watchWith = &dbo4listus.WatchWith{Mode: dbo4listus.WatchWithModeAlone}
	}

	item := dto4listus.CreateListItemRequest{
		// createListItemsTxWorker only auto-generates a random ID when the
		// supplied one collides with an existing item, so an empty ID here
		// would be persisted verbatim and fail ListItemBrief.Validate().
		ID: random.ID(12),
		ListItemBase: dbo4listus.ListItemBase{
			Title:             movie.Title,
			TmdbID:            movie.TmdbID,
			Year:              movie.Year,
			PosterURL:         movie.PosterURL,
			Overview:          movie.Overview,
			TrailerYouTubeKey: movie.TrailerYouTubeKey,
			Cast:              movie.Cast,
			WatchWith:         watchWith,
		},
	}

	createResponse, _, err := CreateListItems(ctx, dto4listus.CreateListItemsRequest{
		ListRequest: dto4listus.ListRequest{
			SpaceRequest: request.SpaceRequest,
			ListID:       dbo4listus.WatchMoviesListID,
		},
		Items: []dto4listus.CreateListItemRequest{item},
	})
	if err != nil {
		err = fmt.Errorf("failed to add movie to watchlist: %w", err)
		return
	}
	if len(createResponse.CreatedItems) > 0 {
		response.Item = createResponse.CreatedItems[0]
	}
	return
}
