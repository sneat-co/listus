// Copyright 2026 Sneat.app

package botapi

import (
	"context"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/update"
	"github.com/sneat-co/listus/backend/dal4listus"
	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/listus/backend/facade4listus"
	"github.com/sneat-co/sneat-core-modules/spaceus/dto4spaceus"
	"github.com/sneat-co/sneat-go-core/coretypes"
)

// GetList is the bot-facing read service. Authorization and persistence stay
// in Listus; delivery code only supplies presentation identifiers.
func GetList(ctx context.Context, userID, spaceID string, listID ListID) (ListEntry, error) {
	return facade4listus.GetList(NewContextWithUserID(ctx, userID), dto4listus.ListRequest{
		SpaceRequest: dto4spaceus.SpaceRequest{SpaceID: coretypes.SpaceID(spaceID)},
		ListID:       listID,
	})
}

// CreateListItems, DeleteListItems, and SetListItemsIsDone are the Listus
// application operations consumed by bot delivery. Their persistence types are
// deliberately contained in this backend-owned package.
func CreateListItems(ctx ContextWithUser, request dto4listus.CreateListItemsRequest) (dto4listus.CreateListItemResponse, ListEntry, error) {
	return facade4listus.CreateListItems(ctx, request)
}

func DeleteListItems(ctx ContextWithUser, request dto4listus.ListItemIDsRequest) ([]*ListItem, ListEntry, error) {
	return facade4listus.DeleteListItems(ctx, request)
}

func SetListItemsIsDone(ctx ContextWithUser, request dto4listus.ListItemsSetIsDoneRequest) ([]*ListItem, ListEntry, error) {
	return facade4listus.SetListItemsIsDone(ctx, request)
}

// ClearListItems clears a list through the Listus worker without exposing a
// transaction or update implementation to bot delivery.
func ClearListItems(ctx ContextWithUser, request dto4listus.ListRequest) (ListEntry, error) {
	var list ListEntry
	err := dal4listus.RunListWorker(ctx, request, func(_ ContextWithUser, _ dal.ReadwriteTransaction, params *dal4listus.ListWorkerParams) error {
		params.List.Data.Items = nil
		params.List.Data.Count = 0
		params.ListUpdates = append(params.ListUpdates, update.DeleteByFieldName("items"))
		params.List.Record.MarkAsChanged()
		list = params.List
		return nil
	})
	return list, err
}

func IsNotFound(err error) bool { return dal.IsNotFound(err) }
