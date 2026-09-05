package facade4listus

import (
	"github.com/dal-go/record"
	"github.com/sneat-co/listus/backend/dbo4listus"
	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/sneat-core-modules/spaceus/dbo4spaceus"
	"sync"
	"testing"
)

func TestCreateListPersistsRequestedTitleInSpaceModuleBrief(t *testing.T) {
	ctx, db := newTestDBWithSpace(t, testSpaceID, testUserID)
	created, err := CreateList(userCtx(ctx, testUserID), dto4listus.CreateListRequest{
		SpaceRequest: spaceRequest(testSpaceID),
		Type:         dbo4listus.ListTypeToDo,
		Title:        "Saturday cleaning",
	})
	if err != nil {
		t.Fatal(err)
	}
	module := new(dbo4listus.ListusSpaceDbo)
	moduleRecord := record.NewRecordWithData(
		dbo4spaceus.NewSpaceModuleKey(testSpaceID, "listus"), module,
	)
	if err = db.Get(ctx, moduleRecord); err != nil {
		t.Fatal(err)
	}
	brief := module.Lists[created.ID]
	if brief == nil || brief.Title != "Saturday cleaning" || brief.Type != dbo4listus.ListTypeToDo {
		t.Fatalf("persisted list brief = %+v", brief)
	}
}

func TestApplyListTemplateAddsRestoresPreservesAndRetriesWithoutMutation(t *testing.T) {
	ctx, _ := newTestDBWithSpace(t, testSpaceID, testUserID)
	template, err := CreateList(userCtx(ctx, testUserID), dto4listus.CreateListRequest{SpaceRequest: spaceRequest(testSpaceID), Type: dbo4listus.ListTypeToBuy, Title: "Regular groceries"})
	if err != nil {
		t.Fatal(err)
	}
	createItems(t, ctx, template.ID, "Milk", "Bread")
	destination := dbo4listus.BuyGroceriesListID
	items := createItems(t, ctx, destination, "Milk", "Eggs")
	_, _, err = SetListItemsIsDone(userCtx(ctx, testUserID), dto4listus.ListItemsSetIsDoneRequest{ListItemIDsRequest: dto4listus.ListItemIDsRequest{ListRequest: listRequest(testSpaceID, destination), ItemIDs: []string{items.CreatedItems[0].ID}}, IsDone: true})
	if err != nil {
		t.Fatal(err)
	}
	request := dto4listus.ApplyListTemplateRequest{SpaceID: testSpaceID, SourceListID: dbo4listus.ListKey(template.ID), DestinationListID: dbo4listus.ListKey(destination), RequestID: "click-1"}
	result, err := ApplyListTemplate(userCtx(ctx, testUserID), request)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Added) != 1 || len(result.Restored) != 1 {
		t.Fatalf("result=%+v", result)
	}
	list := getListData(t, ctx, destination)
	if len(list.Items) != 3 {
		t.Fatalf("destination items=%+v", list.Items)
	}
	_, _, err = SetListItemsIsDone(userCtx(ctx, testUserID), dto4listus.ListItemsSetIsDoneRequest{ListItemIDsRequest: dto4listus.ListItemIDsRequest{ListRequest: listRequest(testSpaceID, destination), ItemIDs: []string{result.Restored[0].ID}}, IsDone: true})
	if err != nil {
		t.Fatal(err)
	}
	replay, err := ApplyListTemplate(userCtx(ctx, testUserID), request)
	if err != nil || replay.Disposition != "reused" {
		t.Fatalf("replay=(%+v,%v)", replay, err)
	}
	list = getListData(t, ctx, destination)
	if !list.Items[0].IsDone() {
		t.Fatal("retry reopened item completed after first apply")
	}
	request.RequestID = "click-2"
	fresh, err := ApplyListTemplate(userCtx(ctx, testUserID), request)
	if err != nil || len(fresh.Restored) != 1 || fresh.Disposition != "applied" {
		t.Fatalf("fresh click=(%+v,%v)", fresh, err)
	}
	if getListData(t, ctx, destination).Items[0].IsDone() {
		t.Fatal("fresh click did not restore the completed regular item")
	}
}

func TestTemplateReceiptIDUsesUnambiguousTupleEncoding(t *testing.T) {
	if templateReceiptID("a", "b\x00c") == templateReceiptID("a\x00b", "c") {
		t.Fatal("distinct user/request tuples produced the same receipt ID")
	}
}

