package facade4listus

import (
	"context"
	"testing"

	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/listus/backend/dbo4listus"
	listuscontract "github.com/sneat-co/sneat-ext-contracts/listus/facade4listus"
	"github.com/sneat-co/sneat-ext-contracts/listus/listusmodels"
	"github.com/sneat-co/sneat-go-core/coretypes"
)

func sourceTodoSpec() listusmodels.SourceTodoSpec {
	return listusmodels.SourceTodoSpec{
		SpaceID: testSpaceID, ListID: dbo4listus.DoTasksListID,
		Source:  coretypes.NewItemRefSameSpace("debtus", "sourceObligations", "invoice-1"),
		Purpose: "payment-due", Title: "Pay Electricity Co.", State: listusmodels.SourceTodoActive,
		DueHappening:       coretypes.NewItemRefSameSpace("calendarius", "happenings", "due-invoice-1"),
		CompletionActionID: "record-payment", CompletionDisposition: listusmodels.SourceTodoRequiresInput,
	}
}

func applySourceTodo(t *testing.T, ctx context.Context, db dal.DB, actor string, spec listusmodels.SourceTodoSpec) listusmodels.SourceTodoPlanView {
	t.Helper()
	var result listusmodels.SourceTodoPlanView
	port := NewSourceTodoPort()
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		prepared, err := port.PlanSourceTodo(ctx, tx, actor, spec)
		if err != nil {
			return err
		}
		result, err = port.ApplySourceTodo(ctx, tx, prepared)
		return err
	}); err != nil {
		t.Fatalf("apply source todo: %v", err)
	}
	return result
}

func TestSourceTodoPortCreatesAndReusesStableLinkedItem(t *testing.T) {
	ctx, db := newTestDBWithSpace(t, testSpaceID, testUserID)
	createItems(t, ctx, dbo4listus.DoTasksListID, "Existing task")

	first := applySourceTodo(t, ctx, db, testUserID, sourceTodoSpec())
	second := applySourceTodo(t, ctx, db, testUserID, sourceTodoSpec())
	if first.ItemID != second.ItemID || first.Item.SubPath != "/items/@id="+first.ItemID {
		t.Fatalf("unstable logical item identity: first=%+v second=%+v", first, second)
	}
	list := getListData(t, ctx, dbo4listus.DoTasksListID)
	if len(list.Items) != 2 {
		t.Fatalf("idempotent apply produced %d items, want 2", len(list.Items))
	}
	item := list.Items[1]
	if item.DateTask == nil || item.DateTask.Happening != first.DueHappening || item.SourceManagement == nil {
		t.Fatalf("missing persisted source/date-task links: %+v", item)
	}
	if item.Linkage == nil || len(item.Linkage.RelatedIDs) == 0 {
		t.Fatalf("missing standard Linkage backlink: %+v", item.Linkage)
	}
}

func TestSourceTodoPortCompletesAndReopensSameItem(t *testing.T) {
	ctx, db := newTestDBWithSpace(t, testSpaceID, testUserID)
	createItems(t, ctx, dbo4listus.DoTasksListID, "Existing task")
	spec := sourceTodoSpec()
	created := applySourceTodo(t, ctx, db, testUserID, spec)
	spec.State = listusmodels.SourceTodoCompleted
	completed := applySourceTodo(t, ctx, db, testUserID, spec)
	if completed.ItemID != created.ItemID || !getListData(t, ctx, dbo4listus.DoTasksListID).Items[1].IsDone() {
		t.Fatal("completion did not update the same source item")
	}
	spec.State = listusmodels.SourceTodoActive
	applySourceTodo(t, ctx, db, testUserID, spec)
	if getListData(t, ctx, dbo4listus.DoTasksListID).Items[1].IsDone() {
		t.Fatal("correction did not reopen the source item")
	}
}

func TestSourceTodoPortRejectsNonMemberAndForeignPreparedPlan(t *testing.T) {
	ctx, db := newTestDBWithSpace(t, testSpaceID, testUserID)
	createItems(t, ctx, dbo4listus.DoTasksListID, "Existing task")
	port := NewSourceTodoPort()
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := port.PlanSourceTodo(ctx, tx, "outsider", sourceTodoSpec())
		return err
	}); err == nil {
		t.Fatal("expected non-member plan to fail")
	}
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := port.ApplySourceTodo(ctx, tx, fakePreparedSourceTodo{})
		return err
	}); err == nil {
		t.Fatal("expected foreign prepared plan to fail")
	}
}

type fakePreparedSourceTodo struct{}

func (fakePreparedSourceTodo) SourceTodoPlanView() listusmodels.SourceTodoPlanView {
	return listusmodels.SourceTodoPlanView{}
}

var _ listuscontract.PreparedSourceTodo = fakePreparedSourceTodo{}
