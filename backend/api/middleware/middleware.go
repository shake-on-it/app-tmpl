package middleware

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/shake-on-it/app-tmpl/backend/api"
	"github.com/shake-on-it/app-tmpl/backend/common"

	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedHeaders: []string{
			api.HeaderAccept,
			api.HeaderAuthorization,
			api.HeaderContentType,
			api.HeaderCredentials,
			api.HeaderXAPP + api.HeaderRequestOrigin,
		},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodPatch,
			http.MethodHead,
		},
		AllowCredentials: true,
		ExposedHeaders: []string{
			api.HeaderXAPP + api.HeaderLocation,
			api.HeaderContentDisposition,
			api.HeaderLocation,
		},
	})
	return corsMiddleware.Handler
}

func RequestCacheBuster(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		next.ServeHTTP(w, r)
	})
}

func RequestLimiter(limit int) func(http.Handler) http.Handler {
	var currentRequests int64
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			total := atomic.AddInt64(&currentRequests, 1)
			defer atomic.AddInt64(&currentRequests, -1)

			if total > int64(limit) {
				api.ErrorResponse(w, r, common.NewErr("server is at capacity", common.ErrCodeServerUnavailable))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequestLogger(logger common.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := primitive.NewObjectID().Hex()

			requestIPAddresses := api.RequestIPAddresses(r)
			var loggerFieldIPAddress interface{}
			if len(requestIPAddresses) == 1 {
				loggerFieldIPAddress = requestIPAddresses[0]
			} else {
				loggerFieldIPAddress = requestIPAddresses
			}

			logger := logger.With(
				common.LoggerFieldRequestID, requestID,
				common.LoggerFieldHTTPMethod, r.Method,
				common.LoggerFieldPath, r.URL.Path,
			)

			writer := api.NewHTTPResponseWriter(w, logger)

			start := time.Now()
			logger.With(
				common.LoggerFieldIPAddress, loggerFieldIPAddress,
				common.LoggerFieldProto, r.Proto,
			).Info("request start")

			defer func() {
				duration := time.Since(start)

				status := http.StatusOK
				if s, ok := writer.Status(); ok {
					status = s
				}

				logger.With(
					common.LoggerFieldHTTPStatus, status,
					common.LoggerFieldDuration, duration.Milliseconds(),
				).Info("request complete")
			}()

			next.ServeHTTP(&writer, r.WithContext(
				api.NewContextBuilder(r.Context()).
					AttachLogger(logger).
					AttachRequestID(requestID).
					Context(),
			))
		})
	}
}

func RequestTimeouter(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequestPanicCatcher(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				if writer, ok := w.(*api.HTTPResponseWriter); ok {
					if _, ok := writer.Status(); !ok {
						writer.WriteHeader(http.StatusInternalServerError)
					}
				}
				logger, _ := api.CtxLogger(r)
				common.WritePanic(err, logger)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
