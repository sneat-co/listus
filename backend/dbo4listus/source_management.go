package dbo4listus

import (
	"strings"

	"github.com/sneat-co/sneat-core-modules/linkage/dbo4linkage"
	"github.com/strongo/validation"
)

type SourceCompletionDisposition string

const (
	SourceCompletionNavigate      SourceCompletionDisposition = "navigate"
	SourceCompletionRequiresInput SourceCompletionDisposition = "requires_input"
)

// SourceManagement tells Listus that completion belongs to the linked source
// owner. It contains no financial state and grants no authority over Source.
type SourceManagement struct {
	Source      dbo4linkage.ItemRef         `json:"source" firestore:"source"`
	Purpose     string                      `json:"purpose" firestore:"purpose"`
	ActionID    string                      `json:"actionID" firestore:"actionID"`
	Disposition SourceCompletionDisposition `json:"disposition" firestore:"disposition"`
}

type DateTaskLink struct {
	Happening dbo4linkage.ItemRef `json:"happening" firestore:"happening"`
	Source    dbo4linkage.ItemRef `json:"source" firestore:"source"`
	Purpose   string              `json:"purpose" firestore:"purpose"`
}

func (v DateTaskLink) Validate() error {
	if err := v.Happening.Validate(); err != nil {
		return validation.NewErrBadRecordFieldValue("happening", err.Error())
	}
	if err := v.Source.Validate(); err != nil {
		return validation.NewErrBadRecordFieldValue("source", err.Error())
	}
	if strings.TrimSpace(v.Purpose) == "" || strings.TrimSpace(v.Purpose) != v.Purpose {
		return validation.NewErrBadRecordFieldValue("purpose", "must be non-empty and trimmed")
	}
	return nil
}

func (v SourceManagement) Validate() error {
	if err := v.Source.Validate(); err != nil {
		return validation.NewErrBadRecordFieldValue("source", err.Error())
	}
	if strings.TrimSpace(v.Purpose) == "" || strings.TrimSpace(v.Purpose) != v.Purpose {
		return validation.NewErrBadRecordFieldValue("purpose", "must be non-empty and trimmed")
	}
	if strings.TrimSpace(v.ActionID) == "" || strings.TrimSpace(v.ActionID) != v.ActionID {
		return validation.NewErrBadRecordFieldValue("actionID", "must be non-empty and trimmed")
	}
	switch v.Disposition {
	case SourceCompletionNavigate, SourceCompletionRequiresInput:
		return nil
	default:
		return validation.NewErrBadRecordFieldValue("disposition", "unknown value: "+string(v.Disposition))
	}
}
