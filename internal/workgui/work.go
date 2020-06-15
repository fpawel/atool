package workgui

import (
	"context"
	"github.com/powerman/structlog"
)

type WorkFunc = func(*structlog.Logger, context.Context) error

type Work struct {
	Name string
	Func WorkFunc
}

type Works []Work

func (x Work) Perform(log *structlog.Logger, ctx context.Context) error {
	return Perform(log, ctx, x.Name, x.Func)
}

func (x Works) Run(log *structlog.Logger, ctx context.Context, name string) error {
	return RunWork(log, ctx, name, func(logger *structlog.Logger, ctx context.Context) error {
		for _, w := range x {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if err := w.Perform(log, ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

type WorkFuncList []WorkFunc

func (xs WorkFuncList) Do(log *structlog.Logger, ctx context.Context) error {
	for _, w := range xs {
		if err := w(log, ctx); err != nil {
			return err
		}
	}
	return nil
}
