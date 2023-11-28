package telegram

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/rs/zerolog/log"
	tgUtils "github.com/teadove/goteleout/internal/presentation/telegram/utils"
)

func (r *Presentation) injectContext(ctx *ext.Context, update *ext.Update) error {
	// TODO fix after issue resolve: https://github.com/celestix/gotgproto/issues/31
	//  ctx.Context = log.With().
	//	Int("update_effective_message_id", update.EffectiveMessage.ID).
	//	Logger().
	//	WithContext(ctx)
	//  _, _ = ctx.Reply(update, update.EffectiveMessage.Message.Message, nil)
	return nil
}

func (r *Presentation) deleteOut(ctx *ext.Context, update *ext.Update) error {
	const silentArgument = "silent"

	args := tgUtils.GetArguments(update.EffectiveMessage.Message.Message)
	_, silent := args[silentArgument]

	if !silent {
		return nil
	}

	err := ctx.DeleteMessages(update.EffectiveChat().GetID(), []int{update.EffectiveMessage.ID})
	if err != nil {
		log.Warn().Str("status", "unable to delete message").Stack().Err(err).Send()
	}

	return nil
}
