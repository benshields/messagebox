package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/benshields/messagebox/internal/pkg/config"
	"go.uber.org/zap"
)

type serverErrorContextKey struct{}

type StartError struct {
	Err error
}

func (e StartError) Error() string {
	return "APIServer.Start() failed unexpectedly: " + e.Err.Error()
}

type ShutdownError struct {
	Err error
}

func (e ShutdownError) Error() string {
	return "APIServer.Shutdown() failed unexpectedly: " + e.Err.Error()
}

type APIServer struct {
	httpServer *http.Server
	log        *zap.Logger
}

func Setup(cfg config.ServerConfiguration, log *zap.Logger) (*APIServer, error) {
	sugar := log.Sugar()
	defer sugar.Sync()
	sugar.Debugw("db.Setup", "config", cfg)

	srv := &APIServer{
		httpServer: &http.Server{},
		log:        log,
	}
	srv.httpServer.Addr = fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	m := http.NewServeMux()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "URL path: %s\n", r.URL.Path)
	})
	srv.httpServer.Handler = m

	return srv, nil
}

func (srv *APIServer) Start(ctx context.Context, errCh chan<- error) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				errCh <- nil
				return
			case errCh <- srv.httpServer.ListenAndServe():
				return
			}
		}
	}()
}

func (srv *APIServer) Shutdown(ctx context.Context) error {
	err := srv.httpServer.Shutdown(ctx)
	if err != nil && err != http.ErrServerClosed {
		return ShutdownError{err}
	}
	return nil
}
