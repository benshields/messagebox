package api

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/benshields/messagebox/internal/pkg/config"
	"github.com/benshields/messagebox/internal/pkg/db"
	"github.com/benshields/messagebox/internal/pkg/logger"
	"github.com/benshields/messagebox/internal/pkg/router"
	"github.com/benshields/messagebox/internal/pkg/server"
)

func Start(configPath string) error {
	cfg, err := config.New(configPath)
	if err != nil {
		return err
	}

	log, err := logger.Setup(cfg.Logger)
	if err != nil {
		return err
	}

	_, err = db.Setup(cfg.Database, log)
	if err != nil {
		return err
	}

	r := router.Setup()

	srv, err := server.Setup(cfg.Server, log, r)
	if err != nil {
		return err
	}

	startErr := make(chan error)
	startCtx, startCancel := context.WithCancel(context.Background())
	defer startCancel()
	go func() {
		srv.Start(startCtx, startErr)
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	select {
	case err := <-startErr:
		if err != nil && err != http.ErrServerClosed {
			return server.StartError{Err: err}
		}
		return nil
	case <-quit:
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		return srv.Shutdown(shutdownCtx)
	}
}
