package facade4listus

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
	"github.com/dal-go/record/update"
	"github.com/sneat-co/listus/backend/const4listus"
	"github.com/sneat-co/listus/backend/dal4listus"
	"github.com/sneat-co/listus/backend/dbo4listus"
	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/sneat-core-modules/linkage/dbo4linkage"
	"github.com/sneat-co/sneat-core-modules/spaceus/dbo4spaceus"
	"github.com/sneat-co/sneat-ext-contracts/calendarius/calendariusmodels"
	calendarfacade "github.com/sneat-co/sneat-ext-contracts/calendarius/facade4calendarius"
	listuscontract "github.com/sneat-co/sneat-ext-contracts/listus/facade4listus"
	"github.com/sneat-co/sneat-ext-contracts/listus/listusmodels"
	"github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/sneat-co/sneat-go-core/facade"
)

type errorApplyingDateTask struct {
	fakePreparedDateTask
}

func (e *errorApplyingDateTask) Apply(context.Context, dal.ReadwriteTransaction) error {
	return errors.New("simulated date task apply error")
}

type errorPlanningProvider struct {
	fakeDateTaskProvider
	applyErr bool
}

func (p *errorPlanningProvider) PlanSourceLinkedDateTask(ctx context.Context, tx dal.ReadwriteTransaction, uid string, req calendariusmodels.MutateSourceLinkedDateTaskRequest) (calendarfacade.PreparedSourceLinkedDateTaskMutation, error) {
	if p.applyErr {
		task := calendariusmodels.SourceLinkedDateTask{HappeningID: "due-list-item", Revision: req.ExpectedRevision + 1}
		return &errorApplyingDateTask{fakePreparedDateTask: fakePreparedDateTask{
			provider: &p.fakeDateTaskProvider,
			mutation: calendariusmodels.SourceLinkedDateTaskMutation{Task: task, Disposition: "changed"},
		}}, nil
	}
	return nil, errors.New("simulated plan error")
}

