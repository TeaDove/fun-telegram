package analitics

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/teadove/fun_telegram/core/repository/ch_repository"
)

func (r *Service) MessageInsert(ctx context.Context, message *Message) error {
	chMessage := &ch_repository.Message{
		CreatedAt:     message.CreatedAt,
		TgChatID:      message.TgChatID,
		TgUserId:      message.TgUserId,
		Text:          message.Text,
		TgId:          message.TgId,
		ReplyToMsgID:  message.ReplyToMsgID,
		ReplyToUserID: message.ReplyToUserID,
	}
	words := strings.Fields(message.Text)

	for _, word := range words {
		word, ok := r.filterAndLemma(word)
		if !ok {
			continue
		}
		chMessage.WordsCount++

		ok, err := r.IsToxic(word)
		if err != nil {
			return errors.Wrap(err, "failed to check if word is toxic")
		}

		if ok {
			chMessage.ToxicWordsCount++
		}
	}

	err := r.chRepository.MessageInsert(ctx, chMessage)
	if err != nil {
		return errors.Wrap(err, "failed to insert message in ch repository")
	}

	return nil
}

func (r *Service) MessageSetReplyToUserId(ctx context.Context, chatId int64) error {
	err := r.chRepository.MessageSetReplyToUserId(ctx, chatId)
	if err != nil {
		return errors.Wrap(err, "failed to set reply to user id in ch repository")
	}

	return nil
}

func (r *Service) DeleteMessagesByChatId(ctx context.Context, chatId int64) (int64, error) {
	count, err := r.mongoRepository.DeleteMessagesByChat(ctx, chatId)
	if err != nil {
		return 0, errors.Wrap(err, "failed to delete messages from mongo repository")
	}

	err = r.chRepository.MessageDeleteByChatId(ctx, chatId)
	if err != nil {
		return 0, errors.Wrap(err, "failed to delete messages from ch repository")
	}

	return count, nil
}

func (r *Service) DeleteAllMessages(ctx context.Context) (int64, error) {
	count, err := r.mongoRepository.DeleteAllMessages(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to delete messages")
	}

	return count, nil
}

func (r *Service) GetLastMessage(ctx context.Context, chatId int64) (Message, error) {
	message, err := r.chRepository.GetLastMessageByChatId(ctx, chatId)
	if err != nil {
		return Message{}, errors.Wrap(err, "failed to get last message")
	}

	return Message{
		CreatedAt: message.CreatedAt,
		TgChatID:  message.TgChatID,
		TgId:      message.TgId,
		TgUserId:  message.TgUserId,
		Text:      message.Text,
	}, nil
}
