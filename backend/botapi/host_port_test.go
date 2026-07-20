// Copyright 2026 Sneat.app

package botapi

import (
	"context"
	"errors"
	"testing"
)

type fakeHostServices struct {
	resolve func(context.Context, string, SpaceRef) (SpaceRef, error)
}

func (f fakeHostServices) ResolveSpace(ctx context.Context, userID string, requested SpaceRef) (SpaceRef, error) {
	return f.resolve(ctx, userID, requested)
}

func TestResolveSpaceFailsClosedWithoutHostBinding(t *testing.T) {
	RegisterHostServices(nil)
	t.Cleanup(func() { RegisterHostServices(nil) })
	_, err := ResolveSpace(t.Context(), "u1", SpaceRef{Type: "family"})
	if err == nil {
		t.Fatal("ResolveSpace() error = nil, want missing host binding error")
	}
}

func TestResolveSpaceDelegatesToHostBinding(t *testing.T) {
	want := SpaceRef{ID: "space1", Type: "family"}
	RegisterHostServices(fakeHostServices{resolve: func(_ context.Context, userID string, requested SpaceRef) (SpaceRef, error) {
		if userID != "u1" || requested.Type != "family" {
			t.Fatalf("unexpected request: userID=%q requested=%+v", userID, requested)
		}
		return want, nil
	}})
	t.Cleanup(func() { RegisterHostServices(nil) })
	got, err := ResolveSpace(t.Context(), "u1", SpaceRef{Type: "family"})
	if err != nil {
		t.Fatalf("ResolveSpace() error = %v", err)
	}
	if got != want {
		t.Errorf("ResolveSpace() = %+v, want %+v", got, want)
	}
}

func TestResolveSpacePropagatesHostError(t *testing.T) {
	wantErr := errors.New("host unavailable")
	RegisterHostServices(fakeHostServices{resolve: func(context.Context, string, SpaceRef) (SpaceRef, error) {
		return SpaceRef{}, wantErr
	}})
	t.Cleanup(func() { RegisterHostServices(nil) })
	_, err := ResolveSpace(t.Context(), "u1", SpaceRef{})
	if !errors.Is(err, wantErr) {
		t.Errorf("ResolveSpace() error = %v, want %v", err, wantErr)
	}
}
