package facade4listus

import (
	"fmt"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
	"github.com/sneat-co/listus/backend/dal4listus"
	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/sneat-go-core/facade"
)

// GetList returns the list record for the given space & list ID.
// Membership is enforced by RunListWorker (via RunModuleSpaceWorkerWithUserCtx).
// Returns a not-found error (detectable via record.IsNotFound) if the list is absent.
func GetList(ctx facade.ContextWithUser, request dto4listus.ListRequest) (list dal4listus.ListEntry, err error) {
	if err = request.Validate(); err != nil {
		return
	}
	err = dal4listus.RunListWorker(ctx, request,
		func(ctx facade.ContextWithUser, tx dal.ReadwriteTransaction, params *dal4listus.ListWorkerParams) (err error) {
			if !params.List.Record.Exists() {
				return dal.NewErrNotFoundByKey(params.List.Record.Key(), record.ErrRecordNotFound)
			}
			list = params.List
			return nil
		},
	)
	if err != nil {
		err = fmt.Errorf("GetList(): %w", err)
	}
	return
}
