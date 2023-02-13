package model

import (
	"time"
)

type FlagBindFunc func(target, val any)

// FlagBinding contains information of a flag associated on the following:
// 1. The name and description of the flag that will appear on the cli.
// 2. The default value of the flag if one isn't provided by the user.
// 3. Where to bind the flag eg. a field in a configuration object.
type FlagBinding struct {
	Usage     string
	Flag      string
	Shorthand rune // One character for a shorthand.
	Target    any
	Default   any
}

type Config struct {
	FilePath  string
	Mode      string
	LogLevel  string
	LogFormat string
	HTTP      HTTPConfig
	Store     StoreConfig
}

type HTTPConfig struct {
	Address          string
	SiteURL          string
	EnableProfiler   bool
	EnablePrometheus bool
	AllowedOrigins   []string
}

type StoreConfig struct {
	Driver string
	DSN    string

	TLS TLS

	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration

	Params map[string]string
}

type TLS struct {
	CA   string
	Cert string
	Key  string
}

// NewConfig creates a new Config object with default values.
func NewConfig() *Config {
	return &Config{
		Mode:      "debug",
		LogLevel:  "info",
		LogFormat: "text",
		HTTP: HTTPConfig{
			Address: ":9000",
			SiteURL: "http://localhost:9000",
		},
		Store: StoreConfig{
			Driver: "sqlite3",
			DSN:    "db.sqlite",
			TLS:    TLS{},
			Params: make(map[string]string),
		},
	}
}

func (t TLS) IsValid() bool {
	return t.CA != "" && t.Cert != "" && t.Key != ""
}
