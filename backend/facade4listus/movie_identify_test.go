package facade4listus

import (
	"context"
	"errors"
	"testing"

	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/listus/backend/movius/aiclient"
)

// fakeIdentifier substitutes the movius/aiclient seam so tests never hit the
// network (the movieClient seam already serves tmdbclient MOCK data because
// TMDB_API_KEY is unset in the test env).
type fakeIdentifier struct {
	isMock  bool
	guesses []aiclient.MovieGuess
	err     error
}

func (f fakeIdentifier) IsMock() bool { return f.isMock }
func (f fakeIdentifier) GuessMovies(ctx context.Context, description string) ([]aiclient.MovieGuess, error) {
	return f.guesses, f.err
}

func swapIdentifyClient(t *testing.T, client movieIdentifier) {
	t.Helper()
	old := identifyClient
	identifyClient = client
	t.Cleanup(func() { identifyClient = old })
}

func TestIdentifyMovies_InvalidRequest(t *testing.T) {
	swapIdentifyClient(t, fakeIdentifier{})
	if _, err := IdentifyMovies(t.Context(), dto4listus.MovieIdentifyRequest{}); err == nil {
		t.Fatal("expected validation error for empty description")
	}
}

func TestIdentifyMovies_AIPath(t *testing.T) {
	// AI path: guesses are grounded via TMDB search, merged & deduped by
	// tmdbID with guess order preserved. "Inception" appears in two guesses
	// but must only be returned once, ahead of "The Matrix".
	swapIdentifyClient(t, fakeIdentifier{guesses: []aiclient.MovieGuess{
		{Title: "Inception", Year: 2010},
		{Title: "The Matrix", Year: 1999},
		{Title: "Inception"}, // duplicate guess - must not duplicate the candidate
	}})
	response, err := IdentifyMovies(t.Context(), dto4listus.MovieIdentifyRequest{Description: "a dream heist movie"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(response.Movies) != 2 {
		t.Fatalf("expected 2 deduped candidates, got %d: %+v", len(response.Movies), response.Movies)
	}
	if response.Movies[0].Title != "Inception" || response.Movies[1].Title != "The Matrix" {
		t.Errorf("expected guess order preserved [Inception, The Matrix], got: %+v", response.Movies)
	}
}

func TestIdentifyMovies_AIPath_GuessError(t *testing.T) {
	swapIdentifyClient(t, fakeIdentifier{err: errors.New("boom")})
	if _, err := IdentifyMovies(t.Context(), dto4listus.MovieIdentifyRequest{Description: "whatever"}); err == nil {
		t.Fatal("expected error when AI guesses fail")
	}
}

func TestIdentifyMovies_DegradedPath(t *testing.T) {
	// No ANTHROPIC_API_KEY (IsMock) - must degrade to a plain TMDB keyword
	// search over the description, never calling GuessMovies.
	swapIdentifyClient(t, fakeIdentifier{isMock: true, err: errors.New("GuessMovies must not be called in mock mode")})
	response, err := IdentifyMovies(t.Context(), dto4listus.MovieIdentifyRequest{Description: "matrix"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(response.Movies) != 1 || response.Movies[0].Title != "The Matrix" {
		t.Fatalf("expected [The Matrix] from degraded keyword search, got: %+v", response.Movies)
	}
}

func TestIdentifyMovies_DefaultClientIsMock(t *testing.T) {
	// The default seam value must be in mock mode when ANTHROPIC_API_KEY is
	// unset, so the degraded fallback works out of the box.
	if !identifyClient.IsMock() {
		t.Error("expected default identifyClient to be mock mode when ANTHROPIC_API_KEY is unset in test env")
	}
}
