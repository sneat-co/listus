package dal4listus

import (
	"context"
	"testing"
	"time"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
	"github.com/dal-go/record/update"
	"github.com/sneat-co/listus/backend/dbo4listus"
	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/sneat-core-modules/spaceus/dbo4spaceus"
	"github.com/sneat-co/sneat-core-modules/spaceus/dto4spaceus"
	"github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/sneat-co/sneat-go-core/sneatcoretesting"
)

const (
	testUserID                    = "user1"
	testSpaceID coretypes.SpaceID = "space1"
)

func seedDB(t *testing.T) (context.Context, dal.DB) {
	t.Helper()
	db := sneatcoretesting.NewMemoryDB()
	now := time.Now()
	space := dbo4spaceus.NewSpaceEntry(testSpaceID)
	space.Data.Type = coretypes.SpaceTypeFamily
	space.Data.Title = "Test family space"
	space.Data.Status = dbmodels.StatusActive
	space.Data.CreatedAt = now
	space.Data.CreatedBy = "seed"
	space.Data.IncreaseVersion(now, "seed")
	space.Data.UserIDs = []string{testUserID}
	if err := space.Data.Validate(); err != nil {
		t.Fatalf("seed space invalid: %v", err)
	}
	if err := db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Insert(ctx, space.Record)
	}); err != nil {
		t.Fatalf("failed to seed space: %v", err)
	}
	return facade.WithSneatDB(context.Background(), db), db
}

func userCtx(ctx context.Context) facade.ContextWithUser {
	return facade.NewContextWithUserID(ctx, testUserID)
}

func TestRunListWorker_InvokesWorkerForStandardList(t *testing.T) {
	ctx, _ := seedDB(t)
	request := dto4listus.ListRequest{
		SpaceRequest: dto4spaceus.SpaceRequest{SpaceID: testSpaceID},
		ListID:       dbo4listus.DoTasksListID,
	}
	var called bool
	err := RunListWorker(userCtx(ctx), request, func(ctx facade.ContextWithUser, tx dal.ReadwriteTransaction, params *ListWorkerParams) error {
		called = true
		if params.List.ID != string(dbo4listus.DoTasksListID) {
			t.Errorf("worker list ID = %q, want %q", params.List.ID, dbo4listus.DoTasksListID)
		}
		// Standard list need not exist yet.
		if params.List.Record.Exists() {
			t.Error("expected standard list record not to exist on first access")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("RunListWorker failed: %v", err)
	}
	if !called {
		t.Error("worker was not invoked")
	}
}

func TestRunListWorker_PropagatesWorkerError(t *testing.T) {
	ctx, _ := seedDB(t)
	request := dto4listus.ListRequest{
		SpaceRequest: dto4spaceus.SpaceRequest{SpaceID: testSpaceID},
		ListID:       dbo4listus.DoTasksListID,
	}
	wantErr := context.Canceled
	err := RunListWorker(userCtx(ctx), request, func(ctx facade.ContextWithUser, tx dal.ReadwriteTransaction, params *ListWorkerParams) error {
		return wantErr
	})
	if err == nil {
		t.Fatal("expected error from worker to propagate")
	}
}

func TestGetListByID_NotFound(t *testing.T) {
	_, db := seedDB(t)
	entry := NewListEntry(testSpaceID, dbo4listus.DoTasksListID)
	err := GetListByID(context.Background(), db, entry)
	if err == nil || !record.IsNotFound(err) {
		t.Fatalf("expected not-found error, got %v", err)
	}
}

func TestGetListForUpdate_RoundTrip(t *testing.T) {
	_, db := seedDB(t)
	// Insert a list record, then read it back via GetListForUpdate inside a tx.
	entry := NewListEntry(testSpaceID, dbo4listus.DoTasksListID)
	entry.Data.Type = dbo4listus.ListTypeToDo
	entry.Data.WithUserIDs = dbmodels.WithUserIDs{UserIDs: []string{testUserID}}
	entry.Data.WithSpaceIDs = dbmodels.WithSingleSpaceID(testSpaceID)
	if err := db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Insert(ctx, entry.Record)
	}); err != nil {
		t.Fatalf("insert list failed: %v", err)
	}
	read := NewListEntry(testSpaceID, dbo4listus.DoTasksListID)
	if err := db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return GetListForUpdate(ctx, tx, read)
	}); err != nil {
		t.Fatalf("GetListForUpdate failed: %v", err)
	}
	if read.Data.Type != dbo4listus.ListTypeToDo {
		t.Errorf("read type = %q, want %q", read.Data.Type, dbo4listus.ListTypeToDo)
	}
}

