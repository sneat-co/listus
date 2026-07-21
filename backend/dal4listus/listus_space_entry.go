package dal4listus

import (
	"github.com/dal-go/record"
	"github.com/sneat-co/listus/backend/dbo4listus"
)

type ListusSpaceEntry = record.DataWithID[string, *dbo4listus.ListusSpaceDbo]
