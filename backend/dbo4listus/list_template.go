package dbo4listus

import (
	"github.com/dal-go/record"
	"github.com/sneat-co/listus/backend/const4listus"
	"github.com/sneat-co/sneat-core-modules/spaceus/dbo4spaceus"
	"github.com/sneat-co/sneat-go-core/coretypes"
	"time"
)

const ListTemplateApplyReceiptsCollection = "templateApplyReceipts"

type ListTemplateApplyReceiptDbo struct {
	UserID            string           `json:"userID" firestore:"userID"`
	RequestID         string           `json:"requestID" firestore:"requestID"`
	SourceListID      string           `json:"sourceListID" firestore:"sourceListID"`
	DestinationListID string           `json:"destinationListID" firestore:"destinationListID"`
	Added             []*ListItemBrief `json:"added" firestore:"added"`
	Restored          []*ListItemBrief `json:"restored" firestore:"restored"`
	Unchanged         []*ListItemBrief `json:"unchanged" firestore:"unchanged"`
	CreatedAt         time.Time        `json:"createdAt" firestore:"createdAt"`
}

func NewListTemplateApplyReceiptKey(spaceID coretypes.SpaceID, id string) *record.Key {
	return dbo4spaceus.NewSpaceModuleItemKey(spaceID, const4listus.ExtensionID, ListTemplateApplyReceiptsCollection, id)
}
