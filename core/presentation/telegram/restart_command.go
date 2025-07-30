package telegram

import (
	"os"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func (r *Presentation) restartCommandHandler(ctx *ext.Context, _ *ext.Update, _ *input) error {
	_, err := ctx.SendMessage(ctx.Self.ID, &tg.MessagesSendMessageRequest{Message: "Restarting..."})
	if err != nil {
		return errors.Wrap(err, "failed to send message")
	}

	zerolog.Ctx(ctx).Warn().Msg("reload.begin")

	os.Exit(0)

	return nil
}
