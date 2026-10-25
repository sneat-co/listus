package facade4listus

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
	"github.com/dal-go/record/update"
	"github.com/sneat-co/listus/backend/dal4listus"
	"github.com/sneat-co/listus/backend/dbo4listus"
	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/sneat-co/sneat-go-core/facade"
)

func TestSetListItemWatchWith_Errors(t *testing.T) {
	ctx, _ := newTestDBWithSpace(t, testSpaceID, testUserID)
	uctx := userCtx(ctx, testUserID)

	// Validation error (empty space/list)
	if _, _, err := SetListItemWatchWith(uctx, dto4listus.SetListItemWatchWithRequest{}); err == nil {
		t.Error("expected validation error")
	}

	// getListWorkerRecords error
	origGet := getListWorkerRecords
	defer func() { getListWorkerRecords = origGet }()
	getListWorkerRecords = func(_ *dal4listus.ListWorkerParams, _ facade.ContextWithUser, _ dal.ReadwriteTransaction) error {
		return errors.New("simulated get records error")
	}
	_, _, err := SetListItemWatchWith(uctx, dto4listus.SetListItemWatchWithRequest{
		ListItemRequest: dto4listus.ListItemRequest{
			ListRequest: listRequest(testSpaceID, dbo4listus.DoTasksListID),
			ItemID:      "i1",
		},
		WatchWith: dbo4listus.WatchWith{Mode: dbo4listus.WatchWithModeAlone},
	})
	if err == nil {
		t.Error("expected error from getListWorkerRecords")
	}
}

func TestSetListItemsIsDone_Errors(t *testing.T) {
	ctx, _ := newTestDBWithSpace(t, testSpaceID, testUserID)
	uctx := userCtx(ctx, testUserID)

	// Validation error
	if _, _, err := SetListItemsIsDone(uctx, dto4listus.ListItemsSetIsDoneRequest{}); err == nil {
		t.Error("expected validation error")
	}

	// getListWorkerRecords error
	origGet := getListWorkerRecords
	defer func() { getListWorkerRecords = origGet }()
	getListWorkerRecords = func(_ *dal4listus.ListWorkerParams, _ facade.ContextWithUser, _ dal.ReadwriteTransaction) error {
		return errors.New("simulated get records error")
	}
	_, _, err := SetListItemsIsDone(uctx, dto4listus.ListItemsSetIsDoneRequest{
		ListItemIDsRequest: dto4listus.ListItemIDsRequest{
			ListRequest: listRequest(testSpaceID, dbo4listus.DoTasksListID),
			ItemIDs:     []string{"i1"},
		},
		IsDone: true,
	})
	if err == nil {
		t.Error("expected error from getListWorkerRecords")
	}
}

