package test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/shake-on-it/app-tmpl/backend/api"
	"github.com/shake-on-it/app-tmpl/backend/api/server"
	"github.com/shake-on-it/app-tmpl/backend/auth"
	"github.com/shake-on-it/app-tmpl/backend/common"
	"github.com/shake-on-it/app-tmpl/backend/common/test/assert"
	u "github.com/shake-on-it/app-tmpl/backend/common/test/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	User     = "user"
	Password = "passw0rd!"
)

type Harness struct {
	APIServer *server.Service
	Config    common.Config

	onClose func()
	logger  common.Logger

	client *http.Client

	authUser    string
	authCookies map[string]map[string]*http.Cookie
	authUsers   map[string]auth.Credentials
}

type HarnessOptions struct {
	common.Config
	Crypter common.Crypter
}

func NewHarness(t *testing.T) Harness {
	t.Helper()
	return NewHarnessWithOptions(t, HarnessOptions{})
}

func NewHarnessWithOptions(t *testing.T, opts HarnessOptions) Harness {
	t.Helper()

	u.SkipUnlessMongoRunning(t)

	config, err := buildConfig(opts)
	assert.Nil(t, err)

	logger := u.NewLogger(t)

	apiServer := server.NewService(config, opts.Crypter, logger)

	ctx, cancel := context.WithCancel(context.Background())
	assert.Nil(t, apiServer.Setup(ctx))

	go apiServer.Start()

	return Harness{
		APIServer: &apiServer,
		Config:    config,

		onClose: cancel,
		logger:  logger,

		client: new(http.Client),

		authCookies: map[string]map[string]*http.Cookie{},
		authUsers:   map[string]auth.Credentials{},
	}
}

func (th *Harness) Close() {
	th.logger.Info("shutting down test harness")
	th.onClose()

	if err := th.APIServer.Shutdown(context.Background()); err != nil {
		th.logger.Errorf("failed to shutdown test harness server: %s", err)
	}
	th.logger.Info("test harness shutdown complete")
}

const (
	testUsername = "test-user"
	testPassword = "p@sSw0rd"
)

func (th *Harness) CreateUser(username string) error {
	creds := auth.Credentials{
		Username: username,
		Password: testPassword,
	}

	_, err := th.APIServer.AdminAPI.AuthService.CreateUser(
		context.Background(),
		auth.Registration{
			Credentials: creds,
			Email:       username + "@domain.com",
		},
	)
	if err != nil {
		return err
	}

	th.authUsers[username] = creds
	return nil
}

func (th *Harness) Login() error {
	_, err := th.APIServer.AdminAPI.UserStore.FindByName(context.Background(), testUsername)
	if err, ok := err.(common.ErrCoder); ok && err.Code() == common.ErrCodeNotFound {
		if err := th.CreateUser(testUsername); err != nil {
			return err
		}
	}
	return th.LoginAs(testUsername)
}

func (th *Harness) LoginAs(username string) error {
	res, err := th.Do(Request{
		Method: http.MethodPost,
		Path:   "/api/admin/v1/user/session",
		Body: auth.Credentials{
			Username: username,
			Password: th.authUsers[username].Password,
		},
	})
	if err != nil {
		return err
	}
	if err := res.Is(http.StatusCreated); err != nil {
		return err
	}

	if _, ok := th.authCookies[username]; !ok {
		th.authCookies[username] = map[string]*http.Cookie{}
	}
	for _, cookie := range res.data.Cookies() {
		th.authCookies[username][cookie.Name] = cookie
	}

	th.authUser = username
	return nil
}

func (th *Harness) SwitchUser(username string) error {
	if th.authUser == username {
		return nil
	}
	if cookies, ok := th.authCookies[username]; ok && len(cookies) > 0 {
		return nil
	}
	return th.LoginAs(username)
}

type Request struct {
	Method  string
	Path    string
	Body    interface{}
	Anon    bool
	Auth    bool
	Refresh bool
}

func (th *Harness) Do(opts Request) (Response, error) {
	method := opts.Method
	if method == "" {
		method = http.MethodGet
	}

	var body io.Reader
	switch b := opts.Body.(type) {
	case io.Reader:
		body = b
	case []byte:
		body = bytes.NewReader(b)
	case string:
		body = bytes.NewReader([]byte(b))
	case interface{}:
		data, err := json.Marshal(b)
		if err != nil {
			return Response{}, err
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, th.Config.Server.BaseURL+opts.Path, body)
	if err != nil {
		return Response{}, err
	}

	if !opts.Anon {
		if cookie, ok := th.authCookies[th.authUser][auth.CookieUserToken]; ok {
			req.AddCookie(cookie)
		}
	}
	if opts.Auth {
		if cookie, ok := th.authCookies[th.authUser][auth.CookieAccessToken]; ok {
			req.AddCookie(cookie)
		}
	}
	if opts.Refresh {
		if cookie, ok := th.authCookies[th.authUser][auth.CookieRefreshToken]; ok {
			req.AddCookie(cookie)
		}
	}

	res, err := th.client.Do(req)
	if err != nil {
		return Response{}, err
	}
	return Response{data: res}, nil
}

type Response struct {
	checked bool
	data    *http.Response
}

func (res *Response) Is(statusCode int) error {
	res.checked = true

	if statusCode == res.data.StatusCode {
		return nil
	}

	defer res.data.Body.Close()

	if statusCode >= 200 && statusCode < 299 {
		return fmt.Errorf("unexpected status: '%s' (expected %d)", res.data.Status, statusCode)
	}

	if res.data.Header.Get(api.HeaderContentType) != api.ContentTypeJSON {
		return fmt.Errorf("request failed: %s (expected %d)", res.data.Status, statusCode)
	}

	return res.Err()
}

func (res *Response) Decode(out interface{}) error {
	if !res.checked {
		return errors.New("must check response status code first")
	}

	defer res.data.Body.Close()
	return json.NewDecoder(res.data.Body).Decode(out)
}

func (res *Response) Err() error {
	if !res.checked {
		defer res.data.Body.Close()
	}
	var err common.ErrResponse
	if err := json.NewDecoder(res.data.Body).Decode(&err); err != nil {
		return fmt.Errorf("failed to parse error response (%s): %s", res.data.Status, err)
	}
	return err
}

func buildConfig(opts HarnessOptions) (common.Config, error) {
	config := common.Config{
		Env: common.EnvTest,
		API: common.APIConfig{
			CORSOrigins:        []string{},
			RequestLimit:       60,
			RequestTimeoutSecs: 55,
		},
		Auth: common.AuthConfig{
			JWTSecret: primitive.NewObjectID().Hex(),
		},
		DB: common.DBConfig{
			URI: u.MongoURI(),
		},
		Server: common.ServerConfig{
			Host: "localhost",
			// Host:       "127.0.0.1",
			Port:       8088,
			SSLEnabled: false,
		},
	}

	if opts.Auth.AccessTokenExpirySecs != 0 {
		config.Auth.AccessTokenExpirySecs = opts.Auth.AccessTokenExpirySecs
	}

	if err := config.Validate(); err != nil {
		return common.Config{}, err
	}
	return config, nil
}
