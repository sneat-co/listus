// Copyright 2026 Sneat.app

package botapi

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
	"github.com/sneat-co/listus/backend/dal4listus"
	"github.com/sneat-co/listus/backend/dbo4listus"
	"github.com/sneat-co/sneat-core-modules/spaceus/dbo4spaceus"
	"github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/sneat-co/sneat-go-core/sneatcoretesting"
)

const (
	serviceTestUserID  = "user1"
	serviceTestSpaceID = "space1"
)

func serviceTestContext(t *testing.T) context.Context {
	t.Helper()
	db := sneatcoretesting.NewMemoryDB()
	space := dbo4spaceus.NewSpaceEntry(serviceTestSpaceID)
	now := time.Now()
	space.Data.Type = coretypes.SpaceTypeFamily
	space.Data.Title = "Test family"
	space.Data.Status = dbmodels.StatusActive
	space.Data.CreatedAt = now
	space.Data.CreatedBy = "seed"
	space.Data.IncreaseVersion(now, "seed")
	space.Data.UserIDs = []string{serviceTestUserID}
	if err := db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Insert(ctx, space.Record)
	}); err != nil {
		t.Fatalf("seed space: %v", err)
	}
	return facade.WithSneatDB(context.Background(), db)
}

func serviceListRef() ListRef {
	return ListRef{
		Space: SpaceRef{ID: serviceTestSpaceID, Type: "family"},
		ID:    TasksListID,
	}
}

func TestListServicePresentationOperations(t *testing.T) {
	ctx := serviceTestContext(t)
	ref := serviceListRef()
	created, err := CreateListItems(ctx, serviceTestUserID, CreateListItemsRequest{
		List:  ref,
		Items: []ListItemInput{{ID: "item1", Title: "Milk"}},
	})
	if err != nil || len(created.Items) != 1 {
		t.Fatalf("CreateListItems() = %+v, %v", created, err)
	}
	if created.Items[0].Title != "Milk" || created.List.Count != 1 {
		t.Fatalf("CreateListItems() = %+v, want rendered item and list count", created)
	}
	list, err := GetList(ctx, serviceTestUserID, ref)
	if err != nil || len(list.Items) != 1 {
		t.Fatalf("GetList() = %+v, %v", list, err)
	}
	changed, err := SetListItemsDone(ctx, serviceTestUserID, SetListItemsDoneRequest{
		ListItemsRequest: ListItemsRequest{List: ref, ItemIDs: []string{"item1"}},
		IsDone:           true,
	})
	if err != nil || len(changed.Items) != 1 || changed.Items[0].Status != "done" {
		t.Fatalf("SetListItemsDone() = %+v, %v", changed, err)
	}
	cleared, err := ClearListItems(ctx, serviceTestUserID, ref)
	if err != nil || len(cleared.Items) != 0 || cleared.Count != 0 {
		t.Fatalf("ClearListItems() = %+v, %v", cleared, err)
	}
	list, err = GetList(ctx, serviceTestUserID, ref)
	if err != nil || len(list.Items) != 0 || list.Count != 0 {
		t.Fatalf("GetList() after ClearListItems = %+v, %v", list, err)
	}
}

func TestBotDataUsesValueOnlyContractTypes(t *testing.T) {
	ref := ListRef{Space: SpaceRef{ID: "space1", Type: "family"}, ID: TasksListID}
	if ref.Space.ID != "space1" || ref.ID != TasksListID {
		t.Fatalf("unexpected list reference: %+v", ref)
	}
}

func TestDeleteListItemsReturnsPresentationValues(t *testing.T) {
	ctx := serviceTestContext(t)
	ref := serviceListRef()
	_, err := CreateListItems(ctx, serviceTestUserID, CreateListItemsRequest{
		List:  ref,
		Items: []ListItemInput{{ID: "item1", Title: "Milk"}},
	})
	if err != nil {
		t.Fatalf("CreateListItems() error = %v", err)
	}
	deleted, err := DeleteListItems(ctx, serviceTestUserID, ListItemsRequest{List: ref, ItemIDs: []string{"item1"}})
	if err != nil {
		t.Fatalf("DeleteListItems() error = %v", err)
	}
	if len(deleted.Items) != 1 || deleted.Items[0].ID != "item1" || deleted.List.Count != 0 {
		t.Fatalf("DeleteListItems() = %+v, want deleted presentation item and empty list", deleted)
	}
}

func TestPresentationProjectionsDoNotExposePersistenceValues(t *testing.T) {
	if got := listView(dal4listus.ListEntry{}); got.ID != "" || got.Count != 0 || len(got.Items) != 0 {
		t.Fatalf("listView(empty) = %+v, want zero view", got)
	}
	if got := listItemView(nil); got.ID != "" || got.Title != "" || len(got.Cast) != 0 || got.WatchWith != nil {
		t.Fatalf("listItemView(nil) = %+v, want zero view", got)
	}
	if got := watchWith(nil); got != nil {
		t.Fatalf("watchWith(nil) = %+v, want nil", got)
	}
	if got := watchWithView(nil); got != nil {
		t.Fatalf("watchWithView(nil) = %+v, want nil", got)
	}

	input := &WatchWith{Mode: "space", Ref: "space2", Title: "Friends"}
	stored := watchWith(input)
	if stored == nil || stored.Mode != input.Mode || stored.Ref != input.Ref || stored.Title != input.Title {
		t.Fatalf("watchWith() = %+v, want value copy of %+v", stored, input)
	}
	view := watchWithView(&dbo4listus.WatchWith{Mode: "contact", Ref: "contact1", Title: "Alex"})
	if view == nil || view.Mode != "contact" || view.Ref != "contact1" || view.Title != "Alex" {
		t.Fatalf("watchWithView() = %+v, want presentation value copy", view)
	}
}

func TestIsNotFound(t *testing.T) {
	if !IsNotFound(record.ErrRecordNotFound) || IsNotFound(errors.New("other")) {
		t.Fatal("IsNotFound() did not preserve not-found semantics")
	}
}
