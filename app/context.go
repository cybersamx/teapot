package app

import (
	"context"
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
)

const (
	CtxObjectKey = contextKey("appObject")
)

type contextKey string

type ContextObject struct {
	Logger *logrus.Logger
}

func NewContextWithSignals(ctx context.Context, signals ...os.Signal) context.Context {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)

	ctx, cancel := context.WithCancel(ctx)

	go func() {
		defer cancel()
		select {
		case <-ctx.Done():
			return
		case sig := <-ch:
			obj, ok := ctx.Value(CtxObjectKey).(ContextObject)
			if ok {
				obj.Logger.WithFields(logrus.Fields{"signal": sig}).Infoln("Received signal")
			}
			return
		}
	}()

	return ctx
}
