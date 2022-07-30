package router

import (
	"net/http"

	"github.com/shake-on-it/app-tmpl/backend/api"
	"github.com/shake-on-it/app-tmpl/backend/api/admin/v1"
)

const (
	pathV1 = "/v1"

	pathUser        = "/user"
	pathUserSession = "/user/session"

	systemStatus  = "/system/status"
	systemVersion = "/system/version"
)

var (
	Versions = []string{pathV1}

	Registry = map[string][]api.RouteRegistration{
		pathV1: {
			// auth routes
			{
				v1.Whoami,
				api.RouteEndpoint{http.MethodGet, pathUser, false},
				api.RouteNeedsSession,
			},
			{
				v1.Register,
				api.RouteEndpoint{http.MethodPost, pathUser, false},
				api.RouteNeedsNothing,
			},
			{
				v1.Login,
				api.RouteEndpoint{http.MethodPost, pathUserSession, true},
				api.RouteNeedsNothing,
			},
			{
				v1.RefreshAccess,
				api.RouteEndpoint{http.MethodPut, pathUserSession, false},
				api.RouteNeedsRefreshToken,
			},
			{
				v1.Logout,
				api.RouteEndpoint{http.MethodDelete, pathUserSession, false},
				api.RouteNeedsNothing,
			},
		},
	}
)
