package telegram

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/rs/zerolog"
)

func (r *Presentation) injectContext(ctx *ext.Context, update *ext.Update) error {
	chatName := GetChatName(update.EffectiveChat())

	tgDict := zerolog.Dict().Str("chat", chatName)
	if update.EffectiveUser() != nil {
		tgDict.Str("username", update.EffectiveUser().Username)
	}

	ctx.Context = zerolog.Ctx(ctx).
		With().
		Dict("tg", tgDict).
		Ctx(ctx.Context).
		Logger().
		WithContext(ctx.Context)

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
		zerolog.Ctx(ctx.Context).
			Warn().
			Str("status", "failed to delete message").
			Stack().
			Err(err).
			Send()
	}

	return nil
}
