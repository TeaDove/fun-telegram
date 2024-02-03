package telegram

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/peers/members"
	"github.com/pkg/errors"
	tgUtils "github.com/teadove/goteleout/internal/presentation/telegram/utils"
	"github.com/teadove/goteleout/internal/service/storage"
	"strconv"
	"strings"
)

func (r *Presentation) isEnabled(chatId int64) (bool, error) {
	_, err := r.storage.Load(strconv.Itoa(int(chatId)))
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			return true, nil
		} else {
			return false, errors.WithStack(err)
		}
	}

	return false, nil
}

func (r *Presentation) isBanned(username string) (bool, error) {
	_, err := r.storage.Load(compileBanKey(strings.ToLower(username)))
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			return false, nil
		} else {
			return false, errors.WithStack(err)
		}
	}

	return true, nil
}

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

func (r *Presentation) disableCommandHandler(ctx *ext.Context, update *ext.Update, input *tgUtils.Input) error {
	chatId := strconv.Itoa(int(update.EffectiveChat().GetID()))

	_, err := r.storage.Load(chatId)
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			err = r.storage.Save(chatId, []byte("1"))
			if err != nil {
				return errors.WithStack(err)
			}

			if !input.Silent {
				_, err = ctx.Reply(update, "Bot disabled in this chat", nil)
				if err != nil {
					return errors.WithStack(err)
				}
			}
		}

		return errors.WithStack(err)
	}

	err = r.storage.Delete(chatId)
	if err != nil {
		return errors.WithStack(err)
	}

	if !input.Silent {
		_, err = ctx.Reply(update, "Bot enabled in this chat", nil)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}
