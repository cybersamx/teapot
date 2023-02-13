package main

import (
	"github.com/cybersamx/teapot/model"
)

const (
	cfgFileName      = "config.yaml"
	cfgFileFlag      = "file"
	cfgFileShorthand = 'f'
)

func rootBindings(bindCfg *model.Config) []model.FlagBinding {
	// Make use of the default values returned from NewConfig().
	defaultCfg := model.NewConfig()

	bindings := []model.FlagBinding{
		{
			Usage:     "config file to load",
			Flag:      cfgFileFlag,
			Shorthand: cfgFileShorthand,
			Target:    &bindCfg.FilePath,
			Default:   defaultCfg.FilePath,
		},
		{
			Usage:     "run mode",
			Flag:      "mode",
			Shorthand: 'm',
			Target:    &bindCfg.Mode,
			Default:   defaultCfg.Mode,
		},
		{
			Usage:   "log level",
			Flag:    "log-level",
			Target:  &bindCfg.LogLevel,
			Default: defaultCfg.LogLevel,
		},
		{
			Usage:   "log format (text|json)",
			Flag:    "log-format",
			Target:  &bindCfg.LogFormat,
			Default: defaultCfg.LogFormat,
		},
	}

	return bindings
}

func serveBindings(bindCfg *model.Config) []model.FlagBinding {
	// Make use of the default values returned from NewConfig().
	defaultCfg := model.NewConfig()

	bindings := []model.FlagBinding{
		{
			Usage:   "http service address",
			Flag:    "http.address",
			Target:  &bindCfg.HTTP.Address,
			Default: defaultCfg.HTTP.Address,
		},
		{
			Usage:   "the service site url",
			Flag:    "http.site-url",
			Target:  &bindCfg.HTTP.SiteURL,
			Default: defaultCfg.HTTP.SiteURL,
		},
		{
			Usage:   "list of allowed domains for cors",
			Flag:    "http.allowed-origins",
			Target:  &bindCfg.HTTP.AllowedOrigins,
			Default: defaultCfg.HTTP.AllowedOrigins,
		},
		{
			Usage:   "enable profiler",
			Flag:    "http.enable-profiler",
			Target:  &bindCfg.HTTP.EnableProfiler,
			Default: defaultCfg.HTTP.EnableProfiler,
		},
		{
			Usage:   "enable prometheus",
			Flag:    "http.enable-prometheus",
			Target:  &bindCfg.HTTP.EnablePrometheus,
			Default: defaultCfg.HTTP.EnablePrometheus,
		},
		{
			Usage:   "sql driver name",
			Flag:    "store.driver",
			Target:  &bindCfg.Store.Driver,
			Default: defaultCfg.Store.Driver,
		},
		{
			Usage:   "data source name",
			Flag:    "store.dsn",
			Target:  &bindCfg.Store.DSN,
			Default: defaultCfg.Store.DSN,
		},
		{
			Usage:   "tls certificate authority (ca)",
			Flag:    "store.tls.ca",
			Target:  &bindCfg.Store.TLS.CA,
			Default: defaultCfg.Store.TLS.CA,
		},
		{
			Usage:   "tls certificate",
			Flag:    "store.tls.cert",
			Target:  &bindCfg.Store.TLS.Cert,
			Default: defaultCfg.Store.TLS.Cert,
		},
		{
			Usage:   "tls key",
			Flag:    "store.tls.key",
			Target:  &bindCfg.Store.TLS.Key,
			Default: defaultCfg.Store.TLS.Key,
		},
		{
			Usage:   "additional params to pass to the data store",
			Flag:    "store.params",
			Target:  &bindCfg.Store.Params,
			Default: defaultCfg.Store.Params,
		},
		{
			Usage:   "max open connections (maybe overridden depending on the driver)",
			Flag:    "store.max-open-conns",
			Target:  &bindCfg.Store.MaxOpenConns,
			Default: defaultCfg.Store.MaxOpenConns,
		},
		{
			Usage:   "max idle connections (maybe overridden depending on the driver)",
			Flag:    "store.max-idle-conns",
			Target:  &bindCfg.Store.MaxIdleConns,
			Default: defaultCfg.Store.MaxIdleConns,
		},
		{
			Usage:   "connection max lifetime (maybe overridden depending on the driver)",
			Flag:    "store.conn-max-lifetime",
			Target:  &bindCfg.Store.ConnMaxLifetime,
			Default: defaultCfg.Store.ConnMaxLifetime,
		},
	}

	return bindings
}
