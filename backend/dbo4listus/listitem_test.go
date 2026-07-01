package dbo4listus

import (
	"testing"
	"time"

	"github.com/sneat-co/listus/backend/const4listus"
	"github.com/strongo/strongoapp/with"
)

func timeNow() time.Time { return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) }

func validCreated() with.CreatedFields {
	return with.CreatedFields{
		CreatedAtField: with.CreatedAtField{CreatedAt: timeNow()},
		CreatedByField: with.CreatedByField{CreatedBy: "u1"},
	}
}

func TestListItemBase_IsDone(t *testing.T) {
	if !(ListItemBase{Status: const4listus.ListItemStatusDone}).IsDone() {
		t.Error("expected IsDone() true for done status")
	}
	if (ListItemBase{Status: const4listus.ListItemStatusActive}).IsDone() {
		t.Error("expected IsDone() false for active status")
	}
	if (ListItemBase{}).IsDone() {
		t.Error("expected IsDone() false for empty status")
	}
}

func TestListItemBase_Validate(t *testing.T) {
	if err := (ListItemBase{Title: "Milk"}).Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := (ListItemBase{Title: "   "}).Validate(); err == nil {
		t.Error("expected error for blank title")
	}
}

func TestWatchWith_Validate(t *testing.T) {
	tests := []struct {
		name    string
		v       WatchWith
		wantErr bool
	}{
		{"alone", WatchWith{Mode: WatchWithModeAlone}, false},
		{"space", WatchWith{Mode: WatchWithModeSpace, Ref: "space1", Title: "Family"}, false},
		{"space_missing_ref", WatchWith{Mode: WatchWithModeSpace}, true},
		{"contact", WatchWith{Mode: WatchWithModeContact, Ref: "contact1", Title: "Alex"}, false},
		{"contact_missing_ref", WatchWith{Mode: WatchWithModeContact}, true},
		{"unknown_mode", WatchWith{Mode: "sometimes"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.v.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("WatchWith.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListItemBase_Validate_WatchWith(t *testing.T) {
	valid := ListItemBase{Title: "Inception", WatchWith: &WatchWith{Mode: WatchWithModeAlone}}
	if err := valid.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	invalid := ListItemBase{Title: "Inception", WatchWith: &WatchWith{Mode: WatchWithModeSpace}}
	if err := invalid.Validate(); err == nil {
		t.Error("expected error for watchWith with missing ref")
	}
}

func TestListItemBrief_Validate(t *testing.T) {
	tests := []struct {
		name    string
		v       ListItemBrief
		wantErr bool
	}{
		{"valid", ListItemBrief{ID: "1", ListItemBase: ListItemBase{Title: "Milk"}, CreatedFields: validCreated()}, false},
		{"missing_id", ListItemBrief{ListItemBase: ListItemBase{Title: "Milk"}, CreatedFields: validCreated()}, true},
		{"missing_title", ListItemBrief{ID: "1", CreatedFields: validCreated()}, true},
		{"missing_created", ListItemBrief{ID: "1", ListItemBase: ListItemBase{Title: "Milk"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.v.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("ListItemBrief.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
