package facade4listus

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

const testUserID = "user1"

// newTestDBWithSpace builds an in-memory dalgo DB seeded with a single Space
// record (members = userIDs) and wires facade.GetSneatDB to it. The module
// space worker reads this Space to enforce membership access.
func newTestDBWithSpace(t *testing.T, spaceID coretypes.SpaceID, userIDs ...string) dal.DB {
	t.Helper()
	db := dalgo2memory.NewDB()
	now := time.Now()
	space := dbo4spaceus.NewSpaceEntry(spaceID)
	space.Data.Type = coretypes.SpaceTypeFamily
	space.Data.Title = "Test family space"
	space.Data.Status = dbmodels.StatusActive
	space.Data.CreatedAt = now
	space.Data.CreatedBy = "seed"
	space.Data.IncreaseVersion(now, "seed")
	space.Data.UserIDs = userIDs
	if err := space.Data.Validate(); err != nil {
		t.Fatalf("seed space invalid: %v", err)
	}
	ctx := context.Background()
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Insert(ctx, space.Record)
	}); err != nil {
		t.Fatalf("failed to seed space: %v", err)
	}
	facade.GetSneatDB = func(context.Context) (dal.DB, error) { return db, nil }
	return db
}

func userCtx(userID string) facade.ContextWithUser {
	return facade.NewContextWithUserID(context.Background(), userID)
}

func spaceRequest(spaceID coretypes.SpaceID) dto4spaceus.SpaceRequest {
	return dto4spaceus.SpaceRequest{SpaceID: spaceID}
}

func listRequest(spaceID coretypes.SpaceID, listID string) dto4listus.ListRequest {
	return dto4listus.ListRequest{
		SpaceRequest: spaceRequest(spaceID),
		ListID:       dbo4listus.ListKey(listID),
	}
}
