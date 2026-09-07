package listusext

import (
	"context"
	"testing"

	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/listus/backend/api4listus"
	"github.com/sneat-co/listus/backend/const4listus"
	"github.com/sneat-co/sneat-ext-contracts/calendarius/calendariusmodels"
	calendarfacade "github.com/sneat-co/sneat-ext-contracts/calendarius/facade4calendarius"
	"github.com/sneat-co/sneat-go-core/extension"
)

type stubSourceLinkedDateTaskProvider struct{}

func (stubSourceLinkedDateTaskProvider) PlanSourceLinkedDateTask(
	context.Context, dal.ReadwriteTransaction, string, calendariusmodels.MutateSourceLinkedDateTaskRequest,
) (calendarfacade.PreparedSourceLinkedDateTaskMutation, error) {
	return nil, nil
}

func TestModule_withoutDependencies(t *testing.T) {
	m := Extension()
	extension.AssertExtension(t, m, extension.Expected{
		ExtID:         const4listus.ExtensionID,
		HandlersCount: 11,
		DelayersCount: 0,
	})
}

func TestModule_withDependencies(t *testing.T) {
	m := Extension(api4listus.Dependencies{SourceLinkedDateTasks: stubSourceLinkedDateTaskProvider{}})
	extension.AssertExtension(t, m, extension.Expected{
		ExtID:         const4listus.ExtensionID,
		HandlersCount: 12,
		DelayersCount: 0,
	})
}
