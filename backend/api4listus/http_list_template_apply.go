package api4listus

import (
	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/listus/backend/facade4listus"
	"github.com/sneat-co/sneat-go-core/apicore"
	"github.com/sneat-co/sneat-go-core/apicore/verify"
	"net/http"
)

func httpPostApplyListTemplate(w http.ResponseWriter, r *http.Request) {
	var request dto4listus.ApplyListTemplateRequest
	ctx, err := apicore.VerifyAuthenticatedRequestAndDecodeBody(w, r, verify.DefaultJsonWithAuthRequired, &request)
	if err != nil {
		return
	}
	result, err := facade4listus.ApplyListTemplate(ctx, request)
	apicore.ReturnJSON(ctx, w, r, http.StatusOK, err, &result)
}
