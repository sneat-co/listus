package dbo4listus

import (
	"testing"

	"github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/strongo/strongoapp/with"
)

func TestNewListKey(t *testing.T) {
	if got := NewListKey(ListTypeToDo, "123"); got != ListKey("do!123") {
		t.Errorf("NewListKey() = %v, want do!123", got)
	}
}

func TestIsStandardList(t *testing.T) {
	tests := []struct {
		key  ListKey
		want bool
	}{
		{BuyGroceriesListID, true},
		{DoTasksListID, true},
		{ReadBooksListID, true},
		{WatchMoviesListID, true},
		{"do!custom", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsStandardList(tt.key); got != tt.want {
			t.Errorf("IsStandardList(%q) = %v, want %v", tt.key, got, tt.want)
		}
	}
}

func TestListKey_ListType(t *testing.T) {
	tests := []struct {
		key  ListKey
		want ListType
	}{
		{"do!123", ListTypeToDo},
		{"buy!groceries", ListTypeToBuy},
		{"noseparator", ""},
		{"", ""},
	}
	for _, tt := range tests {
		if got := tt.key.ListType(); got != tt.want {
			t.Errorf("ListKey(%q).ListType() = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestListKey_ListSubID(t *testing.T) {
	tests := []struct {
		key  ListKey
		want string
	}{
		{"do!123", "123"},
		{"buy!groceries", "groceries"},
		{"noseparator", ""},
		{"", ""},
	}
	for _, tt := range tests {
		if got := tt.key.ListSubID(); got != tt.want {
			t.Errorf("ListKey(%q).ListSubID() = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestListKey_Validate_LeadingTrailingSpaces(t *testing.T) {
	if err := ListKey(" do!123").Validate(); err == nil {
		t.Error("expected error for leading space")
	}
	if err := ListKey("do!123!extra").Validate(); err == nil {
		t.Error("expected error for multiple separators")
	}
	if err := ListKey("!123").Validate(); err == nil {
		t.Error("expected error for empty list type")
	}
}

func TestListBase_Validate(t *testing.T) {
	tests := []struct {
		name    string
		v       ListBase
		wantErr bool
	}{
		{"valid", ListBase{Type: ListTypeToDo, Title: "t"}, false},
		{"missing_type", ListBase{Title: "t"}, true},
		{"unknown_type", ListBase{Type: "nope", Title: "t"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.v.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("ListBase.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListGroup_Validate(t *testing.T) {
	tests := []struct {
		name    string
		v       ListGroup
		wantErr bool
	}{
		{"valid", ListGroup{Type: "t", Title: "Title"}, false},
		{"missing_type", ListGroup{Title: "Title"}, true},
		{"missing_title", ListGroup{Type: "t"}, true},
		{"invalid_brief", ListGroup{Type: "t", Title: "Title", Lists: ListBriefs{
			"do!1": {ListBase: ListBase{Title: "x"}}, // missing type
		}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.v.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("ListGroup.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListBrief_Validate(t *testing.T) {
	tests := []struct {
		name    string
		v       ListBrief
		wantErr bool
	}{
		{"valid", ListBrief{ListBase: ListBase{Type: ListTypeToDo}, ItemsCount: 0}, false},
		{"negative_count", ListBrief{ListBase: ListBase{Type: ListTypeToDo}, ItemsCount: -1}, true},
		{"invalid_base", ListBrief{ListBase: ListBase{}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.v.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("ListBrief.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListDbo_AddListItem(t *testing.T) {
	t.Run("adds new item", func(t *testing.T) {
		dbo := &ListDbo{}
		item := &ListItemBrief{ID: "1", ListItemBase: ListItemBase{Title: "Milk"}}
		got := dbo.AddListItem(item)
		if got != item {
			t.Error("expected returned item to be the added item")
		}
		if dbo.Count != 1 || len(dbo.Items) != 1 {
			t.Errorf("Count=%d len(Items)=%d, want 1/1", dbo.Count, len(dbo.Items))
		}
	})
	t.Run("dedup by title+emoji returns existing", func(t *testing.T) {
		existing := &ListItemBrief{ID: "1", ListItemBase: ListItemBase{Title: "Milk"}}
		dbo := &ListDbo{Items: []*ListItemBrief{existing}, Count: 1}
		got := dbo.AddListItem(&ListItemBrief{ID: "2", ListItemBase: ListItemBase{Title: "Milk"}})
		if got != existing {
			t.Error("expected existing item to be returned")
		}
		if len(dbo.Items) != 1 {
			t.Errorf("len(Items)=%d, want 1 (no append)", len(dbo.Items))
		}
	})
	t.Run("re-adding a done item reactivates it", func(t *testing.T) {
		existing := &ListItemBrief{ID: "1", ListItemBase: ListItemBase{Title: "Milk", Status: "done"}}
		dbo := &ListDbo{Items: []*ListItemBrief{existing}, Count: 1}
		dbo.AddListItem(&ListItemBrief{ID: "2", ListItemBase: ListItemBase{Title: "Milk", Status: "active"}})
		if existing.Status != "active" {
			t.Errorf("status = %q, want active", existing.Status)
		}
	})
}

func validListDbo() *ListDbo {
	return &ListDbo{
		ListBase: ListBase{Type: ListTypeToDo},
		WithModified: dbmodels.WithModified{
			CreatedFields: with.CreatedFields{
				CreatedAtField: with.CreatedAtField{CreatedAt: timeNow()},
				CreatedByField: with.CreatedByField{CreatedBy: "u1"},
			},
		},
		WithUserIDs:  dbmodels.WithUserIDs{UserIDs: []string{"u1"}},
		WithSpaceIDs: dbmodels.WithSpaceIDs{SpaceIDs: []coretypes.SpaceID{"s1"}},
	}
}

func TestListDbo_Validate(t *testing.T) {
	t.Run("valid empty", func(t *testing.T) {
		if err := validListDbo().Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("negative count", func(t *testing.T) {
		dbo := validListDbo()
		dbo.Count = -1
		if err := dbo.Validate(); err == nil {
			t.Error("expected error for negative count")
		}
	})
	t.Run("count mismatch", func(t *testing.T) {
		dbo := validListDbo()
		dbo.Items = []*ListItemBrief{{ID: "1", ListItemBase: ListItemBase{Title: "x"}, CreatedFields: validCreated()}}
		dbo.Count = 5
		if err := dbo.Validate(); err == nil {
			t.Error("expected error for count != len(items)")
		}
	})
	t.Run("invalid item", func(t *testing.T) {
		dbo := validListDbo()
		dbo.Items = []*ListItemBrief{{ID: "", ListItemBase: ListItemBase{Title: "x"}}}
		dbo.Count = 1
		if err := dbo.Validate(); err == nil {
			t.Error("expected error for invalid item")
		}
	})
}

func TestListusSpaceDbo_Validate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		v := ListusSpaceDbo{CreatedFields: validCreated()}
		if err := v.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("invalid brief", func(t *testing.T) {
		v := ListusSpaceDbo{
			CreatedFields: validCreated(),
			Lists:         ListBriefs{"do!1": {ListBase: ListBase{Title: "x"}}},
		}
		if err := v.Validate(); err == nil {
			t.Error("expected error for invalid list brief")
		}
	})
}
