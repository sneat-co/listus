package facade4listus

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
	"github.com/sneat-co/listus/backend/const4listus"
	"github.com/sneat-co/listus/backend/dal4listus"
	"github.com/sneat-co/listus/backend/dbo4listus"
	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/sneat-core-modules/spaceus/dal4spaceus"
	"github.com/sneat-co/sneat-core-modules/spaceus/dbo4spaceus"
	"github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/sneat-co/sneat-go-core/facade"
)

const testSpaceID coretypes.SpaceID = "space1"

func createItems(t *testing.T, ctx context.Context, listID string, titles ...string) dto4listus.CreateListItemResponse {
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
	resp, _, err := CreateListItems(userCtx(ctx, testUserID), dto4listus.CreateListItemsRequest{
		ListRequest: listRequest(testSpaceID, listID),
		Items:       items,
	})
	if err != nil {
		t.Fatalf("CreateListItems failed: %v", err)
	}
	return resp
}

func TestCreateList_Succeeds(t *testing.T) {
	ctx, db := newTestDBWithSpace(t, testSpaceID, testUserID)

	// CreateList builds a ListDbo from the request. It must populate UserIDs
	// (from the requesting user) as well as SpaceIDs, otherwise the formed DTO
	// fails its own Validate(). A valid request must succeed.
	response, err := CreateList(userCtx(ctx, testUserID), dto4listus.CreateListRequest{
		SpaceRequest: spaceRequest(testSpaceID),
		Type:         dbo4listus.ListTypeToDo,
		Title:        "Groceries",
	})
	if err != nil {
		t.Fatalf("CreateList failed: %v", err)
	}
	if response.ID == "" {
		t.Fatal("CreateList returned an empty list ID")
	}
	if !strings.HasPrefix(response.ID, string(dbo4listus.ListTypeToDo)+"!") {
		t.Fatalf("CreateList returned ID %q without the requested list type", response.ID)
	}
	module := dbo4spaceus.NewSpaceModuleEntry(
		testSpaceID,
		const4listus.ExtensionID,
		new(dbo4listus.ListusSpaceDbo),
	)
	if err := db.Get(ctx, module.Record); err != nil {
		t.Fatalf("failed to read Listus Space summary: %v", err)
	}
	brief := module.Data.Lists[response.ID]
	if brief == nil || brief.Title != "Groceries" {
		t.Fatalf("created list brief = %#v, want title Groceries", brief)
	}
}

func TestCreateList_InvalidRequest(t *testing.T) {
	ctx, _ := newTestDBWithSpace(t, testSpaceID, testUserID)
	_, err := CreateList(userCtx(ctx, testUserID), dto4listus.CreateListRequest{
		SpaceRequest: spaceRequest(testSpaceID),
		// missing Type
		Title: "X",
	})
	if err == nil {
		t.Error("expected validation error for missing type")
	}
}

