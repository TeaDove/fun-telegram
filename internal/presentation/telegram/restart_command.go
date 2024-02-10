package telegram

import (
	"context"
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/goteleout/internal/repository/mongo_repository"
	"github.com/teadove/goteleout/internal/service/resource"
	"os"
)

func (r *Presentation) restartCommandHandler(ctx *ext.Context, update *ext.Update, input *Input) error {
	reloadMessage, err := ctx.SendMessage(ctx.Self.ID,
		&tg.MessagesSendMessageRequest{
			Message: r.resourceService.Localize(ctx, resource.CommandRestartRestarting, input.Locale),
		},
	)

	zerolog.Ctx(ctx).Warn().Str("status", "reload.begin").Send()

	err = r.dbRepository.RestartMessageCreate(ctx, &mongo_repository.Message{
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
	messages, err := r.dbRepository.RestartMessageGetAndDelete(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	if len(messages) == 0 {
		return nil
	}

	locale, err := r.getLocale(r.protoClient.Self.ID)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, message := range messages {
		tgCtx := r.protoClient.CreateContext()
		_, err = tgCtx.EditMessage(message.TgChatID, &tg.MessagesEditMessageRequest{
			Peer:    r.protoClient.Self.AsInputPeer(),
			ID:      message.TgId,
			Message: r.resourceService.Localize(ctx, resource.CommandRestartSuccess, locale),
		})
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}
