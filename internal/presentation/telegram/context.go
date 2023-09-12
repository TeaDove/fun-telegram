package telegram

import (
	"github.com/anonyindian/gotgproto/ext"
)

func (r *Presentation) injectContext(ctx *ext.Context, update *ext.Update) error {
	// TODO fix after issue resolve: https://github.com/celestix/gotgproto/issues/31

	//ctx.Context = log.
	//	With().
	//	Int("update_effective_message_id", update.EffectiveMessage.ID).
	//	Logger().
	//	WithContext(ctx)
	//_, _ = ctx.Reply(update, update.EffectiveMessage.Message.Message, nil)
	return nil
}
