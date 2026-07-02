package facade4listus

import (
	"context"
	"fmt"

	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/listus/backend/movius/aiclient"
	"github.com/sneat-co/listus/backend/movius/tmdbclient"
)

// maxIdentifyCandidates caps how many merged TMDB candidates IdentifyMovies returns.
const maxIdentifyCandidates = 10

// movieIdentifier is the seam movius/aiclient is called through - lets tests
// substitute their own client. Defaults to reading ANTHROPIC_API_KEY from
// env; reports IsMock() when the key is absent so IdentifyMovies can degrade
// to a plain TMDB keyword search over the description.
type movieIdentifier interface {
	IsMock() bool
	GuessMovies(ctx context.Context, description string) ([]aiclient.MovieGuess, error)
}

var identifyClient movieIdentifier = aiclient.NewClientFromEnv()

// IdentifyMovies turns a vague natural-language movie description into TMDB
// candidates: Claude proposes up to a few likely title guesses, each guess is
// grounded via TMDB title search, and the results are merged (deduped by
// tmdbID, guess order preserved). Without an ANTHROPIC_API_KEY it degrades to
// a plain TMDB keyword search over the description.
func IdentifyMovies(ctx context.Context, request dto4listus.MovieIdentifyRequest) (response dto4listus.MovieIdentifyResponse, err error) {
	if err = request.Validate(); err != nil {
		return
	}
	if identifyClient.IsMock() { // no AI key - degrade to plain TMDB keyword search
		response.Movies, err = movieClient.SearchMovie(ctx, request.Description)
		if err != nil {
			err = fmt.Errorf("failed to search movies by description: %w", err)
		}
		return
	}
	guesses, err := identifyClient.GuessMovies(ctx, request.Description)
	if err != nil {
		err = fmt.Errorf("failed to get AI movie guesses: %w", err)
		return
	}
	candidatesByGuess := make([][]tmdbclient.MovieSummary, 0, len(guesses))
	for _, guess := range guesses {
		var candidates []tmdbclient.MovieSummary
		if candidates, err = movieClient.SearchMovie(ctx, guess.Title); err != nil {
			err = fmt.Errorf("failed to search movies for AI guess %q: %w", guess.Title, err)
			return
		}
		candidatesByGuess = append(candidatesByGuess, candidates)
	}
	response.Movies = mergeMovieSummaries(candidatesByGuess...)
	if len(response.Movies) > maxIdentifyCandidates {
		response.Movies = response.Movies[:maxIdentifyCandidates]
	}
	return
}
