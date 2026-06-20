package facade4listus

import (
	"testing"

	"github.com/sneat-co/listus/backend/dal4listus"
	"github.com/sneat-co/listus/backend/dbo4listus"
	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/sneat-co/sneat-go-core/facade"
)

const testSpaceID coretypes.SpaceID = "space1"

func createItems(t *testing.T, listID string, titles ...string) dto4listus.CreateListItemResponse {
	t.Helper()
	items := make([]dto4listus.CreateListItemRequest, len(titles))
	for i, title := range titles {
		// Supply explicit IDs: the facade only generates a random ID when the
		// initial ID collides with an existing item, so an empty ID would be
		// kept verbatim and fail list validation.
		items[i] = dto4listus.CreateListItemRequest{
			ID:           "id-" + string(rune('a'+i)),
			ListItemBase: dbo4listus.ListItemBase{Title: title},
		}
	}
	resp, _, err := CreateListItems(userCtx(testUserID), dto4listus.CreateListItemsRequest{
		ListRequest: listRequest(testSpaceID, listID),
		Items:       items,
	})
	if err != nil {
		t.Fatalf("CreateListItems failed: %v", err)
	}
	return resp
}

func TestCreateList_FormedListDboMissingUserIDs(t *testing.T) {
	_ = newTestDBWithSpace(t, testSpaceID, testUserID)

	// NOTE: CreateList builds a ListDbo that sets SpaceIDs but not UserIDs, so
	// the formed DTO fails its own Validate(). This asserts the current
	// behavior (the validation gate is reached and rejects the record).
	_, err := CreateList(userCtx(testUserID), dto4listus.CreateListRequest{
		SpaceRequest: spaceRequest(testSpaceID),
		Type:         dbo4listus.ListTypeToDo,
		Title:        "Groceries",
	})
	if err == nil {
		t.Fatal("expected CreateList to reject formed ListDbo missing userIDs")
	}
}

func TestCreateList_InvalidRequest(t *testing.T) {
	_ = newTestDBWithSpace(t, testSpaceID, testUserID)
	_, err := CreateList(userCtx(testUserID), dto4listus.CreateListRequest{
		SpaceRequest: spaceRequest(testSpaceID),
		// missing Type
		Title: "X",
	})
	if err == nil {
		t.Error("expected validation error for missing type")
	}
}

func TestCreateListItems_StandardListCreatesAndDeductsEmoji(t *testing.T) {
	_ = newTestDBWithSpace(t, testSpaceID, testUserID)

	resp := createItems(t, dbo4listus.BuyGroceriesListID, "Milk", "Banana")
	if len(resp.CreatedItems) != 2 {
		t.Fatalf("created %d items, want 2", len(resp.CreatedItems))
	}
	var banana *dbo4listus.ListItemBrief
	for _, it := range resp.CreatedItems {
		if it.Title == "Banana" {
			banana = it
		}
		if it.ID == "" {
			t.Errorf("item %q got empty ID", it.Title)
		}
	}
	if banana == nil || banana.Emoji != "🍌" {
		t.Errorf("expected deducted banana emoji, got %+v", banana)
	}
}

func TestCreateListItems_DedupSameTitle(t *testing.T) {
	_ = newTestDBWithSpace(t, testSpaceID, testUserID)
	createItems(t, dbo4listus.DoTasksListID, "Task A")
	createItems(t, dbo4listus.DoTasksListID, "Task A")

	// Read back the list and assert a single item.
	list := getListData(t, dbo4listus.DoTasksListID)
	if len(list.Items) != 1 {
		t.Errorf("expected 1 deduped item, got %d", len(list.Items))
	}
}

func TestCreateListItems_NonStandardListNotFound(t *testing.T) {
	_ = newTestDBWithSpace(t, testSpaceID, testUserID)
	_, _, err := CreateListItems(userCtx(testUserID), dto4listus.CreateListItemsRequest{
		ListRequest: listRequest(testSpaceID, "do!custom"),
		Items:       []dto4listus.CreateListItemRequest{{ListItemBase: dbo4listus.ListItemBase{Title: "X"}}},
	})
	if err == nil {
		t.Error("expected error creating items in non-existent non-standard list")
	}
}

