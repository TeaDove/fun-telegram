package telegram

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/utils/logger_utils"
)

func (r *Presentation) injectContext(ctx *ext.Context, update *ext.Update) error {
	chatName := GetChatName(update.EffectiveChat())

	ctx.Context = logger_utils.AddLoggerToCtx(ctx.Context)
	ctx.Context = logger_utils.WithValue(ctx.Context, "chat_name", chatName)
	if update.EffectiveUser() != nil {
		ctx.Context = logger_utils.WithValue(
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
