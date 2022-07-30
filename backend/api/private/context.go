package private

import (
	"context"

	"github.com/shake-on-it/app-tmpl/backend/api"
	"github.com/shake-on-it/app-tmpl/backend/api/admin"
)

type ctxKey int

const (
	ctxKeyServerContext ctxKey = iota
)

type ServerContext struct {
	admin.ServerContext
}

func AttachServerContext(ctx context.Context, srvCtx ServerContext) context.Context {
	return context.WithValue(ctx, ctxKeyServerContext, srvCtx)
}

func MustHaveServerContext(r api.Contexter) ServerContext {
	srvCtx, ok := r.Context().Value(ctxKeyServerContext).(ServerContext)
	if !ok {
		panic("must have private server context")
	}
	return srvCtx
}