func TestApplyListTemplateConcurrentSameRequestHasOneClassification(t *testing.T) {
	ctx, _ := newTestDBWithSpace(t, testSpaceID, testUserID)
	template, err := CreateList(userCtx(ctx, testUserID), dto4listus.CreateListRequest{SpaceRequest: spaceRequest(testSpaceID), Type: dbo4listus.ListTypeToBuy, Title: "Regular groceries"})
	if err != nil {
		t.Fatal(err)
	}
	createItems(t, ctx, template.ID, "Milk", "Bread")
	destination, err := CreateList(userCtx(ctx, testUserID), dto4listus.CreateListRequest{SpaceRequest: spaceRequest(testSpaceID), Type: dbo4listus.ListTypeToBuy, Title: "To buy"})
	if err != nil {
		t.Fatal(err)
	}
	request := dto4listus.ApplyListTemplateRequest{SpaceID: testSpaceID, SourceListID: dbo4listus.ListKey(template.ID), DestinationListID: dbo4listus.ListKey(destination.ID), RequestID: "concurrent-click"}
	results := make([]dto4listus.ApplyListTemplateResult, 2)
	errs := make([]error, 2)
	var ready sync.WaitGroup
	ready.Add(2)
	start := make(chan struct{})
	var done sync.WaitGroup
	done.Add(2)
	for i := range results {
		go func(i int) {
			defer done.Done()
			ready.Done()
			<-start
			results[i], errs[i] = ApplyListTemplate(userCtx(ctx, testUserID), request)
		}(i)
	}
	ready.Wait()
	close(start)
	done.Wait()
	for i := range results {
		if errs[i] != nil {
			t.Fatalf("apply %d: %v", i, errs[i])
		}
		if len(results[i].Added) != 2 {
			t.Fatalf("apply %d classified %d added items, want receipt snapshot of 2: %+v", i, len(results[i].Added), results[i])
		}
	}
	if got := len(getListData(t, ctx, destination.ID).Items); got != 2 {
		t.Fatalf("destination has %d items after concurrent retry, want 2", got)
	}
}

func TestApplyListTemplateSupportsDoListsAndRefusesNonMember(t *testing.T) {
	ctx, _ := newTestDBWithSpace(t, testSpaceID, testUserID)
	template, err := CreateList(userCtx(ctx, testUserID), dto4listus.CreateListRequest{SpaceRequest: spaceRequest(testSpaceID), Type: dbo4listus.ListTypeToDo, Title: "Saturday cleaning"})
	if err != nil {
		t.Fatal(err)
	}
	createItems(t, ctx, template.ID, "Kitchen sink", "Toilet")
	createItems(t, ctx, dbo4listus.DoTasksListID, "Vacuum")
	request := dto4listus.ApplyListTemplateRequest{SpaceID: testSpaceID, SourceListID: dbo4listus.ListKey(template.ID), DestinationListID: dbo4listus.DoTasksListID, RequestID: "do-click"}
	if _, err = ApplyListTemplate(userCtx(ctx, "outsider"), request); err == nil {
		t.Fatal("non-member apply succeeded")
	}
	if result, applyErr := ApplyListTemplate(userCtx(ctx, testUserID), request); applyErr != nil || len(result.Added) != 2 {
		t.Fatalf("do apply=(%+v,%v)", result, applyErr)
	}
}

func TestApplyListTemplateCreatesEmptyStandardDestinationOnFirstUse(t *testing.T) {
	ctx, _ := newTestDBWithSpace(t, testSpaceID, testUserID)
	template, err := CreateList(userCtx(ctx, testUserID), dto4listus.CreateListRequest{SpaceRequest: spaceRequest(testSpaceID), Type: dbo4listus.ListTypeToDo, Title: "Saturday cleaning"})
	if err != nil {
		t.Fatal(err)
	}
	createItems(t, ctx, template.ID, "Kitchen sink", "Toilet")
	request := dto4listus.ApplyListTemplateRequest{SpaceID: testSpaceID, SourceListID: dbo4listus.ListKey(template.ID), DestinationListID: dbo4listus.DoTasksListID, RequestID: "first-standard-click"}
	result, err := ApplyListTemplate(userCtx(ctx, testUserID), request)
	if err != nil || len(result.Added) != 2 {
		t.Fatalf("first apply=(%+v,%v)", result, err)
	}
	destination := getListData(t, ctx, dbo4listus.DoTasksListID)
	if destination.Type != dbo4listus.ListTypeToDo || len(destination.Items) != 2 {
		t.Fatalf("standard destination=%+v", destination)
	}
}
