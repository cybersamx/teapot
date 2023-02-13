package store

import (
	"context"
	"errors"

	"github.com/cybersamx/teapot/model"
)

const (
	DefaultPageSize = 50
)

var (
	ErrInternal = errors.New("internal error")
	ErrNoRows   = errors.New("empty results")
)

type Filter struct {
	Cursor   string
	PageSize int
}

type Store interface {
	Clear(ctx context.Context) error
	Close() error
	Config() *model.Config
	Connect(ctx context.Context, cfg *model.Config) error
	InitDB(ctx context.Context) error
	PingContext(ctx context.Context) error
}
