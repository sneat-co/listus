package dal4listus

import (
	"context"
	"testing"
	"time"

	"github.com/dal-go/dalgo/adapters/dalgo2memory"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/listus/backend/dbo4listus"
	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/sneat-core-modules/spaceus/dbo4spaceus"
	"github.com/sneat-co/sneat-core-modules/spaceus/dto4spaceus"
	"github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
)

const (
	testUserID                    = "user1"
	testSpaceID coretypes.SpaceID = "space1"
)

func seedDB(t *testing.T) dal.DB {
	t.Helper()
	db := dalgo2memory.NewDB()
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
	facade.GetSneatDB = func(context.Context) (dal.DB, error) { return db, nil }
	return db
}

func userCtx() facade.ContextWithUser {
	return facade.NewContextWithUserID(context.Background(), testUserID)
}

func TestRunListWorker_InvokesWorkerForStandardList(t *testing.T) {
	_ = seedDB(t)
	request := dto4listus.ListRequest{
		SpaceRequest: dto4spaceus.SpaceRequest{SpaceID: testSpaceID},
		ListID:       dbo4listus.DoTasksListID,
	}
	var called bool
	err := RunListWorker(userCtx(), request, func(ctx facade.ContextWithUser, tx dal.ReadwriteTransaction, params *ListWorkerParams) error {
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
	_ = seedDB(t)
	request := dto4listus.ListRequest{
		SpaceRequest: dto4spaceus.SpaceRequest{SpaceID: testSpaceID},
		ListID:       dbo4listus.DoTasksListID,
	}
	wantErr := context.Canceled
	err := RunListWorker(userCtx(), request, func(ctx facade.ContextWithUser, tx dal.ReadwriteTransaction, params *ListWorkerParams) error {
		return wantErr
	})
	if err == nil {
		t.Fatal("expected error from worker to propagate")
	}
}

func TestGetListByID_NotFound(t *testing.T) {
	db := seedDB(t)
	entry := NewListEntry(testSpaceID, dbo4listus.DoTasksListID)
	err := GetListByID(context.Background(), db, entry)
	if err == nil || !dal.IsNotFound(err) {
		t.Fatalf("expected not-found error, got %v", err)
	}
}

func TestGetListForUpdate_RoundTrip(t *testing.T) {
	db := seedDB(t)
	// Insert a list record, then read it back via GetListForUpdate inside a tx.
	entry := NewListEntry(testSpaceID, dbo4listus.DoTasksListID)
	entry.Data.Type = dbo4listus.ListTypeToDo
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
