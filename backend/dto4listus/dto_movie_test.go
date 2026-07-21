package dto4listus

import (
	"testing"

	"github.com/sneat-co/listus/backend/dbo4listus"
)

func TestMovieLookupRequestsValidate(t *testing.T) {
	tests := []struct {
		name     string
		validate func() error
		wantErr  bool
	}{
		{"search_valid", func() error { return MovieSearchRequest{Query: "Titanic"}.Validate() }, false},
		{"search_blank", func() error { return MovieSearchRequest{Query: " "}.Validate() }, true},
		{"identify_valid", func() error { return MovieIdentifyRequest{Description: "submarine film"}.Validate() }, false},
		{"identify_blank", func() error { return MovieIdentifyRequest{}.Validate() }, true},
		{"resolve_valid", func() error { return MovieResolveRequest{TmdbID: 1}.Validate() }, false},
		{"resolve_missing", func() error { return MovieResolveRequest{}.Validate() }, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddMovieToWatchlistRequestValidate(t *testing.T) {
	valid := AddMovieToWatchlistRequest{
		SpaceRequest: validListRequest().SpaceRequest,
		ListID:       dbo4listus.WatchMoviesListID,
		TmdbID:       1,
		WatchWith:    &dbo4listus.WatchWith{Mode: dbo4listus.WatchWithModeAlone},
	}
	tests := []struct {
		name    string
		request AddMovieToWatchlistRequest
		wantErr bool
	}{
		{"valid_tmdb", valid, false},
		{"valid_query", AddMovieToWatchlistRequest{SpaceRequest: valid.SpaceRequest, Query: "Titanic"}, false},
		{"missing_movie", AddMovieToWatchlistRequest{SpaceRequest: valid.SpaceRequest}, true},
		{"non_watch_list", AddMovieToWatchlistRequest{SpaceRequest: valid.SpaceRequest, ListID: dbo4listus.DoTasksListID, TmdbID: 1}, true},
		{"invalid_watch_with", AddMovieToWatchlistRequest{SpaceRequest: valid.SpaceRequest, TmdbID: 1, WatchWith: &dbo4listus.WatchWith{Mode: dbo4listus.WatchWithModeContact}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.request.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetListItemWatchWithRequestValidate(t *testing.T) {
	request := SetListItemWatchWithRequest{
		ListItemRequest: ListItemRequest{ListRequest: validListRequest(), ItemID: "item1"},
		WatchWith:       dbo4listus.WatchWith{Mode: dbo4listus.WatchWithModeAlone},
	}
	if err := request.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	request.WatchWith = dbo4listus.WatchWith{Mode: "wrong"}
	if err := request.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want invalid watch-with error")
	}
}
