package router

import (
	"net/http"

	"github.com/shake-on-it/app-tmpl/backend/api"
	"github.com/shake-on-it/app-tmpl/backend/api/private/v1"
)

const (
	pathV1 = "/v1"

	pathHealth  = "/health"
	pathVersion = "/version"

	pathErrorsJSONBasic    = "/errors/json/basic"
	pathErrorsJSONComplete = "/errors/json/complete"
	pathErrorsPayload      = "/errors/payload"
	pathErrorsText         = "/errors/text"
)

var (
	Versions = []string{pathV1}

	Registry = map[string][]api.RouteRegistration{
		pathV1: {
			// system routes
			{
				v1.GetHealth,
				api.RouteEndpoint{http.MethodGet, pathHealth, false},
				api.RouteNeedsNothing,
			},
			{
				v1.GetVersion,
				api.RouteEndpoint{http.MethodGet, pathVersion, false},
				api.RouteNeedsNothing,
			},
			// error routes
			{
				v1.GetJSONBasicError,
				api.RouteEndpoint{http.MethodGet, pathErrorsJSONBasic, false},
				api.RouteNeedsNothing,
			},
			{
				v1.GetJSONCompleteError,
				api.RouteEndpoint{http.MethodGet, pathErrorsJSONComplete, false},
				api.RouteNeedsNothing,
			},
			{
				v1.GetPayloadError,
				api.RouteEndpoint{http.MethodGet, pathErrorsPayload, false},
				api.RouteNeedsNothing,
			},
			{
				v1.GetTextError,
				api.RouteEndpoint{http.MethodGet, pathErrorsText, false},
				api.RouteNeedsNothing,
			},
		},
	}
)
