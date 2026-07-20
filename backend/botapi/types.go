// Copyright 2026 Sneat.app

// Package botapi exposes the narrow Listus data vocabulary consumed by bot
// delivery adapters. It deliberately keeps storage package names out of bot
// code; persistence remains an implementation detail of Listus.
package botapi

import (
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/listus/backend/dal4listus"
	"github.com/sneat-co/listus/backend/dbo4listus"
	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/sneat-co/sneat-go-core/facade"
)

// Presentation DTOs used when rendering a list.
type (
	List                 = dal4listus.ListEntry
	ListEntry            = dal4listus.ListEntry
	ListItem             = dbo4listus.ListItemBrief
	ListItemIn           = dbo4listus.ListItemBase
	ListItemBase         = dbo4listus.ListItemBase
	ListID               = dbo4listus.ListKey
	ListType             = dbo4listus.ListType
	ListWorker           = dal4listus.ListWorker
	ListWorkerParams     = dal4listus.ListWorkerParams
	ReadwriteTransaction = dal.ReadwriteTransaction
)

const (
	GroceriesListID = dbo4listus.BuyGroceriesListID
	TasksListID     = dbo4listus.DoTasksListID
	MoviesListID    = dbo4listus.WatchMoviesListID
	BooksListID     = dbo4listus.ReadBooksListID

	ListTypeToBuy   = dbo4listus.ListTypeToBuy
	ListTypeToDo    = dbo4listus.ListTypeToDo
	ListTypeToWatch = dbo4listus.ListTypeToWatch
	ListTypeToRead  = dbo4listus.ListTypeToRead
)

func NewListID(listType ListType, name string) ListID {
	return dbo4listus.NewListKey(listType, name)
}

func NewList(spaceID coretypes.SpaceID, listID ListID) List {
	return dal4listus.NewListEntry(spaceID, listID)
}

func NewListEntry(spaceID coretypes.SpaceID, listID ListID) ListEntry {
	return dal4listus.NewListEntry(spaceID, listID)
}

// RunListWorker is retained for legacy bot callback rendering while keeping
// the DAL implementation behind this Listus-owned port.
func RunListWorker(ctx facade.ContextWithUser, request dto4listus.ListRequest, worker ListWorker) error {
	return dal4listus.RunListWorker(ctx, request, func(ctx facade.ContextWithUser, tx dal.ReadwriteTransaction, params *dal4listus.ListWorkerParams) error {
		return worker(ctx, tx, params)
	})
}
