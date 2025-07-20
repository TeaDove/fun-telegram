package telegram

import (
	"context"
	"os"

	"github.com/teadove/fun_telegram/core/repository/db_repository"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func (r *Presentation) restartCommandHandler(
	ctx *ext.Context,
	update *ext.Update,
	input *input,
) error {
	reloadMessage, err := ctx.SendMessage(
		ctx.Self.ID,
		&tg.MessagesSendMessageRequest{Message: "Restarting..."},
	)
	if err != nil {
		return errors.Wrap(err, "failed to send message")
	}

	zerolog.Ctx(ctx).Warn().Msg("reload.begin")

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

	for _, message := range messages {
		tgCtx := r.protoClient.CreateContext()

		_, err = tgCtx.EditMessage(message.MessageTgChatID, &tg.MessagesEditMessageRequest{
			Peer:    r.protoClient.Self.AsInputPeer(),
			ID:      message.MessageTgId,
			Message: "Restart success!",
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
