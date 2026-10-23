package dbo4listus

import (
	"strings"

	"github.com/sneat-co/listus/backend/const4listus"
	"github.com/sneat-co/sneat-core-modules/linkage/dbo4linkage"
	"github.com/sneat-co/sneat-ext-contracts/media/models4media"
	"github.com/strongo/strongoapp/with"
	"github.com/strongo/validation"
)

// WatchWith records who a watch-list movie will be (or was) watched with.
// Mirrors the established denormalized-reference convention used across the
// codebase (id + display title, no join needed at read time).
type WatchWith struct {
	// Mode is one of "alone" | "space" | "contact".
	Mode string `json:"mode" firestore:"mode"`
	// Ref is the spaceID (mode=="space") or contactID (mode=="contact"). Empty for "alone".
	Ref string `json:"ref,omitempty" firestore:"ref,omitempty"`
	// Title is a denormalized display name (e.g. space title or contact name).
	Title string `json:"title,omitempty" firestore:"title,omitempty"`
}

const (
	WatchWithModeAlone   = "alone"
	WatchWithModeSpace   = "space"
	WatchWithModeContact = "contact"
)

// Validate returns error if not valid
func (v WatchWith) Validate() error {
	switch v.Mode {
	case WatchWithModeAlone:
		return nil
	case WatchWithModeSpace, WatchWithModeContact:
		if strings.TrimSpace(v.Ref) == "" {
			return validation.NewErrRecordIsMissingRequiredField("ref")
		}
		return nil
	default:
		return validation.NewErrBadRecordFieldValue("mode", "unknown value: "+v.Mode)
	}
}

// ListItemBase DTO
type ListItemBase struct {
	Title string `json:"title" firestore:"title"`
	Emoji string `json:"emoji,omitempty" firestore:"emoji,omitempty"`
	// Photo is a compact forward reference for fast list rendering. The shared
	// media registry and its MediaLink remain the lifecycle authority.
	Photo *models4media.Ref `json:"photo,omitempty" firestore:"photo,omitempty"`

	Status const4listus.ListItemStatus `json:"status,omitempty" firestore:"status,omitempty"`

	// Linkage stores reciprocal standard Sneat relationships for this logical
	// embedded item. Nil keeps legacy items compact. Listus resolves the logical
	// list-item ItemRef to the owning list document; no duplicate item document
	// is created merely to participate in Linkage.
	Linkage *dbo4linkage.WithRelatedAndIDs `json:"linkage,omitempty" firestore:"linkage,omitempty"`

	// SourceManagement is present only for an item placed by another extension
	// through the trusted SourceTodo port. Public Listus item commands must not
	// accept it from a client. The standard Linkage graph remains the navigation
	// authority; this metadata controls how completion is delegated to the owner.
	SourceManagement *SourceManagement `json:"sourceManagement,omitempty" firestore:"sourceManagement,omitempty"`
	DateTask         *DateTaskLink     `json:"dateTask,omitempty" firestore:"dateTask,omitempty"`

	// The following fields are optional and let a "buy"-typed list item (a
	// shopping list) express a structured amount without changing what the
	// item is: the title stays the bare noun (e.g. "milk"), so "bought milk"
	// still matches by title and two differently-sized entries of the same
	// product still collide as the same item. Kept compact & flat as items
	// are rewritten as a whole array on every mutation (see ListDbo.Items).
	// All omitempty so non-buy lists are completely unaffected.

	// Quantity is the numeric amount of Unit (e.g. 2). Zero means unspecified.
	Quantity float64 `json:"quantity,omitempty" firestore:"quantity,omitempty"`
	// Unit is a caller-supplied unit label (e.g. "L", "kg", "pcs"), stored
	// verbatim. No enum, normalisation, or conversion happens at this layer.
	Unit string `json:"unit,omitempty" firestore:"unit,omitempty"`

	// The following fields are optional and only used by "watch"-typed list
	// items (movies). Kept compact & flat as items are rewritten as a whole
	// array on every mutation (see ListDbo.Items). All omitempty so non-watch
	// lists are completely unaffected.

	// TmdbID is The Movie Database (TMDB) movie id.
	TmdbID int `json:"tmdbID,omitempty" firestore:"tmdbID,omitempty"`
	// Year is the movie release year.
	Year int `json:"year,omitempty" firestore:"year,omitempty"`
	// PosterURL is a pre-constructed TMDB poster image URL (w500 size).
	PosterURL string `json:"posterURL,omitempty" firestore:"posterURL,omitempty"`
	// Overview is the movie synopsis/description.
	Overview string `json:"overview,omitempty" firestore:"overview,omitempty"`
	// TrailerYouTubeKey is the YouTube video key for the movie trailer
	// (build the URL as https://www.youtube.com/watch?v={key}).
	TrailerYouTubeKey string `json:"trailerYouTubeKey,omitempty" firestore:"trailerYouTubeKey,omitempty"`
	// Cast holds the top ~5 cast member names (denormalized, no join needed).
	Cast []string `json:"cast,omitempty" firestore:"cast,omitempty"`
	// WatchWith records who this movie is/was watched with.
	WatchWith *WatchWith `json:"watchWith,omitempty" firestore:"watchWith,omitempty"`
}

func (v ListItemBase) IsDone() bool {
	return v.Status == const4listus.ListItemStatusDone
}

// Validate returns error if not valid
func (v ListItemBase) Validate() error {
	if strings.TrimSpace(v.Title) == "" {
		return validation.NewErrRecordIsMissingRequiredField("title")
	}
	if v.Photo != nil {
		if err := v.Photo.Validate(); err != nil {
			return validation.NewErrBadRecordFieldValue("photo", err.Error())
		}
	}
	if v.Quantity < 0 {
		return validation.NewErrBadRecordFieldValue("quantity", "must not be negative")
	}
	if v.Unit != strings.TrimSpace(v.Unit) {
		return validation.NewErrBadRecordFieldValue("unit", "must not have leading or trailing whitespace")
	}
	if v.Unit != "" && v.Quantity == 0 {
		return validation.NewErrBadRecordFieldValue("unit", "requires a quantity")
	}
	if v.WatchWith != nil {
		if err := v.WatchWith.Validate(); err != nil {
			return validation.NewErrBadRecordFieldValue("watchWith", err.Error())
		}
	}
	if v.Linkage != nil {
		if err := v.Linkage.Validate(); err != nil {
			return validation.NewErrBadRecordFieldValue("linkage", err.Error())
		}
	}
	if v.SourceManagement != nil {
		if err := v.SourceManagement.Validate(); err != nil {
			return validation.NewErrBadRecordFieldValue("sourceManagement", err.Error())
		}
	}
	if v.DateTask != nil {
		if err := v.DateTask.Validate(); err != nil {
			return validation.NewErrBadRecordFieldValue("dateTask", err.Error())
		}
	}
	return nil
}

// ListItemBrief DTO
type ListItemBrief struct {
	ID string `json:"id" firestore:"id"`
	ListItemBase
	with.CreatedFields
}

// Validate returns error if not valid
func (v ListItemBrief) Validate() error {
	if strings.TrimSpace(v.ID) == "" {
		return validation.NewErrRecordIsMissingRequiredField("id")
	}
	if err := v.ListItemBase.Validate(); err != nil {
		return err
	}
	if err := v.CreatedFields.Validate(); err != nil {
		return err
	}
	return nil
}

// ListItemDbo DTO
type ListItemDbo struct {
	ListItemBase
}
