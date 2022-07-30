package core_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/shake-on-it/app-tmpl/backend/auth"
	"github.com/shake-on-it/app-tmpl/backend/common"
	"github.com/shake-on-it/app-tmpl/backend/common/test"
	"github.com/shake-on-it/app-tmpl/backend/common/test/assert"
	u "github.com/shake-on-it/app-tmpl/backend/common/test/utils"
	"github.com/shake-on-it/app-tmpl/backend/core/namespaces"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	errInvalidSession = common.ErrResponse{
		Code:    common.ErrCodeInvalidAuth,
		Message: "invalid session",
	}
	errMustAuthenticate = common.ErrResponse{
		Code:    common.ErrCodeInvalidAuth,
		Message: "must authenticate",
	}
)

func TestAuthServiceMultiSession(t *testing.T) {
	th := test.NewHarness(t)
	defer th.Close()

	var httpClient http.Client
	authCookies := map[string]map[string]*http.Cookie{}

	getWhoami := func(session string) error {
		req, err := http.NewRequest(
			http.MethodGet,
			th.Config.Server.BaseURL+"/api/admin/v1/user",
			nil,
		)
		if err != nil {
			return err
		}

		for _, cookie := range authCookies[session] {
			req.AddCookie(cookie)
		}

		res, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		if res.StatusCode == http.StatusOK {
			return nil
		}

		defer res.Body.Close()

		var errRes common.ErrResponse
		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return err
		}
		return errRes
	}

	doLogin := func(session string) error {
		body, err := json.Marshal(auth.Credentials{
			Username: "multi_user",
			Password: "p@sSw0rd",
		})
		if err != nil {
			return err
		}

		req, err := http.NewRequest(
			http.MethodPost,
			th.Config.Server.BaseURL+"/api/admin/v1/user/session",
			bytes.NewReader(body),
		)
		if err != nil {
			return err
		}

		res, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		if res.StatusCode != http.StatusCreated {
			return errors.New("failed to login")
		}

		if _, ok := authCookies[session]; !ok {
			authCookies[session] = map[string]*http.Cookie{}
		}
		for _, cookie := range res.Cookies() {
			authCookies[session][cookie.Name] = cookie
		}
		return nil
	}

	doRefresh := func(oldSession, newSession string) error {
		req, err := http.NewRequest(
			http.MethodPut,
			th.Config.Server.BaseURL+"/api/admin/v1/user/session",
			nil,
		)
		if err != nil {
			return err
		}

		for _, cookie := range authCookies[oldSession] {
			req.AddCookie(cookie)
		}

		res, err := httpClient.Do(req)
		if err != nil {
			return err
		}

		if res.StatusCode == http.StatusCreated {
			if _, ok := authCookies[newSession]; !ok {
				authCookies[newSession] = map[string]*http.Cookie{}
			}
			for _, cookie := range res.Cookies() {
				if cookie.Name == auth.CookieUserToken {
					fmt.Println(cookie.Value)
				}
				authCookies[newSession][cookie.Name] = cookie
			}
			return nil
		}

		defer res.Body.Close()

		var errRes common.ErrResponse
		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return err
		}
		return errRes
	}

	assert.Nil(t, th.CreateUser("multi_user"))

	t.Run("should fail to make requests for each session before logging in", func(t *testing.T) {
		assert.Equal(t, getWhoami("s1"), errMustAuthenticate)
		assert.Equal(t, getWhoami("s2"), errMustAuthenticate)
		assert.Equal(t, getWhoami("s3"), errMustAuthenticate)
	})

	t.Run("should be able to login multiple time", func(t *testing.T) {
		assert.Nil(t, doLogin("s1"))
		assert.Nil(t, doLogin("s2"))
		assert.Nil(t, doLogin("s3"))
	})

	t.Run("should have multiple user sessions for each login attempt", func(t *testing.T) {
		user, err := th.APIServer.AdminAPI.UserStore.FindByName(context.Background(), "multi_user")
		assert.Nil(t, err)
		assert.Equal(t, len(user.Sessions), 3)

		t.Run("and have a corresponding refresh token", func(t *testing.T) {
			mongoClient := u.MongoProvider().Client()

			cursor, err := mongoClient.
				Database(namespaces.DBAuth).
				Collection(namespaces.CollRefreshTokens).
				Find(context.Background(), bson.D{})
			assert.Nil(t, err)
			defer cursor.Close(context.Background())

			var refreshTokens []auth.RefreshToken
			for cursor.Next(context.Background()) {
				var refreshToken auth.RefreshToken
				assert.Nil(t, cursor.Decode(&refreshToken))
				refreshTokens = append(refreshTokens, refreshToken)
			}

			assert.Equal(t, len(refreshTokens), 3)
			for i, refreshToken := range refreshTokens {
				assert.Equal(t, user.Sessions[i], refreshToken.SessionID)
			}
		})
	})

	t.Run("should be able to make requests for each session", func(t *testing.T) {
		assert.Nil(t, getWhoami("s1"))
		assert.Nil(t, getWhoami("s2"))
		assert.Nil(t, getWhoami("s3"))
	})

	t.Run("should be able to refresh access for a session", func(t *testing.T) {
		assert.Nil(t, doRefresh("s2", "s2_new"))

		t.Run("and should be able to make requests for the new session", func(t *testing.T) {
			assert.Nil(t, getWhoami("s2_new"))
		})

		t.Run("and should fail to make requests for the previous session", func(t *testing.T) {
			err, ok := getWhoami("s2").(common.ErrResponse)
			assert.True(t, ok)
			assert.Equal(t, err, errInvalidSession)
		})

		t.Run("and still be able to make requests for the old sessions", func(t *testing.T) {
			assert.Nil(t, getWhoami("s1"))
			assert.Nil(t, getWhoami("s3"))
		})
	})

	t.Run("should be able to logout from one session", func(t *testing.T) {
		req, err := http.NewRequest(
			http.MethodDelete,
			th.Config.Server.BaseURL+"/api/admin/v1/user/session",
			nil,
		)
		assert.Nil(t, err)

		for _, cookie := range authCookies["s3"] {
			req.AddCookie(cookie)
		}

		res, err := httpClient.Do(req)
		assert.Nil(t, err)
		assert.Equal(t, res.StatusCode, http.StatusOK)

		for _, cookie := range res.Cookies() {
			authCookies["s3"][cookie.Name] = cookie
		}

		t.Run("and should invalidate all other sessions", func(t *testing.T) {
			assert.Equal(t, getWhoami("s1"), errInvalidSession)
			assert.Equal(t, getWhoami("s2_new"), errInvalidSession)
			assert.Equal(t, getWhoami("s3"), errMustAuthenticate)
		})
	})

	t.Run("should fail if user token signature is invalid", func(t *testing.T) {
		assert.Nil(t, doLogin("s1"))

		origUserToken := authCookies["s1"][auth.CookieUserToken].Value
		defer func() { authCookies["s1"][auth.CookieUserToken].Value = origUserToken }()
		authCookies["s1"][auth.CookieUserToken].Value = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
			"eyJpZCI6IjYyZmExMWM4MWE5N2E0YWRkNWIwOTMxMCIsIm5hbWUiOiJtdWx0aV91c2VyIiwiZW1haWwiOiJtdWx0aV91c2VyQGRvbWFpbi5jb20ifQ." +
			"xdOmkf5236MCVXiwqw-FGH8U8fDil63ROfb84rcC-qI"

		assert.Equal(t, getWhoami("s1"), common.ErrResponse{
			Code:    common.ErrCodeInvalidAuth,
			Message: "invalid signature",
		})
	})

	t.Run("should fail if user token is malformed", func(t *testing.T) {
		assert.Nil(t, doLogin("s1"))

		origUserToken := authCookies["s1"][auth.CookieUserToken].Value
		defer func() { authCookies["s1"][auth.CookieUserToken].Value = origUserToken }()
		authCookies["s1"][auth.CookieUserToken].Value = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
			"EyJpZCI6IjYyZmExMWM4MWE5N2E0YWRkNWIwOTMxMCIsIm5hbWUiOiJtdWx0aV91c2VyIiwiZW1haWwiOiJtdWx0aV91c2VyQGRvbWFpbi5jb20ifQ." +
			"xdOmkf5236MCVXiwqw-FGH8U8fDil63ROfb84rcC-qI"

		assert.Equal(t, getWhoami("s1"), common.ErrResponse{
			Code:    common.ErrCodeInvalidAuth,
			Message: "token is malformed",
		})
	})

	t.Run("should fail if user token is invalid", func(t *testing.T) {
		assert.Nil(t, doLogin("s1"))

		for _, tc := range []struct {
			errMsg string
			user   auth.User
		}{
			{
				errMsg: "must have name",
				user:   auth.User{Name: ""},
			},
			{
				errMsg: "must have email",
				user:   auth.User{Name: "name", Email: ""},
			},
		} {
			t.Run("because it "+tc.errMsg, func(t *testing.T) {
				invalidToken, err := th.APIServer.AdminAPI.AuthService.SignToken(&tc.user)
				assert.Nil(t, err)

				origUserToken := authCookies["s1"][auth.CookieUserToken].Value
				defer func() { authCookies["s1"][auth.CookieUserToken].Value = origUserToken }()
				authCookies["s1"][auth.CookieUserToken].Value = invalidToken

				assert.Equal(t, getWhoami("s1"), common.ErrResponse{
					Code:    common.ErrCodeInvalidAuth,
					Message: "invalid token: " + tc.errMsg,
				})
			})
		}
	})

	t.Run("should fail if refresh token signature is invalid", func(t *testing.T) {
		assert.Nil(t, doLogin("s1"))

		origRefreshToken := authCookies["s1"][auth.CookieRefreshToken].Value
		defer func() { authCookies["s1"][auth.CookieRefreshToken].Value = origRefreshToken }()
		authCookies["s1"][auth.CookieRefreshToken].Value = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
			"eyJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwODgiLCJzdWIiOiI2MmZhNGYwOGI0OTQ0MGQ4OGU0MzNkNDciLCJhdWQiOlsiYXBpL2FkbWluL3YxIl0sImV4cCI6MTY2MzE2MzQwMCwiaWF0IjoxNjYwNTcxNDAwLCJqdGkiOiI2MmZhNGYwOGI0OTQ0MGQ4OGU0MzNkNGUifQ." +
			"aivJKAzIZs83iOtWppMlyOLOB-fJjDWcfz3N0pqvhSU"

		assert.Equal(t, doRefresh("s1", "s1_new"), common.ErrResponse{
			Code:    common.ErrCodeInvalidAuth,
			Message: "invalid signature",
		})
	})

	t.Run("should fail if refresh token is malformed", func(t *testing.T) {
		assert.Nil(t, doLogin("s1"))

		origRefreshToken := authCookies["s1"][auth.CookieRefreshToken].Value
		defer func() { authCookies["s1"][auth.CookieRefreshToken].Value = origRefreshToken }()
		authCookies["s1"][auth.CookieRefreshToken].Value = "EyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
			"eyJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwODgiLCJzdWIiOiI2MmZhNGYwOGI0OTQ0MGQ4OGU0MzNkNDciLCJhdWQiOlsiYXBpL2FkbWluL3YxIl0sImV4cCI6MTY2MzE2MzQwMCwiaWF0IjoxNjYwNTcxNDAwLCJqdGkiOiI2MmZhNGYwOGI0OTQ0MGQ4OGU0MzNkNGUifQ." +
			"aivJKAzIZs83iOtWppMlyOLOB-fJjDWcfz3N0pqvhSU"

		assert.Equal(t, doRefresh("s1", "s1_new"), common.ErrResponse{
			Code:    common.ErrCodeInvalidAuth,
			Message: "token is malformed",
		})
	})

	t.Run("should fail if refresh token is invalid", func(t *testing.T) {
		assert.Nil(t, doLogin("s1"))

		for _, tc := range []struct {
			errMsg       string
			refreshToken auth.RefreshToken
		}{
			{
				errMsg: "token needs session",
				refreshToken: auth.RefreshToken{
					AccessToken: auth.AccessToken{
						SessionID: primitive.NilObjectID,
					},
				},
			},
			{
				errMsg: "token needs user",
				refreshToken: auth.RefreshToken{
					AccessToken: auth.AccessToken{
						SessionID: primitive.NewObjectID(),
						UserID:    primitive.NilObjectID,
					},
				},
			},
			{
				errMsg: "token is expired",
				refreshToken: auth.RefreshToken{
					AccessToken: auth.AccessToken{
						SessionID: primitive.NewObjectID(),
						UserID:    primitive.NewObjectID(),
						ExpiresAt: time.Now().Add(-1 * time.Minute),
					},
				},
			},
		} {
			t.Run("because it "+tc.errMsg, func(t *testing.T) {
				invalidToken, err := th.APIServer.AdminAPI.AuthService.SignToken(&tc.refreshToken)
				assert.Nil(t, err)

				origRefreshToken := authCookies["s1"][auth.CookieRefreshToken].Value
				defer func() { authCookies["s1"][auth.CookieRefreshToken].Value = origRefreshToken }()
				authCookies["s1"][auth.CookieRefreshToken].Value = invalidToken

				assert.Equal(t, doRefresh("s1", "s1_new"), common.ErrResponse{
					Code:    common.ErrCodeInvalidAuth,
					Message: "invalid token: " + tc.errMsg,
				})
			})
		}
	})

	t.Run("should fail if access token signature is invalid", func(t *testing.T) {
		assert.Nil(t, doLogin("s1"))

		origAccessToken := authCookies["s1"][auth.CookieAccessToken].Value
		defer func() { authCookies["s1"][auth.CookieAccessToken].Value = origAccessToken }()
		authCookies["s1"][auth.CookieAccessToken].Value = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
			"eyJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwODgiLCJzdWIiOiI2MmZhNTAyZmE1YTdiZDA0NmQ0NmNmZGIiLCJhdWQiOlsiYXBpL2FkbWluL3YxIl0sImV4cCI6MTY2MDU3MTk5NSwiaWF0IjoxNjYwNTcxNjk1LCJqdGkiOiI2MmZhNTAyZmE1YTdiZDA0NmQ0NmNmZTIifQ." +
			"VeHFRrTSAc2jldssX5866pkdLqFvxCr3ZZSE6UCoCv4"

		assert.Equal(t, getWhoami("s1"), common.ErrResponse{
			Code:    common.ErrCodeInvalidAuth,
			Message: "invalid signature",
		})
	})

	t.Run("should fail if access token is malformed", func(t *testing.T) {
		assert.Nil(t, doLogin("s1"))

		origAccessToken := authCookies["s1"][auth.CookieAccessToken].Value
		defer func() { authCookies["s1"][auth.CookieAccessToken].Value = origAccessToken }()
		authCookies["s1"][auth.CookieAccessToken].Value = "EyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
			"eyJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwODgiLCJzdWIiOiI2MmZhNTAyZmE1YTdiZDA0NmQ0NmNmZGIiLCJhdWQiOlsiYXBpL2FkbWluL3YxIl0sImV4cCI6MTY2MDU3MTk5NSwiaWF0IjoxNjYwNTcxNjk1LCJqdGkiOiI2MmZhNTAyZmE1YTdiZDA0NmQ0NmNmZTIifQ." +
			"VeHFRrTSAc2jldssX5866pkdLqFvxCr3ZZSE6UCoCv4"

		assert.Equal(t, getWhoami("s1"), common.ErrResponse{
			Code:    common.ErrCodeInvalidAuth,
			Message: "token is malformed",
		})
	})

	t.Run("should fail if access token is invalid", func(t *testing.T) {
		assert.Nil(t, doLogin("s1"))

		for _, tc := range []struct {
			errMsg      string
			accessToken auth.AccessToken
		}{
			{
				errMsg: "token needs session",
				accessToken: auth.AccessToken{
					SessionID: primitive.NilObjectID,
				},
			},
			{
				errMsg: "token needs user",
				accessToken: auth.AccessToken{
					SessionID: primitive.NewObjectID(),
					UserID:    primitive.NilObjectID,
				},
			},
			{
				errMsg: "token is expired",
				accessToken: auth.AccessToken{
					SessionID: primitive.NewObjectID(),
					UserID:    primitive.NewObjectID(),
					ExpiresAt: time.Now().Add(-1 * time.Minute),
				},
			},
		} {
			t.Run("because it "+tc.errMsg, func(t *testing.T) {
				invalidToken, err := th.APIServer.AdminAPI.AuthService.SignToken(&tc.accessToken)
				assert.Nil(t, err)

				origAccessToken := authCookies["s1"][auth.CookieAccessToken].Value
				defer func() { authCookies["s1"][auth.CookieAccessToken].Value = origAccessToken }()
				authCookies["s1"][auth.CookieAccessToken].Value = invalidToken

				assert.Equal(t, getWhoami("s1"), common.ErrResponse{
					Code:    common.ErrCodeInvalidAuth,
					Message: "invalid token: " + tc.errMsg,
				})
			})
		}
	})
}
