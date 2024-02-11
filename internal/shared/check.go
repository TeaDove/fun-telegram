package shared

import (
	"context"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"io"
)

func Check(ctx context.Context, err error) {
	if err != nil {
		FancyPanic(ctx, err)
	}
}

func FancyPanic(ctx context.Context, err error) {
	zerolog.Ctx(ctx).Panic().Stack().Err(err).Msg("check failed!")
}

func CloseOrLog(ctx context.Context, closer io.Closer) {
	err := closer.Close()
	if err != nil {
		zerolog.Ctx(ctx).Error().Stack().Err(errors.WithStack(err)).Str("status", "failed.to.close").Send()
	}
}
