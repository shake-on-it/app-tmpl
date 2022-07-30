package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/shake-on-it/app-tmpl/backend/api"
	"github.com/shake-on-it/app-tmpl/backend/api/admin"
	"github.com/shake-on-it/app-tmpl/backend/auth"
	"github.com/shake-on-it/app-tmpl/backend/common"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	tenYears = 10 * 365 * 24 * time.Hour
)

func Login(w http.ResponseWriter, r *http.Request) {
	srvCtx := admin.MustHaveServerContext(r)

	var creds auth.Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		api.ErrorResponse(w, r, common.NewErr("failed to parse credentials", common.ErrCodeBadRequest))
		return
	}

	user, tokens, err := srvCtx.AuthService.Login(r.Context(), creds)
	if err != nil {
		api.ErrorResponse(w, r, err)
		return
	}

	if err := authResponse(w, srvCtx, user, tokens); err != nil {
		api.ErrorResponse(w, r, err)
		return
	}
	api.Response(w, r, http.StatusCreated)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	srvCtx := admin.MustHaveServerContext(r)

	clearAuth(w, srvCtx)

	var userID primitive.ObjectID
	if user, ok := api.CtxUser(r); ok {
		userID = user.ID
	}
	if refreshToken, ok := api.CtxRefreshToken(r); ok {
		userID = refreshToken.UserID
	}
	if accessToken, ok := api.CtxAccessToken(r); ok {
		userID = accessToken.UserID
	}

	if userID.IsZero() {
		api.Response(w, r, http.StatusNoContent)
		return
	}

	if err := srvCtx.AuthService.Logout(r.Context(), userID); err != nil {
		api.ErrorResponse(w, r, common.WrapErr(fmt.Errorf("failed to logout: %s", err), common.ErrCodeServer))
		return
	}

	api.Response(w, r, 0)
}

func RefreshAccess(w http.ResponseWriter, r *http.Request) {
	refreshToken := api.MustHaveRefreshToken(r)
	srvCtx := admin.MustHaveServerContext(r)

	user, tokens, err := srvCtx.AuthService.RefreshAccess(r.Context(), refreshToken)
	if err != nil {
		clearAuth(w, srvCtx)
		api.ErrorResponse(w, r, err)
		return
	}

	if err := authResponse(w, srvCtx, user, tokens); err != nil {
		api.ErrorResponse(w, r, err)
		return
	}
	api.Response(w, r, http.StatusCreated)
}

func Register(w http.ResponseWriter, r *http.Request) {
	srvCtx := admin.MustHaveServerContext(r)

	var reg auth.Registration
	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		api.ErrorResponse(w, r, common.NewErr("failed to parse registration", common.ErrCodeBadRequest))
		return
	}

	user, err := srvCtx.AuthService.CreateUser(r.Context(), reg)
	if err != nil {
		api.ErrorResponse(w, r, err)
		return
	}

	api.JSONResponse(w, r, http.StatusCreated, user)
}

func Whoami(w http.ResponseWriter, r *http.Request) {
	user := api.MustHaveUser(r)
	api.JSONResponse(w, r, 0, user)
}

func authResponse(w http.ResponseWriter, srvCtx admin.ServerContext, user auth.User, tokens auth.Tokens) error {
	accessToken, err := srvCtx.AuthService.SignToken(&tokens.AccessToken)
	if err != nil {
		return common.WrapErr(fmt.Errorf("failed to sign access token: %w", err), common.ErrCodeServer)
	}

	refreshToken, err := srvCtx.AuthService.SignToken(&tokens.RefreshToken)
	if err != nil {

		return common.WrapErr(fmt.Errorf("failed to sign refresh token: %w", err), common.ErrCodeServer)
	}

	userToken, err := srvCtx.AuthService.SignToken(&user)
	if err != nil {
		return common.WrapErr(fmt.Errorf("failed to sign user token: %w", err), common.ErrCodeServer)
	}

	for _, cookie := range []struct {
		name      string
		token     string
		expiresAt time.Time
	}{
		{auth.CookieAccessToken, accessToken, tokens.AccessToken.ExpiresAt},
		{auth.CookieRefreshToken, refreshToken, tokens.RefreshToken.ExpiresAt},
		{auth.CookieUserToken, userToken, tokens.AccessToken.IssuedAt.Add(tenYears)},
	} {
		http.SetCookie(w, &http.Cookie{
			Name:     cookie.name,
			Value:    cookie.token,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Secure:   srvCtx.Config.Server.SSLEnabled,
			Path:     "/",
			Expires:  cookie.expiresAt,
		})
	}

	return nil
}

func clearAuth(w http.ResponseWriter, srvCtx admin.ServerContext) {
	for _, cookieName := range []string{
		auth.CookieUserToken,
		auth.CookieAccessToken,
		auth.CookieRefreshToken,
	} {
		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    "",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Secure:   srvCtx.Config.Server.SSLEnabled,
			Path:     "/",
			MaxAge:   -1,
		})
	}
}
