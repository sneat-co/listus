package api4listus

import (
	"net/http"

	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/listus/backend/facade4listus"
	"github.com/sneat-co/sneat-go-core/apicore"
	"github.com/sneat-co/sneat-go-core/apicore/verify"
)

var addMovieToWatchlist = facade4listus.AddMovieToWatchlist

// httpPostAddMovieToWatchlist resolves a movie (tmdbID or free-text query) and
// appends it to the space's watch!movies list.
func httpPostAddMovieToWatchlist(w http.ResponseWriter, r *http.Request) {
	var request dto4listus.AddMovieToWatchlistRequest
	ctx, err := apicore.VerifyAuthenticatedRequestAndDecodeBody(w, r, verify.DefaultJsonWithAuthRequired, &request)
	if err != nil {
		return
	}
	response, err := addMovieToWatchlist(ctx, request)
	apicore.ReturnJSON(ctx, w, r, http.StatusCreated, err, &response)
}
