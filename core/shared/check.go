package shared

import (
	"context"
	"io"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func Check(ctx context.Context, err error) {
	if err != nil {
		FancyPanic(ctx, err)
	}
}

func FancyPanic(ctx context.Context, err error) {
	zerolog.Ctx(ctx).Panic().Stack().Err(err).Msg("check failed!")
}

func CheckOfLog(run func(ctx context.Context) error) func(ctx context.Context) {
	return func(ctx context.Context) {
		err := run(ctx)
		if err != nil {
			err = errors.WithStack(err)
			zerolog.Ctx(ctx).Error().Stack().Err(err).Str("status", "failed.to.run.func").Send()
		}
	}
}

func CloseOrLog(ctx context.Context, closer io.Closer) {
	err := closer.Close()
	if err != nil {
		zerolog.Ctx(ctx).
			Error().
			Stack().
			Err(errors.WithStack(err)).
			Str("status", "failed.to.close").
			Send()
	}
}
