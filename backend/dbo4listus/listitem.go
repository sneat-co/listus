package dbo4listus

import (
	"strings"

	"github.com/sneat-co/listus/backend/const4listus"
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

	Status const4listus.ListItemStatus `json:"status,omitempty" firestore:"status,omitempty"`

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
	if v.WatchWith != nil {
		if err := v.WatchWith.Validate(); err != nil {
			return validation.NewErrBadRecordFieldValue("watchWith", err.Error())
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