func TestCreateListItems_AppendsToExistingList(t *testing.T) {
	_ = newTestDBWithSpace(t, testSpaceID, testUserID)
	// First create establishes (inserts) the standard list.
	createItems(t, dbo4listus.DoTasksListID, "First")
	// Second create must take the "list already exists" update branch.
	createItems(t, dbo4listus.DoTasksListID, "Second")

	list := getListData(t, dbo4listus.DoTasksListID)
	if len(list.Items) != 2 {
		t.Errorf("expected 2 items after append, got %d", len(list.Items))
	}
	if list.Count != 2 {
		t.Errorf("Count = %d, want 2", list.Count)
	}
}

func TestSetListItemsIsDone(t *testing.T) {
	_ = newTestDBWithSpace(t, testSpaceID, testUserID)
	resp := createItems(t, dbo4listus.DoTasksListID, "Task A")
	id := resp.CreatedItems[0].ID

	changed, _, err := SetListItemsIsDone(userCtx(testUserID), dto4listus.ListItemsSetIsDoneRequest{
		ListItemIDsRequest: dto4listus.ListItemIDsRequest{ListRequest: listRequest(testSpaceID, dbo4listus.DoTasksListID), ItemIDs: []string{id}},
		IsDone:             true,
	})
	if err != nil {
		t.Fatalf("SetListItemsIsDone failed: %v", err)
	}
	if len(changed) != 1 {
		t.Fatalf("changed %d items, want 1", len(changed))
	}
	if !changed[0].IsDone() {
		t.Error("item should be done")
	}

	// Marking the same again as done changes nothing.
	changed2, _, err := SetListItemsIsDone(userCtx(testUserID), dto4listus.ListItemsSetIsDoneRequest{
		ListItemIDsRequest: dto4listus.ListItemIDsRequest{ListRequest: listRequest(testSpaceID, dbo4listus.DoTasksListID), ItemIDs: []string{id}},
		IsDone:             true,
	})
	if err != nil {
		t.Fatalf("SetListItemsIsDone (2) failed: %v", err)
	}
	if len(changed2) != 0 {
		t.Errorf("expected no change, got %d", len(changed2))
	}

	// Un-done.
	changed3, _, err := SetListItemsIsDone(userCtx(testUserID), dto4listus.ListItemsSetIsDoneRequest{
		ListItemIDsRequest: dto4listus.ListItemIDsRequest{ListRequest: listRequest(testSpaceID, dbo4listus.DoTasksListID), ItemIDs: []string{id}},
		IsDone:             false,
	})
	if err != nil {
		t.Fatalf("SetListItemsIsDone (3) failed: %v", err)
	}
	if len(changed3) != 1 || changed3[0].IsDone() {
		t.Errorf("expected item to be reactivated, got %+v", changed3)
	}
}

func TestDeleteListItems(t *testing.T) {
	_ = newTestDBWithSpace(t, testSpaceID, testUserID)
	resp := createItems(t, dbo4listus.DoTasksListID, "A", "B", "C")
	ids := []string{resp.CreatedItems[0].ID, resp.CreatedItems[1].ID}

	deleted, _, err := DeleteListItems(userCtx(testUserID), dto4listus.ListItemIDsRequest{
		ListRequest: listRequest(testSpaceID, dbo4listus.DoTasksListID),
		ItemIDs:     ids,
	})
	if err != nil {
		t.Fatalf("DeleteListItems failed: %v", err)
	}
	if len(deleted) != 2 {
		t.Fatalf("deleted %d, want 2", len(deleted))
	}
	list := getListData(t, dbo4listus.DoTasksListID)
	if len(list.Items) != 1 {
		t.Errorf("expected 1 remaining item, got %d", len(list.Items))
	}
	if len(list.RecentItems) != 2 {
		t.Errorf("expected 2 recent items, got %d", len(list.RecentItems))
	}
}

func TestDeleteListItems_Wildcard(t *testing.T) {
	_ = newTestDBWithSpace(t, testSpaceID, testUserID)
	createItems(t, dbo4listus.DoTasksListID, "A", "B")

	deleted, _, err := DeleteListItems(userCtx(testUserID), dto4listus.ListItemIDsRequest{
		ListRequest: listRequest(testSpaceID, dbo4listus.DoTasksListID),
		ItemIDs:     []string{"*"},
	})
	if err != nil {
		t.Fatalf("DeleteListItems wildcard failed: %v", err)
	}
	if len(deleted) != 2 {
		t.Errorf("expected 2 deleted, got %d", len(deleted))
	}
	list := getListData(t, dbo4listus.DoTasksListID)
	if len(list.Items) != 0 {
		t.Errorf("expected empty list, got %d items", len(list.Items))
	}
}

