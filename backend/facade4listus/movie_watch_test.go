package facade4listus

import (
	"testing"

	"github.com/sneat-co/listus/backend/dbo4listus"
	"github.com/sneat-co/listus/backend/dto4listus"
)

// These tests rely on movius/tmdbclient's MOCK data (no TMDB_API_KEY set in
// the test environment - see tmdbclient.Client.IsMock()), so they exercise
// the real facade + mock-movie round trip without any network calls.

func TestSearchMovies_ByTitle(t *testing.T) {
	response, err := SearchMovies(userCtx(testUserID), dto4listus.MovieSearchRequest{Query: "inception"})
	if err != nil {
		t.Fatalf("SearchMovies failed: %v", err)
	}
	if len(response.Movies) != 1 || response.Movies[0].Title != "Inception" {
		t.Fatalf("expected 1 result for Inception, got: %+v", response.Movies)
	}
}

func TestSearchMovies_ByActor(t *testing.T) {
	response, err := SearchMovies(userCtx(testUserID), dto4listus.MovieSearchRequest{Query: "Keanu Reeves"})
	if err != nil {
		t.Fatalf("SearchMovies failed: %v", err)
	}
	if len(response.Movies) != 1 || response.Movies[0].Title != "The Matrix" {
		t.Fatalf("expected The Matrix for Keanu Reeves, got: %+v", response.Movies)
	}
}

func TestSearchMovies_InvalidRequest(t *testing.T) {
	if _, err := SearchMovies(userCtx(testUserID), dto4listus.MovieSearchRequest{}); err == nil {
		t.Error("expected error for empty query")
	}
}

func TestResolveMovie_Succeeds(t *testing.T) {
	response, err := ResolveMovie(userCtx(testUserID), dto4listus.MovieResolveRequest{TmdbID: 27205})
	if err != nil {
		t.Fatalf("ResolveMovie failed: %v", err)
	}
	if response.Movie.Title != "Inception" || len(response.Movie.Cast) == 0 || response.Movie.TrailerYouTubeKey == "" {
		t.Errorf("expected fully-enriched movie, got: %+v", response.Movie)
	}
}

func TestResolveMovie_NotFound(t *testing.T) {
	if _, err := ResolveMovie(userCtx(testUserID), dto4listus.MovieResolveRequest{TmdbID: 999999}); err == nil {
		t.Error("expected error for unknown movie id")
	}
}

func TestAddMovieToWatchlist_ByTmdbID(t *testing.T) {
	_ = newTestDBWithSpace(t, testSpaceID, testUserID)

	response, err := AddMovieToWatchlist(userCtx(testUserID), dto4listus.AddMovieToWatchlistRequest{
		SpaceRequest: spaceRequest(testSpaceID),
		TmdbID:       27205, // Inception
	})
	if err != nil {
		t.Fatalf("AddMovieToWatchlist failed: %v", err)
	}
	if response.Item == nil {
		t.Fatal("expected a created item")
	}
	item := response.Item
	if item.Title != "Inception" || item.TmdbID != 27205 || item.PosterURL == "" || item.Overview == "" || item.TrailerYouTubeKey == "" {
		t.Errorf("expected fully-enriched watch item, got: %+v", item)
	}
	if len(item.Cast) == 0 {
		t.Error("expected cast to be populated")
	}
	if item.WatchWith == nil || item.WatchWith.Mode != dbo4listus.WatchWithModeAlone {
		t.Errorf("expected default watchWith=alone, got: %+v", item.WatchWith)
	}
}

func TestAddMovieToWatchlist_ByQuery_WithWatchWith(t *testing.T) {
	_ = newTestDBWithSpace(t, testSpaceID, testUserID)

	response, err := AddMovieToWatchlist(userCtx(testUserID), dto4listus.AddMovieToWatchlistRequest{
		SpaceRequest: spaceRequest(testSpaceID),
		Query:        "The Matrix",
		WatchWith:    &dbo4listus.WatchWith{Mode: dbo4listus.WatchWithModeSpace, Ref: "space1", Title: "Family"},
	})
	if err != nil {
		t.Fatalf("AddMovieToWatchlist failed: %v", err)
	}
	if response.Item == nil || response.Item.Title != "The Matrix" {
		t.Fatalf("expected The Matrix item, got: %+v", response.Item)
	}
	if response.Item.WatchWith == nil || response.Item.WatchWith.Mode != dbo4listus.WatchWithModeSpace || response.Item.WatchWith.Ref != "space1" {
		t.Errorf("expected watchWith to be passed through, got: %+v", response.Item.WatchWith)
	}
}

func TestAddMovieToWatchlist_NoMatchForQuery(t *testing.T) {
	_ = newTestDBWithSpace(t, testSpaceID, testUserID)

	if _, err := AddMovieToWatchlist(userCtx(testUserID), dto4listus.AddMovieToWatchlistRequest{
		SpaceRequest: spaceRequest(testSpaceID),
		Query:        "no such movie exists xyz",
	}); err == nil {
		t.Error("expected error when no movie matches the query")
	}
}

func TestAddMovieToWatchlist_InvalidRequest(t *testing.T) {
	if _, err := AddMovieToWatchlist(userCtx(testUserID), dto4listus.AddMovieToWatchlistRequest{
		SpaceRequest: spaceRequest(testSpaceID),
		// missing both TmdbID and Query
	}); err == nil {
		t.Error("expected error when neither tmdbID nor query is provided")
	}
}

func TestSetListItemWatchWith_Succeeds(t *testing.T) {
	_ = newTestDBWithSpace(t, testSpaceID, testUserID)

	created := createItems(t, "watch!movies", "Inception")
	itemID := created.CreatedItems[0].ID

	item, _, err := SetListItemWatchWith(userCtx(testUserID), dto4listus.SetListItemWatchWithRequest{
		ListItemRequest: dto4listus.ListItemRequest{
			ListRequest: listRequest(testSpaceID, "watch!movies"),
			ItemID:      itemID,
		},
		WatchWith: dbo4listus.WatchWith{Mode: dbo4listus.WatchWithModeContact, Ref: "contact1", Title: "Alex"},
	})
	if err != nil {
		t.Fatalf("SetListItemWatchWith failed: %v", err)
	}
	if item == nil || item.WatchWith == nil || item.WatchWith.Mode != dbo4listus.WatchWithModeContact || item.WatchWith.Ref != "contact1" {
		t.Errorf("expected watchWith to be updated, got: %+v", item)
	}
}

func TestSetListItemWatchWith_ItemNotFound(t *testing.T) {
	_ = newTestDBWithSpace(t, testSpaceID, testUserID)
	_ = createItems(t, "watch!movies", "Inception")

	if _, _, err := SetListItemWatchWith(userCtx(testUserID), dto4listus.SetListItemWatchWithRequest{
		ListItemRequest: dto4listus.ListItemRequest{
			ListRequest: listRequest(testSpaceID, "watch!movies"),
			ItemID:      "does-not-exist",
		},
		WatchWith: dbo4listus.WatchWith{Mode: dbo4listus.WatchWithModeAlone},
	}); err == nil {
		t.Error("expected error for missing item id")
	}
}
