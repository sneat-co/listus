package listusext

import (
	"github.com/sneat-co/listus/backend/api4listus"
	"github.com/sneat-co/listus/backend/const4listus"
	"github.com/sneat-co/sneat-go-core/extension"
)

func Extension() extension.Config {
	return extension.NewExtension(const4listus.ExtensionID,
		extension.RegisterRoutes(api4listus.RegisterHttpRoutes),
	)
}
