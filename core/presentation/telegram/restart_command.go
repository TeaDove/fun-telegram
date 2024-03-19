package telegram

import (
	"context"
	"os"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/repository/mongo_repository"
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

	err = r.mongoRepository.RestartMessageCreate(ctx, &mongo_repository.Message{
		TgChatID: ctx.Self.ID,
		TgUserId: ctx.Self.ID,
		Text:     reloadMessage.Text,
		TgId:     reloadMessage.ID,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	os.Exit(0)

	return nil
}

func (r *Presentation) updateRestartMessages(ctx context.Context) error {
	messages, err := r.mongoRepository.RestartMessageGetAndDelete(ctx)
	if err != nil {
		return errors.WithStack(err)
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

		_, err = tgCtx.EditMessage(message.TgChatID, &tg.MessagesEditMessageRequest{
			Peer: r.protoClient.Self.AsInputPeer(),
			ID:   message.TgId,
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

	return nil
}
