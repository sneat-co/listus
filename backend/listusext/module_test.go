package listusext

import (
	"testing"

	"github.com/sneat-co/listus/backend/const4listus"
	"github.com/sneat-co/sneat-go-core/extension"
)

func TestModule(t *testing.T) {
	m := Extension()
	extension.AssertExtension(t, m, extension.Expected{
		ExtID:         const4listus.ExtensionID,
		HandlersCount: 10,
		DelayersCount: 0,
	})
}
