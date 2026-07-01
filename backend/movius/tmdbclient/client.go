package tmdbclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	// PosterBaseURL is the TMDB image CDN base for w500-sized posters.
	PosterBaseURL = "https://image.tmdb.org/t/p/w500"

	apiBaseURL = "https://api.themoviedb.org/3"

	// EnvAPIKey is the environment variable holding the TMDB v3 API key.
	EnvAPIKey = "TMDB_API_KEY"

	maxCastMembers = 5
)

// Client is a small TMDB v3 API client.
type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a client for the given API key. An empty apiKey puts the
// client in MOCK mode (see mock.go) - realistic sample data instead of live
// TMDB calls, so the feature works without provisioning a key.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    apiBaseURL,
	}
}

// NewClientFromEnv reads TMDB_API_KEY from the environment.
func NewClientFromEnv() *Client {
	return NewClient(os.Getenv(EnvAPIKey))
}

// IsMock reports whether this client has no real API key configured and is
// therefore serving MOCK data.
func (c *Client) IsMock() bool {
	return strings.TrimSpace(c.apiKey) == ""
}

// PosterURL builds the full poster image URL from a TMDB poster_path.
func PosterURL(posterPath string) string {
	if posterPath == "" {
		return ""
	}
	return PosterBaseURL + posterPath
}

func yearFromReleaseDate(releaseDate string) int {
	if len(releaseDate) < 4 {
		return 0
	}
	y, err := strconv.Atoi(releaseDate[:4])
	if err != nil {
		return 0
	}
	return y
}

// --- Search -----------------------------------------------------------

type searchMovieResponse struct {
	Results []struct {
		ID          int    `json:"id"`
		Title       string `json:"title"`
		ReleaseDate string `json:"release_date"`
		PosterPath  string `json:"poster_path"`
	} `json:"results"`
}

// SearchMovie searches TMDB for movies matching the given title query.
func (c *Client) SearchMovie(ctx context.Context, query string) ([]MovieSummary, error) {
	if c.IsMock() {
		return mockSearchMovie(query), nil
	}
	var resp searchMovieResponse
	if err := c.get(ctx, "/search/movie", url.Values{"query": {query}}, &resp); err != nil {
		return nil, fmt.Errorf("tmdb search/movie failed: %w", err)
	}
	summaries := make([]MovieSummary, 0, len(resp.Results))
	for _, r := range resp.Results {
		summaries = append(summaries, MovieSummary{
			TmdbID:    r.ID,
			Title:     r.Title,
			Year:      yearFromReleaseDate(r.ReleaseDate),
			PosterURL: PosterURL(r.PosterPath),
		})
	}
	return summaries, nil
}

type searchPersonResponse struct {
	Results []struct {
		ID       int `json:"id"`
		KnownFor []struct {
			ID          int    `json:"id"`
			Title       string `json:"title"`
			ReleaseDate string `json:"release_date"`
			PosterPath  string `json:"poster_path"`
			MediaType   string `json:"media_type"`
		} `json:"known_for"`
	} `json:"results"`
}

// SearchPerson searches TMDB for an actor by name and returns the movies
// they are best known for (via TMDB's `known_for` on /search/person).
func (c *Client) SearchPerson(ctx context.Context, actor string) ([]MovieSummary, error) {
	if c.IsMock() {
		return mockSearchPerson(actor), nil
	}
	var resp searchPersonResponse
	if err := c.get(ctx, "/search/person", url.Values{"query": {actor}}, &resp); err != nil {
		return nil, fmt.Errorf("tmdb search/person failed: %w", err)
	}
	var summaries []MovieSummary
	for _, person := range resp.Results {
		for _, kf := range person.KnownFor {
			if kf.MediaType != "" && kf.MediaType != "movie" {
				continue
			}
			summaries = append(summaries, MovieSummary{
				TmdbID:    kf.ID,
				Title:     kf.Title,
				Year:      yearFromReleaseDate(kf.ReleaseDate),
				PosterURL: PosterURL(kf.PosterPath),
			})
		}
	}
	return summaries, nil
}

