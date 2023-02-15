package main

import (
	"bytes"
	"testing"
	"time"

	"github.com/cybersamx/teapot/app"
	"github.com/cybersamx/teapot/model"
	"github.com/kylelemons/godebug/pretty"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	wantCfg = model.Config{
		Mode:      "debug",
		LogLevel:  "warning",
		LogFormat: "json",
		HTTP: model.HTTPConfig{
			Address:          ":7000",
			SiteURL:          "https://example.com",
			EnableProfiler:   true,
			EnablePrometheus: true,
		},
		Store: model.StoreConfig{
			Driver: "pgx",
			DSN:    "host=localhost port=5433 user=pguser password=password dbname=db_test sslmode=disable",
			TLS: model.TLS{
				CA:   "ca",
				Cert: "cert",
				Key:  "key",
			},
			MaxOpenConns:    3,
			MaxIdleConns:    4,
			ConnMaxLifetime: 6*time.Minute + 4*time.Second,
			Params: map[string]string{
				"charset": "utf8",
				"loc":     "UTC",
			},
		},
	}
)

func TestBindings_Defaults(t *testing.T) {
	bindCfg := model.NewConfig()
	dupeCfg := *bindCfg

	v := app.NewViper()

	cmd := cobra.Command{
		Use: "authx-test",
	}

	bindings := append(rootBindings(bindCfg), serveBindings(bindCfg)...)

	err := app.BindFlagsToCommand(v, cmd.Flags(), bindings)
	require.NoError(t, err)

	err = cmd.Execute()
	require.NoError(t, err)

	diff := pretty.Compare(bindCfg, dupeCfg)
	assert.Empty(t, diff)
}

func TestBindings_YAMLReader(t *testing.T) {
	rawCfg := []byte(`
mode: debug
log-level: warning
log-format: json
http:
  address: :7000
  site-url: https://example.com
  enable-prometheus: true
  enable-profiler: true
store:
  driver: pgx
  dsn: 'host=localhost port=5433 user=pguser password=password dbname=db_test sslmode=disable'
  user: pguser
  password: password
  host: localhost
  port: 5432
  database: db
  tls:
    ca: ca
    cert: cert
    key: key
  max-open-conns: 3
  max-idle-conns: 4
  conn-max-lifetime: 6m4s
  params:
    charset: utf8
    loc: UTC
oauth2:
  authx:
    client-id: MT2a4b6Ivz
    client-secret: scix5FAv5R
  google:
    client-id: rLmiYczUjd
    client-secret: hXMfaWumTC
`)

	var bindConfig model.Config

	v := app.NewViper()
	v.SetConfigType("yaml")
	err := v.ReadConfig(bytes.NewBuffer(rawCfg))
	require.NoError(t, err)

	cmd := cobra.Command{
		Use: "authx-test",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	bindings := append(rootBindings(&bindConfig), serveBindings(&bindConfig)...)

	err = app.BindFlagsToCommand(v, cmd.Flags(), bindings)
	require.NoError(t, err)

	err = cmd.Execute()
	require.NoError(t, err)

	diff := pretty.Compare(wantCfg, bindConfig)
	assert.Empty(t, diff)
}

func TestBindings_Envs(t *testing.T) {
	t.Setenv("AX_DEBUG", "debug")
	t.Setenv("AX_LOG_LEVEL", "warning")
	t.Setenv("AX_LOG_FORMAT", "json")
	t.Setenv("AX_HTTP_ADDRESS", ":7000")
	t.Setenv("AX_HTTP_SITE_URL", "https://example.com")
	t.Setenv("AX_HTTP_ENABLE_PROMETHEUS", "true")
	t.Setenv("AX_HTTP_ENABLE_PROFILER", "true")
	t.Setenv("AX_STORE_DRIVER", "pgx")
	t.Setenv("AX_STORE_DSN", "host=localhost port=5433 user=pguser password=password dbname=db_test sslmode=disable")
	t.Setenv("AX_STORE_USER", "pguser")
	t.Setenv("AX_STORE_PASSWORD", "password")
	t.Setenv("AX_STORE_HOST", "localhost")
	t.Setenv("AX_STORE_PORT", "5432")
	t.Setenv("AX_STORE_DATABASE", "db")
	t.Setenv("AX_STORE_TLS_CA", "ca")
	t.Setenv("AX_STORE_TLS_CERT", "cert")
	t.Setenv("AX_STORE_TLS_KEY", "key")
	t.Setenv("AX_STORE_MAX_OPEN_CONNS", "3")
	t.Setenv("AX_STORE_MAX_IDLE_CONNS", "4")
	t.Setenv("AX_STORE_CONN_MAX_LIFETIME", "6m4s")
	t.Setenv("AX_STORE_PARAMS", `{"charset":"utf8","loc":"UTC"}`)
	t.Setenv("AX_OAUTH2_AUTHX_CLIENT_ID", "MT2a4b6Ivz")
	t.Setenv("AX_OAUTH2_AUTHX_CLIENT_SECRET", "scix5FAv5R")
	t.Setenv("AX_OAUTH2_GOOGLE_CLIENT_ID", "rLmiYczUjd")
	t.Setenv("AX_OAUTH2_GOOGLE_CLIENT_SECRET", "hXMfaWumTC")

	var bindCfg model.Config

	v := app.NewViper()

	cmd := cobra.Command{
		Use: "authx-test",
	}

	bindings := append(rootBindings(&bindCfg), serveBindings(&bindCfg)...)

	err := app.BindFlagsToCommand(v, cmd.Flags(), bindings)
	require.NoError(t, err)

	err = cmd.Execute()
	require.NoError(t, err)

	diff := pretty.Compare(bindCfg, wantCfg)
	assert.Empty(t, diff)
}

func TestBindings_Flags(t *testing.T) {
	flags := []string{
		"--mode=debug",
		"--log-level=warning",
		"--log-format=json",
		"--http.address=:7000",
		"--http.site-url=https://example.com",
		"--http.enable-prometheus",
		"--http.enable-profiler",
		"--store.driver=pgx",
		"--store.dsn=host=localhost port=5433 user=pguser password=password dbname=db_test sslmode=disable",
		"--store.tls.ca=ca",
		"--store.tls.cert=cert",
		"--store.tls.key=key",
		"--store.max-open-conns=3",
		"--store.max-idle-conns=4",
		"--store.conn-max-lifetime=6m4s",
		"--store.params",
		"charset=utf8",
		"--store.params",
		"loc=UTC",
	}

	var bindCfg model.Config

	v := app.NewViper()

	cmd := cobra.Command{
		Use: "authx-test",
	}
	cmd.SetArgs(flags)

	bindings := append(rootBindings(&bindCfg), serveBindings(&bindCfg)...)

	err := app.BindFlagsToCommand(v, cmd.Flags(), bindings)
	require.NoError(t, err)

	err = cmd.Execute()
	require.NoError(t, err)

	diff := pretty.Compare(bindCfg, wantCfg)
	assert.Empty(t, diff)
}
