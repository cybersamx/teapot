package main

import (
	"context"
	"fmt"
	"syscall"
	"time"

	apiv1 "github.com/cybersamx/teapot/api"
	"github.com/cybersamx/teapot/app"
	"github.com/cybersamx/teapot/model"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	apiPath         = "/api/v1"
	appCloseTimeout = 5 * time.Second
)

func serve(cfg *model.Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		logger := app.NewLogger(cfg.LogLevel, cfg.LogFormat)
		logger.WithFields(map[string]any{"mode": cfg.Mode}).Infof("Running %s", appname)
		logger.WithFields(appVersionMap()).Infoln("App configuration")

		a, err := app.New(cfg, logger)
		if err != nil {
			return err
		}

		// Bind api routes to the server and initialize any services.
		api := apiv1.New()
		api.BindServer(a.Server(), apiPath)

		// Start.
		ctx := context.WithValue(context.Background(), app.CtxObjectKey, app.ContextObject{Logger: logger})
		startCtx := app.NewContextWithSignals(ctx, syscall.SIGINT, syscall.SIGTERM)
		if err := a.Start(startCtx); err != nil {
			return err
		}

		// Wait for termination signals to shut down server.
		<-a.Done()

		// Shutdown.
		closeCtx, cancel := context.WithTimeout(ctx, appCloseTimeout)
		defer cancel()
		a.Close(closeCtx)

		return err
	}
}

func serveCommand(cfg *model.Config, v *viper.Viper) *cobra.Command {
	cmd := cobra.Command{
		Use:     "serve",
		Short:   fmt.Sprintf("Launch %s server", appname),
		Example: fmt.Sprintf("%s serve", appname),
		RunE:    serve(cfg),
	}

	// Flags() comprised of persistent (root) and local (command serve) flags. So we need to bind them to
	// root and serve bindings.
	flags := cmd.Flags()
	bindings := append(rootBindings(cfg), serveBindings(cfg)...)
	err := app.BindFlagsToCommand(v, flags, bindings)
	checkErr(err)

	return &cmd
}
