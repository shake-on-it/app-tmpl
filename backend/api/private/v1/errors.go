package v1

import (
	"errors"
	"net/http"

	"github.com/shake-on-it/app-tmpl/backend/api"
	"github.com/shake-on-it/app-tmpl/backend/common"
)

func GetJSONBasicError(w http.ResponseWriter, r *http.Request) {
	api.ErrorResponse(w, r, errors.New("something bad happened"))
}

func GetJSONCompleteError(w http.ResponseWriter, r *http.Request) {
	api.ErrorResponse(w, r, common.NewErr(
		"something bad happened",
		common.ErrCodeServer,
		common.ErrDatum{"egg", "corn"},
		common.ErrData{
			"a": "ayy",
			"b": true,
			"c": 622,
		},
	))
}

func GetPayloadError(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(api.HeaderContentType, api.ContentTypeJSON)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{"err": "a different error response"}`))
}

func GetTextError(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(api.HeaderContentType, api.ContentTypeText)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("something bad happened"))
}
