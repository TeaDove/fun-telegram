package telegram

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/pkg/errors"
	tgUtils "github.com/teadove/goteleout/internal/presentation/telegram/utils"
)

func (r *Presentation) echoCommandHandler(ctx *ext.Context, update *ext.Update, input *tgUtils.Input) error {
	_, err := ctx.Reply(update, update.EffectiveMessage.Message.Message, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
