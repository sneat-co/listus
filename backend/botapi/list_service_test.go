// Copyright 2026 Sneat.app

package botapi

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/listus/backend/const4listus"
	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/sneat-core-modules/spaceus/dbo4spaceus"
	"github.com/sneat-co/sneat-core-modules/spaceus/dto4spaceus"
	"github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/sneat-co/sneat-go-core/sneatcoretesting"
)

const (
	serviceTestUserID  = "user1"
	serviceTestSpaceID = "space1"
)

func serviceTestContext(t *testing.T) context.Context {
	t.Helper()
	db := sneatcoretesting.NewMemoryDB()
	space := dbo4spaceus.NewSpaceEntry(serviceTestSpaceID)
	now := time.Now()
	space.Data.Type = coretypes.SpaceTypeFamily
	space.Data.Title = "Test family"
	space.Data.Status = dbmodels.StatusActive
	space.Data.CreatedAt = now
	space.Data.CreatedBy = "seed"
	space.Data.IncreaseVersion(now, "seed")
	space.Data.UserIDs = []string{serviceTestUserID}
	if err := db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Insert(ctx, space.Record)
	}); err != nil {
		t.Fatalf("seed space: %v", err)
	}
	return facade.WithSneatDB(context.Background(), db)
}

func serviceRequest() dto4listus.ListRequest {
	return dto4listus.ListRequest{
		SpaceRequest: dto4spaceus.SpaceRequest{SpaceID: serviceTestSpaceID},
		ListID:       TasksListID,
	}
}

func TestListServicePresentationOperations(t *testing.T) {
	ctx := serviceTestContext(t)
	userCtx := NewContextWithUserID(ctx, serviceTestUserID)
	request := serviceRequest()
	created, _, err := CreateListItems(userCtx, dto4listus.CreateListItemsRequest{
		ListRequest: request,
		Items: []dto4listus.CreateListItemRequest{{
			ID:           "item1",
			ListItemBase: ListItemBase{Title: "Milk"},
		}},
	})
	if err != nil || len(created.CreatedItems) != 1 {
		t.Fatalf("CreateListItems() = %+v, %v", created, err)
	}
	list, err := GetList(ctx, serviceTestUserID, serviceTestSpaceID, TasksListID)
	if err != nil || len(list.Data.Items) != 1 {
		t.Fatalf("GetList() = %+v, %v", list, err)
	}
	changed, _, err := SetListItemsIsDone(userCtx, dto4listus.ListItemsSetIsDoneRequest{
		ListItemIDsRequest: dto4listus.ListItemIDsRequest{ListRequest: request, ItemIDs: []string{"item1"}},
		IsDone:             true,
	})
	if err != nil || len(changed) != 1 || changed[0].Status != const4listus.ListItemStatusDone {
		t.Fatalf("SetListItemsIsDone() = %+v, %v", changed, err)
	}
	cleared, err := ClearListItems(userCtx, request)
	if err != nil || len(cleared.Data.Items) != 0 {
		t.Fatalf("ClearListItems() = %+v, %v", cleared, err)
	}
}

func TestIsNotFound(t *testing.T) {
	if !IsNotFound(dal.ErrRecordNotFound) || IsNotFound(errors.New("other")) {
		t.Fatal("IsNotFound() did not preserve DAL not-found semantics")
	}
}
