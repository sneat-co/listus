package facade4listus

import (
	"context"
	"testing"

	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/listus/backend/dbo4listus"
	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/sneat-ext-contracts/calendarius/calendariusmodels"
	calendarfacade "github.com/sneat-co/sneat-ext-contracts/calendarius/facade4calendarius"
)

type fakeDateTaskProvider struct {
	request calendariusmodels.MutateSourceLinkedDateTaskRequest
	applied bool
}

func (v *fakeDateTaskProvider) PlanSourceLinkedDateTask(_ context.Context, _ dal.ReadwriteTransaction, _ string, request calendariusmodels.MutateSourceLinkedDateTaskRequest) (calendarfacade.PreparedSourceLinkedDateTaskMutation, error) {
	v.request = request
	task := calendariusmodels.SourceLinkedDateTask{HappeningID: "due-list-item", Revision: request.ExpectedRevision + 1, Source: request.Source, Title: request.Title, DueDate: request.DueDate, State: request.State, ActionID: request.ActionID, ActionDisposition: request.ActionDisposition}
	return &fakePreparedDateTask{provider: v, mutation: calendariusmodels.SourceLinkedDateTaskMutation{Task: task, Disposition: "changed"}}, nil
}

type fakePreparedDateTask struct {
	provider *fakeDateTaskProvider
	mutation calendariusmodels.SourceLinkedDateTaskMutation
}

func (v *fakePreparedDateTask) Result() calendariusmodels.SourceLinkedDateTaskMutation {
	return v.mutation
}
func (v *fakePreparedDateTask) Apply(context.Context, dal.ReadwriteTransaction) error {
	v.provider.applied = true
	return nil
}

func TestSaveListItemDateTaskPersistsCanonicalCalendarLink(t *testing.T) {
	ctx, _ := newTestDBWithSpace(t, testSpaceID, testUserID)
	created := createItems(t, ctx, dbo4listus.DoTasksListID, "Pay electricity")
	provider := new(fakeDateTaskProvider)
	response, err := SaveListItemDateTask(userCtx(ctx, testUserID), dto4listus.SaveListItemDateTaskRequest{
		SpaceID: testSpaceID, ListID: dbo4listus.DoTasksListID, ItemID: created.CreatedItems[0].ID,
		OperationID: "op-1", DueDate: "2026-09-30", State: "active",
	}, provider)
	if err != nil {
		t.Fatal(err)
	}
	if !provider.applied || provider.request.Source.Namespace != "listus" || len(provider.request.Related) != 1 || provider.request.Related[0].ItemRef.SubPath == "" {
		t.Fatalf("Calendar plan did not receive canonical Listus identity: %+v", provider.request)
	}
	item := getListData(t, ctx, dbo4listus.DoTasksListID).Items[0]
	if item.DateTask == nil || item.DateTask.Revision != 1 || response.DateTask == nil || response.DateTask.Revision != 1 || item.Linkage == nil {
		t.Fatalf("date task link was not persisted: item=%+v response=%+v", item, response)
	}
}

func TestSaveListItemDateTaskRejectsSourceManagedItem(t *testing.T) {
	ctx, db := newTestDBWithSpace(t, testSpaceID, testUserID)
	created := createItems(t, ctx, dbo4listus.DoTasksListID, "Pay invoice")
	markItemSourceManaged(t, ctx, db, created.CreatedItems[0].ID)
	provider := new(fakeDateTaskProvider)
	_, err := SaveListItemDateTask(userCtx(ctx, testUserID), dto4listus.SaveListItemDateTaskRequest{
		SpaceID: testSpaceID, ListID: dbo4listus.DoTasksListID, ItemID: created.CreatedItems[0].ID,
		OperationID: "op-1", DueDate: "2026-09-30", State: "active",
	}, provider)
	if err == nil || provider.applied {
		t.Fatal("source-managed item mutation should fail before Calendar apply")
	}
}

func TestSaveListItemDateTaskCancelsExistingCalendarTask(t *testing.T) {
	ctx, _ := newTestDBWithSpace(t, testSpaceID, testUserID)
	created := createItems(t, ctx, dbo4listus.DoTasksListID, "Pay electricity")
	itemID := created.CreatedItems[0].ID
	provider := new(fakeDateTaskProvider)
	createdTask, err := SaveListItemDateTask(userCtx(ctx, testUserID), dto4listus.SaveListItemDateTaskRequest{
		SpaceID: testSpaceID, ListID: dbo4listus.DoTasksListID, ItemID: itemID,
		OperationID: "op-create", DueDate: "2026-09-30", State: "active",
	}, provider)
	if err != nil {
		t.Fatal(err)
	}
	if createdTask.DateTask == nil {
		t.Fatal("expected persisted date task")
	}
	response, err := SaveListItemDateTask(userCtx(ctx, testUserID), dto4listus.SaveListItemDateTaskRequest{
		SpaceID: testSpaceID, ListID: dbo4listus.DoTasksListID, ItemID: itemID,
		OperationID: "op-cancel", ExpectedTaskRevision: createdTask.DateTask.Revision, State: "canceled",
	}, provider)
	if err != nil {
		t.Fatal(err)
	}
	if response.DateTask != nil {
		t.Fatalf("canceled response retained date task: %+v", response.DateTask)
	}
	item := getListData(t, ctx, dbo4listus.DoTasksListID).Items[0]
	if item.DateTask != nil {
		t.Fatalf("canceled item retained date task: %+v", item.DateTask)
	}
}
