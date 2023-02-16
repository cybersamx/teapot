package app

import (
	"context"
	"fmt"

	"github.com/cybersamx/teapot/httpx"
	"github.com/cybersamx/teapot/model"
	"github.com/cybersamx/teapot/store"
	"github.com/cybersamx/teapot/store/sqlstore"
	"github.com/sirupsen/logrus"
)

type App struct {
	cfg *model.Config

	datastore  store.Store
	httpServer *httpx.Server
	logger     *logrus.Logger

	doneChan <-chan struct{}
	cancel   context.CancelFunc
}

func New(cfg *model.Config, logger *logrus.Logger) (*App, error) {
	// Datastore setup.
	datastore := sqlstore.New(cfg.Store.Driver, logger)

	// HTTP server setup.
	srv := httpx.New(datastore, logger, cfg)

	// App setup.
	a := &App{
		cfg:        cfg,
		datastore:  datastore,
		logger:     logger,
		httpServer: srv,
	}

	return a, nil
}

func (a *App) Server() *httpx.Server {
	return a.httpServer
}

func (a *App) Done() <-chan struct{} {
	return a.doneChan
}

func (a *App) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	a.cancel = cancel
	a.doneChan = ctx.Done()

	// Connect to datastore.
	if err := a.datastore.Connect(ctx, a.cfg); err != nil {
		return fmt.Errorf("(*App).Start; %w", err)
	}

	if err := a.datastore.InitDB(ctx); err != nil {
		return fmt.Errorf("(*App).Start; %w", err)
	}

	// Start http server.
	a.httpServer.Start(ctx)

	return nil
}

func (a *App) Close(ctx context.Context) {
	a.logger.Infoln("Shutting down the app")
	a.httpServer.Close(ctx)
	a.datastore.Close()
}
