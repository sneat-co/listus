package tmdbclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMock_Unknown(t *testing.T) {
	if credits := mockGetCredits(999999); credits != nil {
		t.Errorf("expected nil credits for unknown ID, got: %v", credits)
	}
	if video := mockGetVideos(999999); video != "" {
		t.Errorf("expected empty video for unknown ID, got: %q", video)
	}
}

func TestYearFromReleaseDate_InvalidInt(t *testing.T) {
	if y := yearFromReleaseDate("abcd-01-01"); y != 0 {
		t.Errorf("expected 0 for non-numeric year, got %d", y)
	}
}

func TestClient_Live_AllEndpoints(t *testing.T) {
	mux := http.NewServeMux()

	// 1. Search movie
	mux.HandleFunc("/search/movie", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("query") == "fail" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(`{
			"results": [
				{
					"id": 101,
					"title": "Test Movie",
					"release_date": "2023-01-01",
					"poster_path": "/test.jpg",
					"overview": "Overview"
				}
			]
		}`))
	})

	// 2. Search person
	mux.HandleFunc("/search/person", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("query") == "fail" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(`{
			"results": [
				{
					"id": 201,
					"known_for": [
						{
							"id": 101,
							"title": "Movie 1",
							"release_date": "2021-01-01",
							"poster_path": "/p1.jpg",
							"media_type": "movie",
							"overview": "Overview 1"
						},
						{
							"id": 102,
							"title": "TV Show",
							"release_date": "2022-01-01",
							"poster_path": "/p2.jpg",
							"media_type": "tv",
							"overview": "Overview 2"
						}
					]
				}
			]
		}`))
	})

	// 3. Movie details
	mux.HandleFunc("/movie/101", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"id": 101,
			"title": "Test Movie",
			"overview": "Overview 101",
			"release_date": "2023-05-01",
			"poster_path": "/poster.jpg"
		}`))
	})
	mux.HandleFunc("/movie/500", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	// 4. Credits
	mux.HandleFunc("/movie/101/credits", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"cast": [
				{"name": "Actor 3", "order": 3},
				{"name": "Actor 1", "order": 1},
				{"name": "Actor 2", "order": 2},
				{"name": "Actor 4", "order": 4},
				{"name": "Actor 5", "order": 5},
				{"name": "Actor 6", "order": 6}
			]
		}`))
	})
	mux.HandleFunc("/movie/500/credits", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	// 5. Videos
	mux.HandleFunc("/movie/101/videos", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"results": [
				{"key": "yt-teaser", "site": "YouTube", "type": "Teaser"},
				{"key": "yt-trailer", "site": "YouTube", "type": "Trailer"}
			]
		}`))
	})
	mux.HandleFunc("/movie/102/videos", func(w http.ResponseWriter, r *http.Request) {
		// Fallback: YouTube non-trailer
		_, _ = w.Write([]byte(`{
			"results": [
				{"key": "yt-clip", "site": "YouTube", "type": "Clip"}
			]
		}`))
	})
	mux.HandleFunc("/movie/103/videos", func(w http.ResponseWriter, r *http.Request) {
		// Non-YouTube video
		_, _ = w.Write([]byte(`{
			"results": [
				{"key": "vimeo-clip", "site": "Vimeo", "type": "Trailer"}
			]
		}`))
	})
	mux.HandleFunc("/movie/500/videos", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	// Bad JSON
	mux.HandleFunc("/bad-json", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	c := &Client{
		apiKey:     "test-key",
		httpClient: server.Client(),
		baseURL:    server.URL,
	}
	ctx := context.Background()

	// SearchMovie
	movies, err := c.SearchMovie(ctx, "test")
	if err != nil || len(movies) != 1 || movies[0].Title != "Test Movie" {
		t.Fatalf("SearchMovie failed: %v, got %v", err, movies)
	}
	if _, err := c.SearchMovie(ctx, "fail"); err == nil {
		t.Fatal("expected SearchMovie error")
	}

	// SearchPerson
	persons, err := c.SearchPerson(ctx, "actor")
	if err != nil || len(persons) != 1 || persons[0].Title != "Movie 1" {
		t.Fatalf("SearchPerson failed: %v, got %v", err, persons)
	}
	if _, err := c.SearchPerson(ctx, "fail"); err == nil {
		t.Fatal("expected SearchPerson error")
	}

	// GetMovie
	movie, err := c.GetMovie(ctx, 101)
	if err != nil || movie.Title != "Test Movie" {
		t.Fatalf("GetMovie failed: %v, got %v", err, movie)
	}
	if _, err := c.GetMovie(ctx, 500); err == nil {
		t.Fatal("expected GetMovie error")
	}

	// GetCredits
	credits, err := c.GetCredits(ctx, 101)
	if err != nil || len(credits) != 5 || credits[0] != "Actor 1" || credits[1] != "Actor 2" {
		t.Fatalf("GetCredits failed: %v, got %v", err, credits)
	}
	if _, err := c.GetCredits(ctx, 500); err == nil {
		t.Fatal("expected GetCredits error")
	}

	// GetVideos
	video, err := c.GetVideos(ctx, 101)
	if err != nil || video != "yt-trailer" {
		t.Fatalf("GetVideos Trailer failed: %v, got %q", err, video)
	}
	video2, err := c.GetVideos(ctx, 102)
	if err != nil || video2 != "yt-clip" {
		t.Fatalf("GetVideos Fallback failed: %v, got %q", err, video2)
	}
	video3, err := c.GetVideos(ctx, 103)
	if err != nil || video3 != "" {
		t.Fatalf("GetVideos Non-YouTube failed: %v, got %q", err, video3)
	}
	if _, err := c.GetVideos(ctx, 500); err == nil {
		t.Fatal("expected GetVideos error")
	}

	// Resolve
	details, err := c.Resolve(ctx, 101)
	if err != nil || details.Title != "Test Movie" || len(details.Cast) != 5 || details.TrailerYouTubeKey != "yt-trailer" {
		t.Fatalf("Resolve failed: %v, got %+v", err, details)
	}
	if _, err := c.Resolve(ctx, 500); err == nil {
		t.Fatal("expected Resolve error on GetMovie failure")
	}

	// Resolve errors on credits / videos
	mux.HandleFunc("/movie/200", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id": 200, "title": "M200"}`))
	})
	mux.HandleFunc("/movie/200/credits", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	if _, err := c.Resolve(ctx, 200); err == nil {
		t.Fatal("expected Resolve error on credits failure")
	}

	mux.HandleFunc("/movie/300", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id": 300, "title": "M300"}`))
	})
	mux.HandleFunc("/movie/300/credits", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"cast": []}`))
	})
	mux.HandleFunc("/movie/300/videos", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	if _, err := c.Resolve(ctx, 300); err == nil {
		t.Fatal("expected Resolve error on videos failure")
	}

	// get() with invalid JSON
	var dummy interface{}
	if err := c.get(ctx, "/bad-json", nil, &dummy); err == nil {
		t.Fatal("expected error decoding invalid json")
	}

	// get() with canceled context
	canceledCtx, cancel := context.WithCancel(ctx)
	cancel()
	if err := c.get(canceledCtx, "/search/movie", nil, &dummy); err == nil {
		t.Fatal("expected error with canceled context")
	}

	// get() with invalid baseURL
	badClient := &Client{apiKey: "k", baseURL: "://invalid"}
	if err := badClient.get(ctx, "/test", nil, &dummy); err == nil {
		t.Fatal("expected error with invalid baseURL")
	}
}

