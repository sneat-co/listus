// Copyright 2026 Sneat.app

// Package botapi exposes Listus application operations for bot delivery. Its
// exported data is deliberately value-only: persistence, transactions, and
// facade contexts remain internal to the Listus backend.
package botapi

import botapp "github.com/sneat-co/ext-listus/backend/botapp"

// SpaceRef is the portable host contract used to select a Listus space.
type SpaceRef = botapp.SpaceRef

// ListRef identifies a Listus list without exposing a datastore key or a
// backend record.
type ListRef struct {
	Space SpaceRef
	ID    string
}

// ListItemInput is a bot-supplied list item. It contains presentation data
// only; ownership and audit fields are assigned by Listus.
type ListItemInput struct {
	ID                string
	Title             string
	Emoji             string
	Status            string
	TmdbID            int
	Year              int
	PosterURL         string
	Overview          string
	TrailerYouTubeKey string
	Cast              []string
	WatchWith         *WatchWith
}

// WatchWith is the presentation-safe watch companion reference for a movie.
type WatchWith struct {
	Mode  string
	Ref   string
	Title string
}

// ListItemView is the data bot delivery may render after a Listus operation.
type ListItemView = ListItemInput

// ListView is a presentation-safe projection of a Listus list.
type ListView struct {
	ID    string
	Type  string
	Title string
	Emoji string
	Count int
	Items []ListItemView
}

// CreateListItemsRequest asks Listus to add the supplied items to a list.
type CreateListItemsRequest struct {
	List  ListRef
	Items []ListItemInput
}

// ListItemsRequest addresses one or more items in a list.
type ListItemsRequest struct {
	List    ListRef
	ItemIDs []string
}

// SetListItemsDoneRequest changes completion state for addressed list items.
type SetListItemsDoneRequest struct {
	ListItemsRequest
	IsDone bool
}

// ListItemsResult returns the changed presentation values and their list.
type ListItemsResult struct {
	Items []ListItemView
	List  ListView
}

const (
	GroceriesListID = "buy!groceries"
	TasksListID     = "do!tasks"
	MoviesListID    = "watch!movies"
	BooksListID     = "read!books"
)