// --- Movie details ------------------------------------------------------

type movieResponse struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Overview    string `json:"overview"`
	ReleaseDate string `json:"release_date"`
	PosterPath  string `json:"poster_path"`
}

// GetMovie fetches overview, poster & release year for a movie id.
func (c *Client) GetMovie(ctx context.Context, tmdbID int) (MovieDetails, error) {
	if c.IsMock() {
		return mockGetMovie(tmdbID)
	}
	var resp movieResponse
	if err := c.get(ctx, fmt.Sprintf("/movie/%d", tmdbID), nil, &resp); err != nil {
		return MovieDetails{}, fmt.Errorf("tmdb movie/%d failed: %w", tmdbID, err)
	}
	return MovieDetails{
		MovieSummary: MovieSummary{
			TmdbID:    resp.ID,
			Title:     resp.Title,
			Year:      yearFromReleaseDate(resp.ReleaseDate),
			PosterURL: PosterURL(resp.PosterPath),
		},
		Overview: resp.Overview,
	}, nil
}

type creditsResponse struct {
	Cast []struct {
		Name  string `json:"name"`
		Order int    `json:"order"`
	} `json:"cast"`
}

// GetCredits fetches the top ~5 cast member names (order-sorted) for a movie.
func (c *Client) GetCredits(ctx context.Context, tmdbID int) ([]string, error) {
	if c.IsMock() {
		return mockGetCredits(tmdbID), nil
	}
	var resp creditsResponse
	if err := c.get(ctx, fmt.Sprintf("/movie/%d/credits", tmdbID), nil, &resp); err != nil {
		return nil, fmt.Errorf("tmdb movie/%d/credits failed: %w", tmdbID, err)
	}
	sort.Slice(resp.Cast, func(i, j int) bool { return resp.Cast[i].Order < resp.Cast[j].Order })
	names := make([]string, 0, maxCastMembers)
	for _, member := range resp.Cast {
		if len(names) >= maxCastMembers {
			break
		}
		names = append(names, member.Name)
	}
	return names, nil
}

type videosResponse struct {
	Results []struct {
		Key  string `json:"key"`
		Site string `json:"site"`
		Type string `json:"type"`
	} `json:"results"`
}

// GetVideos fetches the videos for a movie and picks the YouTube trailer key
// (empty string if none found).
func (c *Client) GetVideos(ctx context.Context, tmdbID int) (string, error) {
	if c.IsMock() {
		return mockGetVideos(tmdbID), nil
	}
	var resp videosResponse
	if err := c.get(ctx, fmt.Sprintf("/movie/%d/videos", tmdbID), nil, &resp); err != nil {
		return "", fmt.Errorf("tmdb movie/%d/videos failed: %w", tmdbID, err)
	}
	for _, v := range resp.Results {
		if v.Site == "YouTube" && v.Type == "Trailer" {
			return v.Key, nil
		}
	}
	for _, v := range resp.Results { // fallback: any YouTube video
		if v.Site == "YouTube" {
			return v.Key, nil
		}
	}
	return "", nil
}

// Resolve fetches movie + credits + videos in one round-trip, building the
// fully-enriched MovieDetails used to populate a watch-list item.
func (c *Client) Resolve(ctx context.Context, tmdbID int) (details MovieDetails, err error) {
	if details, err = c.GetMovie(ctx, tmdbID); err != nil {
		return
	}
	if details.Cast, err = c.GetCredits(ctx, tmdbID); err != nil {
		return
	}
	if details.TrailerYouTubeKey, err = c.GetVideos(ctx, tmdbID); err != nil {
		return
	}
	return
}

func (c *Client) get(ctx context.Context, path string, query url.Values, out interface{}) error {
	if query == nil {
		query = url.Values{}
	}
	query.Set("api_key", c.apiKey)
	reqURL := c.baseURL + path + "?" + query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
