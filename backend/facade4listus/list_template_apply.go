package facade4listus

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/dal-go/dalgo/dal"
	dalrecord "github.com/dal-go/record"
	"github.com/dal-go/record/update"
	"github.com/sneat-co/listus/backend/dal4listus"
	"github.com/sneat-co/listus/backend/dbo4listus"
	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/sneat-core-modules/spaceus/dto4spaceus"
	"github.com/sneat-co/sneat-go-core/facade"
)

func ApplyListTemplate(ctx facade.ContextWithUser, request dto4listus.ApplyListTemplateRequest) (result dto4listus.ApplyListTemplateResult, err error) {
	if err = request.Validate(); err != nil {
		return result, err
	}
	destinationRequest := dto4listus.ListRequest{SpaceRequest: dto4spaceus.NewSpaceRequest(request.SpaceID), ListID: request.DestinationListID}
	err = dal4listus.RunListWorker(ctx, destinationRequest, func(ctx facade.ContextWithUser, tx dal.ReadwriteTransaction, params *dal4listus.ListWorkerParams) error {
		if err := params.GetRecords(ctx, tx); err != nil {
			return err
		}
		if !params.List.Record.Exists() {
			return fmt.Errorf("destination list not found")
		}
		receiptID := templateReceiptID(params.UserID(), request.RequestID)
		receipt := new(dbo4listus.ListTemplateApplyReceiptDbo)
		receiptRecord := dalrecord.NewRecordWithData(dbo4listus.NewListTemplateApplyReceiptKey(request.SpaceID, receiptID), receipt)
		if getErr := tx.Get(ctx, receiptRecord); getErr == nil {
			if receipt.SourceListID != string(request.SourceListID) || receipt.DestinationListID != string(request.DestinationListID) {
				return fmt.Errorf("requestID was already used for a different template apply")
			}
			result = resultFromReceipt(receipt, request, "reused")
			return nil
		} else if !dalrecord.IsNotFound(getErr) {
			return getErr
		}
		source := dal4listus.NewListEntry(request.SpaceID, request.SourceListID)
		if getErr := tx.Get(ctx, source.Record); getErr != nil {
			return fmt.Errorf("source template list: %w", getErr)
		}
		for _, templateItem := range source.Data.Items {
			if templateItem == nil {
				continue
			}
			copyItem := *templateItem
			copyItem.Status = ""
			existing := matchingItem(params.List.Data.Items, &copyItem)
			if existing == nil {
				copyItem.ID = ""
				id, idErr := generateRandomListItemID(params.List.Data.Items, "")
				if idErr != nil {
					return idErr
				}
				copyItem.ID, copyItem.CreatedAt, copyItem.CreatedBy = id, params.Started, params.UserID()
				params.List.Data.Items = append(params.List.Data.Items, &copyItem)
				result.Added = append(result.Added, &copyItem)
			} else if existing.IsDone() {
				existing.Status = ""
				result.Restored = append(result.Restored, cloneListItem(existing))
			} else {
				result.Unchanged = append(result.Unchanged, cloneListItem(existing))
			}
		}
		params.List.Data.Count = len(params.List.Data.Items)
		if validateErr := params.List.Data.Validate(); validateErr != nil {
			return fmt.Errorf("resulting destination list: %w", validateErr)
		}
		params.List.Record.MarkAsChanged()
		params.ListUpdates = append(params.ListUpdates, update.ByFieldName("items", params.List.Data.Items), update.ByFieldName("count", params.List.Data.Count))
		result.SourceListID, result.DestinationListID, result.ListType, result.Disposition = string(request.SourceListID), string(request.DestinationListID), request.SourceListID.ListType(), "applied"
		receipt = &dbo4listus.ListTemplateApplyReceiptDbo{UserID: params.UserID(), RequestID: request.RequestID, SourceListID: result.SourceListID, DestinationListID: result.DestinationListID, Added: result.Added, Restored: result.Restored, Unchanged: result.Unchanged, CreatedAt: params.Started}
		return tx.Insert(ctx, dalrecord.NewRecordWithData(dbo4listus.NewListTemplateApplyReceiptKey(request.SpaceID, receiptID), receipt))
	})
	return result, err
}

func matchingItem(items []*dbo4listus.ListItemBrief, wanted *dbo4listus.ListItemBrief) *dbo4listus.ListItemBrief {
	for _, item := range items {
		if item != nil && item.Title == wanted.Title && item.Emoji == wanted.Emoji {
			return item
		}
	}
	return nil
}
func cloneListItem(item *dbo4listus.ListItemBrief) *dbo4listus.ListItemBrief {
	clone := *item
	return &clone
}
func templateReceiptID(userID, requestID string) string {
	h := sha256.New()
	for _, value := range []string{userID, requestID} {
		var length [8]byte
		binary.BigEndian.PutUint64(length[:], uint64(len(value)))
		_, _ = h.Write(length[:])
		_, _ = h.Write([]byte(value))
	}
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
}
func resultFromReceipt(r *dbo4listus.ListTemplateApplyReceiptDbo, request dto4listus.ApplyListTemplateRequest, disposition string) dto4listus.ApplyListTemplateResult {
	return dto4listus.ApplyListTemplateResult{SourceListID: r.SourceListID, DestinationListID: r.DestinationListID, ListType: request.SourceListID.ListType(), Added: r.Added, Restored: r.Restored, Unchanged: r.Unchanged, Disposition: disposition}
}
