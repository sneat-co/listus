package dal4listus

import (
	"context"
	"fmt"

	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/listus/backend/dbo4listus"
	"github.com/sneat-co/sneat-ext-contracts/media/models4media"
	"github.com/sneat-co/sneat-go-core/coretypes"
)

const ListItemMediaTargetType = "list_item"
const ListItemPhotoMediaRole = "photo"

// ListItemMediaTargetAdapter connects a space-scoped, embedded Listus item to
// the shared media lifecycle. Current is the read phase used by MediaLink
// replacement/removal. Set is the write phase and always changes the same
// parent list record in the caller's DALgo transaction.
type ListItemMediaTargetAdapter struct{}

func (ListItemMediaTargetAdapter) Current(
	ctx context.Context,
	tx dal.ReadwriteTransaction,
	target models4media.Target,
	role string,
) (*models4media.Ref, error) {
	item, _, err := getListItemMediaTarget(ctx, tx, target, role)
	if err != nil || item.Photo == nil {
		return nil, err
	}
	copy := *item.Photo
	return &copy, nil
}

func (ListItemMediaTargetAdapter) Set(
	ctx context.Context,
	tx dal.ReadwriteTransaction,
	target models4media.Target,
	role string,
	ref *models4media.Ref,
) error {
	item, list, err := getListItemMediaTarget(ctx, tx, target, role)
	if err != nil {
		return err
	}
	if ref == nil {
		item.Photo = nil
	} else {
		copy := *ref
		item.Photo = &copy
	}
	return tx.Set(ctx, list.Record)
}

func getListItemMediaTarget(
	ctx context.Context,
	tx dal.ReadwriteTransaction,
	target models4media.Target,
	role string,
) (*dbo4listus.ListItemBrief, ListEntry, error) {
	if target.Scope != models4media.TargetScopeSpace || target.Type != ListItemMediaTargetType || target.SpaceID == "" || target.ID == "" || target.ParentID == "" {
		return nil, ListEntry{}, fmt.Errorf("invalid Listus list-item media target")
	}
	if role != ListItemPhotoMediaRole {
		return nil, ListEntry{}, fmt.Errorf("unsupported Listus list-item media role %q", role)
	}
	list := NewListEntry(coretypes.SpaceID(target.SpaceID), dbo4listus.ListKey(target.ParentID))
	if err := tx.Get(ctx, list.Record); err != nil {
		return nil, ListEntry{}, fmt.Errorf("read media target list: %w", err)
	}
	for _, item := range list.Data.Items {
		if item != nil && item.ID == target.ID {
			return item, list, nil
		}
	}
	return nil, ListEntry{}, fmt.Errorf("list item %q not found in list %q", target.ID, target.ParentID)
}
