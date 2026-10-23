package dto4listus

import (
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
}
