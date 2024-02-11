package telegram

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/types"
	"github.com/pkg/errors"
	"github.com/teadove/goteleout/internal/repository/mongo_repository"
	"github.com/teadove/goteleout/internal/shared"
	"strings"
)

func (r *Presentation) catchMessages(ctx *ext.Context, update *ext.Update) error {
	if !shared.AppSettings.Telegram.SaveAllMessages {
		return nil
	}

	ok, err := r.isEnabled(ctx, update.EffectiveChat().GetID())
	if err != nil {
		return errors.WithStack(err)
	}
	if !ok {
		return nil
	}

	ok = filterNonNewMessages(update)
	if !ok {
		return nil
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

	err = r.dbRepository.UserUpsert(ctx, &mongo_repository.User{
		TgUserId:   update.EffectiveUser().GetID(),
		TgUsername: strings.ToLower(update.EffectiveUser().Username),
		TgName:     GetNameFromTgUser(update.EffectiveUser()),
		IsBot:      update.EffectiveUser().Bot,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	if update.EffectiveUser().Bot {
		return nil
	}

	err = r.dbRepository.MessageCreate(ctx, &mongo_repository.Message{
		TgChatID: update.EffectiveChat().GetID(),
		TgUserId: update.EffectiveUser().GetID(),
		Text:     update.EffectiveMessage.Text,
		TgId:     update.EffectiveMessage.GetID(),
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
