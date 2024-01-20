package telegram

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/peers/members"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/goteleout/internal/service/storage"
	"strconv"
)

func (r *Presentation) checkFromAdmin(ctx *ext.Context, update *ext.Update) (ok bool, err error) {
	chatMembers, err := r.getMembers(ctx, update.EffectiveChat())
	if err != nil {
		if errors.Is(err, ErrNotChatOrChannel) {
			// Expects, that in private conversation everyone is admin
			return true, nil
		}
		return false, errors.WithStack(err)
	}

	userMember, ok := chatMembers[update.EffectiveUser().GetID()]
	if !ok {
		return false, errors.New("user not found in members")
	}

	return userMember.Status() == members.Admin ||
		update.EffectiveUser().GetID() == ctx.Self.ID ||
		userMember.Status() == members.Creator, nil
}

func (r *Presentation) disableCommandHandler(ctx *ext.Context, update *ext.Update, input *Input) error {
	ok, err := r.checkFromAdmin(ctx, update)
	if err != nil {
		return errors.WithStack(err)
	}

	if !ok {
		_, err = ctx.Reply(update, "Err: insufficient privilege", nil)
		if err != nil {
			return errors.WithStack(err)
		}

		zerolog.Ctx(ctx).Info().Str("status", "attempt.to.enable.bot.by.non.admin").Send()
	}

	chatId := strconv.Itoa(int(update.EffectiveChat().GetID()))

	_, err = r.storage.Load(chatId)
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