func TestDeleteListItems_Coverage(t *testing.T) {
	ctx, db := newTestDBWithSpace(t, testSpaceID, testUserID)
	uctx := userCtx(ctx, testUserID)

	// Validation error
	if _, _, err := DeleteListItems(uctx, dto4listus.ListItemIDsRequest{}); err == nil {
		t.Error("expected validation error")
	}

	// Create list with items and delete with wildcard "*"
	createItems(t, ctx, "do!tasks", "Item 1", "Item 2")
	deleted, _, err := DeleteListItems(uctx, dto4listus.ListItemIDsRequest{
		ListRequest: listRequest(testSpaceID, "do!tasks"),
		ItemIDs:     []string{"*"},
	})
	if err != nil {
		t.Fatalf("DeleteListItems with wildcard failed: %v", err)
	}
	if len(deleted) != 2 {
		t.Errorf("expected 2 deleted items, got %d", len(deleted))
	}

	// Test isInRecentItems branch: matching Title and Emoji with different ID
	// And test len(list.Data.RecentItems) >= 100
	now := time.Now()
	listEntry := dal4listus.NewListEntry(testSpaceID, "do!recent")
	listEntry.Data.SpaceIDs = []coretypes.SpaceID{testSpaceID}
	listEntry.Data.UserIDs = []string{testUserID}
	listEntry.Data.Type = dbo4listus.ListTypeToDo
	itemNew := &dbo4listus.ListItemBrief{ID: "item-new", ListItemBase: dbo4listus.ListItemBase{Title: "Apple", Emoji: "🍏"}}
	itemNew.CreatedAt = now
	itemNew.CreatedBy = testUserID
	itemExtra := &dbo4listus.ListItemBrief{ID: "item-extra", ListItemBase: dbo4listus.ListItemBase{Title: "Orange", Emoji: "🍊"}}
	itemExtra.CreatedAt = now
	itemExtra.CreatedBy = testUserID
	listEntry.Data.Items = []*dbo4listus.ListItemBrief{itemNew, itemExtra}

	// Seed with 100 recent items, including one with same Title/Emoji as item-new but ID "recent-0"
	listEntry.Data.RecentItems = make([]*dbo4listus.ListItemBrief, 100)
	recent0 := &dbo4listus.ListItemBrief{ID: "recent-0", ListItemBase: dbo4listus.ListItemBase{Title: "Apple", Emoji: "🍏"}}
	recent0.CreatedAt = now
	recent0.CreatedBy = testUserID
	listEntry.Data.RecentItems[0] = recent0
	for i := 1; i < 100; i++ {
		other := &dbo4listus.ListItemBrief{ID: "recent-other", ListItemBase: dbo4listus.ListItemBase{Title: "Other"}}
		other.CreatedAt = now
		other.CreatedBy = testUserID
		listEntry.Data.RecentItems[i] = other
	}
	listEntry.Data.Count = len(listEntry.Data.Items)
	if err := db.RunReadwriteTransaction(ctx, func(txCtx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Insert(txCtx, listEntry.Record)
	}); err != nil {
		t.Fatalf("failed to insert test list: %v", err)
	}

	// Deleting item-new (whose Title and Emoji match recent-0) and item-extra
	deleted2, listRes, err := DeleteListItems(uctx, dto4listus.ListItemIDsRequest{
		ListRequest: listRequest(testSpaceID, "do!recent"),
		ItemIDs:     []string{"item-new", "item-extra"},
	})
	if err != nil {
		t.Fatalf("DeleteListItems failed: %v", err)
	}
	if len(deleted2) != 2 {
		t.Errorf("expected 2 deleted items, got %d", len(deleted2))
	}
	if len(listRes.Data.RecentItems) != 100 {
		t.Errorf("expected 100 recent items max, got %d", len(listRes.Data.RecentItems))
	}
}

func TestReorderListItem_Coverage(t *testing.T) {
	ctx, _ := newTestDBWithSpace(t, testSpaceID, testUserID)
	uctx := userCtx(ctx, testUserID)

	// Empty user ID
	noUser := fakeNoUserCtx{ContextWithUser: uctx}
	err := ReorderListItem(noUser, dto4listus.ReorderListItemsRequest{
		ListItemIDsRequest: dto4listus.ListItemIDsRequest{
			ListRequest: listRequest(testSpaceID, "do!tasks"),
			ItemIDs:     []string{"i1"},
		},
		ToIndex: 0,
	})
	if err == nil {
		t.Error("expected error for empty user ID")
	}

	// Non-existent list -> GetListForUpdate fails
	err = ReorderListItem(uctx, dto4listus.ReorderListItemsRequest{
		ListItemIDsRequest: dto4listus.ListItemIDsRequest{
			ListRequest: listRequest(testSpaceID, "do!nonexistent"),
			ItemIDs:     []string{"i1"},
		},
		ToIndex: 0,
	})
	if err == nil {
		t.Error("expected error for non-existent list")
	}

	// Create a list with 4 items: a, b, c, d
	reorderListID := string(dbo4listus.DoTasksListID)
	createItems(t, ctx, reorderListID, "Item A", "Item B", "Item C", "Item D")

	// toIndex >= len(items) (e.g. 100)
	err = ReorderListItem(uctx, dto4listus.ReorderListItemsRequest{
		ListItemIDsRequest: dto4listus.ListItemIDsRequest{
			ListRequest: listRequest(testSpaceID, reorderListID),
			ItemIDs:     []string{"id-a"},
		},
		ToIndex: 100,
	})
	if err != nil {
		t.Fatalf("ReorderListItem with toIndex >= len failed: %v", err)
	}

	// len(otherItems) < toIndex: move 2 items, remaining otherItems has length 2, but toIndex is 3
	err = ReorderListItem(uctx, dto4listus.ReorderListItemsRequest{
		ListItemIDsRequest: dto4listus.ListItemIDsRequest{
			ListRequest: listRequest(testSpaceID, reorderListID),
			ItemIDs:     []string{"id-b", "id-c"},
		},
		ToIndex: 3,
	})
	if err != nil {
		t.Fatalf("ReorderListItem with len(otherItems) < toIndex failed: %v", err)
	}

	// tx.Update fails via hook
	origUp := updateListRecordInTx
	defer func() { updateListRecordInTx = origUp }()
	updateListRecordInTx = func(ctx context.Context, tx dal.ReadwriteTransaction, key *record.Key, updates []update.Update) error {
		return errors.New("simulated tx update failure")
	}
	err = ReorderListItem(uctx, dto4listus.ReorderListItemsRequest{
		ListItemIDsRequest: dto4listus.ListItemIDsRequest{
			ListRequest: listRequest(testSpaceID, reorderListID),
			ItemIDs:     []string{"id-a"},
		},
		ToIndex: 0,
	})
	if err == nil {
		t.Error("expected error from updateListRecordInTx")
	}
}

