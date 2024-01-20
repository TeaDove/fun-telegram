package telegram

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/teadove/goteleout/internal/presentation/telegram/utils"
	"github.com/teadove/goteleout/internal/repository/db_repository"
)

func (r *Presentation) catchMessages(ctx *ext.Context, update *ext.Update) error {
	_, ok := update.UpdateClass.(*tg.UpdateNewChannelMessage)
	if !ok {
		_, ok = update.UpdateClass.(*tg.UpdateNewMessage)
		if !ok {
			return nil
		}
	}
	text := update.EffectiveMessage.Text
	if text == "" {
		return nil
	}

	err := r.dbRepository.MessageCreate(ctx, &db_repository.Message{
		TgChatID: update.EffectiveChat().GetID(),
		TgUserId: update.EffectiveUser().GetID(),
		Text:     update.EffectiveMessage.Text,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	err = r.dbRepository.UserUpsert(ctx, &db_repository.User{
		TgUserId:   update.EffectiveUser().GetID(),
		TgUsername: update.EffectiveUser().Username,
		TgName:     utils.GetNameFromTgUser(update.EffectiveUser()),
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
