package analitics

import (
	"context"
	"github.com/google/uuid"
	"github.com/kamva/mgm/v3"
	"github.com/pkg/errors"
	"github.com/teadove/goteleout/internal/repository/ch_repository"
	"github.com/teadove/goteleout/internal/repository/mongo_repository"
)

func (r *Service) InsertNewMessage(ctx context.Context, message *Message) error {
	err := r.chRepository.MessageCreate(ctx, &ch_repository.Message{
		Id:        uuid.New(),
		CreatedAt: message.CreatedAt,
		TgChatID:  message.TgChatID,
		TgUserId:  message.TgUserId,
		Text:      message.Text,
		TgId:      message.TgId,
	})
	if err != nil {
		return errors.Wrap(err, "failed to insert message in ch repository")
	}

	err = r.mongoRepository.MessageCreateOrNothingAndSetTime(ctx, &mongo_repository.Message{
		DefaultModel: mgm.DefaultModel{DateFields: mgm.DateFields{CreatedAt: message.CreatedAt}},
		TgChatID:     message.TgChatID,
		TgUserId:     message.TgUserId,
		Text:         message.Text,
		TgId:         message.TgId,
	})
	if err != nil {
		return errors.Wrap(err, "failed to insert message in mongo repository")
	}
	return nil
}
