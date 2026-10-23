package dto4listus

import (
	"strings"
	"testing"

	"github.com/sneat-co/listus/backend/dbo4listus"
)

func TestApplyListTemplateRequestValidation(t *testing.T) {
	valid := ApplyListTemplateRequest{
		SpaceID:           "home",
		SourceListID:      "buy!regular",
		DestinationListID: dbo4listus.BuyGroceriesListID,
		RequestID:         "click-1",
	}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	invalid := valid
	invalid.RequestID = "click\x00one"
	if err := invalid.Validate(); err == nil {
		t.Fatal("control character in requestID was accepted")
	}
	invalid = valid
	invalid.DestinationListID = dbo4listus.DoTasksListID
	if err := invalid.Validate(); err == nil {
		t.Fatal("cross-type template apply was accepted")
	}
	for _, mutate := range []func(*ApplyListTemplateRequest){
		func(r *ApplyListTemplateRequest) { r.SpaceID = "" },
		func(r *ApplyListTemplateRequest) { r.RequestID = "" },
		func(r *ApplyListTemplateRequest) { r.RequestID = strings.Repeat("x", 201) },
		func(r *ApplyListTemplateRequest) { r.SourceListID = "not-a-list" },
		func(r *ApplyListTemplateRequest) { r.DestinationListID = "not-a-list" },
		func(r *ApplyListTemplateRequest) { r.DestinationListID = r.SourceListID },
		func(r *ApplyListTemplateRequest) {
			r.SourceListID = dbo4listus.WatchMoviesListID
			r.DestinationListID = "watch!other"
		},
	} {
		next := valid
		mutate(&next)
		if err := next.Validate(); err == nil {
			t.Fatalf("invalid request was accepted: %+v", next)
		}
	}
}
