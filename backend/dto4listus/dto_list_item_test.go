package dto4listus

import (
	"testing"

	"github.com/sneat-co/listus/backend/dbo4listus"
	"github.com/sneat-co/sneat-core-modules/spaceus/dto4spaceus"
	"github.com/sneat-co/sneat-go-core/coretypes"
)

func validListRequest() ListRequest {
	return ListRequest{
		SpaceRequest: dto4spaceus.SpaceRequest{SpaceID: coretypes.SpaceID("s1")},
		ListID:       dbo4listus.NewListKey(dbo4listus.ListTypeToDo, "123"),
	}
}

func TestCreateListItemRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateListItemRequest
		wantErr bool
	}{
		{"valid_no_id", CreateListItemRequest{ListItemBase: dbo4listus.ListItemBase{Title: "Milk"}}, false},
		{"valid_with_id", CreateListItemRequest{ID: "abc", ListItemBase: dbo4listus.ListItemBase{Title: "Milk"}}, false},
		{"bad_id", CreateListItemRequest{ID: "has space", ListItemBase: dbo4listus.ListItemBase{Title: "Milk"}}, true},
		{"missing_title", CreateListItemRequest{ListItemBase: dbo4listus.ListItemBase{}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateListItemsRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateListItemsRequest
		wantErr bool
	}{
		{"valid", CreateListItemsRequest{
			ListRequest: validListRequest(),
			Items:       []CreateListItemRequest{{ListItemBase: dbo4listus.ListItemBase{Title: "Milk"}}},
		}, false},
		{"empty_items", CreateListItemsRequest{ListRequest: validListRequest()}, false},
		{"invalid_item", CreateListItemsRequest{
			ListRequest: validListRequest(),
			Items:       []CreateListItemRequest{{ListItemBase: dbo4listus.ListItemBase{}}},
		}, true},
		{"missing_space", CreateListItemsRequest{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListItemRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     ListItemRequest
		wantErr bool
	}{
		{"valid", ListItemRequest{ListRequest: validListRequest(), ItemID: "i1"}, false},
		{"missing_item", ListItemRequest{ListRequest: validListRequest()}, true},
		{"bad_list", ListItemRequest{ItemID: "i1"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListItemIDsRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     ListItemIDsRequest
		wantErr bool
	}{
		{"valid", ListItemIDsRequest{ListRequest: validListRequest(), ItemIDs: []string{"a", "b"}}, false},
		{"empty_allowed", ListItemIDsRequest{ListRequest: validListRequest()}, false},
		{"blank_id", ListItemIDsRequest{ListRequest: validListRequest(), ItemIDs: []string{" "}}, true},
		{"bad_list", ListItemIDsRequest{ItemIDs: []string{"a"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReorderListItemsRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     ReorderListItemsRequest
		wantErr bool
	}{
		{"valid", ReorderListItemsRequest{
			ListItemIDsRequest: ListItemIDsRequest{ListRequest: validListRequest(), ItemIDs: []string{"a"}},
			ToIndex:            0,
		}, false},
		{"negative_index", ReorderListItemsRequest{
			ListItemIDsRequest: ListItemIDsRequest{ListRequest: validListRequest(), ItemIDs: []string{"a"}},
			ToIndex:            -1,
		}, true},
		{"bad_ids", ReorderListItemsRequest{
			ListItemIDsRequest: ListItemIDsRequest{ListRequest: validListRequest(), ItemIDs: []string{" "}},
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListItemsSetIsDoneRequest_Validate(t *testing.T) {
	valid := ListItemsSetIsDoneRequest{
		ListItemIDsRequest: ListItemIDsRequest{ListRequest: validListRequest(), ItemIDs: []string{"a"}},
		IsDone:             true,
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	invalid := ListItemsSetIsDoneRequest{
		ListItemIDsRequest: ListItemIDsRequest{ListRequest: validListRequest(), ItemIDs: []string{" "}},
	}
	if err := invalid.Validate(); err == nil {
		t.Error("expected error for blank item id")
	}
}
