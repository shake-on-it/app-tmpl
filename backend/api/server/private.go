package server

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/shake-on-it/app-tmpl/backend/api/middleware"
	"github.com/shake-on-it/app-tmpl/backend/api/private"
	"github.com/shake-on-it/app-tmpl/backend/api/private/router"
	"github.com/shake-on-it/app-tmpl/backend/common"
)

type apiPrivate struct {
	adminAPI *apiAdmin
	logger   common.Logger
}

func (a *apiPrivate) setup(ctx context.Context) error {
	return nil
}

func (api apiPrivate) ServerContext() private.ServerContext {
	return private.ServerContext{
		api.adminAPI.ServerContext(),
	}
}

func (api *apiPrivate) ApplyRoutes(r *mux.Router) {
	for _, version := range router.Versions {
		for _, route := range router.Registry[version] {
			var handler http.Handler = route.Handler

			handler = api.attachServerContext(handler)

			switch route.Endpoint.Method {
			case http.MethodGet, http.MethodHead:
				handler = middleware.RequestCacheBuster(handler)
			}

			methods := make([]string, 0, 2)
			methods = append(methods, route.Endpoint.Method)
			if route.Endpoint.UseCORS {
				methods = append(methods, http.MethodOptions)
			}

			r.Path(version + route.Endpoint.Path).Methods(methods...).Handler(handler)
		}
	}
}

type ctxKey int

const (
	ctxKeyServerContext ctxKey = iota
)

func (a apiPrivate) attachServerContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(private.AttachServerContext(r.Context(), a.ServerContext())))
	})
}
