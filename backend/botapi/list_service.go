// Copyright 2026 Sneat.app

package botapi

import (
	"context"

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
