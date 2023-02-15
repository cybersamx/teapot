package common

import (
	"bytes"
	"context"

	"github.com/sirupsen/logrus"
)

type InsertStorer[T any] interface {
	Insert(ctx context.Context, item T) (T, error)
}

func NewTestLogger(buf *bytes.Buffer, exitFn func(code int)) *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetOutput(buf)
	logger.ExitFunc = exitFn

	return logger
}

func FillStore[T any](ctx context.Context, store InsertStorer[T], items []T) ([]T, error) {
	saves := make([]T, 0, 5)

	for _, item := range items {
		saved, err := store.Insert(ctx, item)
		if err != nil {
			return saves, err
		}

		saves = append(saves, saved)
	}

	return saves, nil
}
