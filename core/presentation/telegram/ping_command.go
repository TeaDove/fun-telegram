package telegram

import (
	"context"
	"time"

	"github.com/teadove/fun_telegram/core/repository/db_repository"

	tgError "github.com/celestix/gotgproto/errors"
	"github.com/celestix/gotgproto/types"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
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
		if !(errors.Is(err, tgError.ErrReplyNotMessage) || errors.Is(err, tgError.ErrMessageNotExist)) {
			return errors.Wrap(err, "failed to set reply to message")
		}

		msgToPing = update.EffectiveMessage
	} else {
		msgToPing = update.EffectiveMessage.ReplyToMessage
	}

	err = r.dbRepository.PingMessageCreate(
		ctx,
		&db_repository.PingMessage{
			MessageTgChatID: update.EffectiveChat().GetID(),
			MessageTgId:     msgToPing.ID,
			DeleteAt:        time.Now().UTC().Add(deletePinAfter),
		},
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
	messages, err := r.dbRepository.PingMessageGet(ctx)
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
			Int("msg_id", message.MessageTgId).
			Int64("chat_id", message.MessageTgChatID).
			Logger()

		inputPeer := r.protoClient.PeerStorage.GetInputPeerById(message.MessageTgChatID)
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
				ID:        message.MessageTgId,
			},
		)
		if err != nil {
			log.Error().Stack().Err(err).Str("status", "failed.to.unpin.message").Send()
			continue
		}

		log.Info().Str("status", "ping.message.unpinned").Send()
	}

	err = r.dbRepository.PingMessageDelete(ctx, messages)
	if err != nil {
		return errors.Wrap(err, "failed to delete ping messages")
	}

	return nil
}
