package telegram

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/teadove/goteleout/internal/presentation/telegram/utils"
	"github.com/teadove/goteleout/internal/repository/db_repository"
	"strings"
	"unicode"
)

func (r *Presentation) catchMessages(ctx *ext.Context, update *ext.Update) error {
	ok, err := r.isEnabled(update.EffectiveMessage.GetID())
	if err != nil {
		return errors.WithStack(err)
	}
	if !ok {
		return nil
	}

	_, ok = update.UpdateClass.(*tg.UpdateNewChannelMessage)
	if !ok {
		_, ok = update.UpdateClass.(*tg.UpdateNewMessage)
		if !ok {
			return nil
		}
	}

	if channel, ok := update.EffectiveChat().(*types.Channel); ok {
		if channel.Broadcast {
			return nil
		}
	}

	text := update.EffectiveMessage.Text
	if text == "" {
		return nil
	}

	if update.EffectiveUser().Bot {
		return nil
	}

	err = r.dbRepository.MessageCreate(ctx, &db_repository.Message{
		TgChatID: update.EffectiveChat().GetID(),
		TgUserId: update.EffectiveUser().GetID(),
		Text:     update.EffectiveMessage.Text,
		TgId:     update.EffectiveMessage.GetID(),
	})
	if err != nil {
		return errors.WithStack(err)
	}

	tgName := utils.GetNameFromTgUser(update.EffectiveUser())
	tgNameCleaned := strings.Map(func(r rune) rune {
		if unicode.IsGraphic(r) {
			return r
		}
		return -1
	}, tgName)

	err = r.dbRepository.UserUpsert(ctx, &db_repository.User{
		TgUserId:   update.EffectiveUser().GetID(),
		TgUsername: update.EffectiveUser().Username,
		TgName:     tgNameCleaned,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
