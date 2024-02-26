package telegram

import (
	"context"
	"fmt"
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/repository/mongo_repository"
	"time"
)

// TODO: fix nolint
// nolint: cyclop
func (r *Presentation) pingCommandHandler(ctx *ext.Context, update *ext.Update, input *input) error {
	var deletePinAfter = 5 * time.Minute

	msg, err := ctx.Reply(update, fmt.Sprintf("Ping requested by %s\n\n", GetNameFromTgUser(update.EffectiveUser())), nil)
	if err != nil {
		return errors.Wrap(err, "failed to send ping messages")
	}

	err = r.mongoRepository.PingMessageCreate(
		ctx,
		&mongo_repository.Message{
			TgChatID: update.EffectiveChat().GetID(),
			TgUserId: ctx.Self.ID,
			Text:     msg.Text,
			TgId:     msg.ID,
		},
		time.Now().UTC().Add(deletePinAfter),
	)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = r.telegramApi.MessagesUpdatePinnedMessage(ctx, &tg.MessagesUpdatePinnedMessageRequest{
		Silent:    false,
		Unpin:     false,
		PmOneside: true,
		Peer:      update.EffectiveChat().GetInputPeer(),
		ID:        msg.ID,
	})
	if err != nil {
		return errors.Wrap(err, "failed to pin message")
	}

	return nil
}

func (r *Presentation) deleteOldPingMessages(ctx context.Context) error {
	messages, err := r.mongoRepository.PingMessageGetAndDeleteForDeletion(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get ping messages")
	}

	if len(messages) == 0 {
		return nil
	}

	zerolog.Ctx(ctx).Debug().Str("status", "ping.messages.deleting").Int("count", len(messages)).Send()

	for _, message := range messages {
		log := zerolog.Ctx(ctx).With().Int("msg_id", message.TgId).Int64("chat_id", message.TgChatID).Logger()

		inputPeer := r.protoClient.PeerStorage.GetInputPeerById(message.TgChatID)
		if inputPeer == nil {
			log.Warn().Str("status", "failed.to.get.peer").Send()
			continue
		}

		_, err = r.telegramApi.MessagesUpdatePinnedMessage(ctx, &tg.MessagesUpdatePinnedMessageRequest{
			Silent:    false,
			Unpin:     true,
			PmOneside: true,
			Peer:      inputPeer,
			ID:        message.TgId,
		})
		if err != nil {
			log.Error().Stack().Err(err).Str("status", "failed.to.unpin.message").Send()
			continue
		}

		err = r.protoClient.CreateContext().DeleteMessages(message.TgChatID, []int{message.TgId})
		if err != nil {
			log.Error().Stack().Err(err).Str("status", "failed.to.delete.message").Send()
			continue
		}

		log.Info().Str("status", "ping.message.deleted").Send()
	}

	return nil
}
