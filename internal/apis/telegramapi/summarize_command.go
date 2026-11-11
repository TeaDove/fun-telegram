package telegramapi

import (
	"fmt"
	"slices"
	"time"

	"github.com/TeaDove/fun-telegram/internal/services/messageservice"
	"github.com/TeaDove/fun-telegram/internal/suppliers/gigachatsupplier"

	"github.com/celestix/gotgproto/ext"
	"github.com/cockroachdb/errors"
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

	messages := []gigachatsupplier.Message{{
		Role: "system",
		Content: "Ты - умный бот, который сумаризирует переписку в телеграмме. " +
			"Ниже тебе будут отправлены сообщения участников некое телеграмм чата. " +
			"В ответе сумаризируй переписку, опиши диалоги, которые происхоидили, темы, что обсуждались",
	}}

	slices.SortFunc(storage.Messages, func(a, b messageservice.Message) int {
		if a.CreatedAt.Before(b.CreatedAt) {
			return -1
		}

		return 1
	})

	for _, message := range storage.Messages {
		messages = append(messages, gigachatsupplier.Message{
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
