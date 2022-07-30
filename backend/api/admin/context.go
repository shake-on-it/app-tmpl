package admin

import (
	"context"

	"github.com/shake-on-it/app-tmpl/backend/api"
	"github.com/shake-on-it/app-tmpl/backend/common"
	"github.com/shake-on-it/app-tmpl/backend/core"
)

type ctxKey int

const (
	ctxKeyServerContext ctxKey = iota
)

type ServerContext struct {
	Config common.Config

	AuthService *core.AuthService

	// RefreshTokenStore *core.RefreshTokenStore
	// PasswordStore *core.PasswordStore
	// UserStore     *core.UserStore
}

func AttachServerContext(ctx context.Context, srvCtx ServerContext) context.Context {
	return context.WithValue(ctx, ctxKeyServerContext, srvCtx)
}

func MustHaveServerContext(r api.Contexter) ServerContext {
	srvCtx, ok := r.Context().Value(ctxKeyServerContext).(ServerContext)
	if !ok {
		panic("must have admin server context")
	}
	return srvCtx
}
