package utils

import (
	"context"
	"io"

	"github.com/rs/zerolog"
)

func CloseOrLog(ctx context.Context, closer io.Closer) {
	err := closer.Close()
	if err != nil {
		zerolog.Ctx(ctx).Warn().Err(err).Msg("close.resource")
	}
}
