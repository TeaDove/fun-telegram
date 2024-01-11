package telegram

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	tgUtils "github.com/teadove/goteleout/internal/presentation/telegram/utils"
	"github.com/teadove/goteleout/internal/service/storage"
	"strconv"
)

func (r *Presentation) disableCommandHandler(ctx *ext.Context, update *ext.Update, input *tgUtils.Input) error {
	if update.EffectiveUser().GetID() != ctx.Self.ID {
		_, err := ctx.Reply(update, "Err: insufficient privilege", nil)
		if err != nil {
			return errors.WithStack(err)
		}

		zerolog.Ctx(ctx).Info().Str("status", "attempt.to.enable.bot.by.non.admin").Send()

		return nil
	}

	chatId := strconv.Itoa(int(update.EffectiveChat().GetID()))

	_, err := r.storage.Load(chatId)
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			err = r.storage.Save(chatId, []byte("1"))
			if err != nil {
				return errors.WithStack(err)
			}

			_, err = ctx.Reply(update, "Bot disabled in this chat", nil)
			if err != nil {
				return errors.WithStack(err)
			}
		}
		return errors.WithStack(err)
	}

	err = r.storage.Delete(chatId)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = ctx.Reply(update, "Bot enabled in this chat", nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
