package facade4listus

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record/update"
	"github.com/sneat-co/listus/backend/const4listus"
	"github.com/sneat-co/listus/backend/dal4listus"
	"github.com/sneat-co/listus/backend/dbo4listus"
	"github.com/sneat-co/sneat-core-modules/linkage/dbo4linkage"
	"github.com/sneat-co/sneat-core-modules/spaceus/dbo4spaceus"
	listuscontract "github.com/sneat-co/sneat-ext-contracts/listus/facade4listus"
	"github.com/sneat-co/sneat-ext-contracts/listus/listusmodels"
	"github.com/sneat-co/sneat-go-core/coretypes"
)

type sourceTodoPort struct{}

func NewSourceTodoPort() listuscontract.SourceTodoPort { return sourceTodoPort{} }

type preparedSourceTodo struct {
	view   listusmodels.SourceTodoPlanView
	entry  dal4listus.ListEntry
	items  []*dbo4listus.ListItemBrief
	insert bool
}

func (v *preparedSourceTodo) SourceTodoPlanView() listusmodels.SourceTodoPlanView { return v.view }

func (sourceTodoPort) PlanSourceTodo(
	ctx context.Context,
	tx dal.ReadwriteTransaction,
	actorUserID string,
	spec listusmodels.SourceTodoSpec,
) (listuscontract.PreparedSourceTodo, error) {
	if err := spec.Validate(); err != nil {
		return nil, err
	}
	space := dbo4spaceus.NewSpaceEntry(spec.SpaceID)
	if err := tx.Get(ctx, space.Record); err != nil {
		return nil, fmt.Errorf("read Space: %w", err)
	}
	if !space.Data.HasUserID(actorUserID) {
		return nil, fmt.Errorf("actor is not a current member of Space %q", spec.SpaceID)
	}

	entry := dal4listus.NewListEntry(spec.SpaceID, dbo4listus.ListKey(spec.ListID))
	if err := tx.Get(ctx, entry.Record); err != nil {
		return nil, fmt.Errorf("read selected list: %w", err)
	}
	if !entry.Data.HasUserID(actorUserID) {
		return nil, fmt.Errorf("actor has no access to selected list %q", spec.ListID)
	}

	itemID := stableSourceTodoItemID(spec.Source, spec.Purpose)
	itemRef := coretypes.NewItemRefSameSpace(const4listus.ExtensionID, dbo4listus.ListsCollection, spec.ListID)
	subPath, err := coretypes.FormatItemSubPath(
		coretypes.ItemSubPathSegment{Field: "items"},
		coretypes.ItemSubPathSegment{Key: "id", Value: itemID},
	)
	if err != nil {
		return nil, fmt.Errorf("format list item reference: %w", err)
	}
	itemRef.SubPath = subPath

	items := append([]*dbo4listus.ListItemBrief(nil), entry.Data.Items...)
	var item *dbo4listus.ListItemBrief
	for i, candidate := range items {
		if candidate.ID == itemID {
			copyOfItem := *candidate
			item = &copyOfItem
			items[i] = item
			break
		}
	}
	insert := item == nil
	if insert {
		item = &dbo4listus.ListItemBrief{ID: itemID}
		item.CreatedAt = time.Now()
		item.CreatedBy = actorUserID
		items = append(items, item)
	} else if item.SourceManagement == nil || item.SourceManagement.Source != spec.Source || item.SourceManagement.Purpose != spec.Purpose {
		return nil, errors.New("stable source todo ID belongs to a different owner")
	}

	item.Title = spec.Title
	item.SourceManagement = &dbo4listus.SourceManagement{
		Source: spec.Source, Purpose: spec.Purpose, ActionID: spec.CompletionActionID,
		Disposition: dbo4listus.SourceCompletionDisposition(spec.CompletionDisposition),
	}
	item.DateTask = &dbo4listus.DateTaskLink{Happening: spec.DueHappening, Source: spec.Source, Purpose: spec.Purpose}
	if item.Linkage == nil {
		item.Linkage = new(dbo4linkage.WithRelatedAndIDs)
	}
	_, err = item.Linkage.AddRelationshipAndID(time.Now(), actorUserID, spec.SpaceID, dbo4linkage.RelationshipItemRolesCommand{
		ItemRef: spec.DueHappening,
		Add:     &dbo4linkage.RolesCommand{RolesOfItem: []string{"date-task"}, RolesToItem: []string{"todo-item"}},
	})
	if err != nil {
		return nil, fmt.Errorf("link Calendar task: %w", err)
	}
	if spec.State == listusmodels.SourceTodoActive {
		item.Status = const4listus.ListItemStatusActive
	} else {
		item.Status = const4listus.ListItemStatusDone
	}
	if err := item.Validate(); err != nil {
		return nil, fmt.Errorf("prepared source todo is invalid: %w", err)
	}

	view := listusmodels.SourceTodoPlanView{
		ListID: spec.ListID, ItemID: itemID, Item: itemRef, DueHappening: spec.DueHappening,
		ExpectedListRevision: 0, State: spec.State,
	}
	return &preparedSourceTodo{view: view, entry: entry, items: items, insert: insert}, nil
}

func (sourceTodoPort) ApplySourceTodo(
	ctx context.Context,
	tx dal.ReadwriteTransaction,
	prepared listuscontract.PreparedSourceTodo,
) (listusmodels.SourceTodoPlanView, error) {
	plan, ok := prepared.(*preparedSourceTodo)
	if !ok || plan == nil {
		return listusmodels.SourceTodoPlanView{}, errors.New("prepared source todo was not created by Listus")
	}
	updates := []update.Update{
		update.ByFieldName("items", plan.items),
		update.ByFieldName("count", len(plan.items)),
	}
	if err := tx.Update(ctx, plan.entry.Key, updates); err != nil {
		return listusmodels.SourceTodoPlanView{}, fmt.Errorf("apply source todo: %w", err)
	}
	return plan.view, nil
}

func stableSourceTodoItemID(source coretypes.ItemRef, purpose string) string {
	sum := sha256.Sum256([]byte(source.ID() + "\x00" + purpose))
	return "source-" + hex.EncodeToString(sum[:12])
}

var _ listuscontract.SourceTodoPort = sourceTodoPort{}
