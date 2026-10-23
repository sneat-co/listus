package dbo4listus

import (
	"strings"
	"testing"

	"github.com/sneat-co/sneat-go-core/coretypes"
)

func TestNewListTemplateApplyReceiptKey(t *testing.T) {
	const spaceID coretypes.SpaceID = "space1"
	const receiptID = "receipt1"
	key := NewListTemplateApplyReceiptKey(spaceID, receiptID)
	if key == nil {
		t.Fatal("expected a receipt key")
	}
	got := key.String()
	for _, part := range []string{string(spaceID), ListTemplateApplyReceiptsCollection, receiptID} {
		if !strings.Contains(got, part) {
			t.Fatalf("key %q does not contain %q", got, part)
		}
	}
}
