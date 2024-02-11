package telegram

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func (r *Presentation) injectContext(ctx *ext.Context, update *ext.Update) error {
	chatName := GetChatName(update.EffectiveChat())

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

	if !GetOpt(update.EffectiveMessage.Message.Message).Silent {
		return nil
	}

	err := ctx.DeleteMessages(update.EffectiveChat().GetID(), []int{update.EffectiveMessage.ID})
	if err != nil {
		zerolog.Ctx(ctx.Context).Warn().Str("status", "failed to delete message").Stack().Err(err).Send()
	}

	return nil
}