func TestSaveListItemDateTask_RemainingBranches(t *testing.T) {
	ctx, db := newTestDBWithSpace(t, testSpaceID, testUserID)
	uctx := userCtx(ctx, testUserID)
	created := createItems(t, ctx, dbo4listus.DoTasksListID, "Task 1")
	itemID := created.CreatedItems[0].ID
	provider := new(fakeDateTaskProvider)

	validReq := dto4listus.SaveListItemDateTaskRequest{
		SpaceID: testSpaceID, ListID: dbo4listus.DoTasksListID, ItemID: itemID,
		OperationID: "op-1", DueDate: "2026-09-30", State: "active",
	}

	// 1. Validation error
	if _, err := SaveListItemDateTask(uctx, dto4listus.SaveListItemDateTaskRequest{}, provider); err == nil {
		t.Error("expected validation error")
	}

	// 2. Calendar is nil
	if _, err := SaveListItemDateTask(uctx, validReq, nil); err == nil {
		t.Error("expected error for nil calendar")
	}

	// 3. DB error from facade.GetSneatDB
	errDBCtx := facade.WithSneatDBProvider(context.Background(), func(context.Context) (dal.DB, error) {
		return nil, errors.New("simulated db provider error")
	})
	if _, err := SaveListItemDateTask(userCtx(errDBCtx, testUserID), validReq, provider); err == nil {
		t.Error("expected error for failing DB provider")
	}

	// 4. Missing space
	missingSpaceReq := validReq
	missingSpaceReq.SpaceID = "nonexistentspace"
	if _, err := SaveListItemDateTask(uctx, missingSpaceReq, provider); err == nil {
		t.Error("expected error for missing space")
	}

	// 5. Actor not in space
	strangerCtx := userCtx(ctx, "stranger")
	if _, err := SaveListItemDateTask(strangerCtx, validReq, provider); err == nil {
		t.Error("expected error when actor not in space")
	}

	// 6. Missing list
	missingListReq := validReq
	missingListReq.ListID = "do!missing"
	if _, err := SaveListItemDateTask(uctx, missingListReq, provider); err == nil {
		t.Error("expected error for missing list")
	}

	// 7. Non-do list or actor cannot update list
	watchCreated := createItems(t, ctx, dbo4listus.WatchMoviesListID, "Movie")
	nonDoReq := validReq
	nonDoReq.ListID = dbo4listus.WatchMoviesListID
	nonDoReq.ItemID = watchCreated.CreatedItems[0].ID
	if _, err := SaveListItemDateTask(uctx, nonDoReq, provider); err == nil {
		t.Error("expected error for non-do list")
	}

	// Helper to mutate the test list record directly
	setListItems := func(items []*dbo4listus.ListItemBrief) {
		entry := dal4listus.NewListEntry(testSpaceID, dbo4listus.ListKey(dbo4listus.DoTasksListID))
		_ = db.RunReadwriteTransaction(ctx, func(txCtx context.Context, tx dal.ReadwriteTransaction) error {
			_ = tx.Get(txCtx, entry.Record)
			entry.Data.Items = items
			return tx.Update(txCtx, entry.Key, []update.Update{update.ByFieldName("items", items)})
		})
	}

	// 8. Nil item in list
	setListItems([]*dbo4listus.ListItemBrief{nil})
	if _, err := SaveListItemDateTask(uctx, validReq, provider); err == nil {
		t.Error("expected error for nil item")
	}

	// 9. Duplicate item IDs
	it1 := &dbo4listus.ListItemBrief{ID: itemID, ListItemBase: dbo4listus.ListItemBase{Title: "T1"}}
	it1.CreatedAt = time.Now()
	it1.CreatedBy = testUserID
	it2 := &dbo4listus.ListItemBrief{ID: itemID, ListItemBase: dbo4listus.ListItemBase{Title: "T2"}}
	it2.CreatedAt = time.Now()
	it2.CreatedBy = testUserID
	setListItems([]*dbo4listus.ListItemBrief{it1, it2})
	if _, err := SaveListItemDateTask(uctx, validReq, provider); err == nil {
		t.Error("expected error for duplicate item IDs")
	}

	// 10. List item not found
	setListItems([]*dbo4listus.ListItemBrief{it1})
	notFoundReq := validReq
	notFoundReq.ItemID = "item-not-found"
	if _, err := SaveListItemDateTask(uctx, notFoundReq, provider); err == nil {
		t.Error("expected error when item not found")
	}

	// 11. Revision conflict: DateTask is nil but ExpectedTaskRevision != 0
	conflictReq := validReq
	conflictReq.ExpectedTaskRevision = 5
	if _, err := SaveListItemDateTask(uctx, conflictReq, provider); err == nil {
		t.Error("expected revision conflict")
	}

	// Set a date task on it1 and test mismatch revision
	it1WithDateTask := &dbo4listus.ListItemBrief{
		ID: itemID, ListItemBase: dbo4listus.ListItemBase{Title: "T1"},
	}
	it1WithDateTask.DateTask = &dbo4listus.DateTaskLink{Revision: 2}
	it1WithDateTask.CreatedAt = time.Now()
	it1WithDateTask.CreatedBy = testUserID
	setListItems([]*dbo4listus.ListItemBrief{it1WithDateTask})
	conflictReq2 := validReq
	conflictReq2.ExpectedTaskRevision = 1
	if _, err := SaveListItemDateTask(uctx, conflictReq2, provider); err == nil {
		t.Error("expected revision conflict for mismatch revision")
	}

	// 12. cloneSourceTodoItem error via jsonMarshal hook
	origMarshal := jsonMarshalSourceTodoItem
	defer func() { jsonMarshalSourceTodoItem = origMarshal }()
	jsonMarshalSourceTodoItem = func(v any) ([]byte, error) {
		return nil, errors.New("simulated marshal error")
	}
	conflictReqMatch := validReq
	conflictReqMatch.ExpectedTaskRevision = 2
	if _, err := SaveListItemDateTask(uctx, conflictReqMatch, provider); err == nil {
		t.Error("expected error from jsonMarshalSourceTodoItem")
	}
	jsonMarshalSourceTodoItem = origMarshal

	// 13. listItemRef error via formatItemSubPath hook
	origSubPath := formatItemSubPath
	defer func() { formatItemSubPath = origSubPath }()
	formatItemSubPath = func(segments ...coretypes.ItemSubPathSegment) (string, error) {
		return "", errors.New("simulated subpath error")
	}
	if _, err := SaveListItemDateTask(uctx, conflictReqMatch, provider); err == nil {
		t.Error("expected error from formatItemSubPath")
	}
	formatItemSubPath = origSubPath

	// 14. PlanSourceLinkedDateTask error
	errProvider := &errorPlanningProvider{}
	if _, err := SaveListItemDateTask(uctx, conflictReqMatch, errProvider); err == nil {
		t.Error("expected error from calendar.PlanSourceLinkedDateTask")
	}

	// 15. addRelationshipAndIDInDateTask error
	origRel := addRelationshipAndIDInDateTask
	defer func() { addRelationshipAndIDInDateTask = origRel }()
	addRelationshipAndIDInDateTask = func(_ *dbo4linkage.WithRelatedAndIDs, _ time.Time, _ string, _ coretypes.SpaceID, _ dbo4linkage.RelationshipItemRolesCommand) ([]update.Update, error) {
		return nil, errors.New("simulated addRelationship error")
	}
	if _, err := SaveListItemDateTask(uctx, conflictReqMatch, provider); err == nil {
		t.Error("expected error from addRelationshipAndIDInDateTask")
	}
	addRelationshipAndIDInDateTask = origRel

	// 16. State = completed -> Status becomes "done"
	completedReq := conflictReqMatch
	completedReq.State = "completed"
	respCompleted, err := SaveListItemDateTask(uctx, completedReq, provider)
	if err != nil {
		t.Fatalf("SaveListItemDateTask completed failed: %v", err)
	}
	if respCompleted.ItemID != itemID {
		t.Errorf("expected itemID %s, got %s", itemID, respCompleted.ItemID)
	}

	// 17. prepared.Apply error
	applyErrProvider := &errorPlanningProvider{applyErr: true}
	conflictReqMatch.ExpectedTaskRevision = 3 // matching revision 3
	if _, err := SaveListItemDateTask(uctx, conflictReqMatch, applyErrProvider); err == nil {
		t.Error("expected error from prepared.Apply")
	}

	// 18. txUpdateInDateTask error
	origUp := txUpdateInDateTask
	defer func() { txUpdateInDateTask = origUp }()
	txUpdateInDateTask = func(_ context.Context, _ dal.ReadwriteTransaction, _ *record.Key, _ []update.Update) error {
		return errors.New("simulated tx update failure")
	}
	if _, err := SaveListItemDateTask(uctx, conflictReqMatch, provider); err == nil {
		t.Error("expected error from txUpdateInDateTask")
	}
	txUpdateInDateTask = origUp
}

