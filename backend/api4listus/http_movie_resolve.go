package api4listus

import (
	"net/http"

	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/listus/backend/facade4listus"
	"github.com/sneat-co/sneat-go-core/apicore"
	"github.com/sneat-co/sneat-go-core/apicore/verify"
)

var resolveMovie = facade4listus.ResolveMovie

// httpPostResolveMovie fully resolves a single movie (overview, poster, cast, trailer) by TMDB id.
func httpPostResolveMovie(w http.ResponseWriter, r *http.Request) {
	var request dto4listus.MovieResolveRequest
	ctx, err := apicore.VerifyAuthenticatedRequestAndDecodeBody(w, r, verify.DefaultJsonWithAuthRequired, &request)
	if err != nil {
		return
	}
	response, err := resolveMovie(ctx, request)
	apicore.ReturnJSON(ctx, w, r, http.StatusOK, err, &response)
}
