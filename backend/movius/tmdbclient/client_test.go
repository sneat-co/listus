package tmdbclient

import (
	"context"
	"testing"
)

func TestNewClient_IsMock(t *testing.T) {
	if !NewClient("").IsMock() {
		t.Error("expected empty api key to be mock mode")
	}
	if NewClient("real-key").IsMock() {
		t.Error("expected non-empty api key to not be mock mode")
	}
	if !NewClientFromEnv().IsMock() {
		t.Error("expected NewClientFromEnv() to be mock mode when TMDB_API_KEY is unset in test env")
	}
}

func TestClient_SearchMovie_Mock(t *testing.T) {
	c := NewClient("")
	results, err := c.SearchMovie(context.Background(), "inception")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Title != "Inception" {
		t.Fatalf("expected 1 result for Inception, got: %+v", results)
	}
	if results[0].PosterURL == "" {
		t.Error("expected posterURL to be set")
	}

	all, err := c.SearchMovie(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != len(mockMovies) {
		t.Errorf("expected empty query to return all mock movies, got %d", len(all))
	}
}

func TestClient_SearchPerson_Mock(t *testing.T) {
	c := NewClient("")
	results, err := c.SearchPerson(context.Background(), "DiCaprio")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Title != "Inception" {
		t.Fatalf("expected Inception for DiCaprio search, got: %+v", results)
	}
}

func TestClient_GetMovie_Mock(t *testing.T) {
	c := NewClient("")
	details, err := c.GetMovie(context.Background(), 27205)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if details.Title != "Inception" || details.Overview == "" {
		t.Errorf("unexpected details: %+v", details)
	}

	if _, err := c.GetMovie(context.Background(), 999999); err == nil {
		t.Error("expected error for unknown mock movie id")
	}
}

func TestClient_GetCredits_Mock(t *testing.T) {
	c := NewClient("")
	cast, err := c.GetCredits(context.Background(), 27205)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cast) == 0 || cast[0] != "Leonardo DiCaprio" {
		t.Errorf("unexpected cast: %+v", cast)
	}
}

func TestClient_GetVideos_Mock(t *testing.T) {
	c := NewClient("")
	key, err := c.GetVideos(context.Background(), 27205)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "YoHD9XEInc0" {
		t.Errorf("unexpected trailer key: %s", key)
	}
}

func TestClient_Resolve_Mock(t *testing.T) {
	c := NewClient("")
	details, err := c.Resolve(context.Background(), 603)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if details.Title != "The Matrix" || len(details.Cast) == 0 || details.TrailerYouTubeKey == "" || details.PosterURL == "" {
		t.Errorf("expected fully-enriched details, got: %+v", details)
	}
}

func TestPosterURL(t *testing.T) {
	if got := PosterURL(""); got != "" {
		t.Errorf("expected empty poster path to yield empty URL, got: %s", got)
	}
	if got := PosterURL("/abc.jpg"); got != "https://image.tmdb.org/t/p/w500/abc.jpg" {
		t.Errorf("unexpected poster URL: %s", got)
	}
}

func TestYearFromReleaseDate(t *testing.T) {
	tests := map[string]int{
		"2010-07-16": 2010,
		"":           0,
		"abc":        0,
		"1999":       1999,
	}
	for in, want := range tests {
		if got := yearFromReleaseDate(in); got != want {
			t.Errorf("yearFromReleaseDate(%q) = %d, want %d", in, got, want)
		}
	}
}
