package telegram

import (
	"context"
	"os"

	"github.com/teadove/fun_telegram/core/repository/db_repository"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/service/resource"
)

func (r *Presentation) restartCommandHandler(
	ctx *ext.Context,
	update *ext.Update,
	input *input,
) error {
	reloadMessage, err := ctx.SendMessage(ctx.Self.ID,
		&tg.MessagesSendMessageRequest{
			Message: r.resourceService.Localize(
				ctx,
				resource.CommandRestartRestarting,
				input.ChatSettings.Locale,
			),
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to send message")
	}

	zerolog.Ctx(ctx).Warn().Str("status", "reload.begin").Send()

	err = r.dbRepository.RestartMessageInsert(ctx, &db_repository.RestartMessage{
		MessageTgChatID: ctx.Self.ID,
		MessageTgId:     reloadMessage.ID,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	os.Exit(0)

	return nil
}

func (r *Presentation) updateRestartMessages(ctx context.Context) error {
	messages, err := r.dbRepository.RestartMessageGet(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get messages")
	}

	if len(messages) == 0 {
		return nil
	}

	chatSetting, err := r.getChatSettings(ctx, r.protoClient.Self.ID)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, message := range messages {
		tgCtx := r.protoClient.CreateContext()

		_, err = tgCtx.EditMessage(message.MessageTgChatID, &tg.MessagesEditMessageRequest{
			Peer: r.protoClient.Self.AsInputPeer(),
			ID:   message.MessageTgId,
			Message: r.resourceService.Localize(
				ctx,
				resource.CommandRestartSuccess,
				chatSetting.Locale,
			),
		})
		if err != nil {
			return errors.WithStack(err)
		}
	}

	err = r.dbRepository.RestartMessageDelete(ctx, messages)
	if err != nil {
		return errors.Wrap(err, "failed to delete restart messages")
	}

	return nil
}
