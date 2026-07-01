package facade4listus

import (
	"fmt"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/update"
	"github.com/sneat-co/listus/backend/dal4listus"
	"github.com/sneat-co/listus/backend/dbo4listus"
	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/sneat-go-core/facade"
)

// SetListItemWatchWith updates the WatchWith field (alone/space/contact) of
// a single existing watch-list item.
func SetListItemWatchWith(ctx facade.ContextWithUser, request dto4listus.SetListItemWatchWithRequest) (item *dbo4listus.ListItemBrief, list dal4listus.ListEntry, err error) {
	if err = request.Validate(); err != nil {
		return
	}
	err = dal4listus.RunListWorker(ctx, request.ListRequest,
		func(ctx facade.ContextWithUser, tx dal.ReadwriteTransaction, params *dal4listus.ListWorkerParams) (err error) {
			if err = params.GetRecords(ctx, tx); err != nil {
				return
			}
			list = params.List
			for _, listItem := range params.List.Data.Items {
				if listItem.ID == request.ItemID {
					watchWith := request.WatchWith
					listItem.WatchWith = &watchWith
					item = listItem
					break
				}
			}
			if item == nil {
				return fmt.Errorf("list item not found: itemID=%s", request.ItemID)
			}
			params.List.Record.MarkAsChanged()
			params.ListUpdates = []update.Update{update.ByFieldName("items", params.List.Data.Items)}
			return nil
		},
	)
	return
}
