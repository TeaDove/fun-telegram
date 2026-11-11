package telegramapi

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/serviceutils/loggerutils"
)

func (r *Presentation) injectContext(ctx *ext.Context, update *ext.Update) error {
	chatName := GetChatName(update.EffectiveChat())

	ctx.Context = loggerutils.AddLoggerToCtx(ctx.Context)
	ctx.Context = loggerutils.WithValue(ctx.Context, "chat_name", chatName)

	if update.EffectiveUser() != nil {
		ctx.Context = loggerutils.WithValue(
			ctx.Context,
			"username",
			update.EffectiveUser().Username,
		)
	}

	return nil
}

func (r *Presentation) deleteOut(ctx *ext.Context, update *ext.Update) error {
	if update.EffectiveUser().GetID() != ctx.Self.ID {
		return nil
	}

	if !getOpt(update.EffectiveMessage.Message.Message).Silent {
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
