package dto4listus

import (
	"fmt"
	"strings"

	"github.com/sneat-co/listus/backend/dbo4listus"
	"github.com/sneat-co/sneat-go-core/coretypes"
)

type ApplyListTemplateRequest struct {
	SpaceID           coretypes.SpaceID  `json:"spaceID"`
	SourceListID      dbo4listus.ListKey `json:"sourceListID"`
	DestinationListID dbo4listus.ListKey `json:"destinationListID"`
	RequestID         string             `json:"requestID"`
}

func (v ApplyListTemplateRequest) Validate() error {
	if v.SpaceID == "" || strings.TrimSpace(v.RequestID) == "" || len(v.RequestID) > 200 {
		return fmt.Errorf("spaceID and bounded requestID are required")
	}
	for _, r := range v.RequestID {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("requestID must not contain control characters")
		}
	}
	if err := v.SourceListID.Validate(); err != nil {
		return fmt.Errorf("sourceListID: %w", err)
	}
	if err := v.DestinationListID.Validate(); err != nil {
		return fmt.Errorf("destinationListID: %w", err)
	}
	if v.SourceListID == v.DestinationListID {
		return fmt.Errorf("source and destination lists must differ")
	}
	if v.SourceListID.ListType() != v.DestinationListID.ListType() {
		return fmt.Errorf("source and destination list types must match")
	}
	t := v.SourceListID.ListType()
	if t != dbo4listus.ListTypeToBuy && t != dbo4listus.ListTypeToDo {
		return fmt.Errorf("templates support buy and do lists")
	}
	return nil
}

type ApplyListTemplateResult struct {
	SourceListID      string                      `json:"sourceListID"`
	DestinationListID string                      `json:"destinationListID"`
	ListType          string                      `json:"listType"`
	Added             []*dbo4listus.ListItemBrief `json:"added"`
	Restored          []*dbo4listus.ListItemBrief `json:"restored"`
	Unchanged         []*dbo4listus.ListItemBrief `json:"unchanged"`
	Disposition       string                      `json:"disposition"`
}
