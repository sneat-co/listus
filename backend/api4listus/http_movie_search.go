package api4listus

import (
	"net/http"

	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/listus/backend/facade4listus"
	"github.com/sneat-co/sneat-go-core/apicore"
	"github.com/sneat-co/sneat-go-core/apicore/verify"
)

var searchMovies = facade4listus.SearchMovies

// httpPostSearchMovies searches TMDB (title + actor) for watch-list candidates.
func httpPostSearchMovies(w http.ResponseWriter, r *http.Request) {
	var request dto4listus.MovieSearchRequest
	ctx, err := apicore.VerifyAuthenticatedRequestAndDecodeBody(w, r, verify.DefaultJsonWithAuthRequired, &request)
	if err != nil {
		return
	}
	response, err := searchMovies(ctx, request)
	apicore.ReturnJSON(ctx, w, r, http.StatusOK, err, &response)
}
