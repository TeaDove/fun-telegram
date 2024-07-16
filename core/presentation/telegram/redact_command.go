package telegram

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/repository/db_repository"
	"github.com/teadove/fun_telegram/core/service/resource"
)

func (r *Presentation) redactCommandHandler(
	ctx *ext.Context,
	update *ext.Update,
	input *input,
) error {
	chatId := update.EffectiveChat().GetID()
	messages, err := r.dbRepository.MessageSelectByChatIdAndUserIdWithWordsCount(
		ctx,
		chatId,
		ctx.Self.ID,
		1,
		500,
	)
	if err != nil {
		return errors.Wrap(err, "failed to get messages to redact")
	}

	zerolog.Ctx(ctx).
		Info().
		Int("count", len(messages)).
		Msg("messages.redacting.begin")

	textToSet := r.resourceService.Localize(ctx, resource.Redacted, input.ChatSettings.Locale)

	var (
		message db_repository.Message
		idx     int
	)
	for idx, message = range messages {
		if textToSet == message.Text {
			continue
		}

		_, err = ctx.EditMessage(chatId, &tg.MessagesEditMessageRequest{
			Peer:    update.EffectiveChat().GetInputPeer(),
			ID:      message.TgId,
			Message: textToSet,
		})
		if err != nil {
			return errors.Wrap(err, "failed to redact")
		}

		if idx%100 == 0 {
			zerolog.Ctx(ctx).
				Info().
				Int("completed", idx).
				Msg("messages.redacting.process")
		}
	}

	// TODO get messages to delete by my self
	zerolog.Ctx(ctx).
		Info().
		Int("completed", idx).
		Msg("messages.redacting.done")

	return nil
}
