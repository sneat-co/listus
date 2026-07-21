// Copyright 2026 Sneat.app

package botapi

import (
	"context"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/update"
	"github.com/sneat-co/listus/backend/const4listus"
	"github.com/sneat-co/listus/backend/dal4listus"
	"github.com/sneat-co/listus/backend/dbo4listus"
	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/listus/backend/facade4listus"
	"github.com/sneat-co/sneat-core-modules/spaceus/dto4spaceus"
	"github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/sneat-co/sneat-go-core/facade"
)

// GetList reads a Listus list. Authorization and persistence remain inside
// Listus; bot delivery supplies only user and presentation identifiers.
func GetList(ctx context.Context, userID string, ref ListRef) (ListView, error) {
	list, err := facade4listus.GetList(contextWithUser(ctx, userID), listRequest(ref))
	return listView(list), err
}

// CreateListItems adds value-only bot input to a Listus list.
func CreateListItems(ctx context.Context, userID string, request CreateListItemsRequest) (ListItemsResult, error) {
	response, list, err := facade4listus.CreateListItems(contextWithUser(ctx, userID), dto4listus.CreateListItemsRequest{
		ListRequest: listRequest(request.List),
		Items:       createListItems(request.Items),
	})
	return ListItemsResult{Items: listItemViews(response.CreatedItems), List: listView(list)}, err
}

// DeleteListItems removes addressed items without exposing Listus persistence
// records to the caller.
func DeleteListItems(ctx context.Context, userID string, request ListItemsRequest) (ListItemsResult, error) {
	items, list, err := facade4listus.DeleteListItems(contextWithUser(ctx, userID), dto4listus.ListItemIDsRequest{
		ListRequest: listRequest(request.List),
		ItemIDs:     request.ItemIDs,
	})
	return ListItemsResult{Items: listItemViews(items), List: listView(list)}, err
}

// SetListItemsDone changes completion state for addressed items.
func SetListItemsDone(ctx context.Context, userID string, request SetListItemsDoneRequest) (ListItemsResult, error) {
	items, list, err := facade4listus.SetListItemsIsDone(contextWithUser(ctx, userID), dto4listus.ListItemsSetIsDoneRequest{
		ListItemIDsRequest: dto4listus.ListItemIDsRequest{
			ListRequest: listRequest(request.List),
			ItemIDs:     request.ItemIDs,
		},
		IsDone: request.IsDone,
	})
	return ListItemsResult{Items: listItemViews(items), List: listView(list)}, err
}

// ClearListItems clears a list through Listus's internal worker. Transaction
// and update types never cross the bot-facing API boundary.
func ClearListItems(ctx context.Context, userID string, ref ListRef) (ListView, error) {
	var list dal4listus.ListEntry
	err := dal4listus.RunListWorker(contextWithUser(ctx, userID), listRequest(ref),
		func(_ facade.ContextWithUser, _ dal.ReadwriteTransaction, params *dal4listus.ListWorkerParams) error {
			params.List.Data.Items = nil
			params.List.Data.Count = 0
			params.ListUpdates = append(params.ListUpdates,
				update.DeleteByFieldName("items"),
				update.ByFieldName("count", 0),
			)
			params.List.Record.MarkAsChanged()
			list = params.List
			return nil
		},
	)
	return listView(list), err
}

// IsNotFound reports whether a Listus operation could not find its requested
// value. It intentionally exposes no DAL type or persistence capability.
func IsNotFound(err error) bool { return dal.IsNotFound(err) }

func contextWithUser(ctx context.Context, userID string) facade.ContextWithUser {
	return facade.NewContextWithUserID(ctx, userID)
}

func listRequest(ref ListRef) dto4listus.ListRequest {
	return dto4listus.ListRequest{
		SpaceRequest: dto4spaceus.SpaceRequest{SpaceID: coretypes.SpaceID(ref.Space.ID)},
		ListID:       dbo4listus.ListKey(ref.ID),
	}
}

func createListItems(inputs []ListItemInput) []dto4listus.CreateListItemRequest {
	items := make([]dto4listus.CreateListItemRequest, len(inputs))
	for i, input := range inputs {
		items[i] = dto4listus.CreateListItemRequest{
			ID: input.ID,
			ListItemBase: dbo4listus.ListItemBase{
				Title:             input.Title,
				Emoji:             input.Emoji,
				Status:            const4listus.ListItemStatus(input.Status),
				TmdbID:            input.TmdbID,
				Year:              input.Year,
				PosterURL:         input.PosterURL,
				Overview:          input.Overview,
				TrailerYouTubeKey: input.TrailerYouTubeKey,
				Cast:              input.Cast,
				WatchWith:         watchWith(input.WatchWith),
			},
		}
	}
	return items
}

func listView(entry dal4listus.ListEntry) ListView {
	if entry.Data == nil {
		return ListView{}
	}
	return ListView{
		ID:    entry.ID,
		Type:  string(entry.Data.Type),
		Title: entry.Data.Title,
		Emoji: entry.Data.Emoji,
		Count: entry.Data.Count,
		Items: listItemViews(entry.Data.Items),
	}
}

func listItemViews(items []*dbo4listus.ListItemBrief) []ListItemView {
	views := make([]ListItemView, len(items))
	for i, item := range items {
		views[i] = listItemView(item)
	}
	return views
}

func listItemView(item *dbo4listus.ListItemBrief) ListItemView {
	if item == nil {
		return ListItemView{}
	}
	return ListItemView{
		ID:                item.ID,
		Title:             item.Title,
		Emoji:             item.Emoji,
		Status:            string(item.Status),
		TmdbID:            item.TmdbID,
		Year:              item.Year,
		PosterURL:         item.PosterURL,
		Overview:          item.Overview,
		TrailerYouTubeKey: item.TrailerYouTubeKey,
		Cast:              item.Cast,
		WatchWith:         watchWithView(item.WatchWith),
	}
}

func watchWith(value *WatchWith) *dbo4listus.WatchWith {
	if value == nil {
		return nil
	}
	return &dbo4listus.WatchWith{Mode: value.Mode, Ref: value.Ref, Title: value.Title}
}

func watchWithView(value *dbo4listus.WatchWith) *WatchWith {
	if value == nil {
		return nil
	}
	return &WatchWith{Mode: value.Mode, Ref: value.Ref, Title: value.Title}
}
