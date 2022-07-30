package common

import (
	"errors"
	"fmt"
)

const (
	ErrCodeUnknownError ErrCode = ""

	ErrCodeBadRequest ErrCode = "bad_request"
	ErrCodeNotFound   ErrCode = "not_found"

	ErrCodeInvalidAuth      ErrCode = "invalid_auth"
	ErrCodeInsufficientAuth ErrCode = "insufficient_auth"

	ErrCodeServer            ErrCode = "server"
	ErrCodeServerUnavailable ErrCode = "server_unavailable"
)

type err struct {
	cause error
	code  ErrCode
	data  map[string]interface{}
}

func NewErr(errMsg string, details ...ErrDetail) error {
	return WrapErr(errors.New(errMsg), details...)
}

func WrapErr(cause error, details ...ErrDetail) error {
	e := err{cause: cause}
	for _, detail := range details {
		detail.applyTo(&e)
	}
	return e
}

func (e err) Error() string {
	return e.cause.Error()
}

func (e err) Code() ErrCode {
	return e.code
}

func (e err) Data() map[string]interface{} {
	return e.data
}

type ErrDetail interface {
	applyTo(err *err)
}

type ErrCodeProvider interface {
	Code() ErrCode
}

type ErrDataProvider interface {
	Data() map[string]interface{}
}

type ErrCode string

func (code ErrCode) applyTo(err *err) {
	err.code = code
}

type ErrData map[string]interface{}

func (data ErrData) applyTo(err *err) {
	if err.data == nil {
		err.data = map[string]interface{}{}
	}
	for k, v := range data {
		err.data[k] = v
	}
}

type ErrDatum struct {
	Key   string
	Value interface{}
}

func (datum ErrDatum) applyTo(err *err) {
	if err.data == nil {
		err.data = map[string]interface{}{}
	}
	err.data[datum.Key] = datum.Value
}

type ErrResponse struct {
	Message   string                 `json:"msg"`
	Code      ErrCode                `json:"code"`
	Data      map[string]interface{} `json:"data,omitempty"`
	RequestID string                 `json:"request_id"`
}

func (err ErrResponse) Error() string {
	return fmt.Sprintf("%s: %s", err.Code, err.Message)
}
