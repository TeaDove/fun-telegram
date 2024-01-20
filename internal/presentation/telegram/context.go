package telegram

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	tgUtils "github.com/teadove/goteleout/internal/presentation/telegram/utils"
)

func (r *Presentation) injectContext(ctx *ext.Context, update *ext.Update) error {
	chatName := tgUtils.GetChatName(update.EffectiveChat())

	ctx.Context = log.
		With().
		Dict("tg", zerolog.Dict().
			Int("message_id", update.EffectiveMessage.ID).
			Str("chat_name", chatName).
			Str("effective_username", update.EffectiveUser().Username)).
		Ctx(ctx.Context).
		Logger().
		WithContext(ctx.Context)

	zerolog.Ctx(ctx.Context).Trace().Str("status", "got.update").Interface("update", update).Send()

	return nil
}

func (r *Presentation) deleteOut(ctx *ext.Context, update *ext.Update) error {
	if update.EffectiveUser().GetID() != ctx.Self.ID {
		return nil
	}

	const silentArgument = "silent"

	args := tgUtils.GetArguments(update.EffectiveMessage.Message.Message)
	_, silent := args[silentArgument]

	if !silent {
		return nil
	}

	err := ctx.DeleteMessages(update.EffectiveChat().GetID(), []int{update.EffectiveMessage.ID})
	if err != nil {
		zerolog.Ctx(ctx.Context).Warn().Str("status", "unable to delete message").Stack().Err(err).Send()
	}

	return nil
}
