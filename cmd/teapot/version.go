package main

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	build     = "dev"
	buildDate = "dev"
)

func appVersionString() string {
	return fmt.Sprintf("Version: %s, build: %s, build-date: %s, os/arch: %s/%s, go version: %s\n",
		version,
		build,
		buildDate,
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Version(),
	)
}

func appVersionMap() map[string]any {
	return map[string]any{
		"version":    version,
		"build":      build,
		"build-date": buildDate,
		"os-arch":    fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		"go-version": runtime.Version(),
	}
}

func versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show the app version",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Printf("%s %s\n", appname, appVersionString())

			return nil
		},
	}
}