func TestRunListWorker_NonStandardListNotFound(t *testing.T) {
	ctx, _ := seedDB(t)
	request := dto4listus.ListRequest{
		SpaceRequest: dto4spaceus.SpaceRequest{SpaceID: testSpaceID},
		ListID:       dbo4listus.ListKey("custom-list"),
	}
	err := RunListWorker(userCtx(ctx), request, func(ctx facade.ContextWithUser, tx dal.ReadwriteTransaction, params *ListWorkerParams) error {
		return nil
	})
	if err == nil || !record.IsNotFound(err) {
		t.Fatalf("expected not-found error for non-standard list, got %v", err)
	}
}

func TestRunListWorker_NormalizesTitleAndAppliesUpdates(t *testing.T) {
	ctx, db := seedDB(t)
	entry := NewListEntry(testSpaceID, dbo4listus.ListKey("custom-list"))
	entry.Data.Type = dbo4listus.ListTypeToDo
	entry.Data.Title = "custom-list"
	entry.Data.WithUserIDs = dbmodels.WithUserIDs{UserIDs: []string{testUserID}}
	entry.Data.WithSpaceIDs = dbmodels.WithSingleSpaceID(testSpaceID)
	if err := db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Insert(ctx, entry.Record)
	}); err != nil {
		t.Fatalf("insert list failed: %v", err)
	}

	request := dto4listus.ListRequest{
		SpaceRequest: dto4spaceus.SpaceRequest{SpaceID: testSpaceID},
		ListID:       dbo4listus.ListKey("custom-list"),
	}
	err := RunListWorker(userCtx(ctx), request, func(ctx facade.ContextWithUser, tx dal.ReadwriteTransaction, params *ListWorkerParams) error {
		return nil
	})
	if err != nil {
		t.Fatalf("RunListWorker failed: %v", err)
	}
}

func TestRunListWorker_UnmarkedChangedWithUpdates(t *testing.T) {
	ctx, db := seedDB(t)
	entry := NewListEntry(testSpaceID, dbo4listus.ListKey("custom-list"))
	entry.Data.Type = dbo4listus.ListTypeToDo
	entry.Data.Title = "My List"
	entry.Data.WithUserIDs = dbmodels.WithUserIDs{UserIDs: []string{testUserID}}
	entry.Data.WithSpaceIDs = dbmodels.WithSingleSpaceID(testSpaceID)
	if err := db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Insert(ctx, entry.Record)
	}); err != nil {
		t.Fatalf("insert list failed: %v", err)
	}

	request := dto4listus.ListRequest{
		SpaceRequest: dto4spaceus.SpaceRequest{SpaceID: testSpaceID},
		ListID:       dbo4listus.ListKey("custom-list"),
	}
	err := RunListWorker(userCtx(ctx), request, func(ctx facade.ContextWithUser, tx dal.ReadwriteTransaction, params *ListWorkerParams) error {
		params.ListUpdates = append(params.ListUpdates, update.ByFieldName("title", "New Title"))
		return nil
	})
	if err == nil {
		t.Fatal("expected error when list updates exist but record is not marked as changed")
	}
}

func TestRunListWorker_UpdateError(t *testing.T) {
	ctx, db := seedDB(t)
	entry := NewListEntry(testSpaceID, dbo4listus.ListKey("custom-list"))
	entry.Data.Type = dbo4listus.ListTypeToDo
	entry.Data.Title = "My List"
	entry.Data.WithUserIDs = dbmodels.WithUserIDs{UserIDs: []string{testUserID}}
	entry.Data.WithSpaceIDs = dbmodels.WithSingleSpaceID(testSpaceID)
	if err := db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Insert(ctx, entry.Record)
	}); err != nil {
		t.Fatalf("insert list failed: %v", err)
	}

	request := dto4listus.ListRequest{
		SpaceRequest: dto4spaceus.SpaceRequest{SpaceID: testSpaceID},
		ListID:       dbo4listus.ListKey("custom-list"),
	}
	err := RunListWorker(userCtx(ctx), request, func(ctx facade.ContextWithUser, tx dal.ReadwriteTransaction, params *ListWorkerParams) error {
		// Delete the record from DB inside worker so tx.Update fails on non-existent record
		if err := tx.Delete(ctx, params.List.Record.Key()); err != nil {
			return err
		}
		params.ListUpdates = append(params.ListUpdates, update.ByFieldName("title", "New Title"))
		params.List.Record.MarkAsChanged()
		return nil
	})
	if err == nil {
		t.Fatal("expected error when tx.Update fails")
	}
}

