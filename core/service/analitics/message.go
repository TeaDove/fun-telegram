package analitics

import (
	"context"
	"strings"

	"fun_telegram/core/repository/db_repository"

	"github.com/pkg/errors"
)

func (r *Service) MessageInsert(ctx context.Context, message *Message) error {
	chMessage := &db_repository.Message{
		WithCreatedAt:   db_repository.WithCreatedAt{CreatedAt: message.CreatedAt},
		TgChatID:        message.TgChatID,
		TgUserID:        message.TgUserID,
		Text:            message.Text,
		TgID:            message.TgID,
		ReplyToTgMsgID:  message.ReplyToMsgID,
		ReplyToTgUserID: message.ReplyToUserID,
	}
	words := strings.Fields(message.Text)

	var (
		ok   bool
		word string
		err  error
	)

	for _, word = range words {
		word, ok = r.filterAndLemma(word)
		if !ok {
			continue
		}

		chMessage.WordsCount++

		ok, err = r.IsToxic(word)
		if err != nil {
			return errors.Wrap(err, "failed to check if word is toxic")
		}

		if ok {
			chMessage.ToxicWordsCount++
		}
	}

	err = r.dbRepository.MessageInsert(ctx, chMessage)
	if err != nil {
		return errors.Wrap(err, "failed to insert message in ch repository")
	}

	return nil
}

func (r *Service) DeleteMessagesByChatID(ctx context.Context, chatID int64) (uint64, error) {
	count, err := r.dbRepository.MessagesDeleteByChat(ctx, chatID)
	if err != nil {
		return 0, errors.Wrap(err, "failed to delete messages from mongo repository")
	}

	return count, nil
}

func (r *Service) GetLastMessage(ctx context.Context, chatID int64) (Message, error) {
	message, err := r.dbRepository.MessageGetLastByChatID(ctx, chatID)
	if err != nil {
		return Message{}, errors.Wrap(err, "failed to get last message")
	}

	return Message{
		CreatedAt: message.CreatedAt,
		TgChatID:  message.TgChatID,
		TgID:      message.TgID,
		TgUserID:  message.TgUserID,
		Text:      message.Text,
	}, nil
}
