package facade4listus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record/update"
	"github.com/sneat-co/listus/backend/const4listus"
	"github.com/sneat-co/listus/backend/dal4listus"
	"github.com/sneat-co/listus/backend/dbo4listus"
	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/sneat-core-modules/linkage/dbo4linkage"
	"github.com/sneat-co/sneat-core-modules/spaceus/dbo4spaceus"
	"github.com/sneat-co/sneat-ext-contracts/calendarius/calendariusmodels"
	calendarfacade "github.com/sneat-co/sneat-ext-contracts/calendarius/facade4calendarius"
	"github.com/sneat-co/sneat-ext-contracts/listus/listusmodels"
	"github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/sneat-co/sneat-go-core/facade"
)

func SaveListItemDateTask(
	ctx facade.ContextWithUser,
	request dto4listus.SaveListItemDateTaskRequest,
	calendar calendarfacade.SourceLinkedDateTaskProvider,
) (response dto4listus.SaveListItemDateTaskResponse, err error) {
	if err = request.Validate(); err != nil {
		return
	}
	if calendar == nil {
		return response, errors.New("Calendar date-task provider is unavailable")
	}
	db, err := facade.GetSneatDB(ctx)
	if err != nil {
		return response, err
	}
	err = db.RunReadwriteTransaction(ctx, func(txCtx context.Context, tx dal.ReadwriteTransaction) error {
		space := dbo4spaceus.NewSpaceEntry(request.SpaceID)
		list := dal4listus.NewListEntry(request.SpaceID, dbo4listus.ListKey(request.ListID))
		if err := tx.Get(txCtx, space.Record); err != nil {
			return err
		}
		if !space.Data.HasUserID(ctx.User().GetUserID()) {
			return errors.New("actor is not a current Space member")
		}
		if err := tx.Get(txCtx, list.Record); err != nil {
			return err
		}
		if list.Data.Type != dbo4listus.ListTypeToDo || !list.Data.HasUserID(ctx.User().GetUserID()) {
			return errors.New("actor cannot update the selected do list")
		}
		index := -1
		for i, item := range list.Data.Items {
			if item == nil {
				return fmt.Errorf("selected list contains a nil item at index %d", i)
			}
			if item.ID == request.ItemID {
				if index >= 0 {
					return errors.New("selected list contains duplicate item IDs")
				}
				index = i
			}
		}
		if index < 0 {
			return errors.New("list item not found")
		}
		original := list.Data.Items[index]
		if original.SourceManagement != nil {
			return errors.New("source-managed items must be changed by their owner")
		}
		if original.DateTask == nil && request.ExpectedTaskRevision != 0 {
			return errors.New("date task revision conflict")
		}
		if original.DateTask != nil && original.DateTask.Revision != request.ExpectedTaskRevision {
			return errors.New("date task revision conflict")
		}
		item, err := cloneSourceTodoItem(original)
		if err != nil {
			return err
		}
		itemRef, err := listItemRef(request.SpaceID, request.ListID, request.ItemID)
		if err != nil {
			return err
		}
		source := calendariusmodels.SourceLinkedDateTaskRef{Namespace: "listus", OwnerSpaceID: string(request.SpaceID), RecordID: request.ListID, LineID: request.ItemID}
		calendarState := calendariusmodels.SourceLinkedDateTaskState(request.State)
		prepared, err := calendar.PlanSourceLinkedDateTask(txCtx, tx, ctx.User().GetUserID(), calendariusmodels.MutateSourceLinkedDateTaskRequest{
			OperationID: request.OperationID, ExpectedRevision: request.ExpectedTaskRevision,
			Source: source, Title: original.Title, DueDate: request.DueDate, State: calendarState,
			ActionID: "open-list-item", ActionDisposition: calendariusmodels.SourceLinkedDateTaskNavigate,
			Related: []calendariusmodels.SourceLinkedDateTaskRelatedRef{{ItemRef: itemRef, Role: "todo-item"}},
		})
		if err != nil {
			return err
		}
		mutation := prepared.Result()
		happeningRef := mutation.Task.HappeningItemRef()
		if item.Linkage == nil {
			item.Linkage = new(dbo4linkage.WithRelatedAndIDs)
		}
		if request.State == listusmodels.SourceTodoCanceled {
			if original.DateTask != nil {
				item.Linkage.RemoveRelatedItem(original.DateTask.Happening)
			}
			item.DateTask = nil
		} else {
			item.DateTask = &dbo4listus.DateTaskLink{Happening: happeningRef, Source: itemRef, Purpose: "due-date", Revision: mutation.Task.Revision}
			if _, err = item.Linkage.AddRelationshipAndID(time.Now(), ctx.User().GetUserID(), request.SpaceID, dbo4linkage.RelationshipItemRolesCommand{ItemRef: happeningRef, Add: &dbo4linkage.RolesCommand{RolesOfItem: []string{"date-task"}, RolesToItem: []string{"todo-item"}}}); err != nil {
				return err
			}
		}
		if request.State == listusmodels.SourceTodoCompleted {
			item.Status = "done"
		} else if request.State == listusmodels.SourceTodoActive {
			item.Status = "active"
		}
		items := append([]*dbo4listus.ListItemBrief(nil), list.Data.Items...)
		items[index] = item
		if err = prepared.Apply(txCtx, tx); err != nil {
			return err
		}
		if err = tx.Update(txCtx, list.Key, []update.Update{update.ByFieldName("items", items)}); err != nil {
			return err
		}
		response = dto4listus.SaveListItemDateTaskResponse{ItemID: item.ID}
		if item.DateTask != nil {
			response.DateTask = &listusmodels.ListItemDateTaskLink{Happening: item.DateTask.Happening, Source: item.DateTask.Source, Purpose: item.DateTask.Purpose, Revision: item.DateTask.Revision}
		}
		return nil
	})
	return
}

func listItemRef(spaceID coretypes.SpaceID, listID, itemID string) (coretypes.ItemRef, error) {
	ref := coretypes.NewFullItemRef(const4listus.ExtensionID, dbo4listus.ListsCollection, spaceID, listID)
	path, err := coretypes.FormatItemSubPath(coretypes.ItemSubPathSegment{Field: "items"}, coretypes.ItemSubPathSegment{Key: "id", Value: itemID})
	if err != nil {
		return coretypes.ItemRef{}, err
	}
	ref.SubPath = path
	return ref, nil
}