func TestReorderListItem(t *testing.T) {
	_ = newTestDBWithSpace(t, testSpaceID, testUserID)
	resp := createItems(t, dbo4listus.DoTasksListID, "A", "B", "C")
	cID := resp.CreatedItems[2].ID

	// Move "C" to index 0.
	err := ReorderListItem(userCtx(testUserID), dto4listus.ReorderListItemsRequest{
		ListItemIDsRequest: dto4listus.ListItemIDsRequest{
			ListRequest: listRequest(testSpaceID, dbo4listus.DoTasksListID),
			ItemIDs:     []string{cID},
		},
		ToIndex: 0,
	})
	if err != nil {
		t.Fatalf("ReorderListItem failed: %v", err)
	}
	list := getListData(t, dbo4listus.DoTasksListID)
	if len(list.Items) != 3 || list.Items[0].Title != "C" {
		titles := make([]string, len(list.Items))
		for i, it := range list.Items {
			titles[i] = it.Title
		}
		t.Errorf("expected C first, got order %v", titles)
	}
}

func TestReorderListItem_InvalidRequest(t *testing.T) {
	_ = newTestDBWithSpace(t, testSpaceID, testUserID)
	err := ReorderListItem(userCtx(testUserID), dto4listus.ReorderListItemsRequest{
		ListItemIDsRequest: dto4listus.ListItemIDsRequest{
			ListRequest: listRequest(testSpaceID, dbo4listus.DoTasksListID),
			ItemIDs:     []string{"a"},
		},
		ToIndex: -1,
	})
	if err == nil {
		t.Error("expected validation error for negative toIndex")
	}
}

func TestDeleteList_NotImplementedWorker(t *testing.T) {
	_ = newTestDBWithSpace(t, testSpaceID, testUserID)
	// deleteListTxWorker always returns "not implemented", so DeleteList must surface an error.
	err := DeleteList(userCtx(testUserID), listRequest(testSpaceID, dbo4listus.DoTasksListID))
	if err == nil {
		t.Error("expected error from DeleteList (worker not implemented)")
	}
}

func TestDeleteList_InvalidRequest(t *testing.T) {
	_ = newTestDBWithSpace(t, testSpaceID, testUserID)
	err := DeleteList(userCtx(testUserID), dto4listus.ListRequest{
		SpaceRequest: spaceRequest(testSpaceID),
		ListID:       "invalid",
	})
	if err == nil {
		t.Error("expected validation error for invalid list id")
	}
}

// getListData reads a list record directly from the seeded DB.
func getListData(t *testing.T, listID string) *dbo4listus.ListDbo {
	t.Helper()
	ctx := userCtx(testUserID)
	db, err := facade.GetSneatDB(ctx)
	if err != nil {
		t.Fatalf("get db: %v", err)
	}
	entry := dal4listus.NewListEntry(testSpaceID, dbo4listus.ListKey(listID))
	if err := db.Get(ctx, entry.Record); err != nil {
		t.Fatalf("get list record: %v", err)
	}
	return entry.Data
}

func TestGenerateRandomListItemID(t *testing.T) {
	items := []*dbo4listus.ListItemBrief{{ID: "a"}, {ID: "b"}}

	// Initial ID not duplicate -> returned as-is.
	id, err := generateRandomListItemID(items, "c")
	if err != nil || id != "c" {
		t.Errorf("got id=%q err=%v, want c/nil", id, err)
	}

	// Initial ID duplicate -> a fresh non-duplicate ID generated.
	id, err = generateRandomListItemID(items, "a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == "a" || id == "b" {
		t.Errorf("generated duplicate id %q", id)
	}
}

func TestDeductListItemEmoji(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{"Apple", "🍏"},
		{"банан", "🍌"},
		{"unknownthing", ""},
		{"Fresh Milk", "🥛"},
	}
	for _, tt := range tests {
		if got := deductListItemEmoji(tt.text); got != tt.want {
			t.Errorf("deductListItemEmoji(%q) = %q, want %q", tt.text, got, tt.want)
		}
	}
}

func TestClearListNoop(t *testing.T) {
	// ClearList is currently a no-op; ensure it can be called without panic.
	ClearList(userCtx(testUserID), nil, "do!tasks")
}
