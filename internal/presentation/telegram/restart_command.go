package telegram

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/goteleout/internal/service/resource"
	"os"
)

func (r *Presentation) restartCommandHandler(ctx *ext.Context, update *ext.Update, input *Input) error {
	err := r.replyIfNotSilentLocalized(ctx, update, input, resource.CommandRestartRestarting)
	if err != nil {
		return errors.WithStack(err)
	}

	zerolog.Ctx(ctx).Warn().Str("status", "reload.begin").Send()

	os.Exit(0)

	return nil
}
