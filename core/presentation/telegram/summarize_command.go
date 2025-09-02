package telegram

import (
	"fmt"
	"fun_telegram/core/service/message_service"
	"fun_telegram/core/supplier/gigachat_supplier"
	"slices"
	"time"

	"github.com/celestix/gotgproto/ext"
	"github.com/pkg/errors"
)

func (r *Presentation) summarizeCommand(c *Context) error {
	storage, err := r.getChatStorage(c, &getChatStorageInput{
		MaxElapsed: time.Hour,
		MaxCount:   200,
		QueryTill:  time.Now().Add(-time.Hour * 24 * 30),
	})
	if err != nil {
		return errors.Wrap(err, "failed to get chat storage")
	}

	messages := []gigachat_supplier.Message{{
		Role: "system",
		Content: "Ты - умный бот, который сумаризирует переписку в телеграмме. " +
			"Ниже тебе будут отправлены сообщения участников некое телеграмм чата. " +
			"В ответе сумаризируй переписку, опиши диалоги, которые происхоидили, темы, что обсуждались",
	}}
	slices.SortFunc(storage.Messages, func(a, b message_service.Message) int {
		if a.CreatedAt.Before(b.CreatedAt) {
			return -1
		}

		return 1
	})

	for _, message := range storage.Messages {
		messages = append(messages, gigachat_supplier.Message{
			Role: "user",
			Content: fmt.Sprintf("Автор: %s, Дата: %s, Сообщение: %s",
				storage.UsersNameGetter.GetNameAndUsername(message.TgUserID),
				message.CreatedAt.String(),
				message.Text,
			),
		})
	}

	resp, err := r.gigachatSupplier.OneMessage(c.extCtx, messages)
	if err != nil {
		return errors.Wrap(err, "gigachat supplier")
	}

	return c.reply(ext.ReplyTextString(resp))
}
