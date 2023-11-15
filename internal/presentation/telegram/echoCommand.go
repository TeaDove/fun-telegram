package telegram

import (
	"github.com/anonyindian/gotgproto/ext"
	"github.com/pkg/errors"
)

func (r *Presentation) echoCommandHandler(ctx *ext.Context, update *ext.Update) error {
	_, err := ctx.Reply(update, update.EffectiveMessage.Message.Message, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
