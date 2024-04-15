package telegram

import (
	"context"
	"time"

	errors2 "github.com/celestix/gotgproto/errors"
	"github.com/celestix/gotgproto/types"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/repository/mongo_repository"
)

// TODO: fix nolint
// nolint: cyclop
func (r *Presentation) pingCommandHandler(
	ctx *ext.Context,
	update *ext.Update,
	input *input,
) error {
	deletePinAfter := 5 * time.Minute

	var msgToPing *types.Message

	err := update.EffectiveMessage.SetRepliedToMessage(
		ctx,
		r.telegramApi,
		r.protoClient.PeerStorage,
	)
	if err != nil {
		if !(errors.Is(err, errors2.ErrReplyNotMessage) || errors.Is(err, errors2.ErrMessageNotExist)) {
			return errors.Wrap(err, "failed to set reply to message")
		}
		msgToPing = update.EffectiveMessage
	} else {
		msgToPing = update.EffectiveMessage.ReplyToMessage
	}

	userId, err := GetSenderId(msgToPing)
	if err != nil {
		return errors.Wrap(err, "failed to get sender id")
	}

	err = r.mongoRepository.PingMessageCreate(
		ctx,
		&mongo_repository.Message{
			TgChatID: update.EffectiveChat().GetID(),
			TgUserId: userId,
			Text:     msgToPing.Text,
			TgId:     msgToPing.ID,
		},
		time.Now().UTC().Add(deletePinAfter),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create ping message")
	}

	_, err = r.telegramApi.MessagesUpdatePinnedMessage(ctx, &tg.MessagesUpdatePinnedMessageRequest{
		Silent:    false,
		Unpin:     false,
		PmOneside: true,
		Peer:      update.EffectiveChat().GetInputPeer(),
		ID:        msgToPing.ID,
	})
	if err != nil {
		return errors.Wrap(err, "failed to pin message")
	}

	_ = ctx.DeleteMessages(update.EffectiveChat().GetID(), []int{update.EffectiveMessage.ID})

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

	zerolog.Ctx(ctx).
		Debug().
		Str("status", "ping.messages.deleting").
		Int("count", len(messages)).
		Send()

	for _, message := range messages {
		log := zerolog.Ctx(ctx).
			With().
			Int("msg_id", message.TgId).
			Int64("chat_id", message.TgChatID).
			Logger()

		inputPeer := r.protoClient.PeerStorage.GetInputPeerById(message.TgChatID)
		if inputPeer == nil {
			log.Warn().Str("status", "failed.to.get.peer").Send()
			continue
		}

		_, err = r.telegramApi.MessagesUpdatePinnedMessage(
			ctx,
			&tg.MessagesUpdatePinnedMessageRequest{
				Silent:    false,
				Unpin:     true,
				PmOneside: true,
				Peer:      inputPeer,
				ID:        message.TgId,
			},
		)
		if err != nil {
			log.Error().Stack().Err(err).Str("status", "failed.to.unpin.message").Send()
			continue
		}

		log.Info().Str("status", "ping.message.unpinned").Send()
	}

	return nil
}
