package facade4listus

import (
	"context"
	"testing"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record/update"
	"github.com/sneat-co/listus/backend/dal4listus"
	"github.com/sneat-co/listus/backend/dbo4listus"
	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/sneat-go-core/coretypes"
)

func sourceManagedItem() *dbo4listus.SourceManagement {
	return &dbo4listus.SourceManagement{
		Source:  dbo4listusItemRef("debtus", "sourceObligations", "invoice-1"),
		Purpose: "payment-due", ActionID: "record-payment",
		Disposition: dbo4listus.SourceCompletionRequiresInput,
	}
}

func dbo4listusItemRef(extID, collection, itemID string) coretypes.ItemRef {
	return coretypes.NewItemRefSameSpace(coretypes.ExtID(extID), collection, itemID)
}

func markItemSourceManaged(t *testing.T, ctx context.Context, db dal.DB, itemID string) {
	t.Helper()
	entry := dal4listus.NewListEntry(testSpaceID, dbo4listus.DoTasksListID)
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		if err := tx.Get(ctx, entry.Record); err != nil {
			return err
		}
		for _, item := range entry.Data.Items {
			if item.ID == itemID {
				item.SourceManagement = sourceManagedItem()
			}
		}
		return tx.Update(ctx, entry.Key, []update.Update{update.ByFieldName("items", entry.Data.Items)})
	}); err != nil {
		t.Fatalf("mark item source managed: %v", err)
	}
}

func markItemDateTask(t *testing.T, ctx context.Context, db dal.DB, itemID string) {
	t.Helper()
	entry := dal4listus.NewListEntry(testSpaceID, dbo4listus.DoTasksListID)
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		if err := tx.Get(ctx, entry.Record); err != nil {
			return err
		}
		for _, item := range entry.Data.Items {
			if item.ID == itemID {
				item.DateTask = &dbo4listus.DateTaskLink{
					Happening: dbo4listusItemRef("calendarius", "happenings", "due-item"),
					Source:    dbo4listusItemRef("listus", "lists", string(dbo4listus.DoTasksListID)),
					Purpose:   "due-date",
					Revision:  1,
				}
			}
		}
		return tx.Update(ctx, entry.Key, []update.Update{update.ByFieldName("items", entry.Data.Items)})
	}); err != nil {
		t.Fatalf("mark item with date task: %v", err)
	}
}

func TestSourceManagedItemRejectsDirectCompletionAndDeletion(t *testing.T) {
	ctx, db := newTestDBWithSpace(t, testSpaceID, testUserID)
	created := createItems(t, ctx, dbo4listus.DoTasksListID, "Pay invoice")
	itemID := created.CreatedItems[0].ID
	markItemSourceManaged(t, ctx, db, itemID)

	_, _, err := SetListItemsIsDone(userCtx(ctx, testUserID), dto4listus.ListItemsSetIsDoneRequest{
		ListItemIDsRequest: dto4listus.ListItemIDsRequest{ListRequest: listRequest(testSpaceID, dbo4listus.DoTasksListID), ItemIDs: []string{itemID}},
		IsDone:             true,
	})
	if err == nil {
		t.Fatal("expected direct completion to be rejected")
	}
	_, _, err = DeleteListItems(userCtx(ctx, testUserID), dto4listus.ListItemIDsRequest{
		ListRequest: listRequest(testSpaceID, dbo4listus.DoTasksListID), ItemIDs: []string{"*"},
	})
	if err == nil {
		t.Fatal("expected wildcard deletion to be rejected")
	}

	item := getListData(t, ctx, dbo4listus.DoTasksListID).Items[0]
	if item.IsDone() || item.SourceManagement == nil {
		t.Fatalf("rejected commands changed protected item: %+v", item)
	}
}

func TestDateTaskItemRejectsDirectCompletionAndDeletion(t *testing.T) {
	ctx, db := newTestDBWithSpace(t, testSpaceID, testUserID)
	created := createItems(t, ctx, dbo4listus.DoTasksListID, "Pay electricity")
	itemID := created.CreatedItems[0].ID
	markItemDateTask(t, ctx, db, itemID)

	_, _, err := SetListItemsIsDone(userCtx(ctx, testUserID), dto4listus.ListItemsSetIsDoneRequest{
		ListItemIDsRequest: dto4listus.ListItemIDsRequest{ListRequest: listRequest(testSpaceID, dbo4listus.DoTasksListID), ItemIDs: []string{itemID}},
		IsDone:             true,
	})
	if err == nil {
		t.Fatal("expected linked date task completion to be rejected")
	}
	_, _, err = DeleteListItems(userCtx(ctx, testUserID), dto4listus.ListItemIDsRequest{
		ListRequest: listRequest(testSpaceID, dbo4listus.DoTasksListID), ItemIDs: []string{itemID},
	})
	if err == nil {
		t.Fatal("expected linked date task deletion to be rejected")
	}

	item := getListData(t, ctx, dbo4listus.DoTasksListID).Items[0]
	if item.IsDone() || item.DateTask == nil {
		t.Fatalf("rejected commands changed dated item: %+v", item)
	}
}

func TestCreateListItemRejectsServerManagedFields(t *testing.T) {
	base := dbo4listus.ListItemBase{Title: "Pay invoice", SourceManagement: sourceManagedItem()}
	request := dto4listus.CreateListItemRequest{ID: "invoice", ListItemBase: base}
	if err := request.Validate(); err == nil {
		t.Fatal("expected sourceManagement injection to be rejected")
	}
}