func TestCreateList_Branches(t *testing.T) {
	ctx, _ := newTestDBWithSpace(t, testSpaceID, testUserID)
	uctx := userCtx(ctx, testUserID)

	// 1. Create first list
	resp1, err := CreateList(uctx, dto4listus.CreateListRequest{
		SpaceRequest: spaceRequest(testSpaceID),
		Type:         dbo4listus.ListTypeToDo,
		Title:        "First List",
	})
	if err != nil {
		t.Fatalf("first CreateList failed: %v", err)
	}

	// 2. Create second list (SpaceModuleEntry already exists, hits lines 111-113)
	resp2, err := CreateList(uctx, dto4listus.CreateListRequest{
		SpaceRequest: spaceRequest(testSpaceID),
		Type:         dbo4listus.ListTypeToDo,
		Title:        "Second List",
	})
	if err != nil {
		t.Fatalf("second CreateList failed: %v", err)
	}
	if resp1.ID == resp2.ID {
		t.Fatalf("expected different IDs, got %q", resp1.ID)
	}

	// 3. Duplicate title (hits line 38)
	_, err = CreateList(uctx, dto4listus.CreateListRequest{
		SpaceRequest: spaceRequest(testSpaceID),
		Type:         dbo4listus.ListTypeToDo,
		Title:        "First List",
	})
	if err == nil {
		t.Fatal("expected error for duplicate title")
	}

	// 4. Invalid UserID in context (hits line 93: listDbo.Validate fails)
	noUserCtx := fakeNoUserCtx{ContextWithUser: uctx}
	_, err = CreateList(noUserCtx, dto4listus.CreateListRequest{
		SpaceRequest: spaceRequest(testSpaceID),
		Type:         dbo4listus.ListTypeToDo,
		Title:        "No User List",
	})
	if err == nil {
		t.Fatal("expected error for empty user ID")
	}

	// 5. getSpaceModuleRecords failure (hits line 36)
	origGetSpace := getSpaceModuleRecords
	defer func() { getSpaceModuleRecords = origGetSpace }()
	getSpaceModuleRecords = func(params *dal4spaceus.ModuleSpaceWorkerParams[*dbo4listus.ListusSpaceDbo], ctx facade.ContextWithUser, tx dal.ReadwriteTransaction) error {
		return errors.New("get space records err")
	}
	_, err = CreateList(uctx, dto4listus.CreateListRequest{
		SpaceRequest: spaceRequest(testSpaceID),
		Type:         dbo4listus.ListTypeToDo,
		Title:        "Get Space Fail",
	})
	if err == nil {
		t.Fatal("expected error on getSpaceModuleRecords failure")
	}
	getSpaceModuleRecords = origGetSpace

	// 6. Random ID collision + success (hits line 51: idGenerationAttempt++)
	firstSubID := dbo4listus.ListKey(resp1.ID).ListSubID()
	origRandom := randomListSubID
	defer func() { randomListSubID = origRandom }()
	call := 0
	randomListSubID = func(length int) string {
		call++
		if call == 1 {
			return firstSubID // duplicate, in params.SpaceModuleEntry.Data.Lists!
		}
		return "sub3"
	}
	resp3, err := CreateList(uctx, dto4listus.CreateListRequest{
		SpaceRequest: spaceRequest(testSpaceID),
		Type:         dbo4listus.ListTypeToDo,
		Title:        "Third List",
	})
	if err != nil {
		t.Fatalf("third CreateList failed: %v", err)
	}
	if resp3.ID == "" {
		t.Fatal("expected non-empty ID")
	}

	// 7. Random ID exhaustion (hits lines 53-54)
	randomListSubID = func(length int) string {
		return firstSubID
	}
	_, err = CreateList(uctx, dto4listus.CreateListRequest{
		SpaceRequest: spaceRequest(testSpaceID),
		Type:         dbo4listus.ListTypeToDo,
		Title:        "Exhaust List",
	})
	if err == nil {
		t.Fatal("expected error for ID generation exhaustion")
	}
	randomListSubID = origRandom

	// 8. Insert failure on listRecord (hits line 98)
	origInsertList := insertListRecord
	defer func() { insertListRecord = origInsertList }()
	insertListRecord = func(ctx context.Context, tx dal.ReadwriteTransaction, r record.Record) error {
		return errors.New("insert list err")
	}
	_, err = CreateList(uctx, dto4listus.CreateListRequest{
		SpaceRequest: spaceRequest(testSpaceID),
		Type:         dbo4listus.ListTypeToDo,
		Title:        "Insert List Fail",
	})
	if err == nil {
		t.Fatal("expected error on insertListRecord failure")
	}
	insertListRecord = origInsertList

	// 9. Insert failure on spaceModuleEntry (hits line 118)
	ctxFresh, _ := newTestDBWithSpace(t, "space2", testUserID)
	uctxFresh := userCtx(ctxFresh, testUserID)
	origInsertSpace := insertSpaceModuleEntry
	defer func() { insertSpaceModuleEntry = origInsertSpace }()
	insertSpaceModuleEntry = func(ctx context.Context, tx dal.ReadwriteTransaction, r record.Record) error {
		return errors.New("insert space module err")
	}
	_, err = CreateList(uctxFresh, dto4listus.CreateListRequest{
		SpaceRequest: spaceRequest("space2"),
		Type:         dbo4listus.ListTypeToDo,
		Title:        "Insert Space Fail",
	})
	if err == nil {
		t.Fatal("expected error on insertSpaceModuleEntry failure")
	}
}

func TestCreateListItems_StandardListCreatesAndDeductsEmoji(t *testing.T) {
	ctx, _ := newTestDBWithSpace(t, testSpaceID, testUserID)

	resp := createItems(t, ctx, dbo4listus.BuyGroceriesListID, "Milk", "Banana")
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
	ctx, _ := newTestDBWithSpace(t, testSpaceID, testUserID)
	createItems(t, ctx, dbo4listus.DoTasksListID, "Task A")
	createItems(t, ctx, dbo4listus.DoTasksListID, "Task A")

	// Read back the list and assert a single item.
	list := getListData(t, ctx, dbo4listus.DoTasksListID)
	if len(list.Items) != 1 {
		t.Errorf("expected 1 deduped item, got %d", len(list.Items))
	}
}

