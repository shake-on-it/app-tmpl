package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/shake-on-it/app-tmpl/backend/api/middleware"
	"github.com/shake-on-it/app-tmpl/backend/common"
	"github.com/shake-on-it/app-tmpl/backend/core/mongodb"

	"github.com/gorilla/mux"
)

const (
	pathAPI = "/api"

	pathAdmin   = "/admin"
	pathPrivate = "/private"
)

type Service struct {
	config  common.Config
	crypter common.Crypter
	logger  common.Logger

	mongoProvider mongodb.Provider

	httpServer *http.Server
	wg         sync.WaitGroup

	AdminAPI   *apiAdmin
	PrivateAPI *apiPrivate
}

func NewService(config common.Config, crypter common.Crypter, logger common.Logger) Service {
	return Service{
		config:  config,
		crypter: crypter,
		logger:  logger,
	}
}

func (s *Service) Setup(ctx context.Context) error {
	if err := s.buildServers(ctx); err != nil {
		return err
	}

	r := mux.NewRouter()
	s.configureRouter(r)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Server.Port),
		Handler: r,

		ReadTimeout:    60 * time.Second,
		WriteTimeout:   s.config.API.RequestTimeout() + (5 * time.Second),
		MaxHeaderBytes: 4_000_000,

		// ConnState:      server.listener.ConnStateTrackingHandler,
		// ErrorLog: log.New(errLogWriter{}, "", 0),
	}
	return nil
}

func (s *Service) Start() {
	s.wg.Add(1)
	defer s.wg.Done()

	s.logger.Infof("listening on %s", s.config.Server.BaseURL)

	err := s.httpServer.ListenAndServe()
	if err != http.ErrServerClosed {
		s.logger.Error("failed to listen and serve http requests: %s", err)
	}
}

func (s *Service) Shutdown(ctx context.Context) error {
	doneCh := make(chan struct{})
	go func() {
		close(doneCh)
		s.wg.Wait()
	}()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}
	s.logger.Info("no longer accepting incoming requests")

	s.mongoProvider.Close(ctx)
	s.logger.Info("disconnected from mongodb")

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-doneCh:
		return nil
	}
}

func (s *Service) buildServers(ctx context.Context) error {
	s.mongoProvider = mongodb.NewProvider(s.config.DB.URI, s.logger)
	if err := s.mongoProvider.Setup(ctx); err != nil {
		return err
	}

	s.AdminAPI = &apiAdmin{
		config:          s.config,
		mongoProvider: s.mongoProvider,
		logger:          s.logger,
	}
	if err := s.AdminAPI.setup(ctx); err != nil {
		return err
	}

	s.PrivateAPI = &apiPrivate{
		adminAPI: s.AdminAPI,
		logger:   s.logger,
	}
	if err := s.PrivateAPI.setup(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Service) configureRouter(router *mux.Router) {
	r := router.PathPrefix(pathAPI).Subrouter()

	r.Use(middleware.RequestLimiter(s.config.API.RequestLimit))
	r.Use(middleware.RequestTimeouter(s.config.API.RequestTimeout()))
	r.Use(middleware.RequestLogger(s.logger))
	r.Use(middleware.RequestPanicCatcher)
	r.Use(middleware.CORS(s.config.API.CORSOrigins))

	s.AdminAPI.ApplyRoutes(r.PathPrefix(pathAdmin).Subrouter())
	s.PrivateAPI.ApplyRoutes(r.PathPrefix(pathPrivate).Subrouter())

	if s.config.Env == common.EnvTest {
		return
	}
	var routeTree strings.Builder
	routeTree.WriteString("registering routes...")

	var routeCount int
	if err := r.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		if route.GetHandler() == nil {
			return nil
		}
		routeCount++
		path, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		methods, err := route.GetMethods()
		if err != nil {
			if err.Error() != "mux: route doesn't have methods" {
				return err
			}
			methods = []string{""}
		}
		routeTree.WriteString(routeString(path, methods))
		return nil
	}); err != nil {
		s.logger.Warnf("failed to walk the route tree: %s", err)
	} else {
		routeTree.WriteString(fmt.Sprintf("\n...registered %d routes", routeCount))
		s.logger.Debug(routeTree.String())
	}
}

func routeString(path string, methods []string) string {
	paths := make([]string, 0, len(methods))
	for _, method := range methods {
		if method == http.MethodOptions {
			continue
		}
		paths = append(paths, fmt.Sprintf("%-8s %s", method, path))
	}
	return "\n"+strings.Join(paths, "\n")
}

// type errLogWriter struct{}
//
// func (w errLogWriter) Write(data []byte) (int, error) {
// 	common.WritePanic(string(data), nil)
// 	return len(data), nil
// }
