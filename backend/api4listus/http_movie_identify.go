package api4listus

import (
	"net/http"

	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/listus/backend/facade4listus"
	"github.com/sneat-co/sneat-go-core/apicore"
	"github.com/sneat-co/sneat-go-core/apicore/verify"
)

var identifyMovies = facade4listus.IdentifyMovies

// httpPostIdentifyMovies identifies movies from a vague description (AI title guesses grounded via TMDB).
func httpPostIdentifyMovies(w http.ResponseWriter, r *http.Request) {
	var request dto4listus.MovieIdentifyRequest
	ctx, err := apicore.VerifyAuthenticatedRequestAndDecodeBody(w, r, verify.DefaultJsonWithAuthRequired, &request)
	if err != nil {
		return
	}
	response, err := identifyMovies(ctx, request)
	apicore.ReturnJSON(ctx, w, r, http.StatusOK, err, &response)
}
