package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/shake-on-it/app-tmpl/backend/api"
	"github.com/shake-on-it/app-tmpl/backend/api/admin"
	"github.com/shake-on-it/app-tmpl/backend/api/admin/router"
	"github.com/shake-on-it/app-tmpl/backend/api/middleware"
	"github.com/shake-on-it/app-tmpl/backend/auth"
	"github.com/shake-on-it/app-tmpl/backend/common"
	"github.com/shake-on-it/app-tmpl/backend/core"
	"github.com/shake-on-it/app-tmpl/backend/core/mongodb"

	"github.com/gorilla/mux"
)

var (
	tenYears = 10 * 365 * 24 * time.Hour
)

type apiAdmin struct {
	config        common.Config
	logger        common.Logger
	mongoProvider mongodb.Provider

	AuthService core.AuthService

	RefreshTokenStore core.RefreshTokenStore
	PasswordStore     core.PasswordStore
	UserStore         core.UserStore
}

func (a *apiAdmin) setup(ctx context.Context) error {
	if a.mongoProvider == nil {
		return errors.New("must have mongodb provider to setup admin api")
	}

	userStore, err := core.NewUserStore(a.mongoProvider.Client())
	if err != nil {
		return err
	}

	passwordStore, err := core.NewPasswordStore(a.mongoProvider.Client())
	if err != nil {
		return err
	}

	refreshTokenStore, err := core.NewRefreshTokenStore(a.mongoProvider.Client())
	if err != nil {
		return err
	}

	a.AuthService = core.NewAuthService(a.config, userStore, passwordStore, refreshTokenStore)
	a.RefreshTokenStore = refreshTokenStore
	a.PasswordStore = passwordStore
	a.UserStore = userStore
	return nil
}

func (a apiAdmin) ServerContext() admin.ServerContext {
	return admin.ServerContext{
		Config:      a.config,
		AuthService: &a.AuthService,
	}
}

func (a *apiAdmin) ApplyRoutes(r *mux.Router) {
	for _, version := range router.Versions {
		for _, route := range router.Registry[version] {
			var handler http.Handler = route.Handler

			// TODO: check access
			// TODO: user request limit (admin only?)

			if route.Needs&api.RouteNeedsUser != 0 {
				handler = a.loadUser(handler)
			}

			if route.Needs&api.RouteNeedsAccessToken != 0 {
				handler = a.loadAccessToken(handler)
			}

			if route.Needs&api.RouteNeedsRefreshToken != 0 {
				handler = a.loadRefreshToken(handler)
			}

			handler = a.attachRefreshToken(handler)
			handler = a.attachAccessToken(handler)
			handler = a.attachUserToken(handler)
			handler = a.attachServerContext(handler)

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

func (a apiAdmin) attachServerContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(admin.AttachServerContext(r.Context(), a.ServerContext())))
	})
}

func (a apiAdmin) attachAccessToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(auth.CookieAccessToken)
		if err != nil {
			if err == http.ErrNoCookie {
				next.ServeHTTP(w, r)
			} else {
				api.ErrorResponse(w, r, auth.ErrMalformedCookie)
			}
			return
		}

		if cookie.Value == "" {
			api.ErrorResponse(w, r, auth.ErrMustAuthenticate)
			return
		}

		var accessToken auth.AccessToken
		if err := a.AuthService.ParseToken(cookie.Value, &accessToken); err != nil {
			api.ErrorResponse(w, r, err)
			return
		}

		next.ServeHTTP(w, r.WithContext(
			api.NewContextBuilder(r.Context()).
				AttachAccessToken(accessToken).
				Context(),
		))
	})
}

func (a apiAdmin) loadAccessToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := api.CtxAccessToken(r)
		if !ok {
			api.ErrorResponse(w, r, auth.ErrMustAuthenticate)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a apiAdmin) attachRefreshToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(auth.CookieRefreshToken)
		if err != nil {
			if err == http.ErrNoCookie {
				next.ServeHTTP(w, r)
			} else {
				api.ErrorResponse(w, r, auth.ErrMalformedCookie)
			}
			return
		}

		if cookie.Value == "" {
			api.ErrorResponse(w, r, auth.ErrMustAuthenticate)
			return
		}

		var refreshToken auth.RefreshToken
		if err := a.AuthService.ParseToken(cookie.Value, &refreshToken); err != nil {
			api.ErrorResponse(w, r, err)
			return
		}

		next.ServeHTTP(w, r.WithContext(
			api.NewContextBuilder(r.Context()).
				AttachRefreshToken(refreshToken).
				Context(),
		))
	})
}

func (a apiAdmin) loadRefreshToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		refreshToken, ok := api.CtxRefreshToken(r)
		if !ok {
			api.ErrorResponse(w, r, auth.ErrMustAuthenticate)
			return
		}
		ok, err := a.RefreshTokenStore.Check(r.Context(), refreshToken.SessionID)
		if err != nil {
			api.ErrorResponse(w, r, common.WrapErr(err, common.ErrCodeServer))
			return
		}
		if !ok {
			api.ErrorResponse(w, r, auth.ErrInvalidSession)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a apiAdmin) attachUserToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(auth.CookieUserToken)
		if err != nil {
			if err == http.ErrNoCookie {
				next.ServeHTTP(w, r)
			} else {
				api.ErrorResponse(w, r, auth.ErrMalformedCookie)
			}
			return
		}

		if cookie.Value == "" {
			api.ErrorResponse(w, r, auth.ErrMustAuthenticate)
			return
		}

		var user auth.User
		if err := a.AuthService.ParseToken(cookie.Value, &user); err != nil {
			api.ErrorResponse(w, r, err)
			return
		}

		next.ServeHTTP(w, r.WithContext(
			api.NewContextBuilder(r.Context()).
				AttachUserToken(user).
				Context(),
		))
	})
}

func (a apiAdmin) loadUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prevUser, ok := api.CtxUser(r)
		if !ok {
			api.ErrorResponse(w, r, auth.ErrMustAuthenticate)
			return
		}

		user, err := a.UserStore.FindByID(r.Context(), prevUser.ID)
		if err != nil {
			api.ErrorResponse(w, r, err)
			return
		}

		accessToken := api.MustHaveAccessToken(r)

		var activeSession bool
		for _, sessionID := range user.Sessions {
			if sessionID == accessToken.SessionID {
				activeSession = true
			}
		}

		if !activeSession {
			api.ErrorResponse(w, r, auth.ErrInvalidSession)
			return
		}

		userToken, err := a.AuthService.SignToken(&user)
		if err != nil {
			api.ErrorResponse(w, r, common.WrapErr(fmt.Errorf("failed to sign user token: %w", err), common.ErrCodeServer))
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     auth.CookieUserToken,
			Value:    userToken,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Secure:   a.config.Server.SSLEnabled,
			Path:     "/",
			Expires:  time.Now().Add(tenYears),
		})

		next.ServeHTTP(w, r.WithContext(
			api.NewContextBuilder(r.Context()).
				AttachUserToken(user).
				Context(),
		))
	})
}
