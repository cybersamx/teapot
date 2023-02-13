package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cybersamx/teapot/app"
	"github.com/cybersamx/teapot/common"
	"github.com/cybersamx/teapot/model"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	ErrConfigFileNotFound = errors.New("config file not found")
)

// getConfigPath returns the config file path if it's found in program args. If the arg is not defined and the config
// file doesn't exist, the function returns the default config file path if the file exists.
func getConfigPath(flags *pflag.FlagSet) string {
	var cfgPath string

	flags.ParseAll(os.Args, func(flag *pflag.Flag, value string) error {
		if flag.Name == cfgFileFlag {
			cfgPath = value
			return nil
		}
		return nil
	})

	if cfgPath != "" {
		if !filepath.IsAbs(cfgPath) {
			cfgPath = filepath.Join(common.WorkDir(), cfgPath)
		}

		if !common.IsFileExist(cfgPath) {
			// If the config file is defined in the args, then it must exist. Otherwise, raise an error.
			checkErr(fmt.Errorf("getConfigPath - %s: %w", cfgPath, ErrConfigFileNotFound))
		}

		return cfgPath
	}

	// Check default config, but it's ok if this file doesn't exist.
	cfgPath = filepath.Join(common.WorkDir(), cfgFileName)
	if !common.IsFileExist(cfgPath) {
		return ""
	}

	return cfgPath
}

func rootCommand() *cobra.Command {
	cfg := model.NewConfig()

	// Root command.
	cmd := cobra.Command{
		Use: appname,
		RunE: func(cmd *cobra.Command, args []string) error {
			// User must enter a command, otherwise display the help menu.
			return cmd.Help()
		},
	}

	//Configure and get flags.
	v := app.NewViper()
	flags := cmd.PersistentFlags()
	bindings := rootBindings(cfg)
	err := app.BindFlagsToCommand(v, flags, bindings)
	checkErr(err)

	// If a config file is specified in args, use it.
	cfgPath := getConfigPath(flags)
	if cfgPath != "" {
		reader, err := os.Open(cfgPath)
		checkErr(err)
		v.SetConfigType("yaml")
		err = v.ReadConfig(reader)
	}

	// CLI commands
	cmd.AddCommand(serveCommand(cfg, v))
	cmd.AddCommand(versionCommand())

	return &cmd
}
