package dbo4listus

import (
	"testing"

	"github.com/sneat-co/sneat-go-core/coretypes"
)

func validItemRef() coretypes.ItemRef {
	return coretypes.NewItemRefSameSpace("contactus", "contacts", "c1")
}

func TestDateTaskLink_Validate(t *testing.T) {
	tests := []struct {
		name    string
		v       DateTaskLink
		wantErr bool
	}{
		{
			name: "valid",
			v: DateTaskLink{
				Happening: validItemRef(),
				Source:    validItemRef(),
				Purpose:   "calendar",
				Revision:  1,
			},
			wantErr: false,
		},
		{
			name: "invalid_happening",
			v: DateTaskLink{
				Happening: coretypes.ItemRef{},
				Source:    validItemRef(),
				Purpose:   "calendar",
				Revision:  1,
			},
			wantErr: true,
		},
		{
			name: "invalid_source",
			v: DateTaskLink{
				Happening: validItemRef(),
				Source:    coretypes.ItemRef{},
				Purpose:   "calendar",
				Revision:  1,
			},
			wantErr: true,
		},
		{
			name: "empty_purpose",
			v: DateTaskLink{
				Happening: validItemRef(),
				Source:    validItemRef(),
				Purpose:   "",
				Revision:  1,
			},
			wantErr: true,
		},
		{
			name: "untrimmed_purpose",
			v: DateTaskLink{
				Happening: validItemRef(),
				Source:    validItemRef(),
				Purpose:   "  calendar ",
				Revision:  1,
			},
			wantErr: true,
		},
		{
			name: "revision_less_than_1",
			v: DateTaskLink{
				Happening: validItemRef(),
				Source:    validItemRef(),
				Purpose:   "calendar",
				Revision:  0,
			},
			wantErr: true,
		},
		{
			name: "revision_too_large",
			v: DateTaskLink{
				Happening: validItemRef(),
				Source:    validItemRef(),
				Purpose:   "calendar",
				Revision:  9_007_199_254_740_992,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.v.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("DateTaskLink.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSourceManagement_Validate(t *testing.T) {
	tests := []struct {
		name    string
		v       SourceManagement
		wantErr bool
	}{
		{
			name: "valid_navigate",
			v: SourceManagement{
				Source:      validItemRef(),
				Purpose:     "todo",
				ActionID:    "act1",
				Disposition: SourceCompletionNavigate,
			},
			wantErr: false,
		},
		{
			name: "valid_requires_input",
			v: SourceManagement{
				Source:      validItemRef(),
				Purpose:     "todo",
				ActionID:    "act1",
				Disposition: SourceCompletionRequiresInput,
			},
			wantErr: false,
		},
		{
			name: "invalid_source",
			v: SourceManagement{
				Source:      coretypes.ItemRef{},
				Purpose:     "todo",
				ActionID:    "act1",
				Disposition: SourceCompletionNavigate,
			},
			wantErr: true,
		},
		{
			name: "empty_purpose",
			v: SourceManagement{
				Source:      validItemRef(),
				Purpose:     "",
				ActionID:    "act1",
				Disposition: SourceCompletionNavigate,
			},
			wantErr: true,
		},
		{
			name: "untrimmed_purpose",
			v: SourceManagement{
				Source:      validItemRef(),
				Purpose:     " todo ",
				ActionID:    "act1",
				Disposition: SourceCompletionNavigate,
			},
			wantErr: true,
		},
		{
			name: "empty_action_id",
			v: SourceManagement{
				Source:      validItemRef(),
				Purpose:     "todo",
				ActionID:    "",
				Disposition: SourceCompletionNavigate,
			},
			wantErr: true,
		},
		{
			name: "untrimmed_action_id",
			v: SourceManagement{
				Source:      validItemRef(),
				Purpose:     "todo",
				ActionID:    " act1 ",
				Disposition: SourceCompletionNavigate,
			},
			wantErr: true,
		},
		{
			name: "unknown_disposition",
			v: SourceManagement{
				Source:      validItemRef(),
				Purpose:     "todo",
				ActionID:    "act1",
				Disposition: "unknown",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.v.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("SourceManagement.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
