package facade4listus

import (
	"context"
	"errors"
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
		DueTaskRevision:    1,
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

func TestSourceTodoPortPlanIsTransactionBoundAndOneShot(t *testing.T) {
	ctx, db := newTestDBWithSpace(t, testSpaceID, testUserID)
	createItems(t, ctx, dbo4listus.DoTasksListID, "Existing task")
	port := NewSourceTodoPort()
	var prepared listuscontract.PreparedSourceTodo
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		var err error
		prepared, err = port.PlanSourceTodo(ctx, tx, testUserID, sourceTodoSpec())
		if err != nil {
			return err
		}
		if _, err = port.ApplySourceTodo(ctx, tx, prepared); err != nil {
			return err
		}
		_, err = port.ApplySourceTodo(ctx, tx, prepared)
		if err == nil {
			return errors.New("second apply unexpectedly succeeded")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := port.ApplySourceTodo(ctx, tx, prepared)
		return err
	}); err == nil {
		t.Fatal("expected applying prepared plan in another transaction to fail")
	}
}

func TestSourceTodoPortRejectsForeignOrNonCalendarRefs(t *testing.T) {
	ctx, db := newTestDBWithSpace(t, testSpaceID, testUserID)
	createItems(t, ctx, dbo4listus.DoTasksListID, "Existing task")
	port := NewSourceTodoPort()
	for name, mutate := range map[string]func(*listusmodels.SourceTodoSpec){
		"foreign source":    func(v *listusmodels.SourceTodoSpec) { v.Source.ItemID += "@other-space" },
		"foreign happening": func(v *listusmodels.SourceTodoSpec) { v.DueHappening.ItemID += "@other-space" },
		"wrong due owner":   func(v *listusmodels.SourceTodoSpec) { v.DueHappening.ExtID = "debtus" },
		"nested happening":  func(v *listusmodels.SourceTodoSpec) { v.DueHappening.SubPath = "/items/@id=x" },
	} {
		t.Run(name, func(t *testing.T) {
			spec := sourceTodoSpec()
			mutate(&spec)
			err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
				_, err := port.PlanSourceTodo(ctx, tx, testUserID, spec)
				return err
			})
			if err == nil {
				t.Fatal("expected invalid owner reference to fail")
			}
		})
	}
}

type fakePreparedSourceTodo struct{}

func (fakePreparedSourceTodo) SourceTodoPlanView() listusmodels.SourceTodoPlanView {
	return listusmodels.SourceTodoPlanView{}
}

var _ listuscontract.PreparedSourceTodo = fakePreparedSourceTodo{}
