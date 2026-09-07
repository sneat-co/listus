package api4listus

import (
	"net/http"

	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/listus/backend/facade4listus"
	calendarfacade "github.com/sneat-co/sneat-ext-contracts/calendarius/facade4calendarius"
	"github.com/sneat-co/sneat-go-core/apicore"
	"github.com/sneat-co/sneat-go-core/apicore/verify"
)

func httpPostSaveListItemDateTask(provider calendarfacade.SourceLinkedDateTaskProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request dto4listus.SaveListItemDateTaskRequest
		ctx, err := apicore.VerifyAuthenticatedRequestAndDecodeBody(w, r, verify.DefaultJsonWithAuthRequired, &request)
		if err != nil {
			return
		}
		response, err := facade4listus.SaveListItemDateTask(ctx, request, provider)
		apicore.ReturnJSON(ctx, w, r, http.StatusOK, err, response)
	}
}