func TestCreateListItems_NonStandardListNotFound(t *testing.T) {
	ctx, _ := newTestDBWithSpace(t, testSpaceID, testUserID)
	_, _, err := CreateListItems(userCtx(ctx, testUserID), dto4listus.CreateListItemsRequest{
		ListRequest: listRequest(testSpaceID, "do!custom"),
		Items:       []dto4listus.CreateListItemRequest{{ListItemBase: dbo4listus.ListItemBase{Title: "X"}}},
	})
	if err == nil {
		t.Error("expected error creating items in non-existent non-standard list")
	}
}

func TestCreateListItems_AppendsToExistingList(t *testing.T) {
	ctx, _ := newTestDBWithSpace(t, testSpaceID, testUserID)
	// First create establishes (inserts) the standard list.
	createItems(t, ctx, dbo4listus.DoTasksListID, "First")
	// Second create must take the "list already exists" update branch.
	createItems(t, ctx, dbo4listus.DoTasksListID, "Second")

	list := getListData(t, ctx, dbo4listus.DoTasksListID)
	if len(list.Items) != 2 {
		t.Errorf("expected 2 items after append, got %d", len(list.Items))
	}
	if list.Count != 2 {
		t.Errorf("Count = %d, want 2", list.Count)
	}
}

func TestSetListItemsIsDone(t *testing.T) {
	ctx, _ := newTestDBWithSpace(t, testSpaceID, testUserID)
	resp := createItems(t, ctx, dbo4listus.DoTasksListID, "Task A")
	id := resp.CreatedItems[0].ID

	changed, _, err := SetListItemsIsDone(userCtx(ctx, testUserID), dto4listus.ListItemsSetIsDoneRequest{
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
	changed2, _, err := SetListItemsIsDone(userCtx(ctx, testUserID), dto4listus.ListItemsSetIsDoneRequest{
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
	changed3, _, err := SetListItemsIsDone(userCtx(ctx, testUserID), dto4listus.ListItemsSetIsDoneRequest{
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
	ctx, _ := newTestDBWithSpace(t, testSpaceID, testUserID)
	resp := createItems(t, ctx, dbo4listus.DoTasksListID, "A", "B", "C")
	ids := []string{resp.CreatedItems[0].ID, resp.CreatedItems[1].ID}

	deleted, _, err := DeleteListItems(userCtx(ctx, testUserID), dto4listus.ListItemIDsRequest{
		ListRequest: listRequest(testSpaceID, dbo4listus.DoTasksListID),
		ItemIDs:     ids,
	})
	if err != nil {
		t.Fatalf("DeleteListItems failed: %v", err)
	}
	if len(deleted) != 2 {
		t.Fatalf("deleted %d, want 2", len(deleted))
	}
	list := getListData(t, ctx, dbo4listus.DoTasksListID)
	if len(list.Items) != 1 {
		t.Errorf("expected 1 remaining item, got %d", len(list.Items))
	}
	if len(list.RecentItems) != 2 {
		t.Errorf("expected 2 recent items, got %d", len(list.RecentItems))
	}
}

func TestDeleteListItems_Wildcard(t *testing.T) {
	ctx, _ := newTestDBWithSpace(t, testSpaceID, testUserID)
	createItems(t, ctx, dbo4listus.DoTasksListID, "A", "B")

	deleted, _, err := DeleteListItems(userCtx(ctx, testUserID), dto4listus.ListItemIDsRequest{
		ListRequest: listRequest(testSpaceID, dbo4listus.DoTasksListID),
		ItemIDs:     []string{"*"},
	})
	if err != nil {
		t.Fatalf("DeleteListItems wildcard failed: %v", err)
	}
	if len(deleted) != 2 {
		t.Errorf("expected 2 deleted, got %d", len(deleted))
	}
	list := getListData(t, ctx, dbo4listus.DoTasksListID)
	if len(list.Items) != 0 {
		t.Errorf("expected empty list, got %d items", len(list.Items))
	}
}

func TestReorderListItem(t *testing.T) {
	ctx, _ := newTestDBWithSpace(t, testSpaceID, testUserID)
	resp := createItems(t, ctx, dbo4listus.DoTasksListID, "A", "B", "C")
	cID := resp.CreatedItems[2].ID

	// Move "C" to index 0.
	err := ReorderListItem(userCtx(ctx, testUserID), dto4listus.ReorderListItemsRequest{
		ListItemIDsRequest: dto4listus.ListItemIDsRequest{
			ListRequest: listRequest(testSpaceID, dbo4listus.DoTasksListID),
			ItemIDs:     []string{cID},
		},
		ToIndex: 0,
	})
	if err != nil {
		t.Fatalf("ReorderListItem failed: %v", err)
	}
	list := getListData(t, ctx, dbo4listus.DoTasksListID)
	if len(list.Items) != 3 || list.Items[0].Title != "C" {
		titles := make([]string, len(list.Items))
		for i, it := range list.Items {
			titles[i] = it.Title
		}
		t.Errorf("expected C first, got order %v", titles)
	}
}

func TestReorderListItem_InvalidRequest(t *testing.T) {
	ctx, _ := newTestDBWithSpace(t, testSpaceID, testUserID)
	err := ReorderListItem(userCtx(ctx, testUserID), dto4listus.ReorderListItemsRequest{
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
	ctx, _ := newTestDBWithSpace(t, testSpaceID, testUserID)
	uctx := userCtx(ctx, testUserID)

	// Create list first so DeleteSpaceItem finds it and invokes worker and briefsAdapter
	createResp, err := CreateList(uctx, dto4listus.CreateListRequest{
		SpaceRequest: spaceRequest(testSpaceID),
		Type:         dbo4listus.ListTypeToDo,
		Title:        "Task List",
	})
	if err != nil {
		t.Fatalf("CreateList failed: %v", err)
	}

	// deleteListTxWorker always returns "not implemented", so DeleteList must surface an error.
	err = DeleteList(uctx, listRequest(testSpaceID, createResp.ID))
	if err == nil {
		t.Error("expected error from DeleteList (worker not implemented)")
	}

	// Success with mocked worker
	orig := deleteListWorker
	defer func() { deleteListWorker = orig }()
	deleteListWorker = func(_ facade.ContextWithUser, _ dal.ReadwriteTransaction, _ *dal4spaceus.SpaceItemWorkerParams[*dbo4listus.ListusSpaceDbo, *dbo4listus.ListDbo]) error {
		return nil
	}
	if err := DeleteList(uctx, listRequest(testSpaceID, createResp.ID)); err != nil {
		t.Fatalf("DeleteList failed with mocked worker: %v", err)
	}
	if count := listBriefsCount(&dbo4listus.ListusSpaceDbo{Lists: dbo4listus.ListBriefs{"a": {}}}); count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	// Empty user ID
	noUserCtx := fakeNoUserCtx{ContextWithUser: uctx}
	if err := DeleteList(noUserCtx, listRequest(testSpaceID, dbo4listus.DoTasksListID)); err == nil {
		t.Error("expected error for empty user ID")
	}
}

type fakeNoUserCtx struct {
	facade.ContextWithUser
}

func (f fakeNoUserCtx) User() facade.UserContext {
	return fakeNoUser{}
}

type fakeNoUser struct {
	facade.UserContext
}

func (fakeNoUser) GetUserID() string { return "" }

func TestDeleteList_InvalidRequest(t *testing.T) {
	ctx, _ := newTestDBWithSpace(t, testSpaceID, testUserID)
	err := DeleteList(userCtx(ctx, testUserID), dto4listus.ListRequest{
		SpaceRequest: spaceRequest(testSpaceID),
		ListID:       "invalid",
	})
	if err == nil {
		t.Error("expected validation error for invalid list id")
	}
}

// getListData reads a list record directly from the seeded DB.
func getListData(t *testing.T, baseCtx context.Context, listID string) *dbo4listus.ListDbo {
	t.Helper()
	ctx := userCtx(baseCtx, testUserID)
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
	orig := randomListItemID
	defer func() { randomListItemID = orig }()

	items := []*dbo4listus.ListItemBrief{{ID: "a"}, {ID: "b"}}

	// Initial ID not duplicate -> returned as-is.
	id, err := generateRandomListItemID(items, "c")
	if err != nil || id != "c" {
		t.Errorf("got id=%q err=%v, want c/nil", id, err)
	}

	// Initial ID duplicate, first random attempt is duplicate, second is unique
	calls := 0
	randomListItemID = func(length int) string {
		calls++
		if calls == 1 {
			return "b" // duplicate
		}
		return "unique"
	}
	id, err = generateRandomListItemID(items, "a")
	if err != nil || id != "unique" {
		t.Fatalf("unexpected id: %q, err: %v", id, err)
	}

	// Loop exhausts 101 attempts with duplicate -> returns error
	randomListItemID = func(length int) string {
		return "a"
	}
	_, err = generateRandomListItemID(items, "a")
	if err == nil {
		t.Fatal("expected error when all attempts fail")
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
	ClearList(userCtx(context.Background(), testUserID), nil, "do!tasks")
}
