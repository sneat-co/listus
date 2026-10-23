package dal4listus

import (
	"context"
	"testing"
	"time"

	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/listus/backend/dbo4listus"
	"github.com/sneat-co/sneat-ext-contracts/media/models4media"
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/strongo/strongoapp/with"
)

func TestListItemMediaTargetAdapter_CurrentAndSet(t *testing.T) {
	_, db := seedDB(t)
	entry := NewListEntry(testSpaceID, dbo4listus.ListKey("groceries"))
	entry.Data.Type = dbo4listus.ListTypeToBuy
	entry.Data.WithSpaceIDs = dbmodels.WithSingleSpaceID(testSpaceID)
	entry.Data.WithUserIDs = dbmodels.WithUserIDs{UserIDs: []string{testUserID}}
	entry.Data.Count = 1
	entry.Data.Items = []*dbo4listus.ListItemBrief{{
		ID:           "oatmeal",
		ListItemBase: dbo4listus.ListItemBase{Title: "Oatmeal"},
		CreatedFields: with.CreatedFields{
			CreatedAtField: with.CreatedAtField{CreatedAt: time.Now()},
			CreatedByField: with.CreatedByField{CreatedBy: testUserID},
		},
	}}
	if err := db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Insert(ctx, entry.Record)
	}); err != nil {
		t.Fatalf("seed list: %v", err)
	}
	target := models4media.Target{Scope: models4media.TargetScopeSpace, SpaceID: string(testSpaceID), Type: ListItemMediaTargetType, ID: "oatmeal", ParentID: "groceries"}
	adapter := ListItemMediaTargetAdapter{}
	ref := &models4media.Ref{MediaID: "media-oatmeal"}
	if err := db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		if current, err := adapter.Current(ctx, tx, target, ListItemPhotoMediaRole); err != nil || current != nil {
			t.Fatalf("Current() = %v, %v; want nil, nil", current, err)
		}
		return adapter.Set(ctx, tx, target, ListItemPhotoMediaRole, ref)
	}); err != nil {
		t.Fatalf("set photo: %v", err)
	}
	if err := db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		current, err := adapter.Current(ctx, tx, target, ListItemPhotoMediaRole)
		if err != nil {
			return err
		}
		if current == nil || current.MediaID != ref.MediaID {
			t.Fatalf("Current() = %#v; want %q", current, ref.MediaID)
		}
		return adapter.Set(ctx, tx, target, ListItemPhotoMediaRole, nil)
	}); err != nil {
		t.Fatalf("read/remove photo: %v", err)
	}
}

func TestListItemMediaTargetAdapter_RejectsOtherRoles(t *testing.T) {
	adapter := ListItemMediaTargetAdapter{}
	_, db := seedDB(t)
	target := models4media.Target{Scope: models4media.TargetScopeSpace, SpaceID: string(testSpaceID), Type: ListItemMediaTargetType, ID: "item", ParentID: "groceries"}
	err := db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := adapter.Current(ctx, tx, target, "avatar")
		return err
	})
	if err == nil {
		t.Fatal("Current() with avatar role returned nil error")
	}
}

func TestListItemMediaTargetAdapter_Errors(t *testing.T) {
	adapter := ListItemMediaTargetAdapter{}
	_, db := seedDB(t)

	// 1. Invalid target in Current
	invalidTarget := models4media.Target{Scope: "invalid"}
	err := db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := adapter.Current(ctx, tx, invalidTarget, ListItemPhotoMediaRole)
		return err
	})
	if err == nil {
		t.Fatal("expected error for invalid target in Current")
	}

	// 2. Invalid target in Set
	err = db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return adapter.Set(ctx, tx, invalidTarget, ListItemPhotoMediaRole, nil)
	})
	if err == nil {
		t.Fatal("expected error for invalid target in Set")
	}

	// 3. List does not exist in DB (tx.Get fails)
	nonExistentTarget := models4media.Target{
		Scope:    models4media.TargetScopeSpace,
		SpaceID:  string(testSpaceID),
		Type:     ListItemMediaTargetType,
		ID:       "item1",
		ParentID: "non-existent-list",
	}
	err = db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := adapter.Current(ctx, tx, nonExistentTarget, ListItemPhotoMediaRole)
		return err
	})
	if err == nil {
		t.Fatal("expected error when list does not exist")
	}

	// 4. List exists, but item is not in list
	entry := NewListEntry(testSpaceID, dbo4listus.ListKey("empty-list"))
	entry.Data.Type = dbo4listus.ListTypeToDo
	entry.Data.WithSpaceIDs = dbmodels.WithSingleSpaceID(testSpaceID)
	entry.Data.WithUserIDs = dbmodels.WithUserIDs{UserIDs: []string{testUserID}}
	if err := db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Insert(ctx, entry.Record)
	}); err != nil {
		t.Fatalf("seed empty-list: %v", err)
	}

	itemNotFoundTarget := models4media.Target{
		Scope:    models4media.TargetScopeSpace,
		SpaceID:  string(testSpaceID),
		Type:     ListItemMediaTargetType,
		ID:       "missing-item",
		ParentID: "empty-list",
	}
	err = db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := adapter.Current(ctx, tx, itemNotFoundTarget, ListItemPhotoMediaRole)
		return err
	})
	if err == nil {
		t.Fatal("expected error when item not in list")
	}
}

