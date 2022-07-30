package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/shake-on-it/app-tmpl/backend/common"
)

func errorStatus(code common.ErrCode) int {
	switch code {

	// 400
	case common.ErrCodeBadRequest:
		return http.StatusBadRequest

	// 401
	case common.ErrCodeInvalidAuth:
		return http.StatusUnauthorized

		// 403
	case common.ErrCodeInsufficientAuth:
		return http.StatusForbidden

		// 403
	case common.ErrCodeNotFound:
		return http.StatusNotFound

		// 500
	case common.ErrCodeServer, common.ErrCodeUnknownError:
		return http.StatusInternalServerError

	// 503
	case common.ErrCodeServerUnavailable:
		return http.StatusServiceUnavailable
	}

	// fallback
	return http.StatusInternalServerError
}

func ErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	logger, _ := CtxLogger(r)
	requestID, _ := CtxRequestID(r)

	if logger != nil {
		logger.Error(fmt.Sprintf("request failed: %s", err))
	}

	body := common.ErrResponse{Message: err.Error(), RequestID: requestID}
	if e, ok := err.(common.ErrCodeProvider); ok {
		body.Code = e.Code()
	}
	if e, ok := err.(common.ErrDataProvider); ok {
		body.Data = e.Data()
	}

	JSONResponse(w, r, errorStatus(body.Code), body)
}

func JSONResponse(w http.ResponseWriter, r *http.Request, status int, body interface{}) {
	if body == nil {
		Response(w, r, status)
		return
	}
	logger, _ := CtxLogger(r)

	data, err := json.Marshal(body)
	if err != nil {
		requestID, _ := CtxRequestID(r)

		if logger != nil {
			logger.Warn(fmt.Sprintf("failed to marshal response body (%T): %s\n%v", body, err, body))
		}

		data = []byte(fmt.Sprintf(`{"msg":%q,"code":%q,"request_id":%q}`, err, common.ErrCodeUnknownError, requestID))
		status = errorStatus(common.ErrCodeUnknownError)
	}

	w.Header().Set(HeaderContentType, ContentTypeJSON)
	Response(w, r, status)
	if _, err := w.Write(data); err != nil {
		if logger != nil {
			logger.Warn(fmt.Sprintf("failed to write response body: %s\n%s", err, string(data)))
		}
	}
}

func Response(w http.ResponseWriter, r *http.Request, status int) {
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
}

const (
	HeaderAccept = "Accept"

	HeaderAuthorization = "Authorization"
	AuthorizationBearer = "Bearer "

	HeaderContentDisposition = "Content-Disposition"

	HeaderContentType = "Content-Type"
	ContentTypeJSON   = "application/json"
	ContentTypeText   = "text/plain"

	HeaderCredentials = "Credentials"

	HeaderLocation = "Location"

	HeaderRequestOrigin = "Request-Origin"

	HeaderXForwardedFor = "X-Forwarded-For"

	HeaderXAPP = "X-APP-"
)

func RequestAuthorization(r *http.Request) (string, error) {
	authorization := r.Header.Get(HeaderAuthorization)
	if authorization == "" {
		return "", common.NewErr("must provide auth", common.ErrCodeInvalidAuth)
	}
	if !strings.HasPrefix(authorization, AuthorizationBearer) {
		return "", common.NewErr("must provide valid auth", common.ErrCodeInvalidAuth)
	}
	return strings.TrimPrefix(authorization, AuthorizationBearer), nil
}

func RequestIPAddresses(r *http.Request) []string {
	ipAddresses := r.Header.Get(HeaderXForwardedFor)
	if ipAddresses == "" {
		return nil
	}
	return strings.Split(ipAddresses, ", ")
}

type HTTPResponseWriter struct {
	writer http.ResponseWriter
	logger common.Logger

	set      bool
	setStack []byte

	status   int
	statusMu sync.Mutex
}

func NewHTTPResponseWriter(w http.ResponseWriter, logger common.Logger) HTTPResponseWriter {
	return HTTPResponseWriter{writer: w, logger: logger}
}

func (w *HTTPResponseWriter) Status() (int, bool) {
	w.statusMu.Lock()
	defer w.statusMu.Unlock()
	return w.status, w.set
}

func (w *HTTPResponseWriter) Header() http.Header {
	w.statusMu.Lock()
	defer w.statusMu.Unlock()
	return w.writer.Header()
}

func (w *HTTPResponseWriter) Write(data []byte) (int, error) {
	w.statusMu.Lock()
	defer w.statusMu.Unlock()
	return w.writer.Write(data)
}

func (w *HTTPResponseWriter) WriteHeader(status int) {
	w.statusMu.Lock()
	defer w.statusMu.Unlock()

	called := w.set
	currentStatus := w.status

	if !called {
		w.set = true
		w.setStack = debug.Stack()

		w.status = status
		w.writer.WriteHeader(status)
		return
	}

	if status == http.StatusServiceUnavailable || currentStatus == http.StatusServiceUnavailable {
		return
	}

	w.logger.Error("write header called more than once",
		"status", status,
		"stack", string(debug.Stack()),
		"set_status", fmt.Sprint(currentStatus),
		"set_stack", string(w.setStack),
	)
}

func (w *HTTPResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.writer.(http.Hijacker)
	if !ok {
		panic("http response writer must be a hijacker")
	}
	return hijacker.Hijack()
}

func (w *HTTPResponseWriter) Flush() {
	flusher, ok := w.writer.(http.Flusher)
	if !ok {
		panic("http response writer must be a flusher")
	}
	flusher.Flush()
}
