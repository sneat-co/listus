package dto4listus

import (
	"strings"

	"github.com/sneat-co/listus/backend/dbo4listus"
	"github.com/sneat-co/listus/backend/movius/tmdbclient"
	"github.com/sneat-co/sneat-core-modules/spaceus/dto4spaceus"
	"github.com/strongo/validation"
)

// MovieSearchRequest searches TMDB by movie title and/or actor name. It is a
// read-only proxy over TMDB, so it does not require a space context - any
// authenticated user may search.
type MovieSearchRequest struct {
	Query string `json:"query"`
}

// Validate returns error if not valid
func (v MovieSearchRequest) Validate() error {
	if strings.TrimSpace(v.Query) == "" {
		return validation.NewErrRequestIsMissingRequiredField("query")
	}
	return nil
}

// MovieSearchResponse returns merged title-search + actor-search candidates.
type MovieSearchResponse struct {
	Movies []tmdbclient.MovieSummary `json:"movies"`
}

// MovieIdentifyRequest identifies movies from a vague natural-language
// description ("that submarine movie with the guy from Titanic") via AI title
// guesses grounded through TMDB search. Like MovieSearchRequest it is a
// read-only proxy, so it does not require a space context.
type MovieIdentifyRequest struct {
	Description string `json:"description"`
}

// Validate returns error if not valid
func (v MovieIdentifyRequest) Validate() error {
	if strings.TrimSpace(v.Description) == "" {
		return validation.NewErrRequestIsMissingRequiredField("description")
	}
	return nil
}

// MovieIdentifyResponse returns TMDB candidates for the AI title guesses,
// merged & deduped, for the user to disambiguate.
type MovieIdentifyResponse struct {
	Movies []tmdbclient.MovieSummary `json:"movies"`
}

// MovieResolveRequest fully resolves a single movie by its TMDB id.
type MovieResolveRequest struct {
	TmdbID int `json:"tmdbID"`
}

// Validate returns error if not valid
func (v MovieResolveRequest) Validate() error {
	if v.TmdbID <= 0 {
		return validation.NewErrRequestIsMissingRequiredField("tmdbID")
	}
	return nil
}

// MovieResolveResponse DTO
type MovieResolveResponse struct {
	Movie tmdbclient.MovieDetails `json:"movie"`
}

// AddMovieToWatchlistRequest resolves a movie (by tmdbID or a free-text
// search query - first search match wins) and appends it to the space's
// canonical watch!movies list (dbo4listus.WatchMoviesListID).
type AddMovieToWatchlistRequest struct {
	dto4spaceus.SpaceRequest
	TmdbID    int                   `json:"tmdbID,omitempty"`
	Query     string                `json:"query,omitempty"`
	WatchWith *dbo4listus.WatchWith `json:"watchWith,omitempty"`
}

// Validate returns error if not valid
func (v AddMovieToWatchlistRequest) Validate() error {
	if err := v.SpaceRequest.Validate(); err != nil {
		return err
	}
	if v.TmdbID <= 0 && strings.TrimSpace(v.Query) == "" {
		return validation.NewValidationError("either tmdbID or query is required")
	}
	if v.WatchWith != nil {
		if err := v.WatchWith.Validate(); err != nil {
			return validation.NewErrBadRequestFieldValue("watchWith", err.Error())
		}
	}
	return nil
}

// AddMovieToWatchlistResponse DTO
type AddMovieToWatchlistResponse struct {
	Item *dbo4listus.ListItemBrief `json:"item"`
}

// SetListItemWatchWithRequest updates the WatchWith field of an existing watch-list item.
type SetListItemWatchWithRequest struct {
	ListItemRequest
	WatchWith dbo4listus.WatchWith `json:"watchWith"`
}

// Validate returns error if not valid
func (v SetListItemWatchWithRequest) Validate() error {
	if err := v.ListItemRequest.Validate(); err != nil {
		return err
	}
	if err := v.WatchWith.Validate(); err != nil {
		return validation.NewErrBadRequestFieldValue("watchWith", err.Error())
	}
	return nil
}
