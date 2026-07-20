// Copyright 2026 Sneat.app

package botapi

import (
	"context"

	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/facade"
)

// ContextWithUser and UserContext are application-port context contracts. Bot
// delivery imports this package rather than the host facade directly.
type (
	ContextWithUser = facade.ContextWithUser
	UserContext     = facade.UserContext
)

func NewContextWithUserID(ctx context.Context, userID string) ContextWithUser {
	return facade.NewContextWithUserID(ctx, userID)
}

func NewContextWithUser(ctx context.Context, user UserContext) ContextWithUser {
	return facade.NewContextWithUser(ctx, user)
}

func NewUserContext(userID string) UserContext {
	return facade.NewUserContext(userID)
}

// WithDBProvider and GetDB exist for legacy callback rendering while the
// persistence dependency remains hidden behind Listus' application port.
func WithDBProvider(ctx context.Context, provider func(context.Context) (dal.DB, error)) context.Context {
	return facade.WithSneatDBProvider(ctx, provider)
}

func GetDB(ctx context.Context) (dal.DB, error) {
	return facade.GetSneatDB(ctx)
}