func TestCreateListItems_Coverage(t *testing.T) {
	ctx, _ := newTestDBWithSpace(t, testSpaceID, testUserID)
	uctx := userCtx(ctx, testUserID)

	// 1. Validation error
	if _, _, err := CreateListItems(uctx, dto4listus.CreateListItemsRequest{}); err == nil {
		t.Error("expected validation error")
	}

	// 2. getListWorkerRecords error
	origGet := getListWorkerRecords
	defer func() { getListWorkerRecords = origGet }()
	getListWorkerRecords = func(_ *dal4listus.ListWorkerParams, _ facade.ContextWithUser, _ dal.ReadwriteTransaction) error {
		return errors.New("simulated get records error")
	}
	_, _, err := CreateListItems(uctx, dto4listus.CreateListItemsRequest{
		ListRequest: listRequest(testSpaceID, "do!tasks"),
		Items:       []dto4listus.CreateListItemRequest{{ID: "1", ListItemBase: dbo4listus.ListItemBase{Title: "T"}}},
	})
	if err == nil {
		t.Error("expected error from getListWorkerRecords")
	}
	getListWorkerRecords = origGet

	// 3. Non-standard list that does not exist -> fails in createListItemsTxWorker
	_ = dal4listus.RunListWorker(uctx, listRequest(testSpaceID, dbo4listus.DoTasksListID),
		func(ctx facade.ContextWithUser, tx dal.ReadwriteTransaction, params *dal4listus.ListWorkerParams) error {
			req := dto4listus.CreateListItemsRequest{
				ListRequest: listRequest(testSpaceID, "do!custom-missing"),
				Items:       []dto4listus.CreateListItemRequest{{ID: "1", ListItemBase: dbo4listus.ListItemBase{Title: "T"}}},
			}
			_, _, err := createListItemsTxWorker(ctx, tx, req, params)
			if err == nil || !strings.Contains(err.Error(), "list not found") {
				t.Errorf("expected list not found error, got: %v", err)
			}
			return nil
		},
	)

	// 4. ListTypeToWatch (watch!movies): standard list sets emoji to "📽️"
	respWatch, listWatch, err := CreateListItems(uctx, dto4listus.CreateListItemsRequest{
		ListRequest: listRequest(testSpaceID, dbo4listus.WatchMoviesListID),
		Items:       []dto4listus.CreateListItemRequest{{ID: "w1", ListItemBase: dbo4listus.ListItemBase{Title: "Movie"}}},
	})
	if err != nil {
		t.Fatalf("CreateListItems on watch list failed: %v", err)
	}
	if listWatch.Data.Emoji != "📽️" {
		t.Errorf("expected emoji 📽️, got %q", listWatch.Data.Emoji)
	}
	_ = respWatch

	// 5. ListTypeToBuy and groceries list: standard list groceries sets brief emoji to "🛒"
	respBuy, _, err := CreateListItems(uctx, dto4listus.CreateListItemsRequest{
		ListRequest: listRequest(testSpaceID, string(dbo4listus.BuyGroceriesListID)),
		Items:       []dto4listus.CreateListItemRequest{{ID: "b1", ListItemBase: dbo4listus.ListItemBase{Title: "Milk"}}},
	})
	if err != nil {
		t.Fatalf("CreateListItems on groceries list failed: %v", err)
	}
	_ = respBuy

	// 6. Title starts with emoji: "🍏 Apple" -> emoji becomes "🍏" and Title becomes "Apple"
	respEmoji, _, err := CreateListItems(uctx, dto4listus.CreateListItemsRequest{
		ListRequest: listRequest(testSpaceID, dbo4listus.DoTasksListID),
		Items:       []dto4listus.CreateListItemRequest{{ID: "e1", ListItemBase: dbo4listus.ListItemBase{Title: "🍏 Apple"}}},
	})
	if err != nil {
		t.Fatalf("CreateListItems with emoji title failed: %v", err)
	}
	if len(respEmoji.CreatedItems) > 0 {
		if respEmoji.CreatedItems[0].Title != "Apple" || respEmoji.CreatedItems[0].Emoji != "🍏" {
			t.Errorf("expected Title 'Apple' and Emoji '🍏', got Title=%q Emoji=%q", respEmoji.CreatedItems[0].Title, respEmoji.CreatedItems[0].Emoji)
		}
	}

	// 7. generateRandomListItemID failure
	origRand := randomListItemID
	defer func() { randomListItemID = origRand }()
	randomListItemID = func(length int) string { return "dup" }
	_, _, err = CreateListItems(uctx, dto4listus.CreateListItemsRequest{
		ListRequest: listRequest(testSpaceID, dbo4listus.DoTasksListID),
		Items: []dto4listus.CreateListItemRequest{
			{ID: "dup", ListItemBase: dbo4listus.ListItemBase{Title: "First"}},
			{ID: "dup", ListItemBase: dbo4listus.ListItemBase{Title: "Second"}},
		},
	})
	if err == nil {
		t.Error("expected error when randomListItemID fails")
	}
	randomListItemID = origRand

	// 8. validateListInCreateListItems error
	origVal := validateListInCreateListItems
	defer func() { validateListInCreateListItems = origVal }()
	validateListInCreateListItems = func(list *dbo4listus.ListDbo) error {
		return errors.New("simulated invalid list")
	}
	_, _, err = CreateListItems(uctx, dto4listus.CreateListItemsRequest{
		ListRequest: listRequest(testSpaceID, dbo4listus.DoTasksListID),
		Items:       []dto4listus.CreateListItemRequest{{ID: "v1", ListItemBase: dbo4listus.ListItemBase{Title: "Task"}}},
	})
	if err == nil {
		t.Error("expected error from validateListInCreateListItems")
	}
	validateListInCreateListItems = origVal

	// 9. User has no access to list and space does not have user
	_ = dal4listus.RunListWorker(uctx, listRequest(testSpaceID, dbo4listus.DoTasksListID),
		func(ctx facade.ContextWithUser, tx dal.ReadwriteTransaction, params *dal4listus.ListWorkerParams) error {
			origGet := getListWorkerRecords
			defer func() { getListWorkerRecords = origGet }()
			getListWorkerRecords = func(_ *dal4listus.ListWorkerParams, _ facade.ContextWithUser, _ dal.ReadwriteTransaction) error {
				return nil
			}
			params.List.Record = record.NewRecordWithData(params.List.Key, params.List.Data)
			params.List.Data.SpaceIDs = []coretypes.SpaceID{testSpaceID}
			params.List.Data.UserIDs = []string{"other-user"}
			params.List.Data.Type = dbo4listus.ListTypeToDo
			params.Space.Data.UserIDs = []string{"someone-else"}
			req := dto4listus.CreateListItemsRequest{
				ListRequest: listRequest(testSpaceID, dbo4listus.DoTasksListID),
				Items:       []dto4listus.CreateListItemRequest{{ID: "s1", ListItemBase: dbo4listus.ListItemBase{Title: "T"}}},
			}
			_, _, err := createListItemsTxWorker(ctx, tx, req, params)
			if err == nil || !strings.Contains(err.Error(), "does not have access") {
				t.Errorf("expected access error, got: %v", err)
			}
			params.Space.Data.UserIDs = []string{testUserID}
			return nil
		},
	)

	// 10. List exists, user in space but not in list -> user gets added to list
	// Seed a space with both testUserID and user2
	ctxMulti, _ := newTestDBWithSpace(t, "spacemulti", testUserID, "user2")
	uctx1 := userCtx(ctxMulti, testUserID)
	uctx2 := userCtx(ctxMulti, "user2")
	// user1 creates item
	_, _, err = CreateListItems(uctx1, dto4listus.CreateListItemsRequest{
		ListRequest: listRequest("spacemulti", dbo4listus.DoTasksListID),
		Items:       []dto4listus.CreateListItemRequest{{ID: "init1", ListItemBase: dbo4listus.ListItemBase{Title: "T1"}}},
	})
	if err != nil {
		t.Fatalf("failed initial CreateListItems: %v", err)
	}
	// user2 creates item on the same list -> user2 gets added to list's UserIDs
	_, listMulti, err := CreateListItems(uctx2, dto4listus.CreateListItemsRequest{
		ListRequest: listRequest("spacemulti", dbo4listus.DoTasksListID),
		Items:       []dto4listus.CreateListItemRequest{{ID: "init2", ListItemBase: dbo4listus.ListItemBase{Title: "T2"}}},
	})
	if err != nil {
		t.Fatalf("expected member of space to be added to list, got error: %v", err)
	}
	if !listMulti.Data.HasUserID("user2") {
		t.Errorf("expected list to have user2")
	}

	// 11. txUpdateInCreateListItems error
	origUpdate := txUpdateInCreateListItems
	defer func() { txUpdateInCreateListItems = origUpdate }()
	txUpdateInCreateListItems = func(ctx context.Context, tx dal.ReadwriteTransaction, key *record.Key, updates []update.Update) error {
		return errors.New("simulated tx update failure")
	}
	_, _, err = CreateListItems(uctx, dto4listus.CreateListItemsRequest{
		ListRequest: listRequest(testSpaceID, dbo4listus.DoTasksListID),
		Items:       []dto4listus.CreateListItemRequest{{ID: "u1", ListItemBase: dbo4listus.ListItemBase{Title: "T"}}},
	})
	if err == nil {
		t.Error("expected error from txUpdateInCreateListItems")
	}
	txUpdateInCreateListItems = origUpdate

	// 12. txInsertInCreateListItems error for list record
	origInsert := txInsertInCreateListItems
	defer func() { txInsertInCreateListItems = origInsert }()
	txInsertInCreateListItems = func(ctx context.Context, tx dal.ReadwriteTransaction, r record.Record) error {
		return errors.New("simulated tx insert failure")
	}
	// For a fresh standard list not yet inserted:
	_, _, err = CreateListItems(uctx, dto4listus.CreateListItemsRequest{
		ListRequest: listRequest(testSpaceID, string(dbo4listus.ReadBooksListID)),
		Items:       []dto4listus.CreateListItemRequest{{ID: "f1", ListItemBase: dbo4listus.ListItemBase{Title: "T"}}},
	})
	if err == nil {
		t.Error("expected error from txInsertInCreateListItems for list record")
	}

	// 13. txInsertInCreateListItems error for space module record
	insertCount := 0
	txInsertInCreateListItems = func(ctx context.Context, tx dal.ReadwriteTransaction, r record.Record) error {
		insertCount++
		if insertCount == 1 {
			// list record insert succeeds
			return tx.Insert(ctx, r)
		}
		// module entry insert fails
		return errors.New("simulated module insert failure")
	}
	// Use fresh space ID so space module entry doesn't exist
	freshSpaceID := coretypes.SpaceID("spacefresh")
	ctx2, _ := newTestDBWithSpace(t, freshSpaceID, testUserID)
	uctxFresh := userCtx(ctx2, testUserID)
	_, _, err = CreateListItems(uctxFresh, dto4listus.CreateListItemsRequest{
		ListRequest: listRequest(freshSpaceID, dbo4listus.DoTasksListID),
		Items:       []dto4listus.CreateListItemRequest{{ID: "m1", ListItemBase: dbo4listus.ListItemBase{Title: "T"}}},
	})
	if err == nil {
		t.Error("expected error from txInsertInCreateListItems for space module entry")
	}
}
