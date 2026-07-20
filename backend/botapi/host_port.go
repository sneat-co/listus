// Copyright 2026 Sneat.app

package botapi

import (
	"context"
	"errors"
	"sync"
)

// SpaceRef is the presentation-level space reference used by Listus bot
// delivery adapters. It intentionally does not expose a host persistence
// model or a Sneat facade context.
type SpaceRef struct {
	ID   string
	Type string
}

// HostServices is the narrow host port Listus needs for bot-originated work.
// The host owns user/profile reads and space creation; Listus owns deciding
// when a missing space must be resolved before operating on a list.
type HostServices interface {
	ResolveSpace(ctx context.Context, userID string, requested SpaceRef) (SpaceRef, error)
}

var (
	hostServicesMu sync.RWMutex
	hostServices   HostServices
)

// RegisterHostServices installs the composition-root implementation. It is
// deliberately explicit: a bot must fail closed instead of silently reading a
// foreign persistence model when the host was not wired.
func RegisterHostServices(services HostServices) {
	hostServicesMu.Lock()
	defer hostServicesMu.Unlock()
	hostServices = services
}

// ResolveSpace resolves a bot's selected space through the registered host.
func ResolveSpace(ctx context.Context, userID string, requested SpaceRef) (SpaceRef, error) {
	hostServicesMu.RLock()
	services := hostServices
	hostServicesMu.RUnlock()
	if services == nil {
		return SpaceRef{}, errors.New("listus bot host services are not registered")
	}
	return services.ResolveSpace(ctx, userID, requested)
}
