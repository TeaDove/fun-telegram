package telegram

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/pkg/errors"
	tgUtils "github.com/teadove/goteleout/internal/presentation/telegram/utils"
)

func (r *Presentation) replyIfNotSilent(ctx *ext.Context, update *ext.Update, input *tgUtils.Input, text any) error {
	if input.Silent {
		return nil
	}

	_, err := ctx.Reply(update, text, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
