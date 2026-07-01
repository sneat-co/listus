package tmdbclient

import (
	"fmt"
	"strings"
)

// mockMovies is a small set of well-known movies used when TMDB_API_KEY is
// not configured (Client.IsMock() == true), so search / resolve / add-to-
// watchlist all work end-to-end without a key. Clearly MOCK sample data -
// poster paths are real TMDB paths, so PosterURL() renders real posters.
var mockMovies = []MovieDetails{
	{
		MovieSummary:      MovieSummary{TmdbID: 27205, Title: "Inception", Year: 2010, PosterURL: PosterURL("/edv5CZvWj09upOsy2Y6IwDhK8bt.jpg")},
		Overview:          "A thief who steals corporate secrets through dream-sharing technology is given the inverse task of planting an idea into the mind of a CEO.",
		TrailerYouTubeKey: "YoHD9XEInc0",
		Cast:              []string{"Leonardo DiCaprio", "Joseph Gordon-Levitt", "Elliot Page", "Tom Hardy", "Ken Watanabe"},
	},
	{
		MovieSummary:      MovieSummary{TmdbID: 603, Title: "The Matrix", Year: 1999, PosterURL: PosterURL("/f89U3ADr1oiB1s9GkdPOEpXUk5H.jpg")},
		Overview:          "A computer hacker learns from mysterious rebels about the true nature of his reality and his role in the war against its controllers.",
		TrailerYouTubeKey: "vKQi3bBA1y8",
		Cast:              []string{"Keanu Reeves", "Laurence Fishburne", "Carrie-Anne Moss", "Hugo Weaving"},
	},
	{
		MovieSummary:      MovieSummary{TmdbID: 278, Title: "The Shawshank Redemption", Year: 1994, PosterURL: PosterURL("/q6y0Go1tsGEsmtFryDOJo3dEmqu.jpg")},
		Overview:          "Two imprisoned men bond over a number of years, finding solace and eventual redemption through acts of common decency.",
		TrailerYouTubeKey: "6hB3S9bIaco",
		Cast:              []string{"Tim Robbins", "Morgan Freeman", "Bob Gunton"},
	},
	{
		MovieSummary:      MovieSummary{TmdbID: 680, Title: "Pulp Fiction", Year: 1994, PosterURL: PosterURL("/d5iIlFn5s0ImszYzBPb8JPIfbXD.jpg")},
		Overview:          "The lives of two mob hitmen, a boxer, a gangster and his wife intertwine in four tales of violence and redemption.",
		TrailerYouTubeKey: "s7EdQ4FqbhY",
		Cast:              []string{"John Travolta", "Samuel L. Jackson", "Uma Thurman", "Bruce Willis"},
	},
	{
		MovieSummary:      MovieSummary{TmdbID: 157336, Title: "Interstellar", Year: 2014, PosterURL: PosterURL("/gEU2QniE6E77NI6lCU6MxlNBvIx.jpg")},
		Overview:          "A team of explorers travel through a wormhole in space in an attempt to ensure humanity's survival.",
		TrailerYouTubeKey: "zSWdZVtXT7E",
		Cast:              []string{"Matthew McConaughey", "Anne Hathaway", "Jessica Chastain", "Michael Caine"},
	},
}

func mockSearchMovie(query string) []MovieSummary {
	q := strings.ToLower(strings.TrimSpace(query))
	var out []MovieSummary
	for _, m := range mockMovies {
		if q == "" || strings.Contains(strings.ToLower(m.Title), q) {
			out = append(out, m.MovieSummary)
		}
	}
	return out
}

func mockSearchPerson(actor string) []MovieSummary {
	q := strings.ToLower(strings.TrimSpace(actor))
	var out []MovieSummary
	for _, m := range mockMovies {
		for _, castName := range m.Cast {
			if q != "" && strings.Contains(strings.ToLower(castName), q) {
				out = append(out, m.MovieSummary)
				break
			}
		}
	}
	return out
}

func mockGetMovie(tmdbID int) (MovieDetails, error) {
	for _, m := range mockMovies {
		if m.TmdbID == tmdbID {
			return MovieDetails{MovieSummary: m.MovieSummary, Overview: m.Overview}, nil
		}
	}
	return MovieDetails{}, fmt.Errorf("mock movie not found: tmdbID=%d", tmdbID)
}

func mockGetCredits(tmdbID int) []string {
	for _, m := range mockMovies {
		if m.TmdbID == tmdbID {
			return m.Cast
		}
	}
	return nil
}

func mockGetVideos(tmdbID int) string {
	for _, m := range mockMovies {
		if m.TmdbID == tmdbID {
			return m.TrailerYouTubeKey
		}
	}
	return ""
}