func TestSourceTodos_Coverage(t *testing.T) {
	ctx, db := newTestDBWithSpace(t, testSpaceID, testUserID)
	port := NewSourceTodoPort()

	// Seed space module entry
	module := dbo4spaceus.NewSpaceModuleEntry(testSpaceID, const4listus.ExtensionID, new(dbo4listus.ListusSpaceDbo))
	module.Data.CreatedAt = time.Now()
	module.Data.CreatedBy = testUserID
	module.Data.Lists = dbo4listus.ListBriefs{
		string(dbo4listus.DoTasksListID): &dbo4listus.ListBrief{
			ListBase: dbo4listus.ListBase{Type: dbo4listus.ListTypeToDo, Title: "Tasks"},
		},
	}
	if err := db.RunReadwriteTransaction(ctx, func(txCtx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Insert(txCtx, module.Record)
	}); err != nil {
		t.Fatalf("failed to insert space module: %v", err)
	}

	// Create list record
	_ = createItems(t, ctx, dbo4listus.DoTasksListID, "T1")

	validSpec := listusmodels.SourceTodoSpec{
		SpaceID:               testSpaceID,
		ListID:                string(dbo4listus.DoTasksListID),
		Source:                coretypes.NewItemRefSameSpace("contactus", "contacts", "c1"),
		DueHappening:          coretypes.NewItemRefSameSpace("calendarius", "happenings", "h1"),
		DueTaskRevision:       1,
		Purpose:               "test-purpose",
		Title:                 "Test Source Todo",
		CompletionActionID:    "complete",
		CompletionDisposition: "navigate",
		State:                 listusmodels.SourceTodoActive,
	}

	runTx := func(fn func(txCtx context.Context, tx dal.ReadwriteTransaction)) {
		_ = db.RunReadwriteTransaction(ctx, func(txCtx context.Context, tx dal.ReadwriteTransaction) error {
			fn(txCtx, tx)
			return nil
		})
	}

	// 1. PlanSourceTodo: spec.Validate() error
	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		if _, err := port.PlanSourceTodo(txCtx, tx, testUserID, listusmodels.SourceTodoSpec{}); err == nil {
			t.Error("expected spec validation error")
		}
	})

	// 2. validateSourceTodoOwnerRefs error (invalid DueHappening)
	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		badRefSpec := validSpec
		badRefSpec.DueHappening = coretypes.NewItemRefSameSpace("invalid", "happenings", "h1")
		if _, err := port.PlanSourceTodo(txCtx, tx, testUserID, badRefSpec); err == nil {
			t.Error("expected error for invalid DueHappening")
		}
	})

	// 3. validateSourceTodoOwnerRefs error (belongs to another space)
	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		otherSpaceRefSpec := validSpec
		otherSpaceRefSpec.Source = coretypes.ItemRef{ExtID: "contactus", Collection: "contacts", ItemID: "c1" + coretypes.SpaceItemIDSeparator + "otherspace"}
		if _, err := port.PlanSourceTodo(txCtx, tx, testUserID, otherSpaceRefSpec); err == nil {
			t.Error("expected error for ref belonging to another space")
		}
	})

	// 4. Missing space
	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		missingSpaceSpec := validSpec
		missingSpaceSpec.SpaceID = "nonexistentspace"
		if _, err := port.PlanSourceTodo(txCtx, tx, testUserID, missingSpaceSpec); err == nil {
			t.Error("expected error for missing space")
		}
	})

	// 5. Actor not in space
	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		if _, err := port.PlanSourceTodo(txCtx, tx, "stranger", validSpec); err == nil {
			t.Error("expected error when actor not in space")
		}
	})

	// 6. Missing list
	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		missingListSpec := validSpec
		missingListSpec.ListID = "do!missing"
		if _, err := port.PlanSourceTodo(txCtx, tx, testUserID, missingListSpec); err == nil {
			t.Error("expected error for missing list")
		}
	})

	// 7. Missing module entry
	missingModuleSpaceID := coretypes.SpaceID("spacemodmissing")
	ctxMod, dbMod := newTestDBWithSpace(t, missingModuleSpaceID, testUserID)
	_ = dbMod.RunReadwriteTransaction(ctxMod, func(tCtx context.Context, txM dal.ReadwriteTransaction) error {
		listEntry := dal4listus.NewListEntry(missingModuleSpaceID, dbo4listus.DoTasksListID)
		listEntry.Data.SpaceIDs = []coretypes.SpaceID{missingModuleSpaceID}
		listEntry.Data.UserIDs = []string{testUserID}
		listEntry.Data.Type = dbo4listus.ListTypeToDo
		listEntry.Data.CreatedAt = time.Now()
		listEntry.Data.CreatedBy = testUserID
		return txM.Insert(tCtx, listEntry.Record)
	})
	_ = dbMod.RunReadwriteTransaction(ctxMod, func(tCtx context.Context, txM dal.ReadwriteTransaction) error {
		specMod := validSpec
		specMod.SpaceID = missingModuleSpaceID
		if _, err := port.PlanSourceTodo(tCtx, txM, testUserID, specMod); err == nil {
			t.Error("expected error for missing module entry")
		}
		return nil
	})

	// 8. Actor has no access to list
	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		listNoAccess := dal4listus.NewListEntry(testSpaceID, "do!noaccess")
		listNoAccess.Data.SpaceIDs = []coretypes.SpaceID{testSpaceID}
		listNoAccess.Data.UserIDs = []string{"other-user"}
		listNoAccess.Data.Type = dbo4listus.ListTypeToDo
		_ = tx.Insert(txCtx, listNoAccess.Record)
	})
	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		noAccessListSpec := validSpec
		noAccessListSpec.ListID = "do!noaccess"
		if _, err := port.PlanSourceTodo(txCtx, tx, testUserID, noAccessListSpec); err == nil {
			t.Error("expected error for actor has no access to list")
		}
	})

	// 9. Non-do list
	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		watchList := dal4listus.NewListEntry(testSpaceID, "watch!todo")
		watchList.Data.SpaceIDs = []coretypes.SpaceID{testSpaceID}
		watchList.Data.UserIDs = []string{testUserID}
		watchList.Data.Type = dbo4listus.ListTypeToWatch
		_ = tx.Insert(txCtx, watchList.Record)
	})
	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		watchSpec := validSpec
		watchSpec.ListID = "watch!todo"
		if _, err := port.PlanSourceTodo(txCtx, tx, testUserID, watchSpec); err == nil {
			t.Error("expected error for non-do list")
		}
	})

	// 10. List missing from module.Data.Lists
	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		orphanList := dal4listus.NewListEntry(testSpaceID, "do!orphan")
		orphanList.Data.SpaceIDs = []coretypes.SpaceID{testSpaceID}
		orphanList.Data.UserIDs = []string{testUserID}
		orphanList.Data.Type = dbo4listus.ListTypeToDo
		_ = tx.Insert(txCtx, orphanList.Record)
	})
	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		orphanSpec := validSpec
		orphanSpec.ListID = "do!orphan"
		if _, err := port.PlanSourceTodo(txCtx, tx, testUserID, orphanSpec); err == nil {
			t.Error("expected error for list missing from module summary")
		}
	})

	// 11. formatItemSubPathInSourceTodo error
	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		origSub := formatItemSubPathInSourceTodo
		defer func() { formatItemSubPathInSourceTodo = origSub }()
		formatItemSubPathInSourceTodo = func(...coretypes.ItemSubPathSegment) (string, error) {
			return "", errors.New("simulated subpath error")
		}
		if _, err := port.PlanSourceTodo(txCtx, tx, testUserID, validSpec); err == nil {
			t.Error("expected error from formatItemSubPathInSourceTodo")
		}
	})

	// Helper to set list items
	setItems := func(items []*dbo4listus.ListItemBrief) {
		runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
			listEntry := dal4listus.NewListEntry(testSpaceID, dbo4listus.DoTasksListID)
			_ = tx.Get(txCtx, listEntry.Record)
			_ = tx.Update(txCtx, listEntry.Key, []update.Update{update.ByFieldName("items", items)})
		})
	}

	// 12. Candidate is nil
	setItems([]*dbo4listus.ListItemBrief{nil})
	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		if _, err := port.PlanSourceTodo(txCtx, tx, testUserID, validSpec); err == nil {
			t.Error("expected error for nil candidate")
		}
	})

	// 13. Candidate matches stable ID, clone fails via jsonMarshal
	stableID := stableSourceTodoItemID(validSpec.Source, validSpec.Purpose)
	cand := &dbo4listus.ListItemBrief{ID: stableID, ListItemBase: dbo4listus.ListItemBase{Title: "Cand"}}
	cand.CreatedAt = time.Now()
	cand.CreatedBy = testUserID
	cand.SourceManagement = &dbo4listus.SourceManagement{
		Source: validSpec.Source, Purpose: validSpec.Purpose, ActionID: "action",
	}
	setItems([]*dbo4listus.ListItemBrief{cand})

	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		origMarshal := jsonMarshalSourceTodoItem
		defer func() { jsonMarshalSourceTodoItem = origMarshal }()
		jsonMarshalSourceTodoItem = func(any) ([]byte, error) {
			return nil, errors.New("simulated marshal error")
		}
		if _, err := port.PlanSourceTodo(txCtx, tx, testUserID, validSpec); err == nil {
			t.Error("expected error when clone fails")
		}
	})

	// 14. Duplicate candidate items in list
	cand2 := &dbo4listus.ListItemBrief{ID: stableID, ListItemBase: dbo4listus.ListItemBase{Title: "Cand 2"}}
	cand2.CreatedAt = time.Now()
	cand2.CreatedBy = testUserID
	setItems([]*dbo4listus.ListItemBrief{cand, cand2})

	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		if _, err := port.PlanSourceTodo(txCtx, tx, testUserID, validSpec); err == nil {
			t.Error("expected error for duplicate items")
		}
	})

	// 15. Candidate has different owner
	candWrongOwner := &dbo4listus.ListItemBrief{ID: stableID, ListItemBase: dbo4listus.ListItemBase{Title: "Wrong"}}
	candWrongOwner.CreatedAt = time.Now()
	candWrongOwner.CreatedBy = testUserID
	candWrongOwner.SourceManagement = &dbo4listus.SourceManagement{
		Source: coretypes.ItemRef{ItemID: "other"}, Purpose: "other",
	}
	setItems([]*dbo4listus.ListItemBrief{candWrongOwner})

	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		if _, err := port.PlanSourceTodo(txCtx, tx, testUserID, validSpec); err == nil {
			t.Error("expected error for wrong owner")
		}
	})

	// 16. Canceled state on existing item (removes item)
	candValid := &dbo4listus.ListItemBrief{ID: stableID, ListItemBase: dbo4listus.ListItemBase{Title: "Valid"}}
	candValid.CreatedAt = time.Now()
	candValid.CreatedBy = testUserID
	candValid.SourceManagement = &dbo4listus.SourceManagement{
		Source: validSpec.Source, Purpose: validSpec.Purpose,
	}
	setItems([]*dbo4listus.ListItemBrief{candValid})

	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		cancelSpec := validSpec
		cancelSpec.State = listusmodels.SourceTodoCanceled
		prepCancel, err := port.PlanSourceTodo(txCtx, tx, testUserID, cancelSpec)
		if err != nil {
			t.Fatalf("PlanSourceTodo cancel failed: %v", err)
		}
		if prepCancel.SourceTodoPlanView().State != listusmodels.SourceTodoCanceled {
			t.Errorf("expected canceled state in view")
		}
	})

	// Reset items to empty
	setItems(nil)

	// 17. addRelationshipAndIDInSourceTodo error for Calendar task
	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		callCount := 0
		origRel := addRelationshipAndIDInSourceTodo
		defer func() { addRelationshipAndIDInSourceTodo = origRel }()
		addRelationshipAndIDInSourceTodo = func(_ *dbo4linkage.WithRelatedAndIDs, _ time.Time, _ string, _ coretypes.SpaceID, _ dbo4linkage.RelationshipItemRolesCommand) ([]update.Update, error) {
			callCount++
			if callCount == 1 {
				return nil, errors.New("simulated calendar link error")
			}
			return nil, nil
		}
		if _, err := port.PlanSourceTodo(txCtx, tx, testUserID, validSpec); err == nil {
			t.Error("expected error linking calendar task")
		}
	})

	// 18. addRelationshipAndIDInSourceTodo error for source owner
	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		callCount := 0
		origRel := addRelationshipAndIDInSourceTodo
		defer func() { addRelationshipAndIDInSourceTodo = origRel }()
		addRelationshipAndIDInSourceTodo = func(_ *dbo4linkage.WithRelatedAndIDs, _ time.Time, _ string, _ coretypes.SpaceID, _ dbo4linkage.RelationshipItemRolesCommand) ([]update.Update, error) {
			callCount++
			if callCount == 2 {
				return nil, errors.New("simulated source owner link error")
			}
			return nil, nil
		}
		if _, err := port.PlanSourceTodo(txCtx, tx, testUserID, validSpec); err == nil {
			t.Error("expected error linking source owner")
		}
	})

	// 19. validateItemInSourceTodo error
	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		origItemVal := validateItemInSourceTodo
		defer func() { validateItemInSourceTodo = origItemVal }()
		validateItemInSourceTodo = func(*dbo4listus.ListItemBrief) error {
			return errors.New("simulated invalid item")
		}
		if _, err := port.PlanSourceTodo(txCtx, tx, testUserID, validSpec); err == nil {
			t.Error("expected error from validateItemInSourceTodo")
		}
	})

	// 20. Successful Plan with State = SourceTodoCompleted
	var prepCompleted listuscontract.PreparedSourceTodo
	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		completedSpec := validSpec
		completedSpec.State = listusmodels.SourceTodoCompleted
		var err error
		prepCompleted, err = port.PlanSourceTodo(txCtx, tx, testUserID, completedSpec)
		if err != nil {
			t.Fatalf("PlanSourceTodo completed failed: %v", err)
		}
		view := prepCompleted.SourceTodoPlanView()
		if view.State != listusmodels.SourceTodoCompleted {
			t.Errorf("expected state completed, got %v", view.State)
		}
	})

	// 21. ApplySourceTodo error cases:
	// a. wrong type or nil
	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		if _, err := port.ApplySourceTodo(txCtx, tx, nil); err == nil {
			t.Error("expected error for nil prepared")
		}
	})

	// b. wrong transaction
	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		if _, err := port.ApplySourceTodo(txCtx, tx, prepCompleted); err == nil {
			t.Error("expected transaction mismatch error")
		}
	})

	// c. txUpdateInSourceTodo error for list entry
	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		plan, err := port.PlanSourceTodo(txCtx, tx, testUserID, validSpec)
		if err != nil {
			t.Fatal(err)
		}
		origTxUp := txUpdateInSourceTodo
		defer func() { txUpdateInSourceTodo = origTxUp }()
		txUpdateInSourceTodo = func(_ context.Context, _ dal.ReadwriteTransaction, _ *record.Key, _ []update.Update) error {
			return errors.New("simulated tx update failure")
		}
		if _, err := port.ApplySourceTodo(txCtx, tx, plan); err == nil {
			t.Error("expected error from txUpdateInSourceTodo on list")
		}
	})

	// d. txUpdateInSourceTodo error for module entry
	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		plan, err := port.PlanSourceTodo(txCtx, tx, testUserID, validSpec)
		if err != nil {
			t.Fatal(err)
		}
		origTxUp := txUpdateInSourceTodo
		defer func() { txUpdateInSourceTodo = origTxUp }()
		upCount := 0
		txUpdateInSourceTodo = func(c context.Context, t dal.ReadwriteTransaction, k *record.Key, u []update.Update) error {
			upCount++
			if upCount == 1 {
				return origTxUp(c, t, k, u)
			}
			return errors.New("simulated module update failure")
		}
		if _, err := port.ApplySourceTodo(txCtx, tx, plan); err == nil {
			t.Error("expected error from txUpdateInSourceTodo on module")
		}
	})

	// e. Successful apply and already applied check
	runTx(func(txCtx context.Context, tx dal.ReadwriteTransaction) {
		prepSuccess, err := port.PlanSourceTodo(txCtx, tx, testUserID, validSpec)
		if err != nil {
			t.Fatalf("PlanSourceTodo failed: %v", err)
		}
		appliedView, err := port.ApplySourceTodo(txCtx, tx, prepSuccess)
		if err != nil {
			t.Fatalf("ApplySourceTodo failed: %v", err)
		}
		if appliedView.ItemID == "" {
			t.Error("expected non-empty item ID")
		}
		// Apply again -> already applied
		if _, err := port.ApplySourceTodo(txCtx, tx, prepSuccess); err == nil {
			t.Error("expected error for already applied plan")
		}
	})

	// 22. sameSourceTodoTransaction edge cases (nil transactions)
	if sameSourceTodoTransaction(nil, nil) {
		t.Error("expected false for nil transactions")
	}

	// 23. cloneSourceTodoItem jsonUnmarshal error
	origUn := jsonUnmarshalSourceTodoItem
	defer func() { jsonUnmarshalSourceTodoItem = origUn }()
	jsonUnmarshalSourceTodoItem = func([]byte, any) error {
		return errors.New("simulated unmarshal error")
	}
	if _, err := cloneSourceTodoItem(&dbo4listus.ListItemBrief{}); err == nil {
		t.Error("expected error from jsonUnmarshalSourceTodoItem")
	}
	jsonUnmarshalSourceTodoItem = origUn
}
